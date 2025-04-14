#!/bin/bash

set -euo pipefail

# Constants
INSTALL_DIR="$HOME/.local/share/ult"
TMP_DIR=$(mktemp -d)

# Cleanup function
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

setup_path() {
  # set envs for FISH SHELL
  if [ -f ~/.config/fish/config.fish ]; then
    echo "set -gx ULT_PATH $INSTALL_DIR" >> ~/.config/fish/config.fish
    echo "set -gx PATH \$PATH \$ULT_PATH" >> ~/.config/fish/config.fish
  else
    echo "Fish config not found. Skipping."
  fi
  # set envs for ZSH SHELL
  if [ -f ~/.zshrc ]; then
    echo "export ULT_PATH=$INSTALL_DIR" >> ~/.zshrc
    echo "export PATH=\$PATH:\$ULT_PATH" >> ~/.zshrc
  else
    echo "Zsh config not found. Skipping."
  fi
}

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
echo "Downloading ult CLI: $DOWNLOAD_URL"
curl -L --progress-bar "$DOWNLOAD_URL" -o "$TMP_DIR/ult.tar.gz"

if command -v file &> /dev/null; then
  echo "Verifying downloaded file..."
  file "$TMP_DIR/ult.tar.gz" || echo "Warning: File check skipped (unknown type)"
else
  echo "Skipping file verification (command 'file' not installed)"
fi
echo "Extracting package..."
tar -xzf "$TMP_DIR/ult.tar.gz" -C "$TMP_DIR"

echo "Installing to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
mv "$TMP_DIR/${OS}-${ARCH}/ult" "$INSTALL_DIR/ult"
chmod +x "$INSTALL_DIR/ult"

setup_path

printf "\n\n"
printf "IMPORTANT --------------------------------------------------------------------------------"
printf "\n"
echo "if not using 'fish' or 'zsh' shell, you need to manually add this environmental variables:"
printf "\n"
echo "ULT_PATH=$INSTALL_DIR"
echo "PATH=\$PATH:\$ULT_PATH"
echo "------------------------------------------------------------------------------------------"
printf "\n\n"

# Verify installation
if command -v "$INSTALL_DIR"/ult &> /dev/null; then
  echo "Successfully installed ult CLI!"
  VERSION=$("$INSTALL_DIR"/ult version)
  echo "Version: $VERSION"
else
  echo "Installation failed: ult not found"
  exit 1
fi
