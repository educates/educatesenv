// Package config loads educatesenv's configuration from the config file in
// the user's educatesenv directory and from EDUCATES_* environment
// variables.
package config

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	// DefaultGithubOrg is the default GitHub organization for educates
	DefaultGithubOrg = "educates"
	// DefaultGithubRepo is the default GitHub repository for educates
	DefaultGithubRepo = "educates-training-platform"
	// ConfigDirName is the name of the directory where educatesenv stores its configuration
	ConfigDirName = ".educatesenv"
	// ConfigFileName is the name of the config file in the educatesenv directory
	ConfigFileName = "config.yaml"
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

// Paths are the places educatesenv keeps its files in.
type Paths struct {
	// Home is the user's home directory.
	Home string
	// Dir is the educatesenv directory in Home.
	Dir string
	// BinDir is the default directory installed versions go in.
	BinDir string
	// File is the config file in Dir.
	File string
}

// homePaths returns the paths in the user's home directory.
func homePaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, fmt.Errorf("could not determine home directory: %w", err)
	}
	return pathsIn(home), nil
}

// pathsIn returns the paths in home.
func pathsIn(home string) Paths {
	dir := filepath.Join(home, ConfigDirName)
	return Paths{
		Home:   home,
		Dir:    dir,
		BinDir: filepath.Join(dir, "bin"),
		File:   filepath.Join(dir, ConfigFileName),
	}
}

// defaults returns the configuration that applies when neither the config
// file nor the environment sets a value.
func defaults(paths Paths) *Config {
	return &Config{
		Github: GithubConfig{
			Org:        DefaultGithubOrg,
			Repository: DefaultGithubRepo,
		},
		Local: LocalConfig{
			Dir: paths.BinDir,
		},
	}
}

