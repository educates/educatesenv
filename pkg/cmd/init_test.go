package cmd_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
)

// defaultConfig returns the config file that init and config init create.
func defaultConfig(env *clitest.Env) string {
	return "github:\n" +
		"    org: educates\n" +
		"    repository: educates-training-platform\n" +
		"local:\n" +
		"    dir: " + env.BinDir() + "\n" +
		"development:\n" +
		"    enabled: false\n" +
		"    binaryLocation: \"\"\n"
}

// assertDefaultConfig checks that the config file holds the default config
// and that only its owner can read it.
func assertDefaultConfig(t *testing.T, env *clitest.Env) {
	t.Helper()
	configPath := filepath.Join(env.ConfigDir(), "config.yaml")
	info, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, defaultConfig(env), string(content))
}

func TestInitWritesTheDefaultConfigReadableOnlyByItsOwner(t *testing.T) {
	env := clitest.New(t)

	res := env.Run("init")

	require.NoError(t, res.Err)
	assertDefaultConfig(t, env)
}

// pathInstructions returns the part of init's output that says how to add
// the bin directory to the PATH.
func pathInstructions(t *testing.T, stdout string) string {
	t.Helper()
	_, rest, found := strings.Cut(stdout, "To use educatesenv, add the bin directory to your PATH:\n")
	require.True(t, found, stdout)
	instructions, _, found := strings.Cut(rest, "\nRestart your terminal or source your profile to apply the changes.\n")
	require.True(t, found, stdout)
	return instructions
}

func TestInitGivesPathInstructionsForTheShell(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows gets PowerShell instructions")
	}

	tests := []struct {
		name  string
		shell string
		want  func(bin string) string
	}{
		{name: "zsh", shell: "/bin/zsh", want: func(bin string) string {
			return "\nFor zsh, run:\n  echo 'export PATH=\"" + bin + ":$PATH\"' >> ~/.zshrc\n  source ~/.zshrc\n"
		}},
		{name: "bash", shell: "/opt/homebrew/bin/bash", want: func(bin string) string {
			return "\nFor bash, run:\n  echo 'export PATH=\"" + bin + ":$PATH\"' >> ~/.bashrc\n  source ~/.bashrc\n"
		}},
		{name: "fish", shell: "/usr/local/bin/fish", want: func(bin string) string {
			return "\nFor fish, run:\n  set -U fish_user_paths " + bin + " $fish_user_paths\n"
		}},
		{name: "another shell", shell: "/bin/sh", want: func(bin string) string {
			return "\nAdd " + bin + " to your PATH in your shell's startup file.\n"
		}},
		{name: "SHELL not set", want: func(bin string) string {
			return "\nAdd " + bin + " to your PATH in your shell's startup file.\n"
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := clitest.New(t)
			if tt.shell != "" {
				env.Setenv("SHELL", tt.shell)
			}

			res := env.Run("init")

			require.NoError(t, res.Err)
			assert.Equal(t, tt.want(env.BinDir()), pathInstructions(t, res.Stdout))
		})
	}
}

func TestCreatingTheConfigWithoutAHomeDirectoryFails(t *testing.T) {
	for _, args := range [][]string{{"init"}, {"config", "init"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			env := clitest.New(t)
			env.Unsetenv("HOME")

			res := env.Run(args...)

			require.Error(t, res.Err)
			assert.Contains(t, res.Err.Error(), "could not determine home directory")
			assert.NoDirExists(t, filepath.Join(env.WorkDir, ".educatesenv"))
		})
	}
}

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
