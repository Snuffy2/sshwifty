# Sshwifty Web SSH & Telnet Client

Sshwifty is a web-based SSH and Telnet client that lets you connect to remote
systems from a browser.

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
      - ./sshwifty.conf.json:/etc/sshwifty/sshwifty.conf.json:ro
    environment:
      SSHWIFTY_CONFIG: /etc/sshwifty/sshwifty.conf.json
```

Then open `http://localhost:8182`.

The container image does not bundle the repository source tree. Published
images include `/SOURCE.md`, the running app source link, and an OCI source label
with an immutable GitHub commit archive URL for the source used to produce that
image. Local builds default to
[`github.com/Snuffy2/sshwifty`](https://github.com/Snuffy2/sshwifty) unless
`SSHWIFTY_SOURCE_URL` is provided as a Docker build argument.

For reverse proxy deployments, publish the service only on localhost:

```yaml
ports:
  - "127.0.0.1:8182:8182"
```

## Configuration

Sshwifty can be configured with a JSON configuration file or environment
variables. See [CONFIGURATION.md](CONFIGURATION.md) for the full configuration
reference.

The Docker Compose example above mounts `./sshwifty.conf.json` and points
`SSHWIFTY_CONFIG` at the mounted file. Start from
[sshwifty.conf.example.json](sshwifty.conf.example.json) when creating your own
configuration.

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
