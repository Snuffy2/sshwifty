#!/bin/sh
set -e

rm -f /tmp/cert /tmp/certkey

if [ -n "$SHELLPORT_DOCKER_TLSCERT" ] && [ -z "$SHELLPORT_DOCKER_TLSCERTKEY" ]; then
    echo "SHELLPORT_DOCKER_TLSCERTKEY is required when SHELLPORT_DOCKER_TLSCERT is set" >&2
    exit 1
fi

if [ -z "$SHELLPORT_DOCKER_TLSCERT" ] && [ -n "$SHELLPORT_DOCKER_TLSCERTKEY" ]; then
    echo "SHELLPORT_DOCKER_TLSCERT is required when SHELLPORT_DOCKER_TLSCERTKEY is set" >&2
    exit 1
fi

if [ -n "$SHELLPORT_DOCKER_TLSCERT" ] && [ -n "$SHELLPORT_DOCKER_TLSCERTKEY" ]; then
    printf "%s" "$SHELLPORT_DOCKER_TLSCERT" > /tmp/cert
    chmod 600 /tmp/cert
    printf "%s" "$SHELLPORT_DOCKER_TLSCERTKEY" > /tmp/certkey
    chmod 600 /tmp/certkey
    exec env SHELLPORT_TLSCERTIFICATEFILE=/tmp/cert SHELLPORT_TLSCERTIFICATEKEYFILE=/tmp/certkey /shellport "$@"
fi

exec /shellport "$@"
