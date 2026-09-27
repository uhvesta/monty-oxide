#!/bin/sh
set -eu

repo_root=$(CDPATH='' cd -P "$(dirname "$0")/../../.." && pwd)

# Older direnv sessions may still put Bazel's XDG cache inside this workspace.
case "${XDG_CACHE_HOME:-}" in
  "$repo_root" | "$repo_root"/*) unset XDG_CACHE_HOME ;;
esac

"$repo_root/tools/scripts/bootstrap/bazelisk.sh" >/dev/null

bazel="$repo_root/.tools/bin/bazel"
tools_dir="$repo_root/.tools"
bin_dir="$tools_dir/bin"
sdk_dir="$tools_dir/sdk"
stamp_file="$tools_dir/devtools.stamp"

cd "$repo_root"
output_base=$("$bazel" info output_base)
stamp=$(cksum \
  Cargo.toml go.mod go.sum MODULE.bazel .bazelversion \
  tools/BUILD.bazel tools/rules/rust/toolchain.bzl \
  tools/rules/tooling/extensions.bzl \
  tools/scripts/bootstrap/devtools.sh \
  tools/scripts/ide/install_devtools.py)
stamp="$stamp
$output_base"

if [ -f "$stamp_file" ] && [ "$(cat "$stamp_file")" = "$stamp" ] &&
  [ -x "$sdk_dir/go/bin/go" ] && [ -x "$sdk_dir/rust/bin/cargo" ] &&
  [ -x "$bin_dir/gopls" ] && [ -x "$bin_dir/rust-analyzer" ] &&
  [ -x "$bin_dir/starpls" ]; then
  exit 0
fi

"$bazel" run //tools:install_devtools
printf '%s\n' "$stamp" >"$stamp_file"
