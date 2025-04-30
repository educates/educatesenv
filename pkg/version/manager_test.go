package version

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/educates/educatesenv/pkg/github"
	"github.com/stretchr/testify/assert"
)

func setupTestManager(t *testing.T) (*Manager, string, func()) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "educatesenv-test")
	assert.NoError(t, err)

	// Create config
	cfg := &config.Config{
		Github: config.GithubConfig{
			Org:        "testorg",
			Repository: "testrepo",
			Token:      "testtoken",
		},
		Local: config.LocalConfig{
			Dir: tmpDir,
		},
		Development: config.DevelopmentConfig{
			Enabled:        false,
			BinaryLocation: "",
		},
	}

	// Create GitHub client
	gh := github.New(cfg)

	// Create manager
	manager := New(cfg, gh)

	cleanup := func() {
		err := os.RemoveAll(tmpDir)
		assert.NoError(t, err)
	}

	return manager, tmpDir, cleanup
}

func TestValidateDevelopmentMode(t *testing.T) {
	manager, tmpDir, cleanup := setupTestManager(t)
	defer cleanup()

	// Test with no symlink
	err := manager.ValidateDevelopmentMode()
	assert.NoError(t, err)

	// Test with regular symlink
	binaryPath := filepath.Join(tmpDir, "educates-v1.0.0")
	err = os.WriteFile(binaryPath, []byte("test binary"), 0755)
	assert.NoError(t, err)

	symlinkPath := filepath.Join(tmpDir, "educates")
	err = os.Symlink(binaryPath, symlinkPath)
	assert.NoError(t, err)

	err = manager.ValidateDevelopmentMode()
	assert.NoError(t, err)

	// Test with development symlink
	err = os.Remove(symlinkPath)
	assert.NoError(t, err)

	devBinaryPath := "/usr/local/bin/educates-dev"
	err = os.Symlink(devBinaryPath, symlinkPath)
	assert.NoError(t, err)

	err = manager.ValidateDevelopmentMode()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "development mode is disabled")
}

func TestUseVersion(t *testing.T) {
	manager, tmpDir, cleanup := setupTestManager(t)
	defer cleanup()

	// Create test binary
	version := "v1.0.0"
	binaryPath := filepath.Join(tmpDir, "educates-"+version)
	err := os.WriteFile(binaryPath, []byte("test binary"), 0755)
	assert.NoError(t, err)

	// Test using version
	err = manager.UseVersion(version)
	assert.NoError(t, err)

	// Verify symlink
	symlinkPath := filepath.Join(tmpDir, "educates")
	target, err := os.Readlink(symlinkPath)
	assert.NoError(t, err)
	assert.Equal(t, "educates-"+version, filepath.Base(target))

	// Test using development version when disabled
	err = manager.UseVersion("develop")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "development mode is not enabled")
}

// TODO: Implement proper mocking for GitHub client
// func TestInstallVersion(t *testing.T) {
// 	...
// }
