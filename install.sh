#!/usr/bin/env sh
set -eu

REPO="madhukraft/nether"
BIN_DIR="$HOME/.local/bin"
BIN_PATH="$BIN_DIR/nether"

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

if command -v curl >/dev/null 2>&1; then
    NEW_VER=$(curl -sIo /dev/null -w '%{redirect_url}' "$URL" | sed 's|.*/download/\([^/]*\)/.*|\1|')
    NEW_VER=${NEW_VER:-unknown}
else
    NEW_VER="unknown"
fi

if [ -f "$BIN_PATH" ]; then
    OLD_VER="$("$BIN_PATH" -V 2>/dev/null || echo "")"
    echo "Existing nether binary found at $BIN_PATH"
    if [ -n "$OLD_VER" ]; then
        echo "  Current version: v$OLD_VER"
    fi
    echo "  New version:     v$NEW_VER"
    printf "Overwrite? [y/N]: "
    read -r response < /dev/tty || true
    case "$response" in
        y|Y) ;;
        *) echo "Aborting."; exit 1 ;;
    esac
fi

TMP_FILE="/tmp/nether-install"
echo "Downloading nether v$NEW_VER for $OS/$ARCH..."

if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$TMP_FILE" "$URL"
elif command -v wget >/dev/null 2>&1; then
    wget -qO "$TMP_FILE" "$URL"
else
    echo "error: curl or wget required"
    exit 1
fi

chmod +x "$TMP_FILE"
mkdir -p "$BIN_DIR"
mv "$TMP_FILE" "$BIN_PATH"
echo "Installed v$NEW_VER to $BIN_PATH"
echo "Run 'nether create' to get started."
