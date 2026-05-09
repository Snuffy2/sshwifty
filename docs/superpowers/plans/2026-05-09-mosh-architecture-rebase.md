# Mosh Architecture Rebase Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild Mosh support cleanly on top of current `origin/main`.

**Architecture:** Treat the old `feature/mosh-support` branch as source
material, not as the final history. Port Mosh into the current Go command
lifecycle, Vite UI entrypoints, current module path, split configuration docs,
and current validation tooling.

**Tech Stack:** Go 1.26, `golang.org/x/crypto/ssh`, `github.com/unixshells/mosh-go`,
Vue/Vite, Vitest, ESLint, `go test`, `npm run generate`.

---

## File Structure

- Modify `go.mod` and `go.sum` to add `github.com/unixshells/mosh-go`.
- Modify `application/command/commander.go` only if a proxy capability flag is
  required for Mosh SOCKS5 rejection.
- Modify `application/application.go` or server wiring only if needed to expose
  the SOCKS5 flag to command configuration.
- Modify `application/commands/commands.go` to register command ID `0x02` as
  `Mosh`.
- Create `application/commands/mosh.go` for the backend command FSM.
- Create `application/commands/mosh_bootstrap.go` for `mosh-server` command
  rendering and bootstrap output parsing.
- Create `application/commands/mosh_session.go` for the `mosh-go` adapter.
- Create or port backend tests beside those files.
- Create `ui/commands/mosh.js` for the frontend Mosh wizard/executor.
- Create `ui/control/mosh.js` for the console control adapter.
- Modify `ui/app.js` to register Mosh command/control factories.
- Create or port frontend tests beside those files.
- Modify `README.md`, `CONFIGURATION.md`, `preset.example.json`, and
  `sshwifty.conf.example.json` for Mosh docs and examples.

---

### Task 1: Rebuild Branch Base

**Files:**
- Inspect only: all current files

- [ ] **Step 1: Confirm current state**

Run: `git status --short --branch`

Expected: current branch is `feature/mosh-support`; only known untracked
generated docs/static assets plus the new plan may appear.

- [ ] **Step 2: Save old branch pointer**

Run: `git branch backup/feature-mosh-support-pre-rebuild HEAD`

Expected: a local backup branch points to the old Mosh branch plus design/plan
docs before the clean rebuild.

- [ ] **Step 3: Create clean rebuild branch from current main**

Run: `git switch -c feature/mosh-support-rebased origin/main`

Expected: new branch starts from current `origin/main`; the active branch is
`feature/mosh-support-rebased`.

- [ ] **Step 4: Restore approved spec and plan**

Run:
`git checkout backup/feature-mosh-support-pre-rebuild -- docs/superpowers/specs/2026-05-09-mosh-architecture-rebase-design.md docs/superpowers/plans/2026-05-09-mosh-architecture-rebase.md`

Expected: the approved spec and this plan exist on the rebuild branch.

- [ ] **Step 5: Commit planning docs**

Run:
`git add docs/superpowers/specs/2026-05-09-mosh-architecture-rebase-design.md docs/superpowers/plans/2026-05-09-mosh-architecture-rebase.md && git commit -m "docs: plan mosh architecture rebase"`

Expected: the rebuild branch has a planning-doc commit on top of `origin/main`.

---

