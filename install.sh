#!/bin/sh

set -eu

# preflight
type curl jq > /dev/null

gh_owner=zachcheung
gh_repo=release-installer
install_dir=/usr/local/bin

kernel=$(uname -s)
# Convert kernel name to lowercase
case "$kernel" in
  Linux) goos="linux" ;;
  Darwin) goos="darwin" ;;
  *) echo "unsupported kernel: $kernel"; exit 1 ;;
esac

platform=$(uname -m)
# Map the platform to Go architecture
case "$platform" in
  x86_64) goarch="amd64" ;;
  i386 | i686) goarch="386" ;;
  armv7l) goarch="armv7" ;;
  aarch64) goarch="arm64" ;;
  *) echo "unsupported platform: $platform"; exit 1 ;;
esac

temp_dir=$(mktemp -d)
trap 'rm -rf "$temp_dir"' EXIT

cd "$temp_dir"
curl -fsSL https://api.github.com/repos/$gh_owner/$gh_repo/releases/latest \
  | jq -r ".assets[] | select(.browser_download_url | test(\"${goos}_${goarch}\")) | .name + \" \" + .browser_download_url" \
  | while IFS=' ' read -r name download_url; do
    echo "downloading $name"
    curl -fsSL -o "$name" "$download_url"
    echo "downloaded $name"
    echo "installing $gh_repo"
    tar xf "$name"
    find . -type f -perm -u=x -exec mv -v {} $install_dir \;
  done
