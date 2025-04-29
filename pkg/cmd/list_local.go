package cmd

import (
	"github.com/spf13/cobra"
)

var listLocalCmd = &cobra.Command{
	Use:           "list-local",
	Short:         "List installed educates versions (alias for list)",
	Hidden:        true,
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE:          listCmd.RunE,
}

func init() {
	rootCmd.AddCommand(listLocalCmd)
}
