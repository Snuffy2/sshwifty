// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/Snuffy2/sshwifty/application/command"
	"github.com/Snuffy2/sshwifty/application/configuration"
	"github.com/Snuffy2/sshwifty/application/log"
	"github.com/Snuffy2/sshwifty/application/network"
	"github.com/Snuffy2/sshwifty/application/rw"
)

const (
	MoshServerRemoteStdOut               = 0x00
	MoshServerHookOutputBeforeConnecting = 0x01
	MoshServerConnectFailed              = 0x02
	MoshServerConnectSucceed             = 0x03
)

const (
	MoshClientStdIn  = 0x00
	MoshClientResize = 0x01
)

const (
	MoshRequestErrorBadUserName      = command.StreamError(0x01)
	MoshRequestErrorBadRemoteAddress = command.StreamError(0x02)
	MoshRequestErrorBadAuthMethod    = command.StreamError(0x03)
	MoshRequestErrorUnsupportedProxy = command.StreamError(0x04)
	MoshRequestErrorBadMetadata      = command.StreamError(0x05)
)

const (
	moshMaxUsernameLen = sshMaxUsernameLen
	moshMaxHostnameLen = sshMaxHostnameLen
)

var (
	ErrMoshUnknownClientSignal = errors.New("unknown client signal")
	ErrMoshSocks5Unsupported   = errors.New(
		"Mosh does not support SOCKS5 proxying in this version because its session uses UDP",
	)
	ErrMoshRemoteSessionUnavailable = errors.New("remote Mosh session is unavailable")
)

type moshSessionBuilder func(host string, port int, key string) (moshSession, error)
type moshHostResolver func(ctx context.Context, host string) ([]net.IP, error)

type moshClient struct {
	w          command.StreamResponder
	l          log.Logger
	hooks      command.Hooks
	cfg        command.Configuration
	meta       map[string]string
	bufferPool *command.BufferPool

	baseCtx       context.Context
	baseCtxCancel func()

	remoteCloseWait                 sync.WaitGroup
	remoteReadTimeoutRetry          bool
	remoteReadForceRetryNextTimeout bool
	remoteReadTimeoutRetryLock      sync.Mutex

	credentialReceive                    chan []byte
	credentialProcessed                  bool
	credentialReceiveClosed              bool
	fingerprintVerifyResultReceive       chan bool
	fingerprintProcessed                 bool
	fingerprintVerifyResultReceiveClosed bool

	sessionReceive chan moshSession
	session        moshSession
	sessionClosed  bool
	sessionLock    sync.Mutex

	sessionBuilder moshSessionBuilder
	hostResolver   moshHostResolver
}

func newMosh(
	l log.Logger,
	hooks command.Hooks,
	w command.StreamResponder,
	cfg command.Configuration,
	bufferPool *command.BufferPool,
) command.FSMMachine {
	ctx, ctxCancel := context.WithCancel(context.Background())
	return &moshClient{
		w:                                    w,
		l:                                    l,
		hooks:                                hooks,
		cfg:                                  cfg,
		bufferPool:                           bufferPool,
		baseCtx:                              ctx,
		baseCtxCancel:                        sync.OnceFunc(ctxCancel),
		credentialReceive:                    make(chan []byte, 1),
		fingerprintVerifyResultReceive:       make(chan bool, 1),
		sessionReceive:                       make(chan moshSession, 1),
		sessionBuilder:                       newMoshGoSession,
		hostResolver:                         defaultMoshHostResolver,
		sessionClosed:                        false,
		remoteReadTimeoutRetryLock:           sync.Mutex{},
		remoteCloseWait:                      sync.WaitGroup{},
		credentialReceiveClosed:              false,
		fingerprintVerifyResultReceiveClosed: false,
	}
}

func parseMoshConfig(p configuration.Preset) (configuration.Preset, error) {
	return parseSSHConfig(p)
}

func moshRemoteHost(hostport string) string {
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		return hostport
	}

	return host
}

