#!/bin/sh
set -eu

# Keep the version and hashes together. Update them from the release assets at
# https://github.com/bazelbuild/bazelisk/releases/tag/v1.29.0.
BAZELISK_VERSION=v1.29.0
BAZELISK_SHA256_DARWIN_AMD64=16c3d7aa15323a9fb69f56c7ec5733ed18bedb786680d0ba13bb12a3c8083007
BAZELISK_SHA256_DARWIN_ARM64=cee851f726789227d5561004e9904a52be45c3efb56f8b38b6993d6adbaa0409
BAZELISK_SHA256_LINUX_AMD64=5a408715e932c0250d28bd84555f12edbf70117de42f9181691c736eacc4a992
BAZELISK_SHA256_LINUX_ARM64=e20e8b0f4f240091b7a55bf17b9398bd4f40ee70ae0208dff95dd4c445fb4010

case "$(uname -s)" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *) printf 'Unsupported operating system: %s\n' "$(uname -s)" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) printf 'Unsupported architecture: %s\n' "$(uname -m)" >&2; exit 1 ;;
esac

case "$os-$arch" in
  darwin-amd64) sha256=$BAZELISK_SHA256_DARWIN_AMD64 ;;
  darwin-arm64) sha256=$BAZELISK_SHA256_DARWIN_ARM64 ;;
  linux-amd64) sha256=$BAZELISK_SHA256_LINUX_AMD64 ;;
  linux-arm64) sha256=$BAZELISK_SHA256_LINUX_ARM64 ;;
esac

sha256_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    output=$(sha256sum "$1")
    printf '%s\n' "${output%% *}"
  elif command -v shasum >/dev/null 2>&1; then
    output=$(shasum -a 256 "$1")
    printf '%s\n' "${output%% *}"
  elif command -v openssl >/dev/null 2>&1; then
    output=$(openssl dgst -sha256 "$1")
    printf '%s\n' "${output##* }"
  else
    printf 'A SHA-256 tool (sha256sum, shasum, or openssl) is required.\n' >&2
    return 1
  fi
}

repo_root=$(CDPATH= cd -P "$(dirname "$0")/../../.." && pwd)
bin_dir="$repo_root/.tools/bin"
destination="$bin_dir/bazelisk"

if [ -f "$destination" ] && [ "$(sha256_file "$destination")" = "$sha256" ]; then
  chmod 0755 "$destination"
  printf 'Bazelisk %s is already installed at %s\n' "$BAZELISK_VERSION" "$destination"
  exit 0
fi

if ! command -v curl >/dev/null 2>&1; then
  printf 'curl is required to download Bazelisk.\n' >&2
  exit 1
fi

mkdir -p "$bin_dir"
temporary_file=$(mktemp "$bin_dir/.bazelisk.XXXXXXXX")
trap 'rm -f "$temporary_file"' 0

asset="bazelisk-$os-$arch"
url="https://github.com/bazelbuild/bazelisk/releases/download/$BAZELISK_VERSION/$asset"
curl --fail --location --silent --show-error --retry 3 --output "$temporary_file" "$url"

actual_sha256=$(sha256_file "$temporary_file")
if [ "$actual_sha256" != "$sha256" ]; then
  printf 'SHA-256 mismatch for %s\nExpected: %s\nActual:   %s\n' "$asset" "$sha256" "$actual_sha256" >&2
  exit 1
fi

chmod 0755 "$temporary_file"
mv -f "$temporary_file" "$destination"
trap - 0
printf 'Installed Bazelisk %s at %s\n' "$BAZELISK_VERSION" "$destination"
