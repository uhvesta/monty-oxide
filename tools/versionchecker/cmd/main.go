package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/uhvesta/monty-oxide/tools/versionchecker/internal"
)

func main() {
	root := os.Getenv("BUILD_WORKSPACE_DIRECTORY")
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			os.Exit(1)
		}
	}
	goSource := &versionchecker.Golang{Root: root}
	bzlmodSource := &versionchecker.Bzlmod{Root: root}
	crateSource := &versionchecker.Crates{Root: root}
	command := newCommand(versionchecker.Runner{Sources: []versionchecker.Source{
		versionchecker.Bazel{Root: root},
		bzlmodSource,
		goSource,
		&versionchecker.Gomod{Go: goSource},
		versionchecker.Rust{Root: root},
		crateSource,
	}})
	var goRunfile, cargoRunfile, bazelBinary string
	command.PersistentFlags().StringVar(&goRunfile, "go-runfile", "", "Go SDK executable runfile")
	command.PersistentFlags().StringVar(&cargoRunfile, "cargo-runfile", "", "Cargo SDK executable runfile")
	command.PersistentFlags().StringVar(&bazelBinary, "bazel-bin", "bazel", "Bazel executable for lockfile updates")
	command.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		if goRunfile == "" {
			return fmt.Errorf("--go-runfile is required")
		}
		var err error
		goSource.GoBinary, err = resolveRunfile(goRunfile)
		if err != nil {
			return err
		}
		if cargoRunfile == "" {
			return fmt.Errorf("--cargo-runfile is required")
		}
		crateSource.CargoBinary, err = resolveRunfile(cargoRunfile)
		bzlmodSource.BazelBinary = bazelBinary
		return err
	}
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

func resolveRunfile(path string) (string, error) {
	if !filepath.IsAbs(path) {
		if path != filepath.Clean(path) || strings.HasPrefix(path, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("invalid runfile path: %q", path)
		}
		runfilesDir := os.Getenv("RUNFILES_DIR")
		if runfilesDir == "" {
			executable, err := os.Executable()
			if err != nil {
				return "", err
			}
			runfilesDir = executable + ".runfiles"
		}
		path = filepath.Join(runfilesDir, path)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("SDK runfile %q: %w", path, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Mode()&0o111 == 0 {
		return "", fmt.Errorf("SDK runfile is not executable: %q", path)
	}
	return resolved, nil
}
