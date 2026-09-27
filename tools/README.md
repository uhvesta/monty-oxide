# Repository tools

Allow the repository with `direnv allow` to install Bazelisk and run
`bazel run //tools:install_devtools` automatically. The setup supports macOS
and Linux on x86-64 and ARM64. Bazel prepares the pinned Go and Rust versions
and builds `gopls`. The installer copies `gopls` and `starpls` into ignored
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

The available commands include `bazel`, `go`, `gofmt`, `cargo`, `rustc`,
`rustfmt`, `gopls`, `rust-analyzer`, and `starpls` (the Starlark/Bazel LSP).
The Rust version comes from `Cargo.toml`. The Go and `gopls` versions come
from `go.mod`.

From the repository root:

```sh
.tools/bin/bazel build //...
.tools/bin/bazel test //...
.tools/bin/bazel run //:gazelle
.tools/bin/bazel run //:format -- path/to/files
.tools/bin/bazel run //:format.check -- path/to/files
```

Builds run the configured lint aspects for Rust, shell, Markdown, and Starlark
targets. The formatter handles Rust, Go, shell, Markdown, and Starlark files.
The `buildifier_format_test` checks the live workspace on every `bazel test`
run; it is a local test with caching disabled.

The Rust compiler version and edition come from the root `Cargo.toml`
workspace package. The local extension that passes these values to
`rules_rust` lives in `tools/rules/rust/toolchain.bzl`.

The host tool extension selects one checked binary for each of Vale, `rumdl`,
`shfmt`, `buildifier`, and `starpls` for the current OS and CPU. Bazel detects
the host platform for builds automatically.
