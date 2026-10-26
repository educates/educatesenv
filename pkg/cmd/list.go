package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/educates/educatesenv/pkg/config"
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

		// Get all files in the bin directory
		files, err := os.ReadDir(cfg.Local.Dir)
		// If the bin directory doesn't exist call init
		if err != nil {
			_, _, _, _, err := config.CreateConfigAndFolders()
			if err != nil {
				return fmt.Errorf("you should run `educatesenv init` first")
			}
		}

		// Get the active version by reading the symlink
		activeVersion := ""
		isDevelopmentActive := false
		symlink := filepath.Join(cfg.Local.Dir, "educates")
		if target, err := os.Readlink(symlink); err == nil {
			// Check if the active version is a development version
			if cfg.Development.Enabled && target == cfg.Development.BinaryLocation {
				isDevelopmentActive = true
			} else {
				activeVersion = filepath.Base(target)
				activeVersion = strings.TrimPrefix(activeVersion, "educates-")
			}
		}

		// Print installed versions
		_, _ = fmt.Fprintln(out, "Installed versions:")

		// Show development version if enabled
		if cfg.Development.Enabled {
			if isDevelopmentActive {
				_, _ = fmt.Fprintf(out, "* develop (active) -> %s\n", cfg.Development.BinaryLocation)
			} else {
				_, _ = fmt.Fprintf(out, "  develop -> %s\n", cfg.Development.BinaryLocation)
			}
		}

		// Print regular versions
		foundVersions := false
		for _, file := range files {
			if file.IsDir() || file.Name() == "educates" {
				continue
			}
			if strings.HasPrefix(file.Name(), "educates-") {
				version := strings.TrimPrefix(file.Name(), "educates-")
				if version == activeVersion {
					_, _ = fmt.Fprintf(out, "* %s (active)\n", version)
				} else {
					_, _ = fmt.Fprintf(out, "  %s\n", version)
				}
				foundVersions = true
			}
		}

		if !foundVersions && !cfg.Development.Enabled {
			_, _ = fmt.Fprintln(out, "No versions installed")
		}

		return nil
	}
}