func defaultMoshHostResolver(ctx context.Context, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, "ip", host)
}

func (d *moshClient) Bootup(
	r *rw.LimitedReader,
	b []byte,
) (command.FSMState, command.FSMError) {
	if err := d.validateProxySupport(); err != nil {
		return nil, command.ToFSMError(err, MoshRequestErrorUnsupportedProxy)
	}

	sBuf := d.bufferPool.Get()
	defer d.bufferPool.Put(sBuf)

	userName, userNameErr := ParseString(r.Read, (*sBuf)[:moshMaxUsernameLen])
	if userNameErr != nil {
		return nil, command.ToFSMError(userNameErr, MoshRequestErrorBadUserName)
	}

	addr, addrErr := ParseAddress(r.Read, (*sBuf)[:moshMaxHostnameLen])
	if addrErr != nil {
		return nil, command.ToFSMError(addrErr, MoshRequestErrorBadRemoteAddress)
	}

	addrStr := addr.String()
	if addrStr == "" {
		return nil, command.ToFSMError(ErrSSHInvalidAddress, MoshRequestErrorBadRemoteAddress)
	}

	rData, rErr := rw.FetchOneByte(r.Fetch)
	if rErr != nil {
		return nil, command.ToFSMError(rErr, MoshRequestErrorBadAuthMethod)
	}

	authMethodBuilder, authMethodBuilderErr := d.buildAuthMethod(rData[0])
	if authMethodBuilderErr != nil {
		return nil, command.ToFSMError(authMethodBuilderErr, MoshRequestErrorBadAuthMethod)
	}

	requestMeta, requestMetaErr := parseMoshRequestMeta(r, (*sBuf)[:])
	if requestMetaErr != nil {
		return nil, command.ToFSMError(requestMetaErr, MoshRequestErrorBadMetadata)
	}
	d.meta = requestMeta

	d.remoteCloseWait.Add(1)
	go d.remote(string(userName.Data()), addrStr, authMethodBuilder)

	return d.local, command.NoFSMError()
}

func (d *moshClient) validateProxySupport() error {
	if d.cfg.Socks5Configured {
		return ErrMoshSocks5Unsupported
	}

	return nil
}

func parseMoshRequestMeta(r *rw.LimitedReader, b []byte) (map[string]string, error) {
	meta := map[string]string{"Mosh Server": ""}
	if r.Completed() {
		return meta, nil
	}

	serverPath, err := ParseString(r.Read, b)
	if err != nil {
		return nil, err
	}

	meta["Mosh Server"] = strings.TrimSpace(string(serverPath.Data()))
	return meta, nil
}

func (d *moshClient) buildAuthMethod(methodType byte) (sshAuthMethodBuilder, error) {
	switch methodType {
	case SSHAuthMethodNone:
		return func(b []byte) []ssh.AuthMethod { return nil }, nil
	case SSHAuthMethodPassphrase:
		return func(b []byte) []ssh.AuthMethod {
			return []ssh.AuthMethod{
				ssh.PasswordCallback(func() (string, error) {
					d.enableRemoteReadTimeoutRetry()
					defer d.disableRemoteReadTimeoutRetry()

					if wErr := d.w.SendManual(SSHServerConnectRequestCredential, b[d.w.HeaderSize():]); wErr != nil {
						return "", wErr
					}

					passphraseBytes, passphraseReceived := <-d.credentialReceive
					if !passphraseReceived {
						return "", ErrSSHAuthCancelled
					}

					return string(passphraseBytes), nil
				}),
			}
		}, nil
	case SSHAuthMethodPrivateKey:
		return func(b []byte) []ssh.AuthMethod {
			return []ssh.AuthMethod{
				ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
					d.enableRemoteReadTimeoutRetry()
					defer d.disableRemoteReadTimeoutRetry()

					if wErr := d.w.SendManual(SSHServerConnectRequestCredential, b[d.w.HeaderSize():]); wErr != nil {
						return nil, wErr
					}

					privateKeyBytes, privateKeyReceived := <-d.credentialReceive
					if !privateKeyReceived {
						return nil, ErrSSHAuthCancelled
					}

					signer, signerErr := ssh.ParsePrivateKey(privateKeyBytes)
					if signerErr != nil {
						return nil, signerErr
					}

					return []ssh.Signer{signer}, nil
				}),
			}
		}, nil
	default:
		return nil, ErrSSHInvalidAuthMethod
	}
}

