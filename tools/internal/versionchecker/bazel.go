package versionchecker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Bazel tracks the Bazel release selected by Bazelisk.
type Bazel struct {
	Root   string
	Client *http.Client
}

func (Bazel) Name() string { return "bazel" }

func (b Bazel) Check(ctx context.Context) ([]Drift, error) {
	data, err := os.ReadFile(filepath.Join(b.Root, ".bazelversion"))
	if err != nil {
		return nil, err
	}
	current := strings.TrimSpace(string(data))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/bazelbuild/bazel/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	client := b.Client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub releases: %s", response.Status)
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return nil, err
	}
	latest := strings.TrimPrefix(release.TagName, "v")
	if latest == "" {
		return nil, fmt.Errorf("GitHub release has no tag")
	}
	if current == latest {
		return nil, nil
	}
	return []Drift{{Name: ".bazelversion", Current: current, Latest: latest}}, nil
}

func (b Bazel) Fix(_ context.Context, drift Drift) error {
	path := filepath.Join(b.Root, ".bazelversion")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if drift.Name != ".bazelversion" || strings.TrimSpace(string(data)) != drift.Current {
		return fmt.Errorf(".bazelversion changed since check")
	}
	return os.WriteFile(path, []byte(drift.Latest+"\n"), 0o644)
}
