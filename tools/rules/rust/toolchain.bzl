"""Register Rust toolchains from the Cargo workspace metadata."""

load("@rules_rust//rust/private:repositories.bzl", "rust_register_toolchains")
load("@toml.bzl", "toml")
load("//tools/rules/platform:host.bzl", "host_binary_key")

_TOOLCHAIN_TRIPLES = {
    "darwin_amd64": "x86_64-apple-darwin",
    "darwin_arm64": "aarch64-apple-darwin",
    "linux_amd64": "x86_64-unknown-linux-gnu",
    "linux_arm64": "aarch64-unknown-linux-gnu",
}

def _rust_metadata_impl(repository_ctx):
    repository_ctx.file("BUILD.bazel", 'exports_files(["edition.bzl"], visibility = ["@//tools/format:__pkg__"])\n')
    repository_ctx.file("edition.bzl", 'RUST_EDITION = "{}"\n'.format(repository_ctx.attr.edition))

_rust_metadata = repository_rule(
    implementation = _rust_metadata_impl,
    attrs = {"edition": attr.string(mandatory = True)},
)

def _rust_from_cargo_impl(module_ctx):
    cargo = toml.decode(module_ctx.read(Label("//:Cargo.toml")))
    workspace_package = cargo["workspace"]["package"]
    triple = _TOOLCHAIN_TRIPLES[host_binary_key(module_ctx.os)]
    toolchain_triples = {triple: "rust_host"}

    rust_register_toolchains(
        hub_name = "rust_toolchains",
        edition = workspace_package["edition"],
        versions = [workspace_package["rust-version"]],
        toolchain_triples = toolchain_triples,
        rustfmt_toolchain_triples = toolchain_triples,
    )
    _rust_metadata(name = "rust_metadata", edition = workspace_package["edition"])
    return module_ctx.extension_metadata(reproducible = True)

rust_from_cargo = module_extension(
    implementation = _rust_from_cargo_impl,
    os_dependent = True,
    arch_dependent = True,
)
