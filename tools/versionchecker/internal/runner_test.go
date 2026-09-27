package versionchecker

import (
	"context"
	"testing"
)

type fakeSource struct {
	name   string
	drifts []Drift
	fixed  []Drift
}

func (s *fakeSource) Name() string { return s.name }

func (s *fakeSource) Check(context.Context) ([]Drift, error) { return s.drifts, nil }

func (s *fakeSource) Fix(_ context.Context, drift Drift) error {
	s.fixed = append(s.fixed, drift)
	return nil
}

func TestRunnerChecksAndRoutesFixes(t *testing.T) {
	first := &fakeSource{name: "bcr", drifts: []Drift{
		{Name: "rules_go", Current: "1.0", Latest: "1.1"},
		{Name: "gazelle", Current: "2.0", Latest: "2.0"},
	}}
	second := &fakeSource{name: "go", drifts: []Drift{
		{Name: "example.org/lib", Current: "v1.0", Latest: "v1.1"},
	}}
	runner := Runner{Sources: []Source{first, second}}

	reports, err := runner.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 2 || reports[0].Source != "bcr" || reports[1].Source != "go" {
		t.Fatalf("unexpected reports: %#v", reports)
	}
	if len(first.fixed) != 0 || len(second.fixed) != 0 {
		t.Fatal("check changed a pin")
	}

	_, err = runner.Fix(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(first.fixed) != 1 || first.fixed[0].Name != "rules_go" || len(second.fixed) != 1 || second.fixed[0].Name != "example.org/lib" {
		t.Fatalf("fixes went to the wrong sources: %#v, %#v", first.fixed, second.fixed)
	}
}
