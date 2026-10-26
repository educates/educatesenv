package cmd_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
)

func TestUseActivatesAnInstalledVersion(t *testing.T) {
	env := clitest.New(t)
	env.WriteBinary("3.7.0")
	env.WriteBinary("3.8.0")
	require.NoError(t, os.Symlink("educates-3.7.0", env.ActiveLink()))

	res := env.Run("use", "3.8.0")

	require.NoError(t, res.Err)
	assert.Equal(t, "Now using educates version 3.8.0\n", res.Stdout)
	assert.Empty(t, res.Stderr)
	target, err := os.Readlink(env.ActiveLink())
	require.NoError(t, err)
	assert.Equal(t, "educates-3.8.0", target)
}

func TestUseOfAVersionThatIsNotInstalledSaysHowToInstallIt(t *testing.T) {
	env := clitest.New(t)
	env.WriteBinary("3.8.0")
	require.NoError(t, os.Symlink("educates-3.8.0", env.ActiveLink()))

	res := env.Run("use", "3.9.0")

	require.EqualError(t, res.Err, "version 3.9.0 is not installed. You should install it first with `educatesenv install 3.9.0`")
	assert.Empty(t, res.Stdout)
	target, err := os.Readlink(env.ActiveLink())
	require.NoError(t, err)
	assert.Equal(t, "educates-3.8.0", target)
}

func TestUseWithAnUnwritableBinDirectoryReportsTheCause(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions do not apply to root")
	}
	env := clitest.New(t)
	env.WriteBinary("3.7.0")
	env.WriteBinary("3.8.0")
	require.NoError(t, os.Symlink("educates-3.7.0", env.ActiveLink()))
	require.NoError(t, os.Chmod(env.BinDir(), 0o555))
	t.Cleanup(func() {
		_ = os.Chmod(env.BinDir(), 0o755)
	})

	res := env.Run("use", "3.8.0")

	require.ErrorIs(t, res.Err, fs.ErrPermission)
	assert.ErrorContains(t, res.Err, "failed to switch to version 3.8.0: ")
	assert.NotContains(t, res.Err.Error(), "not installed")
	assert.Empty(t, res.Stdout)
	target, err := os.Readlink(env.ActiveLink())
	require.NoError(t, err)
	assert.Equal(t, "educates-3.7.0", target)
}

func TestUseDevelopRequiresDevelopmentMode(t *testing.T) {
	env := clitest.New(t)

	res := env.Run("use", "develop")

	require.EqualError(t, res.Err, "development mode is not enabled. Enable it in the config file by setting development.enabled to true")
	assert.NoFileExists(t, env.ActiveLink())
}

func TestUseDevelopActivatesTheDevelopmentBinary(t *testing.T) {
	env := clitest.New(t)
	devBinary := writeDevelopmentBinary(t, env)
	env.WriteConfig("development:\n  enabled: true\n  binaryLocation: " + devBinary + "\n")
	require.NoError(t, os.MkdirAll(env.BinDir(), 0o755))

	res := env.Run("use", "develop")

	require.NoError(t, res.Err)
	assert.Equal(t, "Now using educates development version from "+devBinary+"\n", res.Stdout)
	target, err := os.Readlink(env.ActiveLink())
	require.NoError(t, err)
	assert.Equal(t, devBinary, filepath.Join(env.BinDir(), target))
}
