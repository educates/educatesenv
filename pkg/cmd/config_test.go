package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
	"github.com/educates/educatesenv/internal/fakegithub"
)

// configView runs `config view` and returns what it printed.
func configView(t *testing.T, env *clitest.Env) string {
	t.Helper()
	res := env.Run("config", "view")
	require.NoError(t, res.Err)
	return res.Stdout
}

func TestConfigViewPrintsTheConfiguration(t *testing.T) {
	env := clitest.New(t)
	env.WriteConfig("github:\n  org: acme\n")

	res := env.Run("config", "view")

	require.NoError(t, res.Err)
	assert.Regexp(t, `^Current configuration:\ngithub:\n    org: acme\n`, res.Stdout)
	assert.Empty(t, res.Stderr)
}

func TestConfigViewShowsTheDefaultsWithoutAConfigFile(t *testing.T) {
	env := clitest.New(t)

	out := configView(t, env)

	assert.Contains(t, out, "    org: educates\n")
	assert.Contains(t, out, "    repository: educates-training-platform\n")
	assert.Contains(t, out, "    dir: "+env.BinDir()+"\n")
	assert.Contains(t, out, "    enabled: false\n")
	assert.Contains(t, out, "    binaryLocation: \"\"\n")
}

func TestEachRunReadsTheConfigAfresh(t *testing.T) {
	env := clitest.New(t)
	configPath := env.WriteConfig("github:\n  org: acme\n")
	res := env.Run("config", "view")
	require.NoError(t, res.Err)
	require.Contains(t, res.Stdout, "org: acme\n")

	require.NoError(t, os.Remove(configPath))
	res = env.Run("config", "view")

	require.NoError(t, res.Err)
	assert.Contains(t, res.Stdout, "org: educates\n")
}

func TestAMissingHomeDirectoryWarnsOnce(t *testing.T) {
	env := clitest.New(t)
	env.Unsetenv("HOME")

	res := env.Run("version")

	require.NoError(t, res.Err)
	assert.Equal(t, "Warning: could not determine home directory, using current directory for bin folder.\ndevelop\n", res.Stdout)
}

func TestConfigKeysMatchRegardlessOfCase(t *testing.T) {
	env := clitest.New(t)
	env.WriteConfig(`GitHub:
  ORG: acme
  Repository: widgets
LOCAL:
  Dir: /opt/educates/bin
Development:
  ENABLED: true
  binarylocation: /src/educates/educates
`)

	out := configView(t, env)

	assert.Contains(t, out, "    org: acme\n")
	assert.Contains(t, out, "    repository: widgets\n")
	assert.Contains(t, out, "    dir: /opt/educates/bin\n")
	assert.Contains(t, out, "    enabled: true\n")
	assert.Contains(t, out, "    binaryLocation: /src/educates/educates\n")
}

func TestEnvironmentVariablesOverrideTheConfigFile(t *testing.T) {
	tests := []struct {
		variable string
		value    string
		fromFile string
		fromEnv  string
	}{
		{"EDUCATES_GITHUB_ORG", "envorg", "    org: fileorg\n", "    org: envorg\n"},
		{"EDUCATES_GITHUB_REPOSITORY", "envrepo", "    repository: filerepo\n", "    repository: envrepo\n"},
		{"EDUCATES_LOCAL_DIR", "/env/bin", "    dir: /file/bin\n", "    dir: /env/bin\n"},
		{"EDUCATES_DEVELOPMENT_ENABLED", "false", "    enabled: true\n", "    enabled: false\n"},
		{"EDUCATES_DEVELOPMENT_BINARY_LOCATION", "/env/educates", "    binaryLocation: /file/educates\n", "    binaryLocation: /env/educates\n"},
	}

	for _, tt := range tests {
		t.Run(tt.variable, func(t *testing.T) {
			env := clitest.New(t)
			env.WriteConfig(`github:
  org: fileorg
  repository: filerepo
local:
  dir: /file/bin
development:
  enabled: true
  binaryLocation: /file/educates
`)
			env.Setenv(tt.variable, tt.value)

			out := configView(t, env)

			assert.Contains(t, out, tt.fromEnv)
			assert.NotContains(t, out, tt.fromFile)
		})
	}
}

