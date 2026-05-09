#!/bin/sh
# Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
# Copyright (C) 2026 Snuffy2
# SPDX-License-Identifier: AGPL-3.0-only

set -e

if [ -n "$SSHWIFTY_DOCKER_TLSCERT" ]; then
    printf "%s" "$SSHWIFTY_DOCKER_TLSCERT" > /tmp/cert
    chmod 600 /tmp/cert
fi

if [ -n "$SSHWIFTY_DOCKER_TLSCERTKEY" ]; then
    printf "%s" "$SSHWIFTY_DOCKER_TLSCERTKEY" > /tmp/certkey
    chmod 600 /tmp/certkey
fi

if [ -f "/tmp/cert" ] && [ -f "/tmp/certkey" ]; then
    exec env SSHWIFTY_TLSCERTIFICATEFILE=/tmp/cert SSHWIFTY_TLSCERTIFICATEKEYFILE=/tmp/certkey /sshwifty
fi

exec /sshwifty
