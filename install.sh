#!/usr/bin/env sh
set -eu

REPO="madhukraft/nether"
BIN_DIR="${HOME}/.local/bin"
BIN_PATH="${BIN_DIR}/nether"

detect_os() {
    case "$(uname -s)" in
        Linux)  echo "linux" ;;
        Darwin) echo "darwin" ;;
        *)      echo "" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *)            echo "" ;;
    esac
}

OS=$(detect_os)
ARCH=$(detect_arch)

if [ -z "$OS" ]; then
    echo "Error: Unsupported OS '$(uname -s)'. Expected Linux or Darwin." >&2
    exit 1
fi
if [ -z "$ARCH" ]; then
    echo "Error: Unsupported architecture '$(uname -m)'. Expected x86_64/amd64 or aarch64/arm64." >&2
    exit 1
fi

URL="https://github.com/${REPO}/releases/latest/download/nether-${OS}-${ARCH}"

# Detect latest version from GitHub redirect
NEW_VER="unknown"
if command -v curl >/dev/null 2>&1; then
    REDIRECT=$(curl -sSIo /dev/null -w '%{redirect_url}' "$URL" 2>/dev/null) || true
    if [ -n "$REDIRECT" ]; then
        NEW_VER=$(echo "$REDIRECT" | sed 's|.*/download/\([^/]*\)/.*|\1|')
    fi
elif command -v wget >/dev/null 2>&1; then
    REDIRECT=$(wget -SqO /dev/null "$URL" 2>&1 | grep 'Location:' | tail -1 | sed 's/.*Location: //') || true
    if [ -n "$REDIRECT" ]; then
        NEW_VER=$(echo "$REDIRECT" | sed 's|.*/download/\([^/]*\)/.*|\1|')
    fi
fi
if [ -z "$NEW_VER" ]; then
    NEW_VER="unknown"
fi

# Check existing installation
if [ -f "$BIN_PATH" ]; then
    OLD_VER=$("$BIN_PATH" -V 2>/dev/null || echo "")
    echo "Existing nether binary found at ${BIN_PATH}"
    if [ -n "$OLD_VER" ]; then
        echo "  Current version: v${OLD_VER}"
    fi
    echo "  New version:     v${NEW_VER}"
    printf "Overwrite? [y/N]: "
    read -r response < /dev/tty || response="n"
    case "${response}" in
        y|Y) ;;
        *)   echo "Aborting."; exit 1 ;;
    esac
fi

TMP_FILE=$(mktemp /tmp/nether-install.XXXXXX)
echo "Downloading nether v${NEW_VER} for ${OS}/${ARCH}..."
echo "  URL: ${URL}"

if command -v curl >/dev/null 2>&1; then
    curl -fSL --progress-bar -o "$TMP_FILE" "$URL" || {
        rm -f "$TMP_FILE"
        echo "Error: Download failed. Check the URL above or your network connection." >&2
        exit 1
    }
elif command -v wget >/dev/null 2>&1; then
    wget --show-progress -qO "$TMP_FILE" "$URL" || {
        rm -f "$TMP_FILE"
        echo "Error: Download failed. Check the URL above or your network connection." >&2
        exit 1
    }
else
    echo "Error: curl or wget required but neither was found." >&2
    exit 1
fi

if [ ! -s "$TMP_FILE" ]; then
    rm -f "$TMP_FILE"
    echo "Error: Downloaded file is empty. The release may not exist for ${OS}/${ARCH}." >&2
    exit 1
fi

chmod +x "$TMP_FILE"
mkdir -p "$BIN_DIR"
mv "$TMP_FILE" "$BIN_PATH"

"$BIN_PATH" -V >/dev/null 2>&1 || {
    echo "Warning: Installed binary failed to run. It may be incompatible with your system." >&2
    exit 1
}

echo "Installed v${NEW_VER} to ${BIN_PATH}"
echo "Run 'nether create' to get started."
