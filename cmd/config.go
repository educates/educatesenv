package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage educatesenv configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a default config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		defaultBin := filepath.Join(home, ".educatesenv", "bin")
		configDir := filepath.Join(home, ".educatesenv")
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		configPath := filepath.Join(configDir, "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("config file already exists at %s", configPath)
		}
		defaultConfig := fmt.Sprintf(`github:
  org: educates
  repository: educates-training-platform
  token: ""
local:
  dir: %s
development:
  enabled: false
  binaryLocation: ""
`, defaultBin)
		if err := os.WriteFile(configPath, []byte(defaultConfig), 0o644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
		fmt.Printf("Config file created at %s\n", configPath)
		return nil
	},
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Show the current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		allSettings := viper.AllSettings()
		data, err := yaml.Marshal(allSettings)
		if err != nil {
			return fmt.Errorf("failed to marshal config to YAML: %w", err)
		}
		fmt.Print(string(data))
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configViewCmd)
	rootCmd.AddCommand(configCmd)
}
