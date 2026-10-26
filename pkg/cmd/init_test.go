package cmd_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

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

func TestInitDownloadInstallsTheHighestStableReleaseAfterARunOfPreReleases(t *testing.T) {
	env := clitest.New(t)
	// The twelve newest releases are release candidates of 3.8.0, as
	// during its release-candidate cycle.
	for rc := 12; rc >= 1; rc-- {
		addPublishedRelease(env, fmt.Sprintf("3.8.0-rc.%d", rc), published(time.August, 14+rc), true)
	}
	addPublishedRelease(env, "3.7.3", published(time.August, 1), false)
	addPublishedRelease(env, "3.7.2", published(time.July, 1), false)

	res := env.Run("init", "--download")

	require.NoError(t, res.Err)
	assert.Contains(t, res.Stdout, "Resolved latest to 3.7.3\nInstalling version 3.7.3...\n")
	assert.Contains(t, res.Stdout, "educates 3.7.3 is now active.\n")
	target, err := os.Readlink(env.ActiveLink())
	require.NoError(t, err)
	assert.Equal(t, "educates-3.7.3", target)
	content, err := os.ReadFile(filepath.Join(env.BinDir(), "educates-3.7.3"))
	require.NoError(t, err)
	assert.Equal(t, "educates 3.7.3 binary", string(content))
}

func TestInitDownloadFailsWhenThereIsNoStableRelease(t *testing.T) {
	env := clitest.New(t)
	addPublishedRelease(env, "3.8.0-rc.1", published(time.August, 14), true)

	res := env.Run("init", "--download")

	require.EqualError(t, res.Err, noStableRelease)
	assert.NoFileExists(t, env.ActiveLink())
}
