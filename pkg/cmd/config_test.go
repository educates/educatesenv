package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
)

// configView runs `config view` and returns what it printed.
func configView(t *testing.T, env *clitest.Env) string {
	t.Helper()
	res := env.Run("config", "view")
	require.NoError(t, res.Err)
	return res.Stdout
}

func TestConfigWithoutASubcommandPrintsItsHelp(t *testing.T) {
	env := clitest.New(t)

	res := env.Run("config")

	require.NoError(t, res.Err)
	assert.Contains(t, res.Stdout, "Available Commands:\n")
	assert.Contains(t, res.Stdout, "  init ")
	assert.Contains(t, res.Stdout, "  view ")
	assert.NoDirExists(t, env.ConfigDir())
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

func TestAConfigFileWithoutValuesLeavesTheDefaults(t *testing.T) {
	for name, content := range map[string]string{
		"empty":            "",
		"comments only":    "# educatesenv\n",
		"an empty section": "github:\n",
	} {
		t.Run(name, func(t *testing.T) {
			env := clitest.New(t)
			env.WriteConfig(content)

			out := configView(t, env)

			assert.Contains(t, out, "    org: educates\n")
			assert.Contains(t, out, "    dir: "+env.BinDir()+"\n")
		})
	}
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

func TestAConfigFileInTheWorkingDirectoryIsIgnored(t *testing.T) {
	const workDirConfig = "github:\n  org: elsewhere\ndevelopment:\n  enabled: true\n  binaryLocation: /tmp/educates\n"

	t.Run("with a config in the home directory", func(t *testing.T) {
		env := clitest.New(t)
		env.WriteConfig("github:\n  org: acme\n")
		require.NoError(t, os.WriteFile(filepath.Join(env.WorkDir, "config.yaml"), []byte(workDirConfig), 0o600))

		out := configView(t, env)

		assert.Contains(t, out, "    org: acme\n")
		assert.Contains(t, out, "    enabled: false\n")
		assert.NotContains(t, out, "/tmp/educates")
	})

	t.Run("without a config in the home directory", func(t *testing.T) {
		env := clitest.New(t)
		require.NoError(t, os.WriteFile(filepath.Join(env.WorkDir, "config.yaml"), []byte(workDirConfig), 0o600))

		out := configView(t, env)

		assert.Contains(t, out, "    org: educates\n")
		assert.Contains(t, out, "    enabled: false\n")
		assert.NotContains(t, out, "/tmp/educates")
	})

	t.Run("without a home directory", func(t *testing.T) {
		env := clitest.New(t)
		env.Unsetenv("HOME")
		require.NoError(t, os.WriteFile(filepath.Join(env.WorkDir, "config.yaml"), []byte(workDirConfig), 0o600))
		require.NoError(t, os.MkdirAll(filepath.Join(env.WorkDir, ".educatesenv"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(env.WorkDir, ".educatesenv", "config.yaml"), []byte(workDirConfig), 0o600))

		out := configView(t, env)

		assert.Contains(t, out, "    org: educates\n")
		assert.Contains(t, out, "    dir: "+filepath.Join(".educatesenv", "bin")+"\n")
		assert.Contains(t, out, "    enabled: false\n")
		assert.NotContains(t, out, "/tmp/educates")
	})
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

func TestBooleansInTheirUsualFormsGiveNoWarning(t *testing.T) {
	env := clitest.New(t)
	env.WriteConfig("development:\n  enabled: True\n")
	env.Setenv("EDUCATES_DEVELOPMENT_ENABLED", "1")

	res := env.Run("config", "view")

	require.NoError(t, res.Err)
	assert.Contains(t, res.Stdout, "    enabled: true\n")
	assert.Empty(t, res.Stderr)
}

func TestABooleanInTheEnvironmentInAnotherFormIsReadAsFalseWithAWarning(t *testing.T) {
	for _, value := range []string{"yes", "on", "y", "2", " true"} {
		t.Run(value, func(t *testing.T) {
			env := clitest.New(t)
			env.WriteConfig("development:\n  enabled: true\n")
			env.Setenv("EDUCATES_DEVELOPMENT_ENABLED", value)
			warning := `Warning: EDUCATES_DEVELOPMENT_ENABLED is "` + value + `", which is not a boolean, so it is read as false. Set it to true or false.` + "\n"

			res := env.Run("version")

			require.NoError(t, res.Err)
			assert.Equal(t, "develop\n", res.Stdout)
			assert.Equal(t, warning, res.Stderr)

			res = env.Run("config", "view")

			require.NoError(t, res.Err)
			assert.Contains(t, res.Stdout, "    enabled: false\n")
			assert.Equal(t, warning, res.Stderr)
		})
	}
}

func TestABooleanInTheConfigFileInAnotherFormIsReadAsFalseWithAWarning(t *testing.T) {
	tests := []struct {
		value string
		shown string
	}{
		{"yes", `is "yes", which is not a boolean`},
		{"on", `is "on", which is not a boolean`},
		{"Y", `is "Y", which is not a boolean`},
		{`"yes"`, `is "yes", which is not a boolean`},
		{`""`, `is "", which is not a boolean`},
		{"2026-10-26", `is "2026-10-26", which is not a boolean`},
		{"[true]", "is not a single value"},
		{"{on: true}", "is not a single value"},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			env := clitest.New(t)
			configPath := env.WriteConfig("development:\n  enabled: " + tt.value + "\n")
			warning := "Warning: development.enabled in " + configPath + " " + tt.shown + ", so it is read as false. Set it to true or false.\n"

			res := env.Run("version")

			require.NoError(t, res.Err)
			assert.Equal(t, "develop\n", res.Stdout)
			assert.Equal(t, warning, res.Stderr)

			res = env.Run("config", "view")

			require.NoError(t, res.Err)
			assert.Contains(t, res.Stdout, "    enabled: false\n")
			assert.Equal(t, warning, res.Stderr)
		})
	}
}

func TestAMalformedValueIsAnError(t *testing.T) {
	t.Run("a list in the config file", func(t *testing.T) {
		env := clitest.New(t)
		configPath := env.WriteConfig("github:\n  org: [acme, widgets]\n")

		res := env.Run("config", "view")

		require.EqualError(t, res.Err, `github.org in `+configPath+`: expected a single value`)
	})

	t.Run("YAML that does not parse", func(t *testing.T) {
		env := clitest.New(t)
		configPath := env.WriteConfig("github:\n  org: acme\n bad indentation\n")

		res := env.Run("config", "view")

		require.Error(t, res.Err)
		assert.Contains(t, res.Err.Error(), "error reading config file "+configPath+": ")
	})
}

func TestConfigInitCreatesTheSameConfigAsInit(t *testing.T) {
	env := clitest.New(t)
	configPath := filepath.Join(env.ConfigDir(), "config.yaml")

	res := env.Run("config", "init")

	require.NoError(t, res.Err)
	assert.Equal(t, "Config file created at "+configPath+"\n", res.Stdout)
	assertDefaultConfig(t, env)
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
