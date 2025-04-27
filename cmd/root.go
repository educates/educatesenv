package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config holds all configuration for the CLI
var Config struct {
	Github struct {
		Org        string
		Repository string
		Token      string
	}
	Local struct {
		Folder string
	}
}

var rootCmd = &cobra.Command{
	Use:   "educatesenv",
	Short: "Manage multiple versions of the educates binary",
	Long:  `A version manager for educates, inspired by tfenv.`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize educatesenv: create default config and folders, and show PATH instructions",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		defaultBin := filepath.Join(home, ".educatesenv", "bin")
		configDir := filepath.Join(home, ".educatesenv")
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
			defaultConfig := fmt.Sprintf(`github:\n  org: educates\n  repository: educates-training-platform\n  token: ""\nlocal:\n  folder: %s\n`, defaultBin)
			if err := os.WriteFile(configPath, []byte(defaultConfig), 0o644); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}
			fmt.Printf("Config file created at %s\n", configPath)
		}
		fmt.Printf("Bin directory ensured at %s\n", defaultBin)

		// Print PATH instructions
		fmt.Println("\nTo use educatesenv, add the bin directory to your PATH:")
		var pathCmd string
		switch runtime.GOOS {
		case "darwin", "linux":
			pathCmd = fmt.Sprintf("echo 'export PATH=\"%s:$PATH\"' >> ~/.bashrc\nsource ~/.bashrc", defaultBin)
			fmt.Printf("\nFor bash/zsh, run:\n  %s\n", pathCmd)
			fmt.Printf("\nFor fish shell, run:\n  set -U fish_user_paths %s $fish_user_paths\n", defaultBin)
		case "windows":
			fmt.Printf("\nFor Windows (PowerShell), run:\n  [Environment]::SetEnvironmentVariable(\"Path\", \"%s;\" + $Env:Path, [EnvironmentVariableTarget]::User)\n", defaultBin)
		default:
			fmt.Printf("\nAdd %s to your PATH manually.\n", defaultBin)
		}
		fmt.Println("\nRestart your terminal or source your profile to apply the changes.")
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(initCmd)
}

func initConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Warning: could not determine home directory, using current directory for bin folder.")
		home = "."
	}
	defaultBin := filepath.Join(home, ".educatesenv", "bin")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Join(home, ".educatesenv"))

	// Set defaults
	viper.SetDefault("github.org", "educates")
	viper.SetDefault("github.repository", "educates-training-platform")
	viper.SetDefault("github.token", "")
	viper.SetDefault("local.folder", defaultBin)

	// Bind environment variables
	viper.BindEnv("github.org", "EDUCATES_GITHUB_ORG")
	viper.BindEnv("github.repository", "EDUCATES_GITHUB_REPOSITORY")
	viper.BindEnv("github.token", "EDUCATES_GITHUB_TOKEN")
	viper.BindEnv("local.folder", "EDUCATES_LOCAL_FOLDER")

	// Read config file if present
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	Config.Github.Org = viper.GetString("github.org")
	Config.Github.Repository = viper.GetString("github.repository")
	Config.Github.Token = viper.GetString("github.token")
	Config.Local.Folder = viper.GetString("local.folder")
}
