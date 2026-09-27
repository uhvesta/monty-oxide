package versionchecker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// Rust tracks the toolchain pin in the workspace package manifest.
type Rust struct {
	Root   string
	Client *http.Client
}

func (Rust) Name() string { return "rust" }

func (r Rust) pin() (string, error) {
	_, _, manifest, err := readCargo(r.Root)
	if err != nil {
		return "", err
	}
	if manifest.Workspace.Package.RustVersion == "" {
		return "", fmt.Errorf("Cargo.toml has no workspace rust-version")
	}
	return manifest.Workspace.Package.RustVersion, nil
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
	path, data, manifest, err := readCargo(r.Root)
	if err != nil {
		return err
	}
	if drift.Name != "rust-version" || manifest.Workspace.Package.RustVersion != drift.Current {
		return fmt.Errorf("Cargo.toml rust-version changed since check")
	}
	updated, err := replaceCargoVersion(data, "workspace.package", "rust-version", "", drift.Current, drift.Latest)
	if err != nil {
		return err
	}
	return os.WriteFile(path, updated, 0o644)
}