### Task 2: Backend Command And Bootstrap

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`
- Modify: `application/commands/commands.go`
- Create: `application/commands/mosh.go`
- Create: `application/commands/mosh_bootstrap.go`
- Create: `application/commands/mosh_session.go`
- Create: `application/commands/mosh_bootstrap_test.go`
- Create: `application/commands/mosh_session_test.go`
- Create: `application/commands/mosh_test.go`

- [ ] **Step 1: Port backend source from old branch**

Use these old-branch files as source material:

```text
backup/feature-mosh-support-pre-rebuild:application/commands/mosh.go
backup/feature-mosh-support-pre-rebuild:application/commands/mosh_bootstrap.go
backup/feature-mosh-support-pre-rebuild:application/commands/mosh_session.go
```

Update imports from `github.com/nirui/sshwifty/...` to
`github.com/Snuffy2/sshwifty/...`.

- [ ] **Step 2: Fit current command lifecycle**

Keep these backend constants and wire markers:

```go
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
```

`Bootup` must return `d.local` only after parsing user, address, auth method,
and optional `Mosh Server` metadata. `Close` must close credential/fingerprint
channels, cancel the base context, close any active Mosh session, and wait for
the remote worker to send `HeaderClose`.

- [ ] **Step 3: Add Mosh registration**

Modify `application/commands/commands.go` so `New()` registers:

```go
func New() command.Commands {
	return command.Commands{
		command.Register("Telnet", newTelnet, parseTelnetConfig),
		command.Register("SSH", newSSH, parseSSHConfig),
		command.Register("Mosh", newMosh, parseMoshConfig),
	}
}
```

- [ ] **Step 4: Restore backend tests**

Port old tests and update them for current lifecycle behavior. Required test
coverage:

```text
parseMoshConnectLine parses valid MOSH CONNECT output
parseMoshConnectLine rejects malformed output
renderMoshServerCommand defaults to mosh-server
renderMoshServerCommand uses configured Mosh Server
Mosh command rejects SOCKS5 configuration
Mosh command resolves hostnames to IPv4 before mosh-go
Mosh command rejects IPv6-only targets
Mosh command registration exposes command ID 0x02
```

- [ ] **Step 5: Run focused backend tests**

Run:
`go test ./application/commands ./application/command ./application/configuration`

Expected: all packages pass.

- [ ] **Step 6: Commit backend**

Run:
`git add go.mod go.sum application/commands && git commit -m "feat: add backend mosh command"`

Expected: backend Mosh command and tests are committed.

---

### Task 3: Frontend Command And Control

**Files:**
- Modify: `ui/app.js`
- Create: `ui/commands/mosh.js`
- Create: `ui/commands/mosh_test.js`
- Create: `ui/control/mosh.js`
- Create: `ui/control/mosh_test.js`

- [ ] **Step 1: Port frontend command**

Use the old `ui/commands/mosh.js` as source material and update it to match the
current JSDoc/import style used by `ui/commands/ssh.js`.

Keep command ID `0x02`, default port `22`, and default Mosh server
`mosh-server`.

- [ ] **Step 2: Preserve launcher compatibility**

`Command.launch()` must accept:

```text
user@host|AuthMethod
user@host|AuthMethod|charset
user@host|AuthMethod|charset|encodedMoshServer
```

`Command.launcher()` must omit the fourth field when `moshServer` is empty or
`mosh-server`.

- [ ] **Step 3: Add frontend control**

Create `ui/control/mosh.js` using the SSH control shape: decode stdout through
the selected charset, send encoded stdin, send binary input unchanged, forward
resize events, and reject the receive subscription when the stream completes.

- [ ] **Step 4: Register Mosh in app startup**

Modify `ui/app.js` imports and constructors:

```js
import * as mosh from "./commands/mosh.js";
import * as moshctl from "./control/mosh.js";
```

```js
new Controls([
  new telnetctl.Telnet(uiControlColors),
  new sshctl.SSH(uiControlColors),
  new moshctl.Mosh(uiControlColors),
])
```

```js
new Commands([new telnet.Command(), new ssh.Command(), new mosh.Command()])
```

- [ ] **Step 5: Restore frontend tests**

Port and update tests for launcher parsing, SOCKS5 initialization error text,
Mosh Server field validation, and control send/resize behavior.

- [ ] **Step 6: Run focused frontend tests and lint**

Run:
`npx vitest run ui/commands/mosh_test.js ui/control/mosh_test.js`

Run:
`npm run lint`

Expected: tests and lint pass.

- [ ] **Step 7: Commit frontend**

Run:
`git add ui/app.js ui/commands/mosh.js ui/commands/mosh_test.js ui/control/mosh.js ui/control/mosh_test.js && git commit -m "feat: add frontend mosh command"`

Expected: frontend Mosh command/control are committed.

---

### Task 4: Documentation And Examples

**Files:**
- Modify: `README.md`
- Modify: `CONFIGURATION.md`
- Modify: `preset.example.json`
- Modify: `sshwifty.conf.example.json`

- [ ] **Step 1: Add Mosh examples**

Add a Mosh preset example with:

```json
{
  "Title": "Example Mosh",
  "Type": "Mosh",
  "Host": "ssh.example.com:22",
  "Meta": {
    "User": "guest",
    "Authentication": "Password",
    "Encoding": "utf-8",
    "Mosh Server": "mosh-server"
  }
}
```

- [ ] **Step 2: Document v1 constraints**

Add text stating that Mosh v1:

```text
uses SSH only for bootstrap;
uses UDP only between backend/container and remote host;
requires mosh-server on the remote host;
rejects SOCKS5 proxy configurations;
requires an IPv4 backend-to-remote UDP target with the current mosh-go client;
expects a UTF-8 remote locale.
```

- [ ] **Step 3: Run formatting-sensitive checks**

Run:
`npx prettier --check README.md CONFIGURATION.md preset.example.json sshwifty.conf.example.json`

Expected: files are formatted or the command reports only files that need
formatting; if needed, run `npx prettier --write` on the same paths.

- [ ] **Step 4: Commit docs**

Run:
`git add README.md CONFIGURATION.md preset.example.json sshwifty.conf.example.json && git commit -m "docs: document mosh support"`

Expected: docs/examples are committed.

---

### Task 5: Full Validation And Branch Finish

**Files:**
- Modify if generated: `application/controller/static_assets/*`

- [ ] **Step 1: Run full Go tests**

Run: `go test ./...`

Expected: all Go packages pass.

- [ ] **Step 2: Run frontend validation**

Run: `npm run lint`

Run: `npx vitest run`

Expected: lint and all Vitest tests pass.

- [ ] **Step 3: Regenerate static assets**

Run: `npm run generate`

Expected: Vite builds assets and syncs them into the current static asset
location without restoring Webpack-era generated files.

- [ ] **Step 4: Run final status check**

Run: `git status --short --branch`

Expected: only intentional generated static asset changes remain, or the tree is
clean if generation produced no tracked changes.

- [ ] **Step 5: Commit generated assets if changed**

Run:
`git add application/controller/static_assets && git commit -m "build: regenerate static assets for mosh"`

Expected: commit is created only if tracked generated assets changed.

- [ ] **Step 6: Rename branch if requested**

If the user wants the original branch name rewritten, run:
`git branch -M feature/mosh-support`

Expected: the clean rebuilt branch now carries the original branch name locally.

---

## Self-Review

Spec coverage: backend command, frontend command/control, docs/config examples,
v1 constraints, old launcher compatibility, and validation are all represented
by tasks above.

Placeholder scan: no TBD/TODO/fill-in placeholders are used as task content.

Type consistency: the plan consistently uses `moshServer` in frontend config,
`Mosh Server` for preset/request metadata, command ID `0x02`, and backend marker
names from the old branch.