func TestAnEmptyEnvironmentVariableLeavesTheConfigFileValue(t *testing.T) {
	env := clitest.New(t)
	env.WriteConfig("github:\n  org: acme\n")
	env.Setenv("EDUCATES_GITHUB_ORG", "")

	out := configView(t, env)

	assert.Contains(t, out, "    org: acme\n")
}

func TestBooleansAcceptTheirUsualForms(t *testing.T) {
	trueForms := []string{"1", "t", "T", "true", "TRUE", "True"}
	falseForms := []string{"0", "f", "F", "false", "FALSE", "False"}

	for _, form := range trueForms {
		t.Run("EDUCATES_DEVELOPMENT_ENABLED="+form, func(t *testing.T) {
			env := clitest.New(t)
			env.Setenv("EDUCATES_DEVELOPMENT_ENABLED", form)
			assert.Contains(t, configView(t, env), "    enabled: true\n")
		})
	}
	for _, form := range falseForms {
		t.Run("EDUCATES_DEVELOPMENT_ENABLED="+form, func(t *testing.T) {
			env := clitest.New(t)
			env.WriteConfig("development:\n  enabled: true\n")
			env.Setenv("EDUCATES_DEVELOPMENT_ENABLED", form)
			assert.Contains(t, configView(t, env), "    enabled: false\n")
		})
	}

	fileForms := []struct {
		value string
		want  string
	}{
		{"true", "true"},
		{"True", "true"},
		{"TRUE", "true"},
		{"1", "true"},
		{"2", "true"},
		{`"t"`, "true"},
		{`"true"`, "true"},
		{"false", "false"},
		{"False", "false"},
		{"FALSE", "false"},
		{"0", "false"},
		{`"f"`, "false"},
		{`"false"`, "false"},
		{"", "false"},
	}
	for _, tt := range fileForms {
		t.Run("enabled: "+tt.value, func(t *testing.T) {
			env := clitest.New(t)
			env.WriteConfig("development:\n  enabled: " + tt.value + "\n")
			assert.Contains(t, configView(t, env), "    enabled: "+tt.want+"\n")
		})
	}
}

func TestRequestsCarryTheTokenFromTheEnvironmentOrTheConfigFile(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
		config string
		want   string
	}{
		{name: "EDUCATES_GITHUB_TOKEN", envVar: "envtoken", want: "Bearer envtoken"},
		{name: "github.token", config: "github:\n  token: filetoken\n", want: "Bearer filetoken"},
		{name: "EDUCATES_GITHUB_TOKEN over github.token", envVar: "envtoken", config: "github:\n  token: filetoken\n", want: "Bearer envtoken"},
		{name: "neither", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := clitest.New(t)
			env.GitHub.AddReleases(fakegithub.Release{Tag: "3.8.0"})
			if tt.config != "" {
				env.WriteConfig(tt.config)
			}
			if tt.envVar != "" {
				env.Setenv("EDUCATES_GITHUB_TOKEN", tt.envVar)
			}

			res := env.Run("list-remote")

			require.NoError(t, res.Err)
			requests := env.GitHub.Requests()
			require.NotEmpty(t, requests)
			for _, r := range requests {
				assert.Equal(t, tt.want, r.Header.Get("Authorization"), r.Path)
			}
		})
	}
}

func TestConfigInitCreatesTheConfigFile(t *testing.T) {
	env := clitest.New(t)
	configPath := filepath.Join(env.ConfigDir(), "config.yaml")

	res := env.Run("config", "init")

	require.NoError(t, res.Err)
	assert.Equal(t, "Config file created at "+configPath+"\n", res.Stdout)
	assert.FileExists(t, configPath)
}

func TestConfigInitRefusesToReplaceAnExistingConfig(t *testing.T) {
	env := clitest.New(t)
	configPath := env.WriteConfig("github:\n  org: acme\n")

	res := env.Run("config", "init")

	require.EqualError(t, res.Err, "config file already exists at "+configPath)
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "github:\n  org: acme\n", string(content))
}
