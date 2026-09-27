# monty-oxide

## Bootstrap

### With direnv (suggested)

Install [direnv](https://direnv.net/docs/installation.html). Then for zsh, add this
line to `~/.zshrc`:

```sh
eval "$(direnv hook zsh)"
```

For bash, add this line to `~/.bashrc`:

```sh
eval "$(direnv hook bash)"
```

Restart your shell, enter this repository, and run `direnv allow .` once. On entry,
direnv runs the Bazelisk bootstrap and adds `.tools/bin` to PATH. The `bazel`
command then runs Bazelisk. Direnv removes that PATH entry when you leave the
repository. It will ask you to allow the `.envrc` again if that file changes.

### Manual bootstrap

From the repository root, run:

```sh
./tools/scripts/bootstrap/bazelisk.sh
```

The script supports macOS and Linux on x86-64 or ARM64. It downloads a pinned
version of Bazelisk to `.tools/bin/bazelisk` and creates `.tools/bin/bazel` as a
link to it. Run `.tools/bin/bazel` directly if you do not use direnv.
