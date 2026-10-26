package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
)

// defaultConfig returns the config file that init and config init create.
func defaultConfig(env *clitest.Env) string {
	return "github:\n" +
		"    org: educates\n" +
		"    repository: educates-training-platform\n" +
		"local:\n" +
		"    dir: " + env.BinDir() + "\n" +
		"development:\n" +
		"    enabled: false\n" +
		"    binaryLocation: \"\"\n"
}

// assertDefaultConfig checks that the config file holds the default config
// and that only its owner can read it.
func assertDefaultConfig(t *testing.T, env *clitest.Env) {
	t.Helper()
	configPath := filepath.Join(env.ConfigDir(), "config.yaml")
	info, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, defaultConfig(env), string(content))
}

func TestInitWritesTheDefaultConfigReadableOnlyByItsOwner(t *testing.T) {
	env := clitest.New(t)

	res := env.Run("init")

	require.NoError(t, res.Err)
	assertDefaultConfig(t, env)
}

func TestInitCreatesTheConfigAndBinDirectory(t *testing.T) {
	env := clitest.New(t)
	configPath := filepath.Join(env.ConfigDir(), "config.yaml")

	res := env.Run("init")

	require.NoError(t, res.Err)
	assert.Contains(t, res.Stdout, "Config file created at "+configPath+"\n")
	assert.Contains(t, res.Stdout, "Bin directory ensured at "+env.BinDir()+"\n")
	assert.Contains(t, res.Stdout, "Run 'educatesenv list-remote' to see available versions\n")
	assert.Empty(t, res.Stderr)
	assert.FileExists(t, configPath)
	assert.DirExists(t, env.BinDir())
}

func TestInitKeepsAnExistingConfig(t *testing.T) {
	env := clitest.New(t)
	configPath := env.WriteConfig("github:\n  org: acme\n")

	res := env.Run("init")

	require.NoError(t, res.Err)
	assert.Contains(t, res.Stdout, "Config file already exists at "+configPath+"\n")
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "github:\n  org: acme\n", string(content))
}
