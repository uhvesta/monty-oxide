package versionchecker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// Golang checks the language version pinned in go.mod.
type Golang struct {
	Root     string
	GoBinary string
	Client   *http.Client
}

type goMod struct {
	Go      string `json:"Go"`
	Require []struct {
		Path    string `json:"Path"`
		Version string `json:"Version"`
	} `json:"Require"`
}

func (*Golang) Name() string { return "golang" }

func (g *Golang) goCommand(ctx context.Context, args ...string) ([]byte, error) {
	if g.GoBinary == "" {
		return nil, fmt.Errorf("Go executable is not configured")
	}
	cmd := exec.CommandContext(ctx, g.GoBinary, args...)
	cmd.Dir = g.Root
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("go %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

func (g *Golang) mod(ctx context.Context) (goMod, error) {
	var mod goMod
	output, err := g.goCommand(ctx, "mod", "edit", "-json")
	if err != nil {
		return mod, err
	}
	if err := json.Unmarshal(output, &mod); err != nil {
		return mod, err
	}
	return mod, nil
}

func (g *Golang) Check(ctx context.Context) ([]Drift, error) {
	mod, err := g.mod(ctx)
	if err != nil {
		return nil, err
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
	if mod.Go != latest {
		return []Drift{{Name: "go", Current: mod.Go, Latest: latest}}, nil
	}
	return nil, nil
}

func (g *Golang) Fix(ctx context.Context, drift Drift) error {
	mod, err := g.mod(ctx)
	if err != nil {
		return err
	}
	if drift.Name != "go" || mod.Go != drift.Current {
		return fmt.Errorf("Go version changed since check")
	}
	if _, err := g.goCommand(ctx, "get", "go@"+drift.Latest); err != nil {
		return err
	}
	if _, err := g.goCommand(ctx, "mod", "tidy"); err != nil {
		return err
	}
	return nil
}
