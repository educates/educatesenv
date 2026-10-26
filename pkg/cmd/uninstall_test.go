package cmd_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
)

func TestUninstallRemovesAVersion(t *testing.T) {
	env := clitest.New(t)
	removed := env.WriteBinary("3.7.0")
	kept := env.WriteBinary("3.8.0")
	require.NoError(t, os.Symlink("educates-3.8.0", env.ActiveLink()))

	res := env.Run("uninstall", "3.7.0")

	require.NoError(t, res.Err)
	assert.Equal(t, "educates 3.7.0 uninstalled.\n", res.Stdout)
	assert.Empty(t, res.Stderr)
	assert.NoFileExists(t, removed)
	assert.FileExists(t, kept)
	target, err := os.Readlink(env.ActiveLink())
	require.NoError(t, err)
	assert.Equal(t, "educates-3.8.0", target)
}

func TestUninstallRemovesSeveralVersions(t *testing.T) {
	env := clitest.New(t)
	first := env.WriteBinary("3.6.0")
	second := env.WriteBinary("3.7.0")
	kept := env.WriteBinary("3.8.0")
	require.NoError(t, os.Symlink("educates-3.8.0", env.ActiveLink()))

	res := env.Run("uninstall", "3.7.0", "3.6.0")

	require.NoError(t, res.Err)
	assert.Equal(t, "educates 3.7.0 uninstalled.\neducates 3.6.0 uninstalled.\n", res.Stdout)
	assert.Empty(t, res.Stderr)
	assert.NoFileExists(t, first)
	assert.NoFileExists(t, second)
	assert.FileExists(t, kept)
	target, err := os.Readlink(env.ActiveLink())
	require.NoError(t, err)
	assert.Equal(t, "educates-3.8.0", target)
}

func TestUninstallOfAVersionNamedTwiceRemovesItOnce(t *testing.T) {
	env := clitest.New(t)
	removed := env.WriteBinary("3.7.0")

	res := env.Run("uninstall", "3.7.0", "3.7.0")

	require.NoError(t, res.Err)
	assert.Equal(t, "educates 3.7.0 uninstalled.\n", res.Stdout)
	assert.NoFileExists(t, removed)
}

func TestUninstallRequiresAVersion(t *testing.T) {
	env := clitest.New(t)
	kept := env.WriteBinary("3.8.0")

	res := env.Run("uninstall")

	require.EqualError(t, res.Err, "requires at least 1 arg(s), only received 0")
	assert.Empty(t, res.Stdout)
	assert.FileExists(t, kept)
}

func TestUninstallOfAVersionThatIsNotInstalledFails(t *testing.T) {
	env := clitest.New(t)
	kept := env.WriteBinary("3.8.0")

	res := env.Run("uninstall", "3.9.9")

	require.EqualError(t, res.Err, "version 3.9.9 is not installed. Run 'educatesenv list' to see installed versions")
	assert.Empty(t, res.Stdout)
	assert.FileExists(t, kept)
}

func TestUninstallOfAPathIsNotAnInstalledVersion(t *testing.T) {
	env := clitest.New(t)
	env.WriteBinary("3.8.0")
	configPath := env.WriteConfig("github:\n  org: educates\n")

	// Joined to the bin directory, this argument names the config file.
	res := env.Run("uninstall", "x/../../config.yaml")

	require.EqualError(t, res.Err, "version x/../../config.yaml is not installed. Run 'educatesenv list' to see installed versions")
	assert.FileExists(t, configPath)
}

func TestUninstallWithATypoAmongValidVersionsRemovesNothing(t *testing.T) {
	env := clitest.New(t)
	first := env.WriteBinary("3.6.0")
	second := env.WriteBinary("3.7.0")

	res := env.Run("uninstall", "3.6.0", "3.7.O", "3.7.0")

	require.EqualError(t, res.Err, "version 3.7.O is not installed. Run 'educatesenv list' to see installed versions")
	assert.Empty(t, res.Stdout)
	assert.FileExists(t, first)
	assert.FileExists(t, second)
}

func TestUninstallReportsEveryVersionItCannotUninstall(t *testing.T) {
	env := clitest.New(t)
	valid := env.WriteBinary("3.6.0")
	active := env.WriteBinary("3.8.0")
	require.NoError(t, os.Symlink("educates-3.8.0", env.ActiveLink()))

	res := env.Run("uninstall", "3.6.0", "3.7.O", "develop", "3.8.0")

	require.EqualError(t, res.Err,
		"version 3.7.O is not installed. Run 'educatesenv list' to see installed versions\n"+
			"cannot uninstall develop: educatesenv does not manage the development binary\n"+
			"version 3.8.0 is active. Switch to another version with 'educatesenv use <version>' first, or pass --force to uninstall it")
	assert.Empty(t, res.Stdout)
	assert.FileExists(t, valid)
	assert.FileExists(t, active)
	target, err := os.Readlink(env.ActiveLink())
	require.NoError(t, err)
	assert.Equal(t, "educates-3.8.0", target)
}

func TestUninstallOfDevelopFails(t *testing.T) {
	env := clitest.New(t)
	devBinary := writeDevelopmentBinary(t, env)
	env.WriteConfig("development:\n  enabled: true\n  binaryLocation: " + devBinary + "\n")
	// An educates-develop file in the bin directory does not make develop
	// an installed version.
	shadow := env.WriteBinary("develop")
	require.NoError(t, os.Symlink(devBinary, env.ActiveLink()))

	res := env.Run("uninstall", "develop", "--force")

	require.EqualError(t, res.Err, "cannot uninstall develop: educatesenv does not manage the development binary")
	assert.Empty(t, res.Stdout)
	assert.FileExists(t, devBinary)
	assert.FileExists(t, shadow)
	target, err := os.Readlink(env.ActiveLink())
	require.NoError(t, err)
	assert.Equal(t, devBinary, target)
}

func TestUninstallOfTheActiveVersionFailsWithoutForce(t *testing.T) {
	env := clitest.New(t)
	other := env.WriteBinary("3.7.0")
	active := env.WriteBinary("3.8.0")
	require.NoError(t, os.Symlink("educates-3.8.0", env.ActiveLink()))

	res := env.Run("uninstall", "3.7.0", "3.8.0")

	require.EqualError(t, res.Err, "version 3.8.0 is active. Switch to another version with 'educatesenv use <version>' first, or pass --force to uninstall it")
	assert.Empty(t, res.Stdout)
	assert.FileExists(t, other)
	assert.FileExists(t, active)
	target, err := os.Readlink(env.ActiveLink())
	require.NoError(t, err)
	assert.Equal(t, "educates-3.8.0", target)
}

func TestUninstallOfTheActiveVersionWithForceLeavesNoVersionActive(t *testing.T) {
	env := clitest.New(t)
	kept := env.WriteBinary("3.7.0")
	active := env.WriteBinary("3.8.0")
	require.NoError(t, os.Symlink("educates-3.8.0", env.ActiveLink()))

	res := env.Run("uninstall", "3.8.0", "--force")

	require.NoError(t, res.Err)
	assert.Equal(t,
		"educates 3.8.0 uninstalled.\n"+
			"No educates version is active now. Run 'educatesenv use <version>' to select one.\n",
		res.Stdout)
	assert.Empty(t, res.Stderr)
	assert.NoFileExists(t, active)
	assert.FileExists(t, kept)
	_, err := os.Lstat(env.ActiveLink())
	assert.ErrorIs(t, err, os.ErrNotExist)
}
