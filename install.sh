#!/usr/bin/env sh
set -eu

REPO="madhukraft/nether"
BIN_DIR="/usr/local/bin"

case "$(uname -s)" in
    Linux)  OS="linux" ;;
    Darwin) OS="darwin" ;;
    *)      echo "Unsupported OS"; exit 1 ;;
esac

case "$(uname -m)" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)            echo "Unsupported architecture"; exit 1 ;;
esac

URL="https://github.com/$REPO/releases/latest/download/nether-$OS-$ARCH"
TMP_FILE="/tmp/nether-install"

echo "Downloading nether for $OS/$ARCH..."

if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$TMP_FILE" "$URL"
elif command -v wget >/dev/null 2>&1; then
    wget -qO "$TMP_FILE" "$URL"
else
    echo "error: curl or wget required"
    exit 1
fi

chmod +x "$TMP_FILE"

if [ -w "$BIN_DIR" ]; then
    mv "$TMP_FILE" "$BIN_DIR/nether"
else
    echo "sudo required to install to $BIN_DIR"
    sudo mv "$TMP_FILE" "$BIN_DIR/nether"
fi

echo "Installed to $BIN_DIR/nether"
echo "Run 'nether create' to get started."
