# Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
# Copyright (C) 2026 Snuffy2
# SPDX-License-Identifier: AGPL-3.0-only

# Sshwifty is built as a static Go binary, but the production build also needs
# Node because `npm run build` runs the frontend toolchain before building the
# Go binary with the generated assets embedded.
#
# The Docker build uses a separate Node dependency stage and then a Go builder
# stage. Node is copied from the dependency stage so frontend dependencies stay
# cacheable without installing Node through apt inside the Go image.
#
# The runtime stage is Alpine and contains only the compiled `/sshwifty` binary
# and a small entrypoint wrapper for optional Docker TLS environment variables.
# Source availability is provided by the app's GitHub source link and the OCI
# source metadata label rather than by copying source files into the image.

# Build the frontend dependencies
FROM node:24-trixie AS frontend-deps
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci

# Build the application binary
FROM golang:1.26-trixie AS builder
WORKDIR /src
ARG SSHWIFTY_VERSION=dev
COPY go.mod go.sum ./
RUN go mod download
COPY --from=frontend-deps /usr/local/bin/node /usr/local/bin/node
COPY --from=frontend-deps /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN set -ex && \
    ln -s /usr/local/lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm && \
    ln -s /usr/local/lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx
COPY --from=frontend-deps /src/node_modules ./node_modules
COPY package.json package-lock.json ./
COPY . .
RUN set -ex && \
    SSHWIFTY_VERSION="$SSHWIFTY_VERSION" npm run build && \
    mv ./sshwifty /

# Build the final image for running
FROM alpine:3.23
LABEL org.opencontainers.image.source="https://github.com/Snuffy2/sshwifty" \
    org.opencontainers.image.licenses="AGPL-3.0-only"
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
COPY docker-entrypoint.sh /sshwifty.sh
RUN set -ex && \
    adduser -D sshwifty && \
    chmod +x /sshwifty && \
    chmod +x /sshwifty.sh
USER sshwifty
EXPOSE 8182
ENTRYPOINT [ "/sshwifty.sh" ]
CMD []
