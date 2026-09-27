# Repository tools

Allow the repository with `direnv allow` to install Bazelisk and run
`bazel run //tools:install_devtools` automatically. The setup supports macOS
and Linux on x86-64 and ARM64. Bazel prepares the pinned Go and Rust versions
and builds `gopls`. Bazel runs the installer with its pinned Python toolchain;
the installer copies `gopls` and `starpls` into ignored
`.tools/bin`. It links `rust-analyzer` and the complete Go and Rust
distributions under `.tools/sdk` so their libraries and source files remain
available.

For manual setup from the repository root without `direnv`, run:

```sh
sh tools/scripts/bootstrap/bazelisk.sh
.tools/bin/bazel run //tools:install_devtools
PATH="$PWD/.tools/bin:$PWD/.tools/sdk/go/bin:$PWD/.tools/sdk/rust/bin:$PATH"
export PATH
export GOROOT="$PWD/.tools/sdk/go"
export RUST_SRC_PATH="$PWD/.tools/sdk/rust-analyzer/lib/rustlib/src/library"
```

You can still run only `sh tools/scripts/bootstrap/bazelisk.sh` if you want
just Bazelisk. `sh tools/scripts/bootstrap/devtools.sh` performs both setup
steps and reuses the existing installation when its inputs have not changed.
The shell scripts start Bazel; Bazel runs the Python installer.

The available commands include `bazel`, `go`, `gofmt`, `cargo`, `rustc`,
`rustfmt`, `gopls`, `rust-analyzer`, and `starpls` (the Starlark/Bazel LSP).
The Rust version comes from `Cargo.toml`. The Go and `gopls` versions come
from `go.mod`.

VS Code reads `.vscode/settings.json` to launch these tools from the repository
even when its extension host did not inherit direnv's `PATH`. The Bazel
extension launches `starpls` through `tools/scripts/ide/starpls.sh`,
which adds the repository's Bazelisk to the language server's `PATH`. Run
`sh tools/scripts/bootstrap/devtools.sh` once after cloning, then reload the
VS Code window. The workspace also recommends the Go, Rust, and Bazel
extensions.

The Go and Cargo caches live under ignored `.tools`. Bazel keeps its cache
outside the checkout so its repository cache cannot conflict with the
workspace.

From the repository root:

```sh
.tools/bin/bazel build //...
.tools/bin/bazel test //...
.tools/bin/bazel run //:gazelle
.tools/bin/bazel run //:format -- path/to/files
.tools/bin/bazel run //:format.check -- path/to/files
```

The Bazelisk bootstrap script reports when a different latest release is
available. Its version and four platform checksums are pinned together near
the top of `tools/scripts/bootstrap/bazelisk.sh`.

Builds run the configured lint aspects for Rust, shell, Markdown, and Starlark
targets. The formatter handles Rust, Go, shell, Markdown, and Starlark files.
The `format_test` checks the live workspace for Go, Rust, shell, Markdown, and
Starlark formatting on every `bazel test` run; it is a local test with caching
disabled.

The Rust compiler version and edition come from the root `Cargo.toml`
workspace package. The local extension that passes these values to
`rules_rust` lives in `tools/rules/rust/toolchain.bzl`.

The host tool extension selects one checked binary for each of Vale, `rumdl`,
`shfmt`, `buildifier`, and `starpls` for the current OS and CPU. Bazel detects
the host platform for builds automatically.
