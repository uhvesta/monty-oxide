package versionchecker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bazelbuild/buildtools/build"
)

type modulePin struct {
	name    string
	version *build.StringExpr
}

func modulePins(path string, data []byte) ([]modulePin, error) {
	file, err := build.ParseModule(path, data)
	if err != nil {
		return nil, err
	}
	var pins []modulePin
	for _, statement := range file.Stmt {
		call, ok := statement.(*build.CallExpr)
		if !ok {
			continue
		}
		function, ok := call.X.(*build.Ident)
		if !ok || function.Name != "bazel_dep" {
			continue
		}
		var name, version *build.StringExpr
		for _, argument := range call.List {
			assignment, ok := argument.(*build.AssignExpr)
			if !ok {
				continue
			}
			key, ok := assignment.LHS.(*build.Ident)
			if !ok {
				continue
			}
			value, ok := assignment.RHS.(*build.StringExpr)
			if !ok && (key.Name == "name" || key.Name == "version") {
				return nil, fmt.Errorf("bazel_dep %s must be a string literal", key.Name)
			}
			switch key.Name {
			case "name":
				name = value
			case "version":
				version = value
			}
		}
		if name == nil {
			return nil, fmt.Errorf("bazel_dep must have a name string literal")
		}
		if version == nil || version.Value == "" {
			continue
		}
		pins = append(pins, modulePin{name: name.Value, version: version})
	}
	return pins, nil
}

// Bzlmod checks direct module pins against the Bazel Central Registry.
type Bzlmod struct {
	Root        string
	BazelBinary string
	Client      *http.Client
}

func (Bzlmod) Name() string { return "bzlmod" }

func (b Bzlmod) Check(ctx context.Context) ([]Drift, error) {
	data, err := os.ReadFile(filepath.Join(b.Root, "MODULE.bazel"))
	if err != nil {
		return nil, err
	}
	pins, err := modulePins(filepath.Join(b.Root, "MODULE.bazel"), data)
	if err != nil {
		return nil, err
	}
	client := b.Client
	if client == nil {
		client = http.DefaultClient
	}
	var drifts []Drift
	for _, pin := range pins {
		name, current := pin.name, pin.version.Value
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
	pins, err := modulePins(path, data)
	if err != nil {
		return err
	}
	for _, pin := range pins {
		if pin.name != drift.Name {
			continue
		}
		if pin.version.Value != drift.Current {
			return fmt.Errorf("%s pin changed since check", drift.Name)
		}
		start, end := pin.version.Span()
		updated := make([]byte, 0, len(data)+len(drift.Latest)-len(drift.Current))
		updated = append(updated, data[:start.Byte]...)
		updated = append(updated, strconv.Quote(drift.Latest)...)
		updated = append(updated, data[end.Byte:]...)
		return os.WriteFile(path, updated, 0o644)
	}
	return fmt.Errorf("%s pin changed since check", drift.Name)
}

// FinishFix refreshes every module extension entry after all pins have changed.
func (b Bzlmod) FinishFix(ctx context.Context) error {
	if b.BazelBinary == "" {
		b.BazelBinary = "bazel"
	}
	cmd := exec.CommandContext(ctx, b.BazelBinary, "mod", "deps", "--lockfile_mode=update")
	cmd.Dir = b.Root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("bazel mod deps: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
