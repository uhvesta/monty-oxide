package versionchecker

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBzlmodFixPreservesUnrelatedText(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "MODULE.bazel")
	before := `module(name = "example")
# bazel_dep(name = "commented", version = "0.1.0")
bazel_dep(
    version = '1.2.3',  # current pin
    dev_dependency = True,
    name = "rules_go",
)
bazel_dep(name = "gazelle", version = "0.54.0")
bazel_dep(name = "local_only")
`
	if err := os.WriteFile(path, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}
	pins, err := modulePins(path, []byte(before))
	if err != nil {
		t.Fatal(err)
	}
	if len(pins) != 2 || pins[0].name != "rules_go" || pins[0].version.Value != "1.2.3" {
		t.Fatalf("unexpected pins: %+v", pins)
	}
	if err := (Bzlmod{Root: root}).Fix(context.Background(), Drift{Name: "rules_go", Current: "1.2.3", Latest: "1.3.0"}); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(before, "version = '1.2.3'", `version = "1.3.0"`, 1)
	if string(after) != want {
		t.Fatalf("MODULE.bazel changed unexpectedly:\n%s", after)
	}
}

func TestBzlmodFixRejectsChangedPin(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "MODULE.bazel")
	content := `bazel_dep(name = "rules_go", version = "1.3.0")`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := (Bzlmod{Root: root}).Fix(context.Background(), Drift{Name: "rules_go", Current: "1.2.3", Latest: "1.4.0"}); err == nil {
		t.Fatal("expected stale pin error")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != content {
		t.Fatalf("stale pin was changed: %s", after)
	}
}
