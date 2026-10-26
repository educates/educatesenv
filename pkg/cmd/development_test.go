package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
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

// linkDevelopmentBinary writes a development binary and points the
// educates symlink at it, and returns the binary's path.
func linkDevelopmentBinary(t *testing.T, env *clitest.Env) string {
	t.Helper()
	devBinary := writeDevelopmentBinary(t, env)
	require.NoError(t, os.MkdirAll(env.BinDir(), 0o755))
	require.NoError(t, os.Symlink(devBinary, env.ActiveLink()))
	return devBinary
}

func TestCommandsThatDealWithTheActiveBinaryRemoveTheDevelopmentSymlink(t *testing.T) {
	tests := [][]string{
		{"list"},
		{"list-local"},
		{"use", "3.8.0"},
		{"install", "3.8.0"},
		{"init"},
	}

	for _, command := range tests {
		t.Run(strings.Join(command, " "), func(t *testing.T) {
			env := clitest.New(t)
			env.WriteBinary("3.8.0")
			devBinary := linkDevelopmentBinary(t, env)

			res := env.Run(command...)

			require.NoError(t, res.Err)
			assert.Contains(t, res.Stdout, "Warning: development mode is disabled; removed symlink to development binary.")
			target, _ := os.Readlink(env.ActiveLink())
			assert.NotEqual(t, devBinary, target)
			assert.FileExists(t, devBinary)
		})
	}
}

func TestOtherCommandsKeepTheDevelopmentSymlink(t *testing.T) {
	tests := [][]string{
		{"version"},
		{"config", "view"},
		{"config", "init"},
		{"list-remote"},
	}

	for _, command := range tests {
		t.Run(strings.Join(command, " "), func(t *testing.T) {
			env := clitest.New(t)
			devBinary := linkDevelopmentBinary(t, env)

			res := env.Run(command...)

			require.NoError(t, res.Err)
			assert.NotContains(t, res.Stdout, "development mode is disabled")
			target, err := os.Readlink(env.ActiveLink())
			require.NoError(t, err)
			assert.Equal(t, devBinary, target)
		})
	}
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
