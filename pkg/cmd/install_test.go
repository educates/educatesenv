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

func addRelease(env *clitest.Env, tag string, binary []byte) {
	env.GitHub.AddReleases(fakegithub.Release{
		Tag:    tag,
		Assets: []fakegithub.Asset{clitest.PlatformBinary(binary)},
	})
}

func TestInstallDownloadsANamedVersion(t *testing.T) {
	env := clitest.New(t)
	addRelease(env, "3.8.0", []byte("educates 3.8.0 binary"))

	res := env.Run("install", "3.8.0")

	require.NoError(t, res.Err)
	assert.Contains(t, res.Stdout, "Installing version 3.8.0...\n")
	assert.Contains(t, res.Stdout, "educates 3.8.0 installed successfully.\n")
	assert.Empty(t, res.Stderr)
	binary := filepath.Join(env.BinDir(), "educates-3.8.0")
	content, err := os.ReadFile(binary)
	require.NoError(t, err)
	assert.Equal(t, "educates 3.8.0 binary", string(content))
	info, err := os.Stat(binary)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())
	assert.NoFileExists(t, env.ActiveLink())
}

func TestInstallWithUseActivatesTheVersion(t *testing.T) {
	env := clitest.New(t)
	addRelease(env, "3.8.0", []byte("educates 3.8.0 binary"))

	res := env.Run("install", "3.8.0", "--use")

	require.NoError(t, res.Err)
	assert.Contains(t, res.Stdout, "educates 3.8.0 is now active.\n")
	target, err := os.Readlink(env.ActiveLink())
	require.NoError(t, err)
	assert.Equal(t, "educates-3.8.0", target)
}

func TestInstallOfAnUnknownVersionFails(t *testing.T) {
	env := clitest.New(t)
	addRelease(env, "3.8.0", []byte("educates 3.8.0 binary"))

	res := env.Run("install", "9.9.9")

	require.EqualError(t, res.Err, "failed to install version 9.9.9: version 9.9.9 not found. Run 'educatesenv list-remote' to see available versions")
	assert.NoFileExists(t, filepath.Join(env.BinDir(), "educates-9.9.9"))
}

// binDirEntries returns the names of the files in the bin directory.
func binDirEntries(t *testing.T, env *clitest.Env) []string {
	t.Helper()
	entries, err := os.ReadDir(env.BinDir())
	if os.IsNotExist(err) {
		return nil
	}
	require.NoError(t, err)
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestInstallCutMidTransferLeavesNothingBehindAndARetryInstalls(t *testing.T) {
	env := clitest.New(t)
	addRelease(env, "3.8.0", []byte("educates 3.8.0 binary"))
	env.GitHub.CutNextDownload("3.8.0", clitest.PlatformBinaryName())

	res := env.Run("install", "3.8.0")

	require.Error(t, res.Err)
	assert.Contains(t, res.Err.Error(), "failed to install version 3.8.0: failed to download binary")
	assert.Empty(t, binDirEntries(t, env))

	res = env.Run("install", "3.8.0")

	require.NoError(t, res.Err)
	assert.Contains(t, res.Stdout, "Installing version 3.8.0...\n")
	assert.Contains(t, res.Stdout, "educates 3.8.0 installed successfully.\n")
	content, err := os.ReadFile(filepath.Join(env.BinDir(), "educates-3.8.0"))
	require.NoError(t, err)
	assert.Equal(t, "educates 3.8.0 binary", string(content))
	assert.Equal(t, []string{"educates-3.8.0"}, binDirEntries(t, env))
}