func (d *moshClient) confirmRemoteFingerprint(
	hostname string,
	remote net.Addr,
	key ssh.PublicKey,
	buf []byte,
) error {
	d.enableRemoteReadTimeoutRetry()
	defer d.disableRemoteReadTimeoutRetry()

	fgp := ssh.FingerprintSHA256(key)
	fgpLen := copy(buf[d.w.HeaderSize():], fgp)

	if wErr := d.w.SendManual(SSHServerConnectVerifyFingerprint, buf[:d.w.HeaderSize()+fgpLen]); wErr != nil {
		return wErr
	}

	confirmed, confirmOK := <-d.fingerprintVerifyResultReceive
	if !confirmOK {
		return ErrSSHRemoteFingerprintVerificationCancelled
	}
	if !confirmed {
		return ErrSSHRemoteFingerprintRefused
	}

	return nil
}

func (d *moshClient) enableRemoteReadTimeoutRetry() {
	d.remoteReadTimeoutRetryLock.Lock()
	defer d.remoteReadTimeoutRetryLock.Unlock()
	d.remoteReadTimeoutRetry = true
}

func (d *moshClient) disableRemoteReadTimeoutRetry() {
	d.remoteReadTimeoutRetryLock.Lock()
	defer d.remoteReadTimeoutRetryLock.Unlock()
	d.remoteReadTimeoutRetry = false
	d.remoteReadForceRetryNextTimeout = true
}

func (d *moshClient) dialRemote(
	networkName string,
	addr string,
	config *ssh.ClientConfig,
) (*ssh.Client, net.Addr, func(), error) {
	dialCtx, dialCtxCancel := context.WithTimeout(d.baseCtx, config.Timeout)
	defer dialCtxCancel()

	conn, err := d.cfg.Dial(dialCtx, networkName, addr)
	if err != nil {
		return nil, nil, nil, err
	}
	peerAddr := conn.RemoteAddr()

	sshConn := &sshRemoteConnWrapper{
		Conn:       conn,
		writerConn: network.NewWriteTimeoutConn(conn, d.cfg.DialTimeout),
		requestTimeoutRetry: func(s *sshRemoteConnWrapper) bool {
			d.remoteReadTimeoutRetryLock.Lock()
			defer d.remoteReadTimeoutRetryLock.Unlock()

			if !d.remoteReadTimeoutRetry {
				if !d.remoteReadForceRetryNextTimeout {
					return false
				}
				d.remoteReadForceRetryNextTimeout = false
			}

			s.SetReadDeadline(time.Now().Add(config.Timeout))
			return true
		},
	}

	sshConn.SetWriteDeadline(time.Now().Add(d.cfg.DialTimeout))
	sshConn.SetReadDeadline(time.Now().Add(config.Timeout))

	c, chans, reqs, err := ssh.NewClientConn(sshConn, addr, config)
	if err != nil {
		sshConn.Close()
		return nil, nil, nil, err
	}

	return ssh.NewClient(c, chans, reqs), peerAddr, func() {
		d.remoteReadTimeoutRetryLock.Lock()
		defer d.remoteReadTimeoutRetryLock.Unlock()
		d.remoteReadTimeoutRetry = false
		d.remoteReadForceRetryNextTimeout = true
		sshConn.SetReadDeadline(sshEmptyTime)
	}, nil
}

