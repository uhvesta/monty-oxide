package main

import (
	"os"

	"github.com/uhvesta/monty-oxide/tools/internal/versionchecker"
)

func main() {
	// Register concrete sources here as they are implemented.
	if err := newCommand(versionchecker.Runner{}).Execute(); err != nil {
		os.Exit(1)
	}
}
