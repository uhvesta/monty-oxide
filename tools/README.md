# Repository tools

Run `tools/scripts/bootstrap/bazelisk.sh` once to install Bazelisk under
`.tools/bin`. The bootstrap supports macOS and Linux on x86-64 and ARM64.
The repository's `direnv` setup adds that directory to `PATH` automatically.

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
`shfmt`, and `buildifier` for the current OS and CPU. Bazel detects the host
platform for builds automatically.
