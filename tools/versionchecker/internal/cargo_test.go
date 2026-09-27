package versionchecker

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRustFixUsesParsedCargoManifest(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Cargo.toml")
	before := `[workspace]
members = []
# rust-version = "1.0.0"
[workspace.package]
edition = "2024"
rust-version = '1.98.1' # keep this comment
`
	if err := os.WriteFile(path, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}
	rust := Rust{Root: root}
	if current, err := rust.pin(); err != nil || current != "1.98.1" {
		t.Fatalf("pin = %q, %v", current, err)
	}
	if err := rust.Fix(context.Background(), Drift{Name: "rust-version", Current: "1.98.1", Latest: "1.99.0"}); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(before, "rust-version = '1.98.1'", `rust-version = "1.99.0"`, 1)
	if string(after) != want {
		t.Fatalf("Cargo.toml changed unexpectedly:\n%s", after)
	}
}
