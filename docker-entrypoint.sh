#!/bin/sh
# Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
# Copyright (C) 2026 Snuffy2
# SPDX-License-Identifier: AGPL-3.0-only

set -e

rm -f /tmp/cert /tmp/certkey

if [ -n "$SSHWIFTY_DOCKER_TLSCERT" ] && [ -z "$SSHWIFTY_DOCKER_TLSCERTKEY" ]; then
    echo "SSHWIFTY_DOCKER_TLSCERTKEY is required when SSHWIFTY_DOCKER_TLSCERT is set" >&2
    exit 1
fi

if [ -z "$SSHWIFTY_DOCKER_TLSCERT" ] && [ -n "$SSHWIFTY_DOCKER_TLSCERTKEY" ]; then
    echo "SSHWIFTY_DOCKER_TLSCERT is required when SSHWIFTY_DOCKER_TLSCERTKEY is set" >&2
    exit 1
fi

if [ -n "$SSHWIFTY_DOCKER_TLSCERT" ] && [ -n "$SSHWIFTY_DOCKER_TLSCERTKEY" ]; then
    printf "%s" "$SSHWIFTY_DOCKER_TLSCERT" > /tmp/cert
    chmod 600 /tmp/cert
    printf "%s" "$SSHWIFTY_DOCKER_TLSCERTKEY" > /tmp/certkey
    chmod 600 /tmp/certkey
    exec env SSHWIFTY_TLSCERTIFICATEFILE=/tmp/cert SSHWIFTY_TLSCERTIFICATEKEYFILE=/tmp/certkey /sshwifty
fi

exec /sshwifty
