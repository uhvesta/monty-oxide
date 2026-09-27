#!/bin/sh
set -eu

if [ "$#" -lt 5 ] || [ -z "${BUILD_WORKSPACE_DIRECTORY:-}" ]; then
  printf 'Run this installer with bazel run //tools:install_devtools.\n' >&2
  exit 1
fi

runfiles_dir=${RUNFILES_DIR:-$0.runfiles}
if [ ! -d "$runfiles_dir" ]; then
  printf 'Bazel runfiles directory is missing: %s\n' "$runfiles_dir" >&2
  exit 1
fi

resolve_runfile() {
  path="$runfiles_dir/$1"
  if [ ! -f "$path" ]; then
    printf 'Bazel runfile is missing: %s\n' "$path" >&2
    return 1
  fi
  while [ -L "$path" ]; do
    link=$(readlink "$path")
    case "$link" in
      /*) path=$link ;;
      *) path=$(dirname "$path")/$link ;;
    esac
  done
  directory=$(CDPATH='' cd -P "$(dirname "$path")" && pwd)
  printf '%s/%s\n' "$directory" "$(basename "$path")"
}

go_file=$(resolve_runfile "$1")
cargo_file=$(resolve_runfile "$2")
shift 2

analyzer_path=
gopls_path=
starpls_path=
for path in "$@"; do
  case "$path" in
    */bin/rust-analyzer) analyzer_path=$path ;;
  esac
  gopls_path=$starpls_path
  starpls_path=$path
done
if [ -z "$analyzer_path" ]; then
  printf 'Could not find rust-analyzer in Bazel runfiles.\n' >&2
  exit 1
fi
analyzer_file=$(resolve_runfile "$analyzer_path")
gopls_file=$(resolve_runfile "$gopls_path")
starpls_file=$(resolve_runfile "$starpls_path")

go_root=$(dirname "$(dirname "$go_file")")
rust_root=$(dirname "$(dirname "$cargo_file")")
analyzer_root=$(dirname "$(dirname "$analyzer_file")")

for executable in "$go_root/bin/go" "$rust_root/bin/cargo" \
  "$rust_root/bin/rustc" "$analyzer_root/bin/rust-analyzer" \
  "$gopls_file" "$starpls_file"; do
  if [ ! -x "$executable" ]; then
    printf 'Expected executable is missing: %s\n' "$executable" >&2
    exit 1
  fi
done

tools_dir="$BUILD_WORKSPACE_DIRECTORY/.tools"
bin_dir="$tools_dir/bin"
sdk_dir="$tools_dir/sdk"
mkdir -p "$bin_dir" "$sdk_dir"

ln -sfn "$go_root" "$sdk_dir/go"
ln -sfn "$rust_root" "$sdk_dir/rust"
ln -sfn "$analyzer_root" "$sdk_dir/rust-analyzer"
ln -sfn "$analyzer_root/bin/rust-analyzer" "$bin_dir/rust-analyzer"

copy_tool() (
  source=$1
  destination=$2
  if [ -f "$destination" ] && [ ! -L "$destination" ] && cmp -s "$source" "$destination"; then
    exit 0
  fi
  temporary_file=$(mktemp "$bin_dir/.devtool.XXXXXXXX")
  trap 'rm -f "$temporary_file"' 0
  cp "$source" "$temporary_file"
  chmod 0755 "$temporary_file"
  mv -f "$temporary_file" "$destination"
  trap - 0
)

copy_tool "$gopls_file" "$bin_dir/gopls"
copy_tool "$starpls_file" "$bin_dir/starpls"
printf 'Installed Go, Rust, gopls, rust-analyzer, and starpls under %s\n' "$tools_dir"
