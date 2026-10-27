package cmd

import (
	"github.com/spf13/cobra"
)

func newListLocalCommand(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:           "list-local",
		Short:         "List installed educates versions (alias for list)",
		Hidden:        true,
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun:        d.cleanUpDevelopmentSymlink,
		RunE:          runList(d),
	}
}
