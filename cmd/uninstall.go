package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall [version]",
	Short: "Uninstall a specific educates version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version := args[0]
		binDir := Config.Local.Dir
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

		// Check if this version is currently in use
		inUse := false
		if linkTarget, err := os.Readlink(symlinkPath); err == nil {
			// If the symlink is relative, resolve it to absolute
			if !filepath.IsAbs(linkTarget) {
				linkTarget = filepath.Join(binDir, linkTarget)
			}
			if linkTarget == binaryPath {
				inUse = true
			}
		}

		// Remove the binary
		if err := os.Remove(binaryPath); err != nil {
			return fmt.Errorf("failed to remove binary: %w", err)
		}

		// If in use, remove the symlink and print a message
		if inUse {
			if err := os.Remove(symlinkPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove symlink: %w", err)
			}
			fmt.Printf("Uninstalled version %s, which was the active version. Please select a new version with 'educatesenv use <version>'.\n", version)
		} else {
			fmt.Printf("Uninstalled educates version %s\n", version)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
