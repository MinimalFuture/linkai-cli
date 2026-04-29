#!/bin/sh
# LinkAI CLI installer.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/MinimalFuture/linkai-cli/main/install.sh | sh
#
# Environment overrides:
#   INSTALL_VERSION=v1.2.3   pin a specific release tag
#   INSTALL_PREFIX=/usr/local install into /usr/local/bin (default: $HOME/.local/bin)

set -eu

REPO="MinimalFuture/linkai-cli"
BINARY="linkai"
PREFIX="${INSTALL_PREFIX:-$HOME/.local}"
VERSION="${INSTALL_VERSION:-}"

err() { printf 'error: %s\n' "$*" >&2; exit 1; }
info() { printf '%s\n' "$*" >&2; }

need() {
  command -v "$1" >/dev/null 2>&1 || err "missing required command: $1"
}

need uname
need tar
need mkdir
if command -v curl >/dev/null 2>&1; then
  fetch() { curl -fsSL "$1"; }
  fetch_to() { curl -fsSL -o "$2" "$1"; }
elif command -v wget >/dev/null 2>&1; then
  fetch() { wget -qO- "$1"; }
  fetch_to() { wget -qO "$2" "$1"; }
else
  err "either curl or wget is required"
fi

uname_s=$(uname -s)
case "$uname_s" in
  Linux)  os=linux ;;
  Darwin) os=darwin ;;
  *) err "unsupported OS: $uname_s (use the Windows zip from the GitHub release page)" ;;
esac

uname_m=$(uname -m)
case "$uname_m" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) err "unsupported architecture: $uname_m" ;;
esac

if [ -z "$VERSION" ]; then
  info "==> Resolving latest release..."
  VERSION=$(fetch "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep -o '"tag_name": *"[^"]*"' \
    | head -n1 \
    | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
  [ -n "$VERSION" ] || err "could not resolve latest version (rate limited? set INSTALL_VERSION=vX.Y.Z)"
fi

# Strip leading 'v' for archive name (matches goreleaser's {{ .Version }} default).
ver_no_v=${VERSION#v}
archive="${BINARY}_${ver_no_v}_${os}_${arch}.tar.gz"
url="https://github.com/${REPO}/releases/download/${VERSION}/${archive}"
checksum_url="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

tmp=$(mktemp -d 2>/dev/null || mktemp -d -t linkai)
trap 'rm -rf "$tmp"' EXIT

info "==> Downloading ${archive}"
fetch_to "$url" "$tmp/$archive"

info "==> Verifying checksum"
fetch_to "$checksum_url" "$tmp/checksums.txt"

if command -v shasum >/dev/null 2>&1; then
  expected=$(grep " $archive\$" "$tmp/checksums.txt" | awk '{print $1}')
  actual=$(shasum -a 256 "$tmp/$archive" | awk '{print $1}')
elif command -v sha256sum >/dev/null 2>&1; then
  expected=$(grep " $archive\$" "$tmp/checksums.txt" | awk '{print $1}')
  actual=$(sha256sum "$tmp/$archive" | awk '{print $1}')
else
  info "warning: no shasum/sha256sum found, skipping checksum verification"
  expected=""
  actual=""
fi

if [ -n "$expected" ] && [ -n "$actual" ] && [ "$expected" != "$actual" ]; then
  err "checksum mismatch (expected $expected, got $actual)"
fi

info "==> Extracting"
tar -xzf "$tmp/$archive" -C "$tmp"

bindir="$PREFIX/bin"
mkdir -p "$bindir"
install_path="$bindir/$BINARY"

if [ -w "$bindir" ]; then
  mv "$tmp/$BINARY" "$install_path"
else
  info "==> $bindir not writable, using sudo"
  sudo mv "$tmp/$BINARY" "$install_path"
fi
chmod +x "$install_path" 2>/dev/null || sudo chmod +x "$install_path"

info "==> Installed $install_path ($VERSION)"

case ":$PATH:" in
  *":$bindir:"*) ;;
  *)
    info ""
    info "Note: $bindir is not on your PATH. Add this to your shell profile:"
    info "    export PATH=\"$bindir:\$PATH\""
    ;;
esac

"$install_path" --version || true
