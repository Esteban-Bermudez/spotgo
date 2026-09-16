#!/bin/sh
set -eu

REPO="Esteban-Bermudez/spotgo"
BIN="spotgo"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

detect_os_arch() {
  OS=$(uname -s)
  ARCH=$(uname -m)

  case "$ARCH" in
    x86_64|amd64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
  esac

  case "$OS" in
    Darwin|Linux) ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
  esac
}

fetch_latest_version() {
  if command -v curl >/dev/null 2>&1; then
    DOWNLOADER="curl -sfL"
  elif command -v wget >/dev/null 2>&1; then
    DOWNLOADER="wget -qO-"
  else
    echo "Need curl or wget to install."
    exit 1
  fi

  VERSION=$($DOWNLOADER "https://api.github.com/repos/$REPO/releases/latest" |
    grep '"tag_name"' | sed 's/.*"tag_name": "\(.*\)",/\1/')

  if [ -z "$VERSION" ]; then
    echo "Could not fetch latest version."
    exit 1
  fi
}

download_and_install() {
  ARCHIVE="${BIN}_${OS}_${ARCH}.tar.gz"
  URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE"
  CHECKSUM_URL="https://github.com/$REPO/releases/download/$VERSION/checksums.txt"

  TMPDIR=$(mktemp -d)
  cd "$TMPDIR"

  echo "Downloading $BIN $VERSION for $OS/$ARCH..."

  $DOWNLOADER "$URL" > "$ARCHIVE"
  $DOWNLOADER "$CHECKSUM_URL" > checksums.txt 2>/dev/null || true

  if [ -f checksums.txt ]; then
    if command -v sha256sum >/dev/null 2>&1; then
      EXPECTED=$(grep "$ARCHIVE" checksums.txt | cut -d' ' -f1)
      COMPUTED=$(sha256sum "$ARCHIVE" | cut -d' ' -f1)
      [ "$EXPECTED" = "$COMPUTED" ] || echo "Warning: checksum mismatch"
    elif command -v shasum >/dev/null 2>&1; then
      EXPECTED=$(grep "$ARCHIVE" checksums.txt | cut -d' ' -f1)
      COMPUTED=$(shasum -a 256 "$ARCHIVE" | cut -d' ' -f1)
      [ "$EXPECTED" = "$COMPUTED" ] || echo "Warning: checksum mismatch"
    fi
  fi

  tar -xzf "$ARCHIVE"
  chmod +x "$BIN"
  if [ -f "$BIN" ]; then
    mkdir -p "$INSTALL_DIR"
    mv "$BIN" "$INSTALL_DIR/$BIN"
    echo "Installed $BIN to $INSTALL_DIR/$BIN"
  fi

  rm -rf "$TMPDIR"
}

detect_os_arch
fetch_latest_version
download_and_install

case ":$PATH:" in
  *:"$INSTALL_DIR":*) ;;
  *) echo "Add $INSTALL_DIR to your PATH: export PATH=\"\$HOME/.local/bin:\$PATH\"" ;;
esac

echo "Run '$BIN' to start."
