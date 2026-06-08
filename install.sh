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
    ARCHIVE_NAME="${BINARY}-${OS}-${ARCH}.tar.gz"
    if [ "$VERSION" = "latest" ]; then
        DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/$ARCHIVE_NAME"
    else
        DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE_NAME"
    fi

    TMP_DIR=$(mktemp -d)
    TMP_ARCHIVE="$TMP_DIR/$ARCHIVE_NAME"

    info "Downloading $BINARY for $OS/$ARCH..."
    echo "  $DOWNLOAD_URL"

    DOWNLOAD_OK=false
    if command -v curl &>/dev/null; then
        curl -sSL "$DOWNLOAD_URL" -o "$TMP_ARCHIVE" --progress-bar && DOWNLOAD_OK=true
    elif command -v wget &>/dev/null; then
        wget -q --show-progress "$DOWNLOAD_URL" -O "$TMP_ARCHIVE" && DOWNLOAD_OK=true
    fi

    if [ "$DOWNLOAD_OK" = false ] || [ ! -s "$TMP_ARCHIVE" ]; then
        rm -rf "$TMP_DIR"
        if [ "$DOWNLOAD_OK" = false ]; then
            warn "GitHub release download failed (CDN may still be propagating)."
        fi
        go_install_fallback
        return
    fi

    info "Extracting..."
    tar xzf "$TMP_ARCHIVE" -C "$TMP_DIR"

    TMP_FILE="$TMP_DIR/$BINARY"
    EXTRACTED=$(ls "$TMP_DIR" | grep -v "^$(basename "$TMP_ARCHIVE")$" | head -1)
    if [ -n "$EXTRACTED" ] && [ "$EXTRACTED" != "$BINARY" ]; then
        mv -f "$TMP_DIR/$EXTRACTED" "$TMP_FILE"
    elif [ ! -f "$TMP_FILE" ]; then
        err "Could not find binary in extracted archive."
        rm -rf "$TMP_DIR"
        exit 1
    fi

    chmod +x "$TMP_FILE"
    rm -f "$TMP_ARCHIVE"
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
    mv -f "$TMP_FILE" "$INSTALL_DIR/$BINARY" 2>/dev/null || {
        sudo mv -f "$TMP_FILE" "$INSTALL_DIR/$BINARY" 2>/dev/null || {
            err "Failed to install. Try: sudo cp $TMP_FILE $INSTALL_DIR/$BINARY"
            rm -rf "$TMP_DIR"
            exit 1
        }
    }
    rm -rf "$TMP_DIR" 2>/dev/null || true

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

go_install_fallback() {
    warn "Pre-built binary not found (no GitHub release yet)."
    warn "Falling back to building from source with 'go install'..."
    echo ""

    if ! command -v go &>/dev/null; then
        err "Go is not installed. Install Go from https://go.dev/dl/"
        err "Then run: go install github.com/$REPO@latest"
        exit 1
    fi

    info "Building $BINARY from source..."
    go install "github.com/$REPO@latest" 2>&1 || {
        err "Build failed."
        exit 1
    }

    GOPATH_BIN="$(go env GOPATH)/bin"
    if [ -f "$GOPATH_BIN/$BINARY" ]; then
        ok "Built and installed to $GOPATH_BIN/$BINARY"
        TMP_FILE="$GOPATH_BIN/$BINARY"
        install_binary
        TMP_FILE=""
        TMP_DIR=""
        return
        return
    else
        ok "Built successfully via 'go install'"
        ok "Binary should be at $GOPATH_BIN/$BINARY"
        if ! echo "$PATH" | tr ':' '\n' | grep -qx "$GOPATH_BIN"; then
            warn "Add it to PATH: export PATH=\"\$PATH:$GOPATH_BIN\""
        fi
        return
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

    # download_binary may have already installed via go_install_fallback
    if [ -n "${TMP_FILE:-}" ] && [ -f "$TMP_FILE" ]; then
        install_binary
    fi

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
