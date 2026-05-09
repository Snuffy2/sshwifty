# Go Command And Session Trace

## HTTP And Socket Entry

- `application/controller/controller.go` registers `/sshwifty/socket` and hands requests to the socket controller.
- `application/controller/socket.go` upgrades the HTTP request to WebSocket, derives the shared AES-GCM session keys, and constructs `command.New(...)` to process framed command traffic.
- The socket controller creates a `command.Handler` with the shared `command.BufferPool`, lifecycle hooks, and per-connection read/write wrappers.

## Command Core Frame Flow

1. `application/command/handler.go` reads one packet header byte from the socket transport.
2. Control frames stay in `Handler.handleControl`, which echoes, pauses, or resumes outbound stream writes.
3. Stream frames go through `Handler.handleStream`, which uses `application/command/streams.go` to boot, tick, close, and release per-stream FSM instances.
4. `application/command/commands.go` resolves the command ID to the registered command factory and preset reloader.
5. `application/command/bufferpool.go` provides reusable fixed-size `[]byte` buffers to handlers and command implementations through `Get() *[]byte` / `Put(*[]byte)`.

## SSH Session Path

1. `application/commands/ssh.go` parses boot parameters such as remote address, username, and auth method.
2. `sshClient.Bootup` starts the remote SSH goroutine and returns the client FSM state for later stream ticks.
3. The remote side runs before-connect hooks, dials the TCP target with explicit dial timeout, then performs host-key verification and authentication.
4. Host-key approval and credential prompts flow back to the browser as stream markers; replies arrive as later stream ticks handled by the local FSM state.
5. Once the SSH session is active, stdout/stderr are forwarded back through `command.StreamResponder`, stdin and resize events flow from the client into the SSH session, and close tears down both the session and the underlying connection.

## Telnet Session Path

1. `application/commands/telnet.go` parses the remote address in `Bootup`.
2. `telnetClient.remote` runs before-connect hooks, dials the target with the configured timeout, and signals connect success or failure back to the client.
3. After connect, remote Telnet bytes are forwarded as stream frames and client bytes are written straight to the remote TCP connection.
4. Telnet keeps protocol setup simpler than SSH: no host-key verification or credential callback loop, but it still shares the same command stream lifecycle and buffer pool.

## Close And Error Paths To Preserve

- Hook execution must keep using sanitized environments from `application/command/hook_exec.go`, excluding `SSHWIFTY_*` process secrets while preserving safe ambient variables like `PATH`.
- Dial, read, write, initial, and heartbeat timeouts remain explicit at the controller, handler, and command layers.
- Stream close, release, and command teardown must tolerate repeated or racing shutdown paths without panics.
- Shared mechanics should stay limited to command/session framing, buffer reuse, and lifecycle signaling; SSH- and Telnet-specific setup stays in their protocol implementations.

## Stage 5 Decision

`BufferPool` remains in place for now because it is used across command handler, SSH, and Telnet session code, and this stage did not produce a measured basis for removal.
