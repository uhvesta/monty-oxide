package main

import (
	"os"

	"github.com/uhvesta/monty-oxide/tools/internal/versionchecker"
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
	if err := newCommand(versionchecker.Runner{Sources: []versionchecker.Source{
		versionchecker.Bazel{Root: root},
		versionchecker.Bzlmod{Root: root},
	}}).Execute(); err != nil {
		os.Exit(1)
	}
}
