// Command versions will report or fix version drift once sources are registered.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/uhvesta/monty-oxide/tools/internal/versions"
)

func main() {
	if len(os.Args) != 2 || (os.Args[1] != "check" && os.Args[1] != "fix") {
		fmt.Fprintln(os.Stderr, "usage: versions check|fix")
		os.Exit(2)
	}

	// Register concrete sources here as they are implemented.
	runner := versions.Runner{}
	var reports []versions.Report
	var err error
	if os.Args[1] == "fix" {
		reports, err = runner.Fix(context.Background())
	} else {
		reports, err = runner.Check(context.Background())
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "versions:", err)
		os.Exit(1)
	}
	for _, report := range reports {
		fmt.Printf("%s/%s: %s -> %s\n", report.Source, report.Drift.Name, report.Drift.Current, report.Drift.Latest)
	}
}
