#!/usr/bin/env bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[${PLUGIN_NAME}]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[${PLUGIN_NAME} Warning]${NC} $1"
}

error() {
    echo -e "${RED}[${PLUGIN_NAME} Error]${NC} $1"
    exit 1
}

get_plugin_info() {
    PLUGIN_NAME=$(awk '/name:/ {print $2}' "$HELM_PLUGIN_DIR/plugin.yaml" | tr -d '"')
    PLUGIN_VERSION=$(awk '/version:/ {print $2}' "$HELM_PLUGIN_DIR/plugin.yaml" | tr -d '"')
    GITHUB_REPO="cstanislawski/helm-${PLUGIN_NAME}"

    if [ -z "$PLUGIN_NAME" ] || [ -z "$PLUGIN_VERSION" ]; then
        error "Failed to extract plugin information from plugin.yaml"
    fi

    log "Installing ${PLUGIN_NAME} version ${PLUGIN_VERSION}"
}

get_system_info() {
    OS=$(uname | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv8*) ARCH="arm64" ;;
        armv7*) ARCH="arm" ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac

    case $OS in
        darwin)
            OS="darwin"
            if [ "$ARCH" = "arm64" ]; then
                log "Detected Apple Silicon (M1/M2) Mac"
            fi
            ;;
        linux) OS="linux" ;;
        mingw*|msys*) OS="windows" ;;
        *)
            error "Unsupported operating system: $OS"
            ;;
    esac

    log "Detected system: ${OS}-${ARCH}"
}

verify_helm() {
    if ! command -v helm &> /dev/null; then
        error "Helm is not installed. Please install Helm first: https://helm.sh/docs/intro/install/"
    fi
    HELM_VERSION=$(helm version --template="{{ .Version }}" | cut -d "+" -f 1 | cut -c 2-)
    log "Detected Helm version: $HELM_VERSION"
}

download_plugin() {
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/v${PLUGIN_VERSION}/helm-${PLUGIN_NAME}-${OS}-${ARCH}.tar.gz"
    log "Attempting to download from: $download_url"

    if command -v curl &> /dev/null; then
        if ! curl -sS -L -f "$download_url" -o "helm-${PLUGIN_NAME}.tar.gz"; then
            error "Failed to download plugin. Please check if the release for ${OS}-${ARCH} exists."
        fi
    elif command -v wget &> /dev/null; then
        if ! wget -q "$download_url" -O "helm-${PLUGIN_NAME}.tar.gz"; then
            error "Failed to download plugin. Please check if the release for ${OS}-${ARCH} exists."
        fi
    else
        error "Neither curl nor wget found. Please install one of them and try again."
    fi

    if [ ! -f "helm-${PLUGIN_NAME}.tar.gz" ]; then
        error "Download seemed to succeed, but helm-${PLUGIN_NAME}.tar.gz not found."
    fi

    log "Download completed. Verifying archive..."
    if ! tar tf "helm-${PLUGIN_NAME}.tar.gz" &> /dev/null; then
        error "Downloaded file is not a valid tar.gz archive. Please check the release file on GitHub."
    fi
}

install_plugin() {
    local install_dir="${HELM_PLUGIN_DIR}"

    log "Extracting plugin..."
    if ! tar -xzf "helm-${PLUGIN_NAME}.tar.gz" -C "$install_dir"; then
        error "Failed to extract plugin. The archive may be corrupted."
    fi
    rm "helm-${PLUGIN_NAME}.tar.gz"

    local bin_path="$install_dir/bin/helm-${PLUGIN_NAME}"
    if [ ! -f "$bin_path" ]; then
        error "Plugin binary not found after extraction. Expected at: $bin_path"
    fi

    log "Setting execute permissions..."
    chmod +x "$bin_path"

    log "Plugin installed successfully to: $install_dir"
}

verify_installation() {
    if helm plugin list | grep -q "${PLUGIN_NAME}"; then
        log "Verified: ${PLUGIN_NAME} is now installed and ready to use!"
        log "Try running: helm ${PLUGIN_NAME} --help"
    else
        error "Verification failed. The plugin doesn't seem to be properly installed."
    fi
}

main() {
    get_plugin_info
    get_system_info
    verify_helm
    download_plugin
    install_plugin
    verify_installation

    log "Installation complete!"
}

main "$@"
