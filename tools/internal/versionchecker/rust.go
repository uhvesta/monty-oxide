package versionchecker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var workspaceRustVersion = regexp.MustCompile(`(?m)^rust-version\s*=\s*"([^"]+)"\s*$`)

// Rust tracks the toolchain pin in the workspace package manifest.
type Rust struct {
	Root   string
	Client *http.Client
}

func (Rust) Name() string { return "rust" }

func (r Rust) pin() (string, error) {
	data, err := os.ReadFile(filepath.Join(r.Root, "Cargo.toml"))
	if err != nil {
		return "", err
	}
	section := strings.SplitN(string(data), "[workspace.package]", 2)
	if len(section) != 2 {
		return "", fmt.Errorf("Cargo.toml has no [workspace.package] section")
	}
	workspace := strings.SplitN(section[1], "\n[", 2)[0]
	match := workspaceRustVersion.FindStringSubmatch(workspace)
	if match == nil {
		return "", fmt.Errorf("Cargo.toml has no workspace rust-version")
	}
	return match[1], nil
}

func (r Rust) Check(ctx context.Context) ([]Drift, error) {
	current, err := r.pin()
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/rust-lang/rust/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	client := r.Client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Rust releases: %s", response.Status)
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return nil, err
	}
	latest := strings.TrimPrefix(release.TagName, "v")
	if latest == "" {
		return nil, fmt.Errorf("Rust release has no tag")
	}
	if current == latest {
		return nil, nil
	}
	return []Drift{{Name: "rust-version", Current: current, Latest: latest}}, nil
}

func (r Rust) Fix(_ context.Context, drift Drift) error {
	current, err := r.pin()
	if err != nil {
		return err
	}
	if drift.Name != "rust-version" || current != drift.Current {
		return fmt.Errorf("Cargo.toml rust-version changed since check")
	}
	path := filepath.Join(r.Root, "Cargo.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	old := `rust-version = "` + drift.Current + `"`
	newPin := `rust-version = "` + drift.Latest + `"`
	return os.WriteFile(path, []byte(strings.Replace(string(data), old, newPin, 1)), 0o644)
}
