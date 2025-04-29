package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/educates/educatesenv/pkg/platform"
)

var (
	downloadLatest bool
	overwrite      bool
)

var initCmd = &cobra.Command{
	Use:           "init",
	Short:         "Initialize educatesenv: create default config and folders, and show PATH instructions",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		configDir := filepath.Join(home, config.ConfigDirName)
		defaultBin := filepath.Join(configDir, "bin")
		configPath := filepath.Join(configDir, "config.yaml")

		if err := os.MkdirAll(defaultBin, 0o755); err != nil {
			return fmt.Errorf("failed to create bin directory: %w", err)
		}
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("Config file already exists at %s\n", configPath)
		} else {
			// Create new config with defaults
			configFile := config.New()

			yamlBytes, err := yaml.Marshal(&configFile)
			if err != nil {
				return fmt.Errorf("failed to marshal config to YAML: %w", err)
			}
			if err := os.WriteFile(configPath, yamlBytes, 0o644); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}
			fmt.Printf("Config file created at %s\n", configPath)
		}
		fmt.Printf("Bin directory ensured at %s\n", defaultBin)

		// Print PATH instructions
		fmt.Println("\nTo use educatesenv, add the bin directory to your PATH:")
		var pathCmd string
		switch runtime.GOOS {
		case platform.Darwin, platform.Linux:
			pathCmd = fmt.Sprintf("echo 'export PATH=\"%s:$PATH\"' >> ~/.bashrc\nsource ~/.bashrc", defaultBin)
			fmt.Printf("\nFor bash/zsh, run:\n  %s\n", pathCmd)
			fmt.Printf("\nFor fish shell, run:\n  set -U fish_user_paths %s $fish_user_paths\n", defaultBin)
		case platform.Windows:
			fmt.Printf("\nFor Windows (PowerShell), run:\n  [Environment]::SetEnvironmentVariable(\"Path\", \"%s;\" + $Env:Path, [EnvironmentVariableTarget]::User)\n", defaultBin)
		default:
			fmt.Printf("\nAdd %s to your PATH manually.\n", defaultBin)
		}
		fmt.Println("\nRestart your terminal or source your profile to apply the changes.")

		if downloadLatest {
			fmt.Println("\nFetching latest educates version...")
			latest, err := gh.GetLatestReleaseVersion()
			if err != nil {
				return fmt.Errorf("failed to get latest release version: %w", err)
			}
			fmt.Printf("Latest version: %s\n", latest)

			if err := manager.InstallVersion(latest, overwrite, true); err != nil {
				return err
			}
		} else {
			fmt.Println("\nRun 'educatesenv install <version>' to install a specific version")
			fmt.Println("Run 'educatesenv list-remote' to see available versions")
		}
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&downloadLatest, "download", false, "Download and set as active the latest stable version")
	initCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Force download even if the version already exists")
	rootCmd.AddCommand(initCmd)
}
