package cmd

import (
	"github.com/spf13/cobra"
)

func newListLocalCommand(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:           "list-local",
		Short:         "List installed educates versions (alias for list)",
		Hidden:        true,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          runList(d),
	}
}
