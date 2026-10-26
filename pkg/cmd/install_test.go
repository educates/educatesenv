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

// addRelease adds a release with the platform binary and a checksums.txt
// that lists it, as every Educates release has.
func addRelease(env *clitest.Env, tag string, binary []byte) {
	asset := clitest.PlatformBinary(binary)
	env.GitHub.AddReleases(fakegithub.Release{
		Tag:    tag,
		Assets: []fakegithub.Asset{asset, fakegithub.Checksums(asset)},
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

// releasedSum is the SHA-256 checksum of "educates 3.8.0 binary".
const releasedSum = "21163080cb987ff30bc789604b421676c7065e47b4c77415af76459db9bfc4bd"

// addTamperedRelease adds a release whose binary is not the one its
// checksums.txt lists.
func addTamperedRelease(env *clitest.Env, tag string) {
	env.GitHub.AddReleases(fakegithub.Release{Tag: tag, Assets: []fakegithub.Asset{
		clitest.PlatformBinary([]byte("tampered binary")),
		{Name: "checksums.txt", Content: []byte(releasedSum + "  " + clitest.PlatformBinaryName() + "\n")},
	}})
}

func TestInstallFailsOnAChecksumMismatchAndLeavesNothingBehind(t *testing.T) {
	env := clitest.New(t)
	addTamperedRelease(env, "3.8.0")

	res := env.Run("install", "3.8.0", "--use")

	require.Error(t, res.Err)
	assert.Contains(t, res.Err.Error(), "failed to install version 3.8.0: checksum mismatch for "+clitest.PlatformBinaryName())
	assert.NotContains(t, res.Stdout, "installed successfully")
	assert.Empty(t, binDirEntries(t, env))
}

func TestInstallOverwriteWithAChecksumMismatchKeepsTheInstalledBinary(t *testing.T) {
	env := clitest.New(t)
	installed := env.WriteBinary("3.8.0")
	require.NoError(t, os.Symlink("educates-3.8.0", env.ActiveLink()))
	addTamperedRelease(env, "3.8.0")

	res := env.Run("install", "3.8.0", "--overwrite")

	require.Error(t, res.Err)
	assert.Contains(t, res.Err.Error(), "checksum mismatch for "+clitest.PlatformBinaryName())
	assert.Contains(t, res.Stdout, "Reinstalling version 3.8.0...\n")
	content, err := os.ReadFile(installed)
	require.NoError(t, err)
	assert.Equal(t, "educates 3.8.0", string(content))
	info, err := os.Stat(installed)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())
	assert.ElementsMatch(t, []string{"educates", "educates-3.8.0"}, binDirEntries(t, env))
}

func TestInstallFailsWhenChecksumsTxtHasNoEntryForTheBinary(t *testing.T) {
	env := clitest.New(t)
	env.GitHub.AddReleases(fakegithub.Release{Tag: "3.8.0", Assets: []fakegithub.Asset{
		clitest.PlatformBinary([]byte("educates 3.8.0 binary")),
		{Name: "checksums.txt", Content: []byte(releasedSum + "  educates-plan9-mips\n")},
	}})

	res := env.Run("install", "3.8.0")

	require.Error(t, res.Err)
	assert.Contains(t, res.Err.Error(), "failed to install version 3.8.0: checksums.txt has no entry for "+clitest.PlatformBinaryName())
	assert.Empty(t, binDirEntries(t, env))
}

func TestInstallWarnsAndContinuesWhenTheReleaseHasNoChecksumsTxt(t *testing.T) {
	env := clitest.New(t)
	env.GitHub.AddReleases(fakegithub.Release{Tag: "3.8.0", Assets: []fakegithub.Asset{
		clitest.PlatformBinary([]byte("educates 3.8.0 binary")),
	}})

	res := env.Run("install", "3.8.0")

	require.NoError(t, res.Err)
	assert.Contains(t, res.Stderr, "Warning: release 3.8.0 has no checksums.txt")
	assert.NotContains(t, res.Stdout, "Warning")
	assert.Contains(t, res.Stdout, "educates 3.8.0 installed successfully.\n")
	content, err := os.ReadFile(filepath.Join(env.BinDir(), "educates-3.8.0"))
	require.NoError(t, err)
	assert.Equal(t, "educates 3.8.0 binary", string(content))
	assert.Equal(t, []string{"educates-3.8.0"}, binDirEntries(t, env))
}
