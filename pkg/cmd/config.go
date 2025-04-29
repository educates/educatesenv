package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/educates/educatesenv/pkg/config"
)

var configCmd = &cobra.Command{
	Use:           "config",
	Short:         "Manage educatesenv configuration",
	SilenceErrors: true,
	SilenceUsage:  true,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a default config file",
	// Disable the automatic error printing
	SilenceErrors: true,
	// Disable automatic usage printing on error
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		configDir := filepath.Join(home, config.ConfigDirName)
		configPath := filepath.Join(configDir, "config.yaml")

		// Check if config already exists
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("config file already exists at %s", configPath)
		}

		// Create config directory if it doesn't exist
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Create new config with defaults
		configFile := config.New()

		// Write configuration to file
		yamlBytes, err := yaml.Marshal(&configFile)
		if err != nil {
			return fmt.Errorf("failed to marshal config to YAML: %w", err)
		}

		if err := os.WriteFile(configPath, yamlBytes, 0o644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		fmt.Printf("Config file created at %s\n", configPath)
		return nil
	},
}

var configViewCmd = &cobra.Command{
	Use:           "view",
	Short:         "Show the current configuration",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Convert config struct to YAML
		yamlBytes, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal config to YAML: %w", err)
		}

		// Print the configuration
		fmt.Println("Current configuration:")
		fmt.Println(string(yamlBytes))
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configViewCmd)
	rootCmd.AddCommand(configCmd)
}
