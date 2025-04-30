package cmd

import (
	"fmt"

	"github.com/educates/educatesenv/pkg/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:           "version",
	Short:         "Print client version",
	Args:          cobra.ExactArgs(0),
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println(version.Version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
