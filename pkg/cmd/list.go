package cmd

import (
	"fmt"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/educates/educatesenv/pkg/releases"
	"github.com/educates/educatesenv/pkg/version"
	"github.com/spf13/cobra"
)

func newListCommand(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:           "list",
		Short:         "List installed educates versions",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun:        d.cleanUpDevelopmentSymlink,
		RunE:          runList(d),
	}
}

// runList returns the run function that list and list-local share.
func runList(d *deps) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cfg := d.cfg
		out := cmd.OutOrStdout()

		versions, activeVersion, err := d.manager.Installed()
		// If the bin directory cannot be read, create the default config
		// and bin directory, as init does
		if err != nil {
			if _, _, err := config.CreateDefault(); err != nil {
				return fmt.Errorf("you should run `educatesenv init` first")
			}
		}

		// Print installed versions
		_, _ = fmt.Fprintln(out, "Installed versions:")

		// Show development version if enabled
		if cfg.Development.Enabled {
			if activeVersion == version.Develop {
				_, _ = fmt.Fprintf(out, "* develop (active) -> %s\n", cfg.Development.BinaryLocation)
			} else {
				_, _ = fmt.Fprintf(out, "  develop -> %s\n", cfg.Development.BinaryLocation)
			}
		}

		// Print regular versions, newest first
		for _, name := range releases.SortInstalled(versions) {
			if name == activeVersion {
				_, _ = fmt.Fprintf(out, "* %s (active)\n", name)
			} else {
				_, _ = fmt.Fprintf(out, "  %s\n", name)
			}
		}

		if len(versions) == 0 && !cfg.Development.Enabled {
			_, _ = fmt.Fprintln(out, "No versions installed")
		}

		return nil
	}
}
