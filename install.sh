#!/usr/bin/env bash

set -euo pipefail

REPO="Dev-syphax/lanserve"
BINARY="lanserve"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux)
        OS="linux"
        ;;
    Darwin)
        OS="darwin"
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

VERSION="$(
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" |
    grep '"tag_name":' |
    sed -E 's/.*"([^"]+)".*/\1/'
)"

VERSION="${VERSION#v}"

ARCHIVE="${BINARY}_${VERSION}_${OS}_${ARCH}.tar.gz"

URL="https://github.com/${REPO}/releases/download/v${VERSION}/${ARCHIVE}"

echo "Installing ${BINARY} v${VERSION}..."
echo "Downloading ${URL}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

curl -fL "$URL" -o "$TMP_DIR/$ARCHIVE"

tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR"

mkdir -p "$INSTALL_DIR"

install -m 755 "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"

echo
echo " ${BINARY} installed successfully."
echo "  Location: ${INSTALL_DIR}/${BINARY}"
echo

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "Add this to your shell:"
    echo
    echo 'export PATH="$HOME/.local/bin:$PATH"'
fi