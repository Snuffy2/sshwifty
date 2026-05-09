# Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
# Copyright (C) 2026 Snuffy2
# SPDX-License-Identifier: AGPL-3.0-only

# Sshwifty is built as a static Go binary, but the production build also needs
# Node because `npm run build` first runs Vite and then `go generate ./...` to
# embed the generated frontend assets into Go source.
#
# The builder stage therefore starts from the official Go image and installs
# Node 24 for the frontend toolchain. It installs npm and Go dependencies before
# copying the full source tree so Docker can reuse those dependency layers when
# only application code changes.
#
# The runtime stage is Alpine and contains only the compiled `/sshwifty` binary,
# a small entrypoint wrapper for optional Docker TLS environment variables, and a
# curated source bundle under `/sshwifty-src` for license/source availability.
# The source bundle is intentionally explicit instead of `COPY .` so local
# operator files such as real config JSON are not accidentally baked into images.

# Build the application binary
FROM golang:1.26-trixie AS builder
WORKDIR /src
ARG SSHWIFTY_VERSION=dev
RUN set -eux; \
    export DEBIAN_FRONTEND=noninteractive; \
    apt-get update; \
    apt-get install -y --no-install-recommends npm; \
    npm install -g n; \
    n 24; \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/*
COPY go.mod go.sum package.json package-lock.json ./
RUN set -ex && \
    npm ci && \
    go mod download
COPY . .
RUN set -ex && \
    SSHWIFTY_VERSION="$SSHWIFTY_VERSION" npm run build && \
    mv ./sshwifty /

# Build the final image for running
FROM alpine:3.23
ENV SSHWIFTY_DIALTIMEOUT=10 \
    SSHWIFTY_HOOKTIMEOUT=30 \
    SSHWIFTY_LISTENINTERFACE=0.0.0.0 \
    SSHWIFTY_LISTENPORT=8182 \
    SSHWIFTY_INITIALTIMEOUT=0 \
    SSHWIFTY_READTIMEOUT=0 \
    SSHWIFTY_WRITETIMEOUT=0 \
    SSHWIFTY_HEARTBEATTIMEOUT=0 \
    SSHWIFTY_READDELAY=0 \
    SSHWIFTY_WRITEDELAY=0
COPY --from=builder /sshwifty /
COPY application /sshwifty-src/application
COPY ui /sshwifty-src/ui
COPY LICENSE.md README.md CONFIGURATION.md DEPENDENCIES.md /sshwifty-src/
COPY go.mod go.sum package.json package-lock.json /sshwifty-src/
COPY sshwifty.go vite.config.js eslint.config.mjs /sshwifty-src/
COPY scripts /sshwifty-src/scripts
COPY preset.example.json sshwifty.conf.example.json /sshwifty-src/
RUN set -ex && \
    adduser -D sshwifty && \
    chmod +x /sshwifty && \
    printf '%s\n' \
        '#!/bin/sh' \
        'set -e' \
        '[ -z "$SSHWIFTY_DOCKER_TLSCERT" ] || { printf "%s" "$SSHWIFTY_DOCKER_TLSCERT" > /tmp/cert && chmod 600 /tmp/cert; }' \
        '[ -z "$SSHWIFTY_DOCKER_TLSCERTKEY" ] || { printf "%s" "$SSHWIFTY_DOCKER_TLSCERTKEY" > /tmp/certkey && chmod 600 /tmp/certkey; }' \
        'if [ -f "/tmp/cert" ] && [ -f "/tmp/certkey" ]; then' \
        '    exec env SSHWIFTY_TLSCERTIFICATEFILE=/tmp/cert SSHWIFTY_TLSCERTIFICATEKEYFILE=/tmp/certkey /sshwifty' \
        'fi' \
        'exec /sshwifty' \
        > /sshwifty.sh && \
    chmod +x /sshwifty.sh
USER sshwifty
EXPOSE 8182
ENTRYPOINT [ "/sshwifty.sh" ]
CMD []
