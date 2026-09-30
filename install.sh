```bash
#!/usr/bin/env bash

set -euo pipefail

REPO="Dev-syphax/lanserve"
BINARY="lanserve"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  linux)
    GOOS="Linux"
    ;;
  darwin)
    GOOS="Darwin"
    ;;
  *)
    echo "Unsupported operating system: $OS"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64)
    GOARCH="x86_64"
    ;;
  arm64|aarch64)
    GOARCH="arm64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name":' \
  | sed -E 's/.*"([^"]+)".*/\1/')"

if [ -z "$VERSION" ]; then
  echo "Could not determine latest version."
  exit 1
fi

ARCHIVE="${BINARY}_${VERSION#v}_${GOOS}_${GOARCH}.tar.gz"

URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

echo "Installing ${BINARY} ${VERSION}..."
echo "Downloading: ${URL}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

curl -fsSL "$URL" -o "$TMP_DIR/$ARCHIVE"

mkdir -p "$INSTALL_DIR"

tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR"

install -m 755 "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"

echo
echo "✓ ${BINARY} ${VERSION} installed to:"
echo "  ${INSTALL_DIR}/${BINARY}"
echo

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  echo "Add this to your shell configuration:"
  echo
  echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
  echo
fi

echo "Run:"
echo "  ${BINARY} --help"
```
