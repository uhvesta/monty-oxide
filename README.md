# monty-oxide

- **Bootstrap:** `sh tools/scripts/bootstrap/bazelisk.sh` installs pinned Bazelisk as `bazel` on macOS or Linux (x86-64 or ARM64). It reports newer releases without changing the pin.
- **Automatic tools:** Install [direnv](https://direnv.net/docs/installation.html), add `eval "$(direnv hook zsh)"` or `eval "$(direnv hook bash)"` to your shell startup file, then run `direnv allow`. This bootstraps Bazel and adds installed Go and Rust tools to `PATH` while you are here.
- **Manual tools:** Run `sh tools/scripts/bootstrap/devtools.sh` to install Go, Rust, and language servers. Add `.tools/bin` to `PATH` to use `bazel` without direnv. Tools live under ignored `.tools`.
- **VS Code:** Run the setup script once and reload the window. `.vscode/settings.json` points the Go, Rust, and Bazel extensions at repository tools.
- **Pinned versions:** Go comes from `go.mod`, Rust from `Cargo.toml`, Bazel from `.bazelversion`, and Bazelisk from its bootstrap script.
- **Build and test:** `bazel build //...` and `bazel test //...`.
- **Generate Go BUILD files:** `bazel run //:gazelle`.
- **Version checks:** `bazel run check-deps` reports all drift;
  `bazel run update-deps` applies it. Scope either with `-- --bazel`,
  `--bzlmod`, `--golang`, `--gomod`, `--rust`, or `--crates`; `--all`
  is explicit.
- **Format:** `bazel run //:format -- path/to/files`; check with `bazel run //:format.check -- path/to/files`.
