package versionchecker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// Gomod checks module requirements and regenerates go.sum after upgrades.
type Gomod struct {
	Go      *Golang
	pending []Drift
}

func (*Gomod) Name() string { return "gomod" }

func (g *Gomod) pins(ctx context.Context) (map[string]string, error) {
	mod, err := g.Go.mod(ctx)
	if err != nil {
		return nil, err
	}
	pins := make(map[string]string, len(mod.Require))
	for _, require := range mod.Require {
		pins[require.Path] = require.Version
	}
	return pins, nil
}

func (g *Gomod) Check(ctx context.Context) ([]Drift, error) {
	pins, err := g.pins(ctx)
	if err != nil {
		return nil, err
	}
	if len(pins) == 0 {
		return nil, nil
	}
	output, err := g.Go.goCommand(ctx, "list", "-m", "-u", "-json", "all")
	if err != nil {
		return nil, err
	}
	var drifts []Drift
	decoder := json.NewDecoder(bytes.NewReader(output))
	for decoder.More() {
		var module struct {
			Path   string `json:"Path"`
			Update *struct {
				Version string `json:"Version"`
			} `json:"Update"`
		}
		if err := decoder.Decode(&module); err != nil {
			return nil, err
		}
		if current, pinned := pins[module.Path]; pinned && module.Update != nil && current != module.Update.Version {
			drifts = append(drifts, Drift{Name: module.Path, Current: current, Latest: module.Update.Version})
		}
	}
	return drifts, nil
}

func (g *Gomod) Fix(ctx context.Context, drift Drift) error {
	pins, err := g.pins(ctx)
	if err != nil {
		return err
	}
	if pins[drift.Name] != drift.Current {
		return fmt.Errorf("%s pin changed since check", drift.Name)
	}
	g.pending = append(g.pending, drift)
	return nil
}

func (g *Gomod) FinishFix(ctx context.Context) error {
	if len(g.pending) == 0 {
		return nil
	}
	args := []string{"get"}
	for _, drift := range g.pending {
		args = append(args, drift.Name+"@"+drift.Latest)
	}
	if _, err := g.Go.goCommand(ctx, args...); err != nil {
		return err
	}
	if _, err := g.Go.goCommand(ctx, "mod", "tidy"); err != nil {
		return err
	}
	g.pending = nil
	return nil
}
