package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/uhvesta/monty-oxide/tools/versionchecker/internal"
)

type testSource struct {
	name    string
	checked bool
	fixed   bool
}

func (s *testSource) Name() string {
	if s.name != "" {
		return s.name
	}
	return "bcr"
}

func (s *testSource) Check(context.Context) ([]versionchecker.Drift, error) {
	s.checked = true
	return []versionchecker.Drift{{Name: "rules_go", Current: "1.0", Latest: "1.1"}}, nil
}

func TestSourceFlags(t *testing.T) {
	first := &testSource{name: "bazel"}
	second := &testSource{name: "crates"}
	cmd := newCommand(versionchecker.Runner{Sources: []versionchecker.Source{first, second}})
	cmd.SetArgs([]string{"check", "--crates"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if first.checked || !second.checked {
		t.Fatalf("source selection: bazel=%t crates=%t", first.checked, second.checked)
	}

	first.checked, second.checked = false, false
	cmd = newCommand(versionchecker.Runner{Sources: []versionchecker.Source{first, second}})
	cmd.SetArgs([]string{"check", "--all"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !first.checked || !second.checked {
		t.Fatalf("all selection: bazel=%t crates=%t", first.checked, second.checked)
	}
}

func (s *testSource) Fix(context.Context, versionchecker.Drift) error {
	s.fixed = true
	return nil
}

func TestCommands(t *testing.T) {
	for _, action := range []string{"check", "fix"} {
		t.Run(action, func(t *testing.T) {
			source := &testSource{}
			cmd := newCommand(versionchecker.Runner{Sources: []versionchecker.Source{source}})
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetArgs([]string{action})
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != "bcr/rules_go: 1.0 -> 1.1\n" {
				t.Fatalf("unexpected output: %q", got)
			}
			if source.fixed != (action == "fix") {
				t.Fatalf("fix called for %s: %t", action, source.fixed)
			}
		})
	}
}

func TestHelpListsActions(t *testing.T) {
	cmd := newCommand(versionchecker.Runner{})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"check", "fix"} {
		if !strings.Contains(output.String(), action) {
			t.Fatalf("help does not list %s: %q", action, output.String())
		}
	}
}

func TestResolveRunfile(t *testing.T) {
	runfiles := t.TempDir()
	goPath := filepath.Join(runfiles, "go_sdk", "bin", "go")
	if err := os.MkdirAll(filepath.Dir(goPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(goPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RUNFILES_DIR", runfiles)
	got, err := resolveRunfile("go_sdk/bin/go")
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(goPath)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("resolved %q, want %q", got, want)
	}
}
