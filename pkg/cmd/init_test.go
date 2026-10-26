package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
)

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
