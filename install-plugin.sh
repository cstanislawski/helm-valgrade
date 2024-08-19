#!/usr/bin/env bash

set -euo pipefail

[ -z "$HELM_BIN" ] && HELM_BIN=$(command -v helm)
[ -z "$HELM_HOME" ] && HELM_HOME=$(helm env | grep 'HELM_DATA_HOME' | cut -d '=' -f2 | tr -d '"')
mkdir -p "$HELM_HOME"
: "${HELM_PLUGIN_DIR:="$HELM_HOME/plugins/helm-valgrade"}"

PLUGIN_NAME=$(awk '/name:/ {print $2}' "$HELM_PLUGIN_DIR/plugin.yaml" | tr -d '"')
PLUGIN_VERSION=$(awk '/version:/ {print $2}' "$HELM_PLUGIN_DIR/plugin.yaml" | tr -d '"')
GITHUB_REPO="cstanislawski/helm-${PLUGIN_NAME}"

log() { echo -e "\033[0;32m[${PLUGIN_NAME}]\033[0m $1"; }
warn() { echo -e "\033[1;33m[${PLUGIN_NAME} Warning]\033[0m $1"; }
error() { echo -e "\033[0;31m[${PLUGIN_NAME} Error]\033[0m $1"; exit 1; }

get_system_info() {
    OS=$(uname | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv8*) ARCH="arm64" ;;
        armv7*) ARCH="arm" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac

    case $OS in
        darwin) ;;
        linux) ;;
        mingw*|msys*) OS="windows" ;;
        *) error "Unsupported operating system: $OS" ;;
    esac

    log "Detected system: ${OS}-${ARCH}"
}

verify_helm() {
    command -v helm >/dev/null 2>&1 || error "Helm is not installed. Please install Helm first: https://helm.sh/docs/intro/install/"
    HELM_VERSION=$(helm version --template="{{ .Version }}" | cut -d "+" -f 1 | cut -c 2-)
    log "Detected Helm version: $HELM_VERSION"
}

download_plugin() {
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/v${PLUGIN_VERSION}/helm-${PLUGIN_NAME}-${OS}-${ARCH}.tar.gz"
    log "Downloading from: $download_url"

    if command -v curl >/dev/null 2>&1; then
        curl -sSL "$download_url" -o "helm-${PLUGIN_NAME}.tar.gz" || error "Failed to download plugin. Please check if the release for ${OS}-${ARCH} exists."
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$download_url" -O "helm-${PLUGIN_NAME}.tar.gz" || error "Failed to download plugin. Please check if the release for ${OS}-${ARCH} exists."
    else
        error "Neither curl nor wget found. Please install one of them and try again."
    fi

    [ -f "helm-${PLUGIN_NAME}.tar.gz" ] || error "Download seemed to succeed, but helm-${PLUGIN_NAME}.tar.gz not found."
    tar tf "helm-${PLUGIN_NAME}.tar.gz" >/dev/null 2>&1 || error "Downloaded file is not a valid tar.gz archive. Please check the release file on GitHub."
}

install_plugin() {
    local install_dir="${HELM_PLUGIN_DIR}"
    log "Extracting plugin..."
    tar -xzvf "helm-${PLUGIN_NAME}.tar.gz" -C "$install_dir" || error "Failed to extract plugin. The archive may be corrupted."
    rm "helm-${PLUGIN_NAME}.tar.gz"

    log "Setting up plugin binary..."
    mkdir -p "$install_dir/bin"
    [ -f "$install_dir/helm-${PLUGIN_NAME}" ] || error "Plugin binary not found after extraction."
    mv "$install_dir/helm-${PLUGIN_NAME}" "$install_dir/bin/helm-${PLUGIN_NAME}"

    local bin_path="$install_dir/bin/helm-${PLUGIN_NAME}"
    chmod +x "$bin_path" || error "Failed to set execute permissions on the plugin binary."
    log "Plugin installed successfully to: $bin_path"
}

verify_installation() {
    if helm plugin list | grep -q "${PLUGIN_NAME}"; then
        log "Verified: ${PLUGIN_NAME} is now installed and ready to use!"
        log "Try running: helm ${PLUGIN_NAME} --help"
    else
        error "Verification failed. The plugin doesn't seem to be properly installed. Current plugin list:\n$(helm plugin list)"
    fi
}

main() {
    [ -z "$PLUGIN_NAME" ] || [ -z "$PLUGIN_VERSION" ] && error "Failed to extract plugin information from plugin.yaml"
    log "Installing ${PLUGIN_NAME} version ${PLUGIN_VERSION}"
    get_system_info
    verify_helm
    download_plugin
    install_plugin
    verify_installation
    log "Installation complete!"
}

main "$@"
