package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/go-github/v71/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the CLI
var Config struct {
	Github struct {
		Org        string
		Repository string
		Token      string
	}
	Local struct {
		Dir string
	}
	Development struct {
		Enabled        bool
		BinaryLocation string
	}
}

var (
	downloadLatest bool
	overwrite      bool
)

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
			type configFileStruct struct {
				Github struct {
					Org        string `yaml:"org"`
					Repository string `yaml:"repository"`
					Token      string `yaml:"token"`
				} `yaml:"github"`
				Local struct {
					Dir string `yaml:"dir"`
				} `yaml:"local"`
				Development struct {
					Enabled        bool   `yaml:"enabled"`
					BinaryLocation string `yaml:"binaryLocation"`
				} `yaml:"development"`
			}
			var configFile configFileStruct
			configFile.Github.Org = "educates"
			configFile.Github.Repository = "educates-training-platform"
			configFile.Github.Token = ""
			configFile.Local.Dir = defaultBin
			configFile.Development.Enabled = false
			configFile.Development.BinaryLocation = ""
			yamlBytes, err := yaml.Marshal(&configFile)
			if err != nil {
				return fmt.Errorf("failed to marshal config to YAML: %w", err)
			}
			if err := os.WriteFile(configPath, yamlBytes, 0o644); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}
			fmt.Printf("Config file created at %s\n", configPath)
			initConfig()
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

		if downloadLatest {
			fmt.Println("\nFetching latest educates version...")
			latest, err := getLatestReleaseVersion()
			if err != nil {
				return fmt.Errorf("failed to get latest release version: %w", err)
			}
			fmt.Printf("Latest version: %s\n", latest)

			if err := installAndOptionallyUse(latest, overwrite, true); err != nil {
				return err
			}
		} else {
			fmt.Println("\nRun 'educatesenv install <version>' to install a specific version")
			fmt.Println("Run 'educatesenv list-remote' to see available versions")
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	initCmd.Flags().BoolVar(&downloadLatest, "download", false, "Download and set as active the latest stable version")
	initCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Force download even if the version already exists")
	rootCmd.AddCommand(initCmd)
}

func initConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Warning: could not determine home directory, using current directory for bin folder.")
		home = "."
	}
	configDir := filepath.Join(home, ".educatesenv")
	defaultBin := filepath.Join(configDir, "bin")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath(configDir)

	// Set defaults
	viper.SetDefault("github.org", "educates")
	viper.SetDefault("github.repository", "educates-training-platform")
	viper.SetDefault("github.token", "")
	viper.SetDefault("local.dir", defaultBin)
	viper.SetDefault("development.enabled", false)
	viper.SetDefault("development.binaryLocation", "")

	// Bind environment variables
	viper.BindEnv("github.org", "EDUCATES_GITHUB_ORG")
	viper.BindEnv("github.repository", "EDUCATES_GITHUB_REPOSITORY")
	viper.BindEnv("github.token", "EDUCATES_GITHUB_TOKEN")
	viper.BindEnv("local.dir", "EDUCATES_LOCAL_DIR")
	viper.BindEnv("development.enabled", "EDUCATES_DEVELOPMENT_ENABLED")
	viper.BindEnv("development.binaryLocation", "EDUCATES_DEVELOPMENT_BINARY_LOCATION")

	// Read config file if present
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	Config.Github.Org = viper.GetString("github.org")
	Config.Github.Repository = viper.GetString("github.repository")
	Config.Github.Token = viper.GetString("github.token")
	Config.Local.Dir = viper.GetString("local.dir")
	Config.Development.Enabled = viper.GetBool("development.enabled")
	Config.Development.BinaryLocation = viper.GetString("development.binaryLocation")
}

// createGitHubClient creates a GitHub client with optional authentication
func createGitHubClient() *github.Client {
	ctx := context.Background()
	if Config.Github.Token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: Config.Github.Token})
		return github.NewClient(oauth2.NewClient(ctx, ts))
	}
	return github.NewClient(nil)
}

// --- Helper to get latest release version ---
func getLatestReleaseVersion() (string, error) {
	client := createGitHubClient()
	releases, _, err := client.Repositories.ListReleases(context.Background(), Config.Github.Org, Config.Github.Repository, &github.ListOptions{PerPage: 10})
	if err != nil {
		return "", err
	}
	for _, rel := range releases {
		if rel.TagName != nil && !rel.GetPrerelease() {
			return *rel.TagName, nil
		}
	}
	return "", fmt.Errorf("no stable releases found")
}

// getPlatformBinaryName returns the platform-specific binary name
func getPlatformBinaryName() (string, error) {
	platform, arch := runtime.GOOS, runtime.GOARCH
	switch platform {
	case "darwin":
		if arch == "arm64" {
			return "educates-darwin-arm64", nil
		}
		return "educates-darwin-amd64", nil
	case "linux":
		if arch == "arm64" {
			return "educates-linux-arm64", nil
		}
		return "educates-linux-amd64", nil
	default:
		return "", fmt.Errorf("unsupported platform: %s-%s", platform, arch)
	}
}

// installAndOptionallyUse installs a version and optionally sets it as active
// If overwrite is true, it will reinstall even if the version exists
// If activate is true, it will set the version as active after installation
func installAndOptionallyUse(version string, overwrite bool, activate bool) error {
	binDir := Config.Local.Dir
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Check if version already exists
	binaryPath := filepath.Join(binDir, fmt.Sprintf("educates-%s", version))
	_, err := os.Stat(binaryPath)
	versionExists := err == nil

	// Handle installation
	if versionExists && !overwrite {
		fmt.Printf("Version %s is already installed.\n", version)
	} else {
		if versionExists {
			fmt.Printf("Reinstalling version %s...\n", version)
		}

		assetName, err := getPlatformBinaryName()
		if err != nil {
			return err
		}

		client := createGitHubClient()
		release, _, err := client.Repositories.GetReleaseByTag(context.Background(), Config.Github.Org, Config.Github.Repository, version)
		if err != nil {
			return fmt.Errorf("failed to fetch release info: %w", err)
		}

		var downloadURL string
		for _, a := range release.Assets {
			if a.GetName() == assetName {
				downloadURL = a.GetBrowserDownloadURL()
				break
			}
		}
		if downloadURL == "" {
			return fmt.Errorf("could not find asset %s in release %s", assetName, version)
		}

		fmt.Printf("Downloading %s...\n", downloadURL)
		if err := downloadFile(downloadURL, binaryPath); err != nil {
			return fmt.Errorf("failed to download binary: %w", err)
		}
		if err := os.Chmod(binaryPath, 0o755); err != nil {
			return fmt.Errorf("failed to set executable permission: %w", err)
		}
		fmt.Printf("educates %s installed successfully.\n", version)
	}

	// Handle activation if requested
	if activate {
		if err := useVersion(version); err != nil {
			return fmt.Errorf("failed to set version %s as active: %w", version, err)
		}
		fmt.Printf("educates %s is now active.\n", version)
	}

	return nil
}

// --- Helper to install a version (reusing logic from install.go) ---
func installVersion(version string) error {
	return installAndOptionallyUse(version, false, false)
}

// --- Helper to set a version as active (reusing logic from use.go) ---
func useVersion(version string) error {
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
}
