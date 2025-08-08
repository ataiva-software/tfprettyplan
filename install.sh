#!/usr/bin/env bash

set -e

REPO="ao/tfprettyplan"
ARCH=$(uname -m)
OS=$(uname -s)

# Translate OS and ARCH
case "$OS" in
  Darwin) PLATFORM="Darwin" ;;
  Linux) PLATFORM="Linux" ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "Fetching latest version..."
LATEST=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep tag_name | cut -d '"' -f 4)
echo "Latest version is $LATEST"

FILENAME="tfprettyplan_${LATEST#v}_${PLATFORM}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST/$FILENAME"

echo "Downloading $URL..."
curl -L "$URL" -o "$FILENAME"

echo "Extracting..."
tar -xzf "$FILENAME"
rm "$FILENAME"

echo "Installing..."
chmod +x tfprettyplan
sudo mv tfprettyplan /usr/local/bin/

echo "✅ Installed tfprettyplan $LATEST to /usr/local/bin"
