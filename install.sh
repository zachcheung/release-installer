#!/bin/sh

set -eu

# preflight
type curl jq > /dev/null

gh_owner=zachcheung
gh_repo=release-installer
install_dir=/usr/local/bin
bin=$gh_repo

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
  | jq -r ".tag_name as \$tag | .assets[] | select(.name | test(\"${goos}_${goarch}\")) | .name + \" \" + .browser_download_url + \" \" + \$tag" \
  | while IFS=' ' read -r name download_url tag_name; do
    if [ -x $install_dir/$bin ]; then
      local_version=$($install_dir/$bin -version)
      if [ "$local_version" = "$tag_name" ]; then
        echo "$tag_name is the latest version, no need to upgrade"
        exit 0
      else
        echo "upgrading from $local_version to $tag_name"
      fi
    fi
    echo "downloading $name"
    curl -fsSL -o "$name" "$download_url"
    echo "downloaded $name"
    echo "installing $bin"
    tar xf "$name"
    find . -type f -perm -u=x -exec mv -v {} $install_dir \;
  done
