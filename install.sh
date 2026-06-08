#!/usr/bin/env bash
set -euo pipefail

REPO="IshpreetSingh8264/gndec-qp"
BINARY="qp"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

info()  { echo -e "${BLUE}::${NC} $1"; }
ok()    { echo -e "${GREEN}✓${NC} $1"; }
warn()  { echo -e "${YELLOW}!${NC} $1"; }
err()   { echo -e "${RED}✗${NC} $1"; }

detect_os_arch() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"

    case "$OS" in
        linux)   OS="linux" ;;
        darwin)  OS="darwin" ;;
        mingw*|msys*|cygwin*) OS="windows" ;;
        *)       err "Unsupported OS: $OS"; exit 1 ;;
    esac

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *)            err "Unsupported architecture: $ARCH"; exit 1 ;;
    esac

    if [ "$OS" = "windows" ]; then
        err "Windows detected. Please use install.ps1 instead."
        err "Run the following in PowerShell:"
        err '  powershell -c "iwr -Uri https://raw.githubusercontent.com/IshpreetSingh8264/gndec-qp/main/install.ps1 -OutFile install.ps1; .\install.ps1"'
        exit 1
    fi

    info "Detected: $OS / $ARCH"
}

fetch_latest_version() {
    if command -v curl &>/dev/null; then
        VERSION=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
    elif command -v wget &>/dev/null; then
        VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
    else
        err "Neither curl nor wget found. Please install one of them."
        exit 1
    fi

    if [ -z "$VERSION" ] || [ "$VERSION" = "null" ]; then
        warn "Could not fetch latest release from GitHub API."
        warn "Falling back to 'latest' tag."
        VERSION="latest"
    fi
    info "Release: $VERSION"
}

download_binary() {
    if [ "$VERSION" = "latest" ]; then
        DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/${BINARY}-${OS}-${ARCH}"
    else
        DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/${BINARY}-${OS}-${ARCH}"
    fi

    TMP_DIR=$(mktemp -d)
    TMP_FILE="$TMP_DIR/$BINARY"

    info "Downloading $BINARY for $OS/$ARCH..."
    echo "  $DOWNLOAD_URL"

    if command -v curl &>/dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE" --progress-bar
    elif command -v wget &>/dev/null; then
        wget -q --show-progress "$DOWNLOAD_URL" -O "$TMP_FILE"
    fi

    if [ ! -s "$TMP_FILE" ]; then
        err "Download failed. Binary may not exist for this platform yet."
        err "Try building from source: go install github.com/$REPO@latest"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    chmod +x "$TMP_FILE"
    ok "Downloaded successfully"
}

install_binary() {
    if [ -w "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
    elif [ -w "$HOME/.local/bin" ]; then
        INSTALL_DIR="$HOME/.local/bin"
    else
        INSTALL_DIR="$HOME/bin"
    fi

    mkdir -p "$INSTALL_DIR"
    mv "$TMP_FILE" "$INSTALL_DIR/$BINARY"
    rm -rf "$TMP_DIR"

    ok "Installed to $INSTALL_DIR/$BINARY"

    if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
        warn "$INSTALL_DIR is not in your PATH."
        warn "Add it by running:"
        case "$SHELL" in
            */zsh) echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.zshrc && source ~/.zshrc" ;;
            */bash) echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.bashrc && source ~/.bashrc" ;;
            *)      echo "  export PATH=\"\$PATH:$INSTALL_DIR\"" ;;
        esac
    fi
}

verify() {
    if command -v "$BINARY" &>/dev/null; then
        ok "Installation verified: $($BINARY --help 2>&1 | head -1)"
    else
        warn "Binary installed but not found in PATH."
        warn "Run 'qp --help' after adding the directory to PATH."
    fi
}

main() {
    echo ""
    echo -e "${GREEN}┌──────────────────────────────────────────┐${NC}"
    echo -e "${GREEN}│  GNDEC Question Paper Downloader Install  │${NC}"
    echo -e "${GREEN}└──────────────────────────────────────────┘${NC}"
    echo ""

    detect_os_arch
    fetch_latest_version
    download_binary
    install_binary
    verify

    echo ""
    ok "Installation complete!"
    echo ""
    echo "  Run '$BINARY' to launch the interactive TUI"
    echo "  Run '$BINARY --code PCIT-114' for CLI mode"
    echo "  Run '$BINARY --code PCIT-114 --auto' to auto-open PDFs"
    echo ""
}

main
