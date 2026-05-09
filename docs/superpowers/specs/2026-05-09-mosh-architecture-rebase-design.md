# Mosh Architecture Rebase Design

## Goal

Rebuild the `feature/mosh-support` branch on top of current `origin/main` as a
clean implementation of Mosh support that fits the Vite-era frontend, current Go
command lifecycle, split configuration documentation, and Snuffy2 module path.

## Scope

This feature adds Mosh as the third supported remote type alongside Telnet and
SSH. It preserves the behavior already proven on the old branch: SSH-based
bootstrap, remote `mosh-server` startup, SSH host-key and credential prompts,
terminal resize and input forwarding, and browser-visible connection failures.

The rebuild does not try to make browser-to-backend traffic use Mosh. The
browser still talks to Sshwifty over the existing WebSocket protocol. Mosh is
used only between the Sshwifty backend and the remote host after SSH bootstrap.

## Backend Design

Register a new backend command named `Mosh` with command ID `0x02`. The command
uses the current `command.FSMMachine` lifecycle from `origin/main`: `Bootup`
parses the initial request, starts a remote worker goroutine, returns a non-nil
local FSM state on success, and ensures `Close` waits for the remote worker to
signal `HeaderClose`.

Mosh reuses the SSH authentication and fingerprint semantics where practical,
but keeps its own command file and tests so Mosh-specific failure modes stay
obvious. Bootstrap opens an SSH connection, executes the configured
`mosh-server` command, parses the `MOSH CONNECT` line, resolves the remote host
to an IPv4 literal, and dials `github.com/unixshells/mosh-go`.

Version 1 intentionally rejects SOCKS5 proxy configurations with a clear
request error. The Mosh data channel is UDP, while the existing SOCKS5 dialer is
TCP-oriented. Version 1 also requires an IPv4 backend-to-remote UDP path because
the current `mosh-go` dependency dials IPv4.

## Frontend Design

Add `ui/commands/mosh.js` and `ui/control/mosh.js` in the current Vite-era UI
style, then register them in `ui/app.js` beside Telnet and SSH. The command
wizard collects Host, User, Authentication, Encoding, and optional `Mosh Server`
fields. It reuses the SSH-style fingerprint and credential prompts while
surfacing Mosh-specific backend rejection for SOCKS5.

The Mosh control uses the same console widget shape as SSH. It decodes stdout
through the configured charset, sends stdin bytes to the backend, forwards
terminal resize messages, and closes through the shared stream sender.

Launcher compatibility is preserved with the old branch format:
`user@host|AuthMethod|charset|encodedMoshServer`. The `moshServer` launcher
field remains optional and defaults to `mosh-server`.

## Documentation And Configuration

Update `CONFIGURATION.md`, `README.md`, `preset.example.json`, and
`sshwifty.conf.example.json` so operators can configure Mosh presets and
understand the v1 constraints:

- remote hosts need `mosh-server`;
- the Sshwifty backend/container needs UDP access to the remote Mosh port range;
- SOCKS5 is rejected for Mosh;
- the backend-to-remote Mosh leg is IPv4-only with the current dependency;
- Mosh expects a UTF-8 remote locale in this first implementation.

## Testing And Validation

Backend tests cover request parsing, `mosh-server` command rendering,
bootstrap-output parsing, SOCKS5 rejection, IPv4 host resolution, command
registration, and command lifecycle closure. Tests avoid real remote SSH or UDP
networking by injecting resolvers and session builders.

Frontend tests cover launcher parsing, field validation, initialization failure
mapping, connect failure handling, and control data/resize behavior. The final
validation uses the current main-branch tooling: Go tests, JavaScript tests and
lint, `npm run generate`, and targeted ESLint for Vite/build files if those
files are touched.

Generated static assets follow the current `origin/main` architecture and are
not copied from the old branch.
