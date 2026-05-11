# Sshwifty Web SSH, Telnet & Mosh Client

Sshwifty is a web-based SSH, Telnet, and Mosh client that lets you connect to
remote systems from a browser.

This repository is a fork of
[nirui/sshwifty](https://github.com/nirui/sshwifty). The original project and
its design are the work of [@nirui](https://github.com/nirui), whose excellent
work made this fork possible.

![Screenshot](Screenshot.png)

## Supported browsers

ES2020-era browsers and newer, including Chrome 80+, Edge 80+, Firefox 78+, and Safari 14+.

## Docker

Run Sshwifty with Docker Compose:

```yaml
services:
  sshwifty:
    image: ghcr.io/snuffy2/sshwifty:latest
    container_name: sshwifty
    restart: unless-stopped
    ports:
      - "8182:8182"
    volumes:
      - ./config:/etc/sshwifty
    environment:
      SSHWIFTY_CONFIG: /etc/sshwifty/sshwifty.conf.json
      # Optional: base64-encoded 32-byte key for encrypted preset passwords.
      # Generate with: openssl rand -base64 32
      # SSHWIFTY_PRESET_SECRET_KEY: "replace-with-generated-key"
      # Optional admin key. Users enter this in the same prompt as SharedKey.
      # SSHWIFTY_ADMIN_KEY: "replace-with-admin-key"
```

Then open `http://localhost:8182`.

The container image does not bundle the repository source tree. Published
images include `/SOURCE.md`, the running app source link, and an OCI source label
with an immutable GitHub commit archive URL for the source used to produce that
image. Local Docker builds default to
[`github.com/Snuffy2/sshwifty`](https://github.com/Snuffy2/sshwifty) unless
`SSHWIFTY_SOURCE_URL` is provided as a Docker build argument for `/SOURCE.md` and
the in-app source link.

For reverse proxy deployments, publish the service only on localhost:

```yaml
ports:
  - "127.0.0.1:8182:8182"
```

## Configuration

Sshwifty can be configured with a JSON configuration file or environment
variables. See [CONFIGURATION.md](CONFIGURATION.md) for the full configuration
reference.

The Docker Compose example above mounts `./config` as a writable configuration
directory and points `SSHWIFTY_CONFIG` at `sshwifty.conf.json` inside it. Start
from [sshwifty.conf.example.json](sshwifty.conf.example.json) when creating your
own configuration.

Writable file-backed configuration enables preset updates from the UI, such as
saving SSH/Mosh fingerprints. If `SSHWIFTY_PRESET_SECRET_KEY` is set, plaintext
preset `Password` values are migrated on startup to `Encrypted Password` and the
plaintext value is removed from the JSON file. That key must be set through the
environment, not in JSON. Without that key, plaintext password presets continue
to work as before. Full preset add/edit/remove API writes require admin access.
Users enter either `SharedKey` or `AdminKey` in the same authentication prompt:
`SharedKey` grants user access, while `AdminKey` grants admin access. If
`AdminKey` is blank, `SharedKey` grants admin access. If both keys are blank,
everyone gets admin access without authentication. Fingerprint saves from the
current UI remain limited to the selected preset's fingerprint.

Generate a preset secret key with one of these commands:

```sh
# macOS/Linux
openssl rand -base64 32
```

```powershell
# Windows PowerShell
$rng = [Security.Cryptography.RandomNumberGenerator]::Create()
$bytes = New-Object byte[] 32
$rng.GetBytes($bytes)
[Convert]::ToBase64String($bytes)
```

Mosh support is available in v1 with SSH used for bootstrap only. The browser
connection to Sshwifty still uses WebSocket, while Mosh data flows over UDP
between the backend container and the remote host. Remote hosts need
`mosh-server` installed, SOCKS5 is not supported for Mosh, the backend-to-host
Mosh leg is IPv4-only, and terminal encoding is fixed to UTF-8.

## Running From Source

Use this path for development.

Prerequisites:

- `git`
- `go`
- `node` 24 or newer
- `npm`

Build the frontend assets and backend binary:

```sh
git clone https://github.com/Snuffy2/sshwifty.git
cd sshwifty
npm ci
npm run build
```

Run the development server:

```sh
npm run dev
```

The development command starts the Go backend with `sshwifty.conf.example.json`
and serves the frontend through Vite with HMR. Vite proxies backend routes such
as `/sshwifty/socket` to the Go process.

The generated production binary is written to `./sshwifty` by `npm run build`.

Useful development checks:

```sh
npm run generate
npm run testonly
npm run lint
go test ./...
```

`npm run generate` produces the Vite assets and then refreshes the embedded Go
static assets.

## License

Code in this project is licensed under AGPL-3.0-only. See
[LICENSE.md](LICENSE.md) for details.

Third-party components are licensed under their respective licenses. See
[DEPENDENCIES.md](DEPENDENCIES.md) for dependency copyright and license details.