func (d *moshClient) remote(user string, address string, authMethodBuilder sshAuthMethodBuilder) {
	u := d.bufferPool.Get()
	defer d.bufferPool.Put(u)

	var session moshSession
	defer func() {
		if session != nil {
			session.Close()
		}
		d.w.Signal(command.HeaderClose)
		close(d.sessionReceive)
		d.baseCtxCancel()
		d.remoteCloseWait.Done()
	}()

	err := d.hooks.Run(
		d.baseCtx,
		configuration.HOOK_BEFORE_CONNECTING,
		command.NewHookParameters(2).
			Insert("Remote Type", "Mosh").
			Insert("Remote Address", address),
		command.NewDefaultHookOutput(d.l, func(b []byte) (int, error) {
			dLen := copy((*u)[d.w.HeaderSize():], b) + d.w.HeaderSize()
			return len(b), d.w.SendManual(MoshServerHookOutputBeforeConnecting, (*u)[:dLen])
		}),
	)
	if err != nil {
		d.sendConnectFailed((*u)[:], err)
		return
	}

	conn, peerAddr, clearConnInitialDeadline, err := d.dialRemote("tcp", address, &ssh.ClientConfig{
		User: user,
		Auth: authMethodBuilder((*u)[:]),
		HostKeyCallback: func(h string, r net.Addr, k ssh.PublicKey) error {
			return d.confirmRemoteFingerprint(h, r, k, (*u)[:])
		},
		Timeout: d.cfg.DialTimeout,
	})
	if err != nil {
		d.sendConnectFailed((*u)[:], err)
		d.l.Debug("Unable to connect to remote machine: %s", err)
		return
	}
	defer conn.Close()

	output, err := d.bootstrapRemoteMoshServer(conn)
	clearConnInitialDeadline()
	if err != nil {
		d.sendConnectFailed((*u)[:], fmt.Errorf("failed to bootstrap mosh-server: %w", err))
		d.l.Debug("Unable to bootstrap remote mosh-server: %s", err)
		return
	}

	connectInfo, err := parseMoshConnectLine(output)
	if err != nil {
		d.sendConnectFailed((*u)[:], fmt.Errorf("failed to parse mosh-server bootstrap output: %w", err))
		d.l.Debug("Unable to parse remote mosh-server bootstrap output: %s", err)
		return
	}

	session, err = d.buildRemoteSession(address, peerAddr, connectInfo.Port, connectInfo.Key)
	if err != nil {
		d.sendConnectFailed((*u)[:], fmt.Errorf("failed to connect to remote mosh session: %w", err))
		d.l.Debug("Unable to connect to remote mosh session: %s", err)
		return
	}

	d.cacheSession(session)

	initialOutput, err := d.awaitRemoteSessionReady(session)
	if err != nil {
		d.sendConnectFailed((*u)[:], fmt.Errorf("failed to verify remote mosh session readiness: %w", err))
		d.l.Debug("Unable to verify remote mosh session readiness: %s", err)
		return
	}

	d.sessionReceive <- session
	if err = d.w.SendManual(MoshServerConnectSucceed, (*u)[:d.w.HeaderSize()]); err != nil {
		return
	}

	if err = d.sendRemoteOutput((*u)[:], initialOutput); err != nil {
		return
	}

	for {
		output, recvErr := session.Recv(d.recvTimeout())
		if recvErr != nil {
			d.l.Debug("Failed to receive mosh output: %s", recvErr)
			return
		}

		select {
		case <-d.baseCtx.Done():
			return
		default:
		}

		if err = d.sendRemoteOutput((*u)[:], output); err != nil {
			return
		}
	}
}

func (d *moshClient) buildRemoteSession(address string, peerAddr net.Addr, port int, key string) (moshSession, error) {
	if err := d.validateMoshRemoteAllowed(address); err != nil {
		return nil, err
	}

	host, err := d.resolveMoshSessionHost(address, peerAddr)
	if err != nil {
		return nil, err
	}

	return d.sessionBuilder(host, port, key)
}

