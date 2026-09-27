// Command versionchecker reports and fixes version drift.
package main

import (
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
		newActionCommand("check", "Report version drift", runner),
		newActionCommand("fix", "Update drifted versions", runner),
	)
	return root
}

func newActionCommand(name, description string, runner versionchecker.Runner) *cobra.Command {
	selected := make(map[string]*bool, len(runner.Sources))
	cmd := &cobra.Command{
		Use:   name,
		Short: description,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			all, _ := cmd.Flags().GetBool("all")
			var sources []versionchecker.Source
			for _, source := range runner.Sources {
				if *selected[source.Name()] {
					sources = append(sources, source)
				}
			}
			if all && len(sources) > 0 {
				return fmt.Errorf("--all cannot be combined with a source flag")
			}
			if all || len(sources) == 0 {
				sources = runner.Sources
			}
			scoped := versionchecker.Runner{Sources: sources}
			var reports []versionchecker.Report
			var err error
			if name == "fix" {
				reports, err = scoped.Fix(cmd.Context())
			} else {
				reports, err = scoped.Check(cmd.Context())
			}
			if err != nil {
				return err
			}
			for _, report := range reports {
				fmt.Fprintf(cmd.OutOrStdout(), "%s/%s: %s -> %s\n", report.Source, report.Drift.Name, report.Drift.Current, report.Drift.Latest)
			}
			return nil
		},
	}
	cmd.Flags().Bool("all", false, "Check or update every source (the default)")
	for _, source := range runner.Sources {
		flag := new(bool)
		selected[source.Name()] = flag
		cmd.Flags().BoolVar(flag, source.Name(), false, "Include "+source.Name()+" pins")
	}
	return cmd
}
