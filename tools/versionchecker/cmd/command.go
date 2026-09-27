// Command versionchecker reports and fixes version drift.
package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uhvesta/monty-oxide/tools/versionchecker/internal"
)

func newCommand(runner versionchecker.Runner) *cobra.Command {
	root := &cobra.Command{
		Use:          "versionchecker",
		Short:        "Check and update pinned versions",
		SilenceUsage: true,
	}
	root.AddCommand(
		newActionCommand("check", "Report version drift", runner.Check),
		newActionCommand("fix", "Update drifted versions", runner.Fix),
	)
	return root
}

func newActionCommand(name, description string, run func(context.Context) ([]versionchecker.Report, error)) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: description,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			reports, err := run(cmd.Context())
			if err != nil {
				return err
			}
			for _, report := range reports {
				fmt.Fprintf(cmd.OutOrStdout(), "%s/%s: %s -> %s\n", report.Source, report.Drift.Name, report.Drift.Current, report.Drift.Latest)
			}
			return nil
		},
	}
}
