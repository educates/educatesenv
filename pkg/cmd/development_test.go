package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
)

// writeDevelopmentBinary writes an executable outside the bin directory,
// standing in for a locally built educates.
func writeDevelopmentBinary(t *testing.T, env *clitest.Env) string {
	t.Helper()
	path := filepath.Join(env.Home, "src", "educates", "educates")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("educates develop"), 0o755))
	return path
}

func TestDisabledDevelopmentModeRemovesTheDevelopmentSymlink(t *testing.T) {
	env := clitest.New(t)
	devBinary := writeDevelopmentBinary(t, env)
	require.NoError(t, os.MkdirAll(env.BinDir(), 0o755))
	require.NoError(t, os.Symlink(devBinary, env.ActiveLink()))

	res := env.Run("list")

	require.NoError(t, res.Err)
	assert.Equal(t,
		"Warning: development mode is disabled; removed symlink to development binary. Please use 'educatesenv use <version>' to select a version\n"+
			"Installed versions:\nNo versions installed\n",
		res.Stdout)
	assert.NoFileExists(t, env.ActiveLink())
	assert.FileExists(t, devBinary)
}

func TestEnabledDevelopmentModeKeepsTheDevelopmentSymlink(t *testing.T) {
	env := clitest.New(t)
	devBinary := writeDevelopmentBinary(t, env)
	env.WriteConfig("development:\n  enabled: true\n  binaryLocation: " + devBinary + "\n")
	require.NoError(t, os.MkdirAll(env.BinDir(), 0o755))
	require.NoError(t, os.Symlink(devBinary, env.ActiveLink()))

	res := env.Run("list")

	require.NoError(t, res.Err)
	assert.Equal(t, "Installed versions:\n* develop (active) -> "+devBinary+"\n", res.Stdout)
	target, err := os.Readlink(env.ActiveLink())
	require.NoError(t, err)
	assert.Equal(t, devBinary, target)
}
