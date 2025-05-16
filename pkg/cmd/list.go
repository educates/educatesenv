package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:           "list",
	Short:         "List installed educates versions",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get all files in the bin directory
		files, err := os.ReadDir(cfg.Local.Dir)
		// If the bin directory doesn't exist call init
		if err != nil {
			_, _, _, _, err := config.CreateConfigAndFolders()
			if err != nil {
				return fmt.Errorf("You should run `educatesenv init` first")
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
		fmt.Println("Installed versions:")

		// Show development version if enabled
		if cfg.Development.Enabled {
			if isDevelopmentActive {
				fmt.Printf("* develop (active) -> %s\n", cfg.Development.BinaryLocation)
			} else {
				fmt.Printf("  develop -> %s\n", cfg.Development.BinaryLocation)
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
					fmt.Printf("* %s (active)\n", version)
				} else {
					fmt.Printf("  %s\n", version)
				}
				foundVersions = true
			}
		}

		if !foundVersions && !cfg.Development.Enabled {
			fmt.Println("No versions installed")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
