package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"

	"github.com/educates/educatesenv/pkg/version"
)

func newUseCommand(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:           "use <version|develop>",
		Short:         "Switch to a specific educates version",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun:        d.cleanUpDevelopmentSymlink,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Handle development version
			if name == "develop" {
				if !d.cfg.Development.Enabled {
					return fmt.Errorf("development mode is not enabled. Enable it in the config file by setting development.enabled to true")
				}
				if d.cfg.Development.BinaryLocation == "" {
					return fmt.Errorf("development binary location is not set. Set development.binaryLocation in the config file")
				}

				// Check if the development binary exists
				if _, err := os.Stat(d.cfg.Development.BinaryLocation); err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf("development binary not found at %s", d.cfg.Development.BinaryLocation)
					}
					return fmt.Errorf("failed to check development binary: %w", err)
				}

				if err := d.manager.UseVersion(name); err != nil {
					return fmt.Errorf("failed to switch to development version: %w", err)
				}

				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Now using educates development version from %s\n", d.cfg.Development.BinaryLocation)
				return nil
			}

			// Handle regular version
			if err := d.manager.UseVersion(name); err != nil {
				switch {
				// install resolves latest to a version, while use only
				// switches to a version that is installed.
				case errors.Is(err, fs.ErrNotExist) && name == version.Latest:
					return fmt.Errorf("use takes an installed version, not %[1]s. "+
						"Run 'educatesenv install %[1]s --use' to install the latest stable release and switch to it", version.Latest)
				case errors.Is(err, fs.ErrNotExist):
					return fmt.Errorf("version %s is not installed. You should install it first with 'educatesenv install %s'", name, name)
				}
				return fmt.Errorf("failed to switch to version %s: %w", name, err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Now using educates version %s\n", name)
			return nil
		},
	}
}
