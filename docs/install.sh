#!/usr/bin/env bash
set -euo pipefail

### ===== CONFIG =====
APP_NAME="flux"
VERSION="v0.0.1"

BASE_URL="https://github.com/Dishank-Sen/Flux-CLI/releases/download/"

INSTALL_DIR="/usr/local/bin"
TMP_DIR="$(mktemp -d)"
### ==================

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT


echo "Installing $APP_NAME..."

### Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"

if [[ "$OS" != "linux" ]]; then
  echo "Only Linux is supported for now."
  exit 1
fi


### Detect architecture
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac


FILE="$APP_NAME-$OS-$ARCH"
DOWNLOAD_URL="$BASE_URL/$VERSION/$FILE"

echo "Detected: $OS/$ARCH"
echo "Downloading from: $DOWNLOAD_URL"


### Download binary
curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$APP_NAME"


### Optional: checksum verification
if curl -fsSL "$DOWNLOAD_URL.sha256" -o "$TMP_DIR/$APP_NAME.sha256" 2>/dev/null; then
  echo "Verifying checksum..."
  (cd "$TMP_DIR" && sha256sum -c "$APP_NAME.sha256")
fi


chmod +x "$TMP_DIR/$APP_NAME"


### Install
if [[ -w "$INSTALL_DIR" ]]; then
  mv "$TMP_DIR/$APP_NAME" "$INSTALL_DIR/$APP_NAME"
else
  sudo mv "$TMP_DIR/$APP_NAME" "$INSTALL_DIR/$APP_NAME"
fi


echo ""
echo "Installed successfully!"
echo "Run:"
echo "  $APP_NAME --help"
