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

Restart your shell (or source the rc files), enter this repository, and follow
the prompts to allow the hook. `bazel` is then on your path while you work in
the repository.

### Manual bootstrap

From the repository root, run:

```sh
./tools/scripts/bootstrap/bazelisk.sh
```

1. The script supports macOS and Linux on x86-64 or ARM64.
2. It downloads a pinned version of Bazelisk to `.tools/bin/bazelisk`.
3. It creates `.tools/bin/bazel` as a link to it.
4. Run `.tools/bin/bazel` directly if you do not use `direnv`.
