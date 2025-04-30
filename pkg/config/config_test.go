package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cfg := New()
	assert.NotNil(t, cfg)
	assert.Equal(t, DefaultGithubOrg, cfg.Github.Org)
	assert.Equal(t, DefaultGithubRepo, cfg.Github.Repository)
	assert.Empty(t, cfg.Github.Token)
	assert.False(t, cfg.Development.Enabled)
	assert.Empty(t, cfg.Development.BinaryLocation)

	// Test that Local.Dir is set to a path in the home directory
	home, err := os.UserHomeDir()
	if err == nil {
		expectedPath := filepath.Join(home, ConfigDirName, "bin")
		assert.Equal(t, expectedPath, cfg.Local.Dir)
	}
}

func TestLoad(t *testing.T) {
	// Setup temporary directory for test config
	tmpDir, err := os.MkdirTemp("", "educatesenv-test")
	assert.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tmpDir)
		assert.NoError(t, err)
	}()

	// Create a test config file
	configContent := []byte(`
github:
  org: testorg
  repository: testrepo
  token: testtoken
local:
  dir: /test/dir
development:
  enabled: true
  binaryLocation: /test/binary
`)
	err = os.WriteFile(filepath.Join(tmpDir, "config.yaml"), configContent, 0644)
	assert.NoError(t, err)

	// Set working directory to temp dir for config loading
	oldWd, err := os.Getwd()
	assert.NoError(t, err)
	err = os.Chdir(tmpDir)
	assert.NoError(t, err)
	defer func() {
		err := os.Chdir(oldWd)
		assert.NoError(t, err)
	}()

	// Test loading config from file
	cfg := New()
	err = cfg.Load()
	assert.NoError(t, err)
	assert.Equal(t, "testorg", cfg.Github.Org)
	assert.Equal(t, "testrepo", cfg.Github.Repository)
	assert.Equal(t, "testtoken", cfg.Github.Token)
	assert.Equal(t, "/test/dir", cfg.Local.Dir)
	assert.True(t, cfg.Development.Enabled)
	assert.Equal(t, "/test/binary", cfg.Development.BinaryLocation)
}

func TestLoadWithEnvVars(t *testing.T) {
	// Set environment variables
	envVars := map[string]string{
		"EDUCATES_GITHUB_ORG":                  "envorg",
		"EDUCATES_GITHUB_REPOSITORY":           "envrepo",
		"EDUCATES_GITHUB_TOKEN":                "envtoken",
		"EDUCATES_LOCAL_DIR":                   "/env/dir",
		"EDUCATES_DEVELOPMENT_ENABLED":         "true",
		"EDUCATES_DEVELOPMENT_BINARY_LOCATION": "/env/binary",
	}

	// Set environment variables
	for k, v := range envVars {
		err := os.Setenv(k, v)
		assert.NoError(t, err)
		defer func() {
			err := os.Unsetenv(k)
			assert.NoError(t, err)
		}()
	}

	// Test loading config with environment variables
	cfg := New()
	err := cfg.Load()
	assert.NoError(t, err)
	assert.Equal(t, "envorg", cfg.Github.Org)
	assert.Equal(t, "envrepo", cfg.Github.Repository)
	assert.Equal(t, "envtoken", cfg.Github.Token)
	assert.Equal(t, "/env/dir", cfg.Local.Dir)
	assert.True(t, cfg.Development.Enabled)
	assert.Equal(t, "/env/binary", cfg.Development.BinaryLocation)
}