func (d *moshClient) awaitRemoteSessionReady(session moshSession) ([]byte, error) {
	return session.AwaitReady(d.baseCtx, d.recvTimeout())
}

func (d *moshClient) sendRemoteOutput(buf []byte, output []byte) error {
	if len(output) == 0 {
		return nil
	}

	return d.w.Send(MoshServerRemoteStdOut, output, buf)
}

func (d *moshClient) validateMoshRemoteAllowed(address string) error {
	if !d.cfg.OnlyAllowPresetRemotes {
		return nil
	}

	for _, preset := range d.cfg.Presets {
		if preset.Host == address {
			return nil
		}
	}

	return network.ErrAccessControlDialTargetHostNotAllowed
}

func (d *moshClient) resolveMoshSessionHost(address string, peerAddr net.Addr) (string, error) {
	if peerAddr != nil {
		return normalizeMoshPeerHost(peerAddr)
	}

	host := moshRemoteHost(address)
	if ip := net.ParseIP(host); ip != nil {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}

		return "", fmt.Errorf(
			"Mosh v1 requires an IPv4 target because the current mosh-go UDP client is IPv4-only: %q is IPv6",
			host,
		)
	}

	resolveCtx := d.baseCtx
	if resolveCtx == nil {
		resolveCtx = context.Background()
	}

	cancel := func() {}
	if d.cfg.DialTimeout > 0 {
		resolveCtx, cancel = context.WithTimeout(resolveCtx, d.cfg.DialTimeout)
	}
	defer cancel()

	resolver := d.hostResolver
	if resolver == nil {
		resolver = defaultMoshHostResolver
	}

	ips, err := resolver(resolveCtx, host)
	if err != nil {
		return "", fmt.Errorf("resolve mosh host %q: %w", host, err)
	}

	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}
	}

	return "", fmt.Errorf(
		"Mosh v1 requires an IPv4 target because the current mosh-go UDP client is IPv4-only: %q did not resolve to IPv4",
		host,
	)
}

func normalizeMoshPeerHost(peerAddr net.Addr) (string, error) {
	if tcpAddr, ok := peerAddr.(*net.TCPAddr); ok {
		if ipv4 := tcpAddr.IP.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}

		return "", fmt.Errorf(
			"Mosh v1 requires an IPv4 target because the current mosh-go UDP client is IPv4-only: %q is not IPv4",
			tcpAddr.IP.String(),
		)
	}

	host := peerAddr.String()
	if splitHost, _, err := net.SplitHostPort(host); err == nil {
		host = splitHost
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return "", fmt.Errorf("remote SSH peer address %q is not an IP literal", peerAddr.String())
	}

	if ipv4 := ip.To4(); ipv4 != nil {
		return ipv4.String(), nil
	}

	return "", fmt.Errorf(
		"Mosh v1 requires an IPv4 target because the current mosh-go UDP client is IPv4-only: %q is not IPv4",
		host,
	)
}

func (d *moshClient) bootstrapRemoteMoshServer(conn *ssh.Client) (string, error) {
	session, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(renderMoshServerCommand(d.meta))
	outputText := strings.TrimSpace(string(output))
	if err != nil {
		if outputText != "" {
			return outputText, fmt.Errorf("%w: %s", err, outputText)
		}

		return outputText, err
	}

	return outputText, nil
}

func (d *moshClient) recvTimeout() time.Duration {
	if d.cfg.DialTimeout > 0 {
		return d.cfg.DialTimeout
	}

	return time.Second
}

func (d *moshClient) sendConnectFailed(buf []byte, err error) {
	errLen := copy(buf[d.w.HeaderSize():], err.Error()) + d.w.HeaderSize()
	d.w.SendManual(MoshServerConnectFailed, buf[:errLen])
}

func (d *moshClient) cacheSession(session moshSession) moshSession {
	d.sessionLock.Lock()
	defer d.sessionLock.Unlock()

	if d.session == nil {
		d.session = session
		d.sessionClosed = false
	}

	return d.session
}

