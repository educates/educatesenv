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

// GithubConfig holds GitHub-related configuration.
type GithubConfig struct {
	Org        string `yaml:"org"`
	Repository string `yaml:"repository"`
	// Token is github.token from the config file, the last and deprecated
	// source of the GitHub token. The token package resolves the token
	// from every source. An empty token is left out of the YAML, so the
	// default config file holds no token key.
	Token string `yaml:"token,omitempty"`
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

	file string
}

// File returns the path of the config file Load read the configuration
// from, whether or not the file exists. It is empty when Load could not
// determine the home directory.
func (c *Config) File() string {
	return c.file
}

// Paths are the places educatesenv keeps its files in.
type Paths struct {
	// Home is the user's home directory.
	Home string
	// ConfigDir is the educatesenv directory in Home.
	ConfigDir string
	// BinDir is the default directory installed versions go in.
	BinDir string
	// ConfigFile is the config file in ConfigDir.
	ConfigFile string
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
	configDir := filepath.Join(home, ConfigDirName)
	return Paths{
		Home:       home,
		ConfigDir:  configDir,
		BinDir:     filepath.Join(configDir, "bin"),
		ConfigFile: filepath.Join(configDir, ConfigFileName),
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
// empty string counts as not set. The GitHub token is not among them: the
// token package reads EDUCATES_GITHUB_TOKEN.
//
// A boolean value in none of the forms boolOf accepts reads as false, and
// Load writes a warning that names the value's key and file, or its
// environment variable, to errOut, instead of failing.
//
// Without a home directory, Load writes a warning to out, reads no config
// file, and puts the bin directory under the current directory.
func Load(out, errOut io.Writer) (*Config, error) {
	paths, err := homePaths()
	if err != nil {
		_, _ = fmt.Fprintln(out, "Warning: could not determine home directory, using current directory for bin folder.")
		cfg := defaults(pathsIn("."))
		cfg.applyEnv(errOut)
		return cfg, nil
	}

	cfg := defaults(paths)
	cfg.file = paths.ConfigFile
	if err := cfg.applyFile(paths.ConfigFile, errOut); err != nil {
		return nil, err
	}
	cfg.applyEnv(errOut)
	return cfg, nil
}

// setting is one value of the configuration.
type setting struct {
	// key is the value's key in the config file, as section.key. Keys
	// match regardless of case.
	key string
	// env is the environment variable that overrides the config file, or
	// "" for a value only the config file sets.
	env string
	// Exactly one of str and boolean points at the field the value sets.
	str     *string
	boolean *bool
}

func (c *Config) settings() []setting {
	return []setting{
		{key: "github.org", env: "EDUCATES_GITHUB_ORG", str: &c.Github.Org},
		{key: "github.repository", env: "EDUCATES_GITHUB_REPOSITORY", str: &c.Github.Repository},
		{key: "github.token", str: &c.Github.Token},
		{key: "local.dir", env: "EDUCATES_LOCAL_DIR", str: &c.Local.Dir},
		{key: "development.enabled", env: "EDUCATES_DEVELOPMENT_ENABLED", boolean: &c.Development.Enabled},
		{key: "development.binaryLocation", env: "EDUCATES_DEVELOPMENT_BINARY_LOCATION", str: &c.Development.BinaryLocation},
	}
}

// boolWarning is the warning for a boolean value that reads as false
// because boolOf does not accept it. Its verbs are where the value is set
// and what is wrong with it.
const boolWarning = "Warning: %s %s, so it is read as false. Set it to true or false.\n"

// notABoolean says that value is not a boolean, for boolWarning.
func notABoolean(value string) string {
	return fmt.Sprintf("is %q, which is not a boolean", value)
}

// applyEnv sets every value whose environment variable is set and not
// empty, and writes a warning to warnings for each boolean that reads as
// false because boolOf does not accept it.
func (c *Config) applyEnv(warnings io.Writer) {
	for _, s := range c.settings() {
		if s.env == "" {
			continue
		}
		value := os.Getenv(s.env)
		if value == "" {
			continue
		}
		if s.str != nil {
			*s.str = value
			continue
		}
		b, ok := boolOf(value)
		if !ok {
			_, _ = fmt.Fprintf(warnings, boolWarning, s.env, notABoolean(value))
		}
		*s.boolean = b
	}
}

// applyFile sets every value the config file at path sets, and writes a
// warning to warnings for each boolean that reads as false because boolOf
// does not accept it. A missing file sets nothing, and so does a key whose
// value is null.
func (c *Config) applyFile(path string, warnings io.Writer) error {
	values, err := readFile(path)
	if err != nil {
		return err
	}

	for _, s := range c.settings() {
		n, ok := values[strings.ToLower(s.key)]
		if !ok || n.ShortTag() == "!!null" {
			continue
		}
		if s.boolean != nil {
			b, ok := fileBool(n)
			if !ok {
				problem := "is not a single value"
				if n.Kind == yaml.ScalarNode {
					problem = notABoolean(n.Value)
				}
				_, _ = fmt.Fprintf(warnings, boolWarning, s.key+" in "+path, problem)
			}
			*s.boolean = b
			continue
		}
		if n.Kind != yaml.ScalarNode {
			return fmt.Errorf("%s in %s: expected a single value", s.key, path)
		}
		*s.str = n.Value
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

// boolOf reads value as a boolean: a bool as itself, a number as true when
// it is not zero, and a string in the forms strconv.ParseBool accepts: 1,
// t, T, TRUE, true, True, 0, f, F, FALSE, false and False. Any other value
// reads as false, and ok is false. These are the forms and the readings of
// viper's GetBool, so a config written for educatesenv 1.0 reads the same.
func boolOf(value any) (b, ok bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case int:
		return v != 0, true
	case int64:
		return v != 0, true
	case uint64:
		return v != 0, true
	case float64:
		return v != 0, true
	case string:
		b, err := strconv.ParseBool(v)
		return b, err == nil
	default:
		return false, false
	}
}

// fileBool reads a boolean from the config file by boolOf, from the value
// YAML decodes n to: a bool for true or false, a number for an integer or
// a float, and a string for any other single value, such as yes or on.
func fileBool(n *yaml.Node) (b, ok bool) {
	var value any
	if err := n.Decode(&value); err != nil {
		return false, false
	}
	return boolOf(value)
}

// CreateDefault creates the educatesenv directory in the user's home
// directory and its bin directory, and writes the default config file there
// unless a config file exists. The file is readable only by its owner and
// holds no token. CreateDefault returns the paths and whether it wrote the
// file, and fails when the home directory cannot be determined.
func CreateDefault() (Paths, bool, error) {
	paths, err := homePaths()
	if err != nil {
		return Paths{}, false, err
	}

	if err := os.MkdirAll(paths.BinDir, 0o755); err != nil {
		return paths, false, fmt.Errorf("failed to create bin directory: %w", err)
	}

	content, err := yaml.Marshal(defaults(paths))
	if err != nil {
		return paths, false, fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	f, err := os.OpenFile(paths.ConfigFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return paths, false, nil
	}
	if err != nil {
		return paths, false, fmt.Errorf("failed to write config file: %w", err)
	}
	_, err = f.Write(content)
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(paths.ConfigFile)
		return paths, false, fmt.Errorf("failed to write config file: %w", err)
	}
	return paths, true, nil
}