// Load returns the configuration: the defaults, overridden by the config
// file in the user's educatesenv directory, overridden in turn by the
// EDUCATES_* environment variables. An environment variable set to the
// empty string counts as not set.
//
// Without a home directory, Load writes a warning to out, reads no config
// file, and puts the bin directory under the current directory.
func Load(out io.Writer) (*Config, error) {
	paths, err := homePaths()
	if err != nil {
		_, _ = fmt.Fprintln(out, "Warning: could not determine home directory, using current directory for bin folder.")
		cfg := defaults(pathsIn("."))
		if err := cfg.applyEnv(); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	cfg := defaults(paths)
	if err := cfg.applyFile(paths.File); err != nil {
		return nil, err
	}
	if err := cfg.applyEnv(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// setting is one value of the configuration.
type setting struct {
	// key is the value's key in the config file, as section.key. Keys
	// match regardless of case.
	key string
	// env is the environment variable that overrides the config file.
	env string
	// Exactly one of str and boolean points at the field the value sets.
	str     *string
	boolean *bool
}

func (c *Config) settings() []setting {
	return []setting{
		{key: "github.org", env: "EDUCATES_GITHUB_ORG", str: &c.Github.Org},
		{key: "github.repository", env: "EDUCATES_GITHUB_REPOSITORY", str: &c.Github.Repository},
		{key: "github.token", env: "EDUCATES_GITHUB_TOKEN", str: &c.Github.Token},
		{key: "local.dir", env: "EDUCATES_LOCAL_DIR", str: &c.Local.Dir},
		{key: "development.enabled", env: "EDUCATES_DEVELOPMENT_ENABLED", boolean: &c.Development.Enabled},
		{key: "development.binaryLocation", env: "EDUCATES_DEVELOPMENT_BINARY_LOCATION", str: &c.Development.BinaryLocation},
	}
}

// applyEnv sets every value whose environment variable is set and not
// empty.
func (c *Config) applyEnv() error {
	for _, s := range c.settings() {
		value := os.Getenv(s.env)
		if value == "" {
			continue
		}
		if s.str != nil {
			*s.str = value
			continue
		}
		b, err := parseBool(value)
		if err != nil {
			return fmt.Errorf("%s: %w", s.env, err)
		}
		*s.boolean = b
	}
	return nil
}

// applyFile sets every value the config file at path sets. A missing file
// sets nothing, and so does a key whose value is null.
func (c *Config) applyFile(path string) error {
	values, err := readFile(path)
	if err != nil {
		return err
	}

	for _, s := range c.settings() {
		n, ok := values[strings.ToLower(s.key)]
		if !ok || n.ShortTag() == "!!null" {
			continue
		}
		if n.Kind != yaml.ScalarNode {
			return fmt.Errorf("%s in %s: expected a single value", s.key, path)
		}
		if s.str != nil {
			*s.str = n.Value
			continue
		}
		b, err := fileBool(n)
		if err != nil {
			return fmt.Errorf("%s in %s: %w", s.key, path, err)
		}
		*s.boolean = b
	}
	return nil
}

// readFile returns the values in the config file at path, by their key as
// section.key in lower case. When a key appears more than once in
// different cases, the last one wins.
func readFile(path string) (map[string]*yaml.Node, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", path, err)
	}
	if len(doc.Content) == 0 {
		return nil, nil
	}

	root := resolve(doc.Content[0])
	if root.ShortTag() == "!!null" {
		return nil, nil
	}
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("error reading config file %s: expected a mapping of sections", path)
	}

	values := map[string]*yaml.Node{}
	for i := 0; i+1 < len(root.Content); i += 2 {
		section, body := root.Content[i].Value, resolve(root.Content[i+1])
		if body.Kind != yaml.MappingNode {
			continue
		}
		for j := 0; j+1 < len(body.Content); j += 2 {
			key := strings.ToLower(section + "." + body.Content[j].Value)
			values[key] = resolve(body.Content[j+1])
		}
	}
	return values, nil
}

// resolve returns the node an alias points at, or n itself.
func resolve(n *yaml.Node) *yaml.Node {
	for n.Kind == yaml.AliasNode {
		n = n.Alias
	}
	return n
}

// parseBool parses the boolean forms strconv.ParseBool accepts: 1, t, T,
// TRUE, true, True, 0, f, F, FALSE, false and False.
func parseBool(value string) (bool, error) {
	b, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%q is not a boolean; use true or false", value)
	}
	return b, nil
}

// fileBool parses a boolean from the config file. A number is true when it
// is not zero, and anything else takes the forms parseBool accepts.
func fileBool(n *yaml.Node) (bool, error) {
	switch n.ShortTag() {
	case "!!int", "!!float":
		var f float64
		if err := n.Decode(&f); err != nil {
			return false, fmt.Errorf("%q is not a boolean; use true or false", n.Value)
		}
		return f != 0, nil
	default:
		return parseBool(n.Value)
	}
}

// CreateConfigAndFolders ensures the config and bin directories exist, and creates a default config.yaml if not present.
// Returns (configDir, binDir, configPath, configCreated, error)
func CreateConfigAndFolders() (string, string, string, bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	configDir := filepath.Join(home, ConfigDirName)
	binDir := filepath.Join(configDir, "bin")
	configPath := filepath.Join(configDir, "config.yaml")

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return configDir, binDir, configPath, false, fmt.Errorf("failed to create bin directory: %w", err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return configDir, binDir, configPath, false, fmt.Errorf("failed to create config directory: %w", err)
	}

	configCreated := false
	if _, err := os.Stat(configPath); err == nil {
		// Config file already exists
		configCreated = false
	} else {
		// Create new config with defaults
		configFile := defaults(pathsIn(home))
		yamlBytes, err := yaml.Marshal(&configFile)
		if err != nil {
			return configDir, binDir, configPath, false, fmt.Errorf("failed to marshal config to YAML: %w", err)
		}
		if err := os.WriteFile(configPath, yamlBytes, 0o644); err != nil {
			return configDir, binDir, configPath, false, fmt.Errorf("failed to write config file: %w", err)
		}
		configCreated = true
	}
	return configDir, binDir, configPath, configCreated, nil
}
