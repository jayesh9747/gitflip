#!/usr/bin/env sh
# Install gitflip from GitHub Releases (no Go required).
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/jayesh9747/gitflip/main/scripts/install.sh | sh
# Optional:
#   VERSION=v0.1.0 curl -fsSL ... | sh
#   INSTALL_DIR=$HOME/.local/bin curl -fsSL ... | sh

set -eu

REPO_OWNER="jayesh9747"
REPO_NAME="gitflip"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

die() {
  printf '%s\n' "$*" >&2
  exit 1
}

command -v curl >/dev/null 2>&1 || die "curl is required"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
x86_64 | amd64) arch=amd64 ;;
aarch64 | arm64) arch=arm64 ;;
*) die "unsupported architecture: $arch (need amd64 or arm64)" ;;
esac

case "$os" in
linux) goos=linux ;;
darwin) goos=darwin ;;
*) die "unsupported OS: $os (need Linux or Darwin)" ;;
esac

if [ "${VERSION:-}" = "" ]; then
  json=$(curl -fsSL "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest") || die "failed to fetch latest release (is the repo public and has a release?)"
  VERSION=$(printf '%s' "$json" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1)
  [ -n "$VERSION" ] || die "could not parse latest release tag"
fi

ver_num=${VERSION#v}
asset_base="gitflip_${ver_num}_${goos}_${arch}"
if [ "$goos" = "windows" ]; then
  die "use the .zip from Releases on Windows, or install via WSL"
fi
asset="${asset_base}.tar.gz"
url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/${asset}"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

curl -fsSL "$url" -o "$tmp/$asset" || die "download failed: $url (wrong VERSION or asset missing for ${goos}/${arch}?)"

tar -xzf "$tmp/$asset" -C "$tmp"
bin="$tmp/gitflip"
[ -f "$bin" ] || die "expected gitflip binary inside archive"

if [ "$INSTALL_DIR" = "/usr/local/bin" ] && [ ! -w "$INSTALL_DIR" ]; then
  die "cannot write to $INSTALL_DIR. Run: sudo sh, or INSTALL_DIR=\$HOME/.local/bin sh (and put that dir on PATH)"
fi

mkdir -p "$INSTALL_DIR"
mv "$bin" "$INSTALL_DIR/gitflip"
chmod +x "$INSTALL_DIR/gitflip"

printf 'Installed gitflip %s -> %s/gitflip\n' "$VERSION" "$INSTALL_DIR"
command -v gitflip >/dev/null 2>&1 || printf 'Note: add %s to PATH if needed.\n' "$INSTALL_DIR"
