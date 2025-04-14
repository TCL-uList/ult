#!/bin/bash

set -euo pipefail

# Constants
INSTALL_DIR="/usr/local/bin"
TMP_DIR=$(mktemp -d)

# Cleanup function
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
echo "Detected operating system: $OS"

ARCH=$(uname -m)
case "$ARCH" in
  "x86_64") ARCH="amd64" ;;
  "aarch64"|"arm64") ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac
echo "Detected architecture: $ARCH"

# Download and install
DOWNLOAD_URL="https://github.com/TCL-uList/ult/releases/latest/download/${OS}-${ARCH}.tar.gz"
echo "⬇️ Downloading ult CLI: $DOWNLOAD_URL"
curl -L --progress-bar "$DOWNLOAD_URL" -o "$TMP_DIR/ult.tar.gz"
file "$TMP_DIR/ult.tar.gz"

echo "Extracting package..."
tar -xzf "$TMP_DIR/ult.tar.gz" -C "$TMP_DIR"

echo "Installing to $INSTALL_DIR..."
sudo mv "$TMP_DIR/${OS}-${ARCH}/ult" "$INSTALL_DIR/ult"
sudo chmod +x "$INSTALL_DIR/ult"

# Verify installation
if command -v ult &> /dev/null; then
  echo "Successfully installed ult CLI!"
  echo "Version: $(ult version)"
else
  echo "Installation failed: ult not found in PATH"
  exit 1
fi
