package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use [version]",
	Short: "Select the active educates version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version := args[0]
		binDir := Config.Local.Folder
		binaryName := fmt.Sprintf("educates-%s", version)
		binaryPath := filepath.Join(binDir, binaryName)
		symlinkPath := filepath.Join(binDir, "educates")

		// Check if the binary exists
		if _, err := os.Stat(binaryPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("version %s is not installed (expected at %s)", version, binaryPath)
			}
			return fmt.Errorf("failed to stat binary: %w", err)
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
		relTarget, err := filepath.Rel(binDir, binaryPath)
		if err != nil {
			relTarget = binaryPath // fallback to absolute path
		}
		if err := os.Symlink(relTarget, symlinkPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}

		fmt.Printf("Now using educates version %s\n", version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
