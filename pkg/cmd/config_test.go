package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
)

func TestConfigViewPrintsTheConfiguration(t *testing.T) {
	env := clitest.New(t)
	env.WriteConfig("github:\n  org: acme\n")

	res := env.Run("config", "view")

	require.NoError(t, res.Err)
	assert.Regexp(t, `^Current configuration:\ngithub:\n    org: acme\n`, res.Stdout)
	assert.Empty(t, res.Stderr)
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

func TestConfigInitCreatesTheConfigFile(t *testing.T) {
	env := clitest.New(t)
	configPath := filepath.Join(env.ConfigDir(), "config.yaml")

	res := env.Run("config", "init")

	require.NoError(t, res.Err)
	assert.Equal(t, "Config file created at "+configPath+"\n", res.Stdout)
	assert.FileExists(t, configPath)
}
