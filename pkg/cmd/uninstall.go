package cmd

import (
	"github.com/spf13/cobra"
)

func newUninstallCommand(d *deps) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:           "uninstall <version>...",
		Short:         "Uninstall one or more versions of educates",
		Args:          cobra.MinimumNArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.manager.UninstallVersions(args, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Uninstall the active version too, leaving no version active")
	return cmd
}