func (d *moshClient) getSession() (moshSession, error) {
	d.sessionLock.Lock()
	if d.session != nil {
		session := d.session
		d.sessionLock.Unlock()
		return session, nil
	}
	d.sessionLock.Unlock()

	session, ok := <-d.sessionReceive
	if !ok {
		return nil, ErrMoshRemoteSessionUnavailable
	}

	return d.cacheSession(session), nil
}

func (d *moshClient) closeSession() error {
	d.sessionLock.Lock()
	defer d.sessionLock.Unlock()

	if d.session == nil || d.sessionClosed {
		return nil
	}

	d.sessionClosed = true
	return d.session.Close()
}

func (d *moshClient) local(
	f *command.FSM,
	r *rw.LimitedReader,
	h command.StreamHeader,
	b []byte,
) error {
	switch h.Marker() {
	case MoshClientStdIn:
		session, sessionErr := d.getSession()
		if sessionErr != nil {
			return sessionErr
		}

		for !r.Completed() {
			rData, rErr := r.Buffered()
			if rErr != nil {
				return rErr
			}

			if wErr := session.Send(rData); wErr != nil {
				session.Close()
				d.l.Debug("Failed to write data to remote mosh session: %s", wErr)
			}
		}

		return nil

	case MoshClientResize:
		session, sessionErr := d.getSession()
		if sessionErr != nil {
			return sessionErr
		}

		if _, rErr := io.ReadFull(r, b[:4]); rErr != nil {
			return rErr
		}

		rows := uint16(b[0])<<8 | uint16(b[1])
		cols := uint16(b[2])<<8 | uint16(b[3])
		if resizeErr := session.Resize(cols, rows); resizeErr != nil {
			d.l.Debug("Failed to resize mosh session to cols=%d rows=%d: %s", cols, rows, resizeErr)
		}
		return nil

	case SSHClientRespondFingerprint:
		if d.fingerprintProcessed {
			return ErrSSHUnexpectedFingerprintVerificationRespond
		}
		d.fingerprintProcessed = true

		rData, rErr := rw.FetchOneByte(r.Fetch)
		if rErr != nil {
			return rErr
		}

		d.fingerprintVerifyResultReceive <- (rData[0] == 0)
		return nil

	case SSHClientRespondCredential:
		if d.credentialProcessed {
			return ErrSSHUnexpectedCredentialDataRespond
		}
		d.credentialProcessed = true

		credentialDataBufSize := min(r.Remains(), sshCredentialMaxSize)
		credentialDataBuf := make([]byte, 0, credentialDataBufSize)
		totalCredentialRead := 0

		for !r.Completed() {
			rData, rErr := r.Buffered()
			if rErr != nil {
				return rErr
			}

			totalCredentialRead += len(rData)
			if totalCredentialRead > credentialDataBufSize {
				return ErrSSHCredentialDataTooLarge
			}

			credentialDataBuf = append(credentialDataBuf, rData...)
		}

		d.credentialReceive <- credentialDataBuf
		return nil
	default:
		return ErrMoshUnknownClientSignal
	}
}

func (d *moshClient) Close() error {
	d.credentialProcessed = true
	d.fingerprintProcessed = true

	if !d.credentialReceiveClosed {
		close(d.credentialReceive)
		d.credentialReceiveClosed = true
	}
	if !d.fingerprintVerifyResultReceiveClosed {
		close(d.fingerprintVerifyResultReceive)
		d.fingerprintVerifyResultReceiveClosed = true
	}

	d.baseCtxCancel()
	if closeErr := d.closeSession(); closeErr != nil {
		d.remoteCloseWait.Wait()
		return closeErr
	}
	d.remoteCloseWait.Wait()
	return nil
}

func (d *moshClient) Release() error {
	d.baseCtxCancel()
	return nil
}
