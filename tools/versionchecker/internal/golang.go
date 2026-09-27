package versionchecker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Golang uses the pinned Go toolchain to inspect and upgrade go.mod dependencies.
type Golang struct {
	Root    string
	Client  *http.Client
	pending []Drift
}

func (*Golang) Name() string { return "golang" }

func (g *Golang) goCommand(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, filepath.Join(g.Root, ".tools/sdk/go/bin/go"), args...)
	cmd.Dir = g.Root
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("go %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

func (g *Golang) pins(ctx context.Context) (string, map[string]string, error) {
	output, err := g.goCommand(ctx, "mod", "edit", "-json")
	if err != nil {
		return "", nil, err
	}
	var mod struct {
		Go      string `json:"Go"`
		Require []struct {
			Path    string `json:"Path"`
			Version string `json:"Version"`
		} `json:"Require"`
	}
	if err := json.Unmarshal(output, &mod); err != nil {
		return "", nil, err
	}
	pins := make(map[string]string, len(mod.Require))
	for _, require := range mod.Require {
		pins[require.Path] = require.Version
	}
	return mod.Go, pins, nil
}

func (g *Golang) Check(ctx context.Context) ([]Drift, error) {
	goVersion, pins, err := g.pins(ctx)
	if err != nil {
		return nil, err
	}
	output, err := g.goCommand(ctx, "list", "-m", "-u", "-json", "all")
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
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://go.dev/dl/?mode=json", nil)
	if err != nil {
		return nil, err
	}
	client := g.Client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Go releases: %s", response.Status)
	}
	var releases []struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(response.Body).Decode(&releases); err != nil {
		return nil, err
	}
	if len(releases) == 0 {
		return nil, fmt.Errorf("Go releases: empty response")
	}
	latest := strings.TrimPrefix(releases[0].Version, "go")
	if goVersion != latest {
		drifts = append(drifts, Drift{Name: "go", Current: goVersion, Latest: latest})
	}
	return drifts, nil
}

func (g *Golang) Fix(ctx context.Context, drift Drift) error {
	goVersion, pins, err := g.pins(ctx)
	if err != nil {
		return err
	}
	current := goVersion
	if drift.Name != "go" {
		var ok bool
		current, ok = pins[drift.Name]
		if !ok {
			return fmt.Errorf("%s is no longer pinned", drift.Name)
		}
	}
	if current != drift.Current {
		return fmt.Errorf("%s pin changed since check", drift.Name)
	}
	g.pending = append(g.pending, drift)
	return nil
}

// FinishFix lets Go resolve all updates together, then regenerate go.sum.
func (g *Golang) FinishFix(ctx context.Context) error {
	if len(g.pending) == 0 {
		return nil
	}
	args := []string{"get"}
	for _, drift := range g.pending {
		args = append(args, drift.Name+"@"+drift.Latest)
	}
	if _, err := g.goCommand(ctx, args...); err != nil {
		return err
	}
	if _, err := g.goCommand(ctx, "mod", "tidy"); err != nil {
		return err
	}
	g.pending = nil
	return nil
}
