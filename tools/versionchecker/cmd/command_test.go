package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/uhvesta/monty-oxide/tools/versionchecker/internal"
)

type testSource struct{ fixed bool }

func (*testSource) Name() string { return "bcr" }

func (*testSource) Check(context.Context) ([]versionchecker.Drift, error) {
	return []versionchecker.Drift{{Name: "rules_go", Current: "1.0", Latest: "1.1"}}, nil
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
