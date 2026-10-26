package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newUseCommand(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:           "use [version|develop]",
		Short:         "Switch to a specific educates version",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]

			// Handle development version
			if version == "develop" {
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

				if err := d.manager.UseVersion(version); err != nil {
					return fmt.Errorf("failed to switch to development version: %w", err)
				}

				fmt.Printf("Now using educates development version from %s\n", d.cfg.Development.BinaryLocation)
				return nil
			}

			// Handle regular version
			if err := d.manager.UseVersion(version); err != nil {
				return fmt.Errorf("version %s is not installed. You should install it first with `educatesenv install %s`", version, version)
			}

			fmt.Printf("Now using educates version %s\n", version)
			return nil
		},
	}
}
