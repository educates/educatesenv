package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// switchVersion handles the common logic for switching between educates versions
// sourcePath is the path to the binary to link to
// If sourcePath is relative, it will be resolved relative to binDir
func switchVersion(sourcePath string, description string) error {
	symlinkPath := filepath.Join(Config.Local.Dir, "educates")

	// Check if the source binary exists
	if _, err := os.Stat(sourcePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s not found at %s", description, sourcePath)
		}
		return fmt.Errorf("failed to check %s: %w", description, err)
	}

	// Remove existing symlink if it exists
	if fi, err := os.Lstat(symlinkPath); err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(symlinkPath); err != nil {
				return fmt.Errorf("failed to remove existing symlink: %w", err)
			}
		} else {
			return fmt.Errorf("%s exists and is not a symlink", symlinkPath)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check symlink: %w", err)
	}

	// Create new symlink
	relTarget, err := filepath.Rel(Config.Local.Dir, sourcePath)
	if err != nil {
		relTarget = sourcePath // fallback to absolute path
	}
	if err := os.Symlink(relTarget, symlinkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

var useCmd = &cobra.Command{
	Use:   "use [version]",
	Short: "Select the active educates version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version := args[0]

		// Handle development version
		if version == "develop" {
			if !Config.Development.Enabled {
				return fmt.Errorf("development version is not enabled. Enable it in the config file by setting development.enabled to true")
			}
			if Config.Development.BinaryLocation == "" {
				return fmt.Errorf("development binary location is not set. Set development.binaryLocation in the config file")
			}

			if err := switchVersion(Config.Development.BinaryLocation, "development binary"); err != nil {
				return err
			}

			fmt.Printf("Now using educates development version from %s\n", Config.Development.BinaryLocation)
			return nil
		}

		binaryPath := filepath.Join(Config.Local.Dir, fmt.Sprintf("educates-%s", version))
		if err := switchVersion(binaryPath, fmt.Sprintf("version %s", version)); err != nil {
			return err
		}

		fmt.Printf("Now using educates version %s\n", version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
