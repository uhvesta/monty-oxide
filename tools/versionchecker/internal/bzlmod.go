package versionchecker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var bazelDep = regexp.MustCompile(`bazel_dep\s*\(\s*name\s*=\s*"([^"]+)"\s*,\s*version\s*=\s*"([^"]+)"\s*\)`)

// Bzlmod checks direct module pins against the Bazel Central Registry.
type Bzlmod struct {
	Root   string
	Client *http.Client
}

func (Bzlmod) Name() string { return "bzlmod" }

func (b Bzlmod) Check(ctx context.Context) ([]Drift, error) {
	data, err := os.ReadFile(filepath.Join(b.Root, "MODULE.bazel"))
	if err != nil {
		return nil, err
	}
	client := b.Client
	if client == nil {
		client = http.DefaultClient
	}
	var drifts []Drift
	for _, match := range bazelDep.FindAllStringSubmatch(string(data), -1) {
		name, current := match[1], match[2]
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://bcr.bazel.build/modules/"+name+"/metadata.json", nil)
		if err != nil {
			return nil, err
		}
		response, err := client.Do(request)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			return nil, fmt.Errorf("%s: registry returned %s", name, response.Status)
		}
		var metadata struct {
			Versions []string          `json:"versions"`
			Yanked   map[string]string `json:"yanked_versions"`
		}
		err = json.NewDecoder(response.Body).Decode(&metadata)
		response.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		latest := ""
		for i := len(metadata.Versions) - 1; i >= 0; i-- {
			if _, yanked := metadata.Yanked[metadata.Versions[i]]; !yanked {
				latest = metadata.Versions[i]
				break
			}
		}
		if latest == "" {
			return nil, fmt.Errorf("%s: no available version", name)
		}
		if current != latest {
			drifts = append(drifts, Drift{Name: name, Current: current, Latest: latest})
		}
	}
	return drifts, nil
}

func (b Bzlmod) Fix(_ context.Context, drift Drift) error {
	path := filepath.Join(b.Root, "MODULE.bazel")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	old := fmt.Sprintf(`bazel_dep(name = %q, version = %q)`, drift.Name, drift.Current)
	newPin := fmt.Sprintf(`bazel_dep(name = %q, version = %q)`, drift.Name, drift.Latest)
	if !strings.Contains(string(data), old) {
		return fmt.Errorf("%s pin changed since check", drift.Name)
	}
	return os.WriteFile(path, []byte(strings.Replace(string(data), old, newPin, 1)), 0o644)
}

// FinishFix refreshes every module extension entry after all pins have changed.
func (b Bzlmod) FinishFix(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, filepath.Join(b.Root, ".tools/bin/bazel"), "mod", "deps", "--lockfile_mode=update")
	cmd.Dir = b.Root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("bazel mod deps: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
