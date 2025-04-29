package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	// DefaultGithubOrg is the default GitHub organization for educates
	DefaultGithubOrg = "educates"
	// DefaultGithubRepo is the default GitHub repository for educates
	DefaultGithubRepo = "educates-training-platform"
	// ConfigDirName is the name of the directory where educatesenv stores its configuration
	ConfigDirName = ".educatesenv"
)

// GithubConfig holds GitHub-related configuration
type GithubConfig struct {
	Org        string `yaml:"org"`
	Repository string `yaml:"repository"`
	Token      string `yaml:"token"`
}

// LocalConfig holds local directory configuration
type LocalConfig struct {
	Dir string `yaml:"dir"`
}

// DevelopmentConfig holds development mode configuration
type DevelopmentConfig struct {
	Enabled        bool   `yaml:"enabled"`
	BinaryLocation string `yaml:"binaryLocation"`
}

// Config holds all configuration for the CLI
type Config struct {
	Github      GithubConfig      `yaml:"github"`
	Local       LocalConfig       `yaml:"local"`
	Development DevelopmentConfig `yaml:"development"`
}

// New returns a new Config instance with defaults set
func New() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Warning: could not determine home directory, using current directory for bin folder.")
		home = "."
	}
	configDir := filepath.Join(home, ConfigDirName)
	defaultBin := filepath.Join(configDir, "bin")

	return &Config{
		Github: GithubConfig{
			Org:        DefaultGithubOrg,
			Repository: DefaultGithubRepo,
			Token:      "",
		},
		Local: LocalConfig{
			Dir: defaultBin,
		},
		Development: DevelopmentConfig{
			Enabled:        false,
			BinaryLocation: "",
		},
	}
}

// Load initializes the configuration from file and environment variables
func (c *Config) Load() error {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Warning: could not determine home directory, using current directory for bin folder.")
		home = "."
	}
	configDir := filepath.Join(home, ConfigDirName)
	defaultBin := filepath.Join(configDir, "bin")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath(configDir)

	// Set defaults
	viper.SetDefault("github.org", DefaultGithubOrg)
	viper.SetDefault("github.repository", DefaultGithubRepo)
	viper.SetDefault("github.token", "")
	viper.SetDefault("local.dir", defaultBin)
	viper.SetDefault("development.enabled", false)
	viper.SetDefault("development.binaryLocation", "")

	// Bind environment variables
	if err := viper.BindEnv("github.org", "EDUCATES_GITHUB_ORG"); err != nil {
		return fmt.Errorf("failed to bind env EDUCATES_GITHUB_ORG: %w", err)
	}
	if err := viper.BindEnv("github.repository", "EDUCATES_GITHUB_REPOSITORY"); err != nil {
		return fmt.Errorf("failed to bind env EDUCATES_GITHUB_REPOSITORY: %w", err)
	}
	if err := viper.BindEnv("github.token", "EDUCATES_GITHUB_TOKEN"); err != nil {
		return fmt.Errorf("failed to bind env EDUCATES_GITHUB_TOKEN: %w", err)
	}
	if err := viper.BindEnv("local.dir", "EDUCATES_LOCAL_DIR"); err != nil {
		return fmt.Errorf("failed to bind env EDUCATES_LOCAL_DIR: %w", err)
	}
	if err := viper.BindEnv("development.enabled", "EDUCATES_DEVELOPMENT_ENABLED"); err != nil {
		return fmt.Errorf("failed to bind env EDUCATES_DEVELOPMENT_ENABLED: %w", err)
	}
	if err := viper.BindEnv("development.binaryLocation", "EDUCATES_DEVELOPMENT_BINARY_LOCATION"); err != nil {
		return fmt.Errorf("failed to bind env EDUCATES_DEVELOPMENT_BINARY_LOCATION: %w", err)
	}

	// Read config file if present
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Map the configuration to our struct
	c.Github.Org = viper.GetString("github.org")
	c.Github.Repository = viper.GetString("github.repository")
	c.Github.Token = viper.GetString("github.token")
	c.Local.Dir = viper.GetString("local.dir")
	c.Development.Enabled = viper.GetBool("development.enabled")
	c.Development.BinaryLocation = viper.GetString("development.binaryLocation")

	return nil
}
