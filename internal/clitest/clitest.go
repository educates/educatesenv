// Package clitest runs educatesenv commands in-process, isolated from the
// machine they run on.
//
// New gives each test a temporary home directory and working directory, a
// PATH that holds only an optional fake gh script, an environment without
// the variables educatesenv or gh read, and a fake GitHub server. Run then
// builds a fresh command tree pointed at that server, runs it with the
// given arguments, and returns what it wrote and the error it returned.
//
// The environment is changed with t.Setenv, so tests that use an Env cannot
// run in parallel.
package clitest

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/educates/educatesenv/internal/fakegithub"
	"github.com/educates/educatesenv/pkg/cmd"
)

// scrubbedPrefixes and scrubbedVariables name the environment variables
// New removes, so that the developer's own settings and tokens never reach
// a test.
var (
	scrubbedPrefixes  = []string{"EDUCATES_"}
	scrubbedVariables = []string{
		"GH_TOKEN",
		"GITHUB_TOKEN",
		"GH_ENTERPRISE_TOKEN",
		"GITHUB_ENTERPRISE_TOKEN",
		"GH_HOST",
		"GH_CONFIG_DIR",
		"XDG_CONFIG_HOME",
		"SHELL",
	}
)

// Env is an isolated environment for running educatesenv commands.
type Env struct {
	t *testing.T

	// Home is the temporary home directory, set as HOME.
	Home string
	// WorkDir is the temporary working directory the commands run in.
	WorkDir string
	// PathDir is the only directory on PATH. It holds nothing until FakeGH
	// writes a gh script into it.
	PathDir string
	// GitHub is the fake GitHub server the commands call.
	GitHub *fakegithub.Server
}

// Result is what one run of a command tree produced.
type Result struct {
	Stdout string
	Stderr string
	Err    error
}

// New sets up an isolated environment for the test.
func New(t *testing.T) *Env {
	t.Helper()

	e := &Env{
		t:       t,
		Home:    t.TempDir(),
		WorkDir: t.TempDir(),
		PathDir: t.TempDir(),
		GitHub:  fakegithub.New(t),
	}

	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		for _, prefix := range scrubbedPrefixes {
			if strings.HasPrefix(key, prefix) {
				e.Unsetenv(key)
			}
		}
	}
	for _, key := range scrubbedVariables {
		e.Unsetenv(key)
	}

	t.Setenv("HOME", e.Home)
	t.Setenv("PATH", e.PathDir)
	t.Chdir(e.WorkDir)
	return e
}

// Run builds a fresh command tree and runs it as `educatesenv <args>`.
func (e *Env) Run(args ...string) Result {
	e.t.Helper()

	root := cmd.NewRootCommand(cmd.WithGitHubBaseURL(e.GitHub.URL))
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetIn(strings.NewReader(""))
	// cobra reads os.Args when the arguments are nil.
	root.SetArgs(append([]string{}, args...))

	err := root.Execute()
	return Result{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
}

// Setenv sets an environment variable for the rest of the test.
func (e *Env) Setenv(key, value string) {
	e.t.Helper()
	e.t.Setenv(key, value)
}

// Unsetenv removes an environment variable for the rest of the test.
func (e *Env) Unsetenv(key string) {
	e.t.Helper()
	// t.Setenv restores the original value when the test ends.
	e.t.Setenv(key, "")
	if err := os.Unsetenv(key); err != nil {
		e.t.Fatalf("unset %s: %v", key, err)
	}
}

// FakeGH writes an executable gh script into PathDir, so commands find it
// on PATH. script is the body of a /bin/sh script. PATH holds only
// PathDir, so the script can use shell builtins and absolute paths only.
func (e *Env) FakeGH(script string) {
	e.t.Helper()
	path := filepath.Join(e.PathDir, "gh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script+"\n"), 0o755); err != nil {
		e.t.Fatalf("write fake gh: %v", err)
	}
}

// ConfigDir returns the educatesenv directory in the home directory.
func (e *Env) ConfigDir() string {
	return filepath.Join(e.Home, ".educatesenv")
}

// BinDir returns the default bin directory, where versions are installed.
func (e *Env) BinDir() string {
	return filepath.Join(e.ConfigDir(), "bin")
}

// WriteConfig writes content as the config file in ConfigDir and returns
// its path.
func (e *Env) WriteConfig(content string) string {
	e.t.Helper()
	path := filepath.Join(e.ConfigDir(), "config.yaml")
	e.writeFile(path, []byte(content), 0o600)
	return path
}

// WriteBinary writes an executable that stands in for an installed
// version into BinDir, and returns its path. Its content is
// "educates <version>".
func (e *Env) WriteBinary(version string) string {
	e.t.Helper()
	path := filepath.Join(e.BinDir(), "educates-"+version)
	e.writeFile(path, []byte("educates "+version), 0o755)
	return path
}

// ActiveLink returns the path of the educates symlink in BinDir, which
// points at the active version.
func (e *Env) ActiveLink() string {
	return filepath.Join(e.BinDir(), "educates")
}

// PlatformBinaryName returns the name of the release asset that holds the
// educates binary for the platform the tests run on.
func PlatformBinaryName() string {
	return "educates-" + runtime.GOOS + "-" + runtime.GOARCH
}

// PlatformBinary returns a release asset holding content as the educates
// binary for the platform the tests run on.
func PlatformBinary(content []byte) fakegithub.Asset {
	return fakegithub.Asset{Name: PlatformBinaryName(), Content: content}
}

func (e *Env) writeFile(path string, content []byte, perm os.FileMode) {
	e.t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		e.t.Fatalf("create %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, content, perm); err != nil {
		e.t.Fatalf("write %s: %v", path, err)
	}
}
