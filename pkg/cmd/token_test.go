package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
	"github.com/educates/educatesenv/internal/fakegithub"
)

// The tokens each source holds in these tests.
const (
	educatesEnvToken = "educates-env-token"
	ghEnvToken       = "gh-env-token"
	githubEnvToken   = "github-env-token"
	ghCLIToken       = "gh-cli-token"
	fileToken        = "file-token"
)

// tokenSource sets up one source of the GitHub token in env.
type tokenSource struct {
	name  string
	token string
	set   func(env *clitest.Env)
}

// tokenSources are the sources of the GitHub token, in the order
// educatesenv checks them.
var tokenSources = []tokenSource{
	{name: "EDUCATES_GITHUB_TOKEN", token: educatesEnvToken, set: func(env *clitest.Env) {
		env.Setenv("EDUCATES_GITHUB_TOKEN", educatesEnvToken)
	}},
	{name: "GH_TOKEN", token: ghEnvToken, set: func(env *clitest.Env) {
		env.Setenv("GH_TOKEN", ghEnvToken)
	}},
	{name: "GITHUB_TOKEN", token: githubEnvToken, set: func(env *clitest.Env) {
		env.Setenv("GITHUB_TOKEN", githubEnvToken)
	}},
	{name: "gh auth token", token: ghCLIToken, set: func(env *clitest.Env) {
		fakeGHWithToken(env, ghCLIToken)
	}},
	{name: "config file", token: fileToken, set: func(env *clitest.Env) {
		env.WriteConfig("github:\n  token: " + fileToken + "\n")
	}},
}

// fakeGHWithToken puts a gh on the PATH that prints token for
// `gh auth token --hostname github.com` and fails for anything else.
func fakeGHWithToken(env *clitest.Env, token string) {
	env.FakeGH(`if [ "$*" = "auth token --hostname github.com" ]; then printf '%s\n' '` + token + `'; exit 0; fi
exit 1`)
}

// authorizations returns the Authorization header of every request the
// fake GitHub server received, failing the test when it received none.
func authorizations(t *testing.T, env *clitest.Env) []string {
	t.Helper()
	requests := env.GitHub.Requests()
	require.NotEmpty(t, requests)
	var headers []string
	for _, r := range requests {
		headers = append(headers, r.Header.Get("Authorization"))
	}
	return headers
}

func TestTheTokenComesFromTheFirstSourceThatHasOne(t *testing.T) {
	for i, winner := range tokenSources {
		t.Run(winner.name, func(t *testing.T) {
			env := clitest.New(t)
			env.GitHub.AddReleases(fakegithub.Release{Tag: "3.8.0"})
			for _, source := range tokenSources[i:] {
				source.set(env)
			}

			res := env.Run("list-remote")

			require.NoError(t, res.Err)
			for _, header := range authorizations(t, env) {
				assert.Equal(t, "Bearer "+winner.token, header)
			}
		})
	}
}

func TestAGHThatFailsPrintsNothingOrIsMissingGivesNoTokenSilently(t *testing.T) {
	scripts := map[string]string{
		"exits non-zero":         "printf 'gho_partial\\n'\nprintf 'gh: not logged in\\n' >&2\nexit 1",
		"prints nothing":         "exit 0",
		"prints only whitespace": "printf ' \\n'",
		"is not on the PATH":     "",
	}

	for name, script := range scripts {
		t.Run(name, func(t *testing.T) {
			env := clitest.New(t)
			env.GitHub.AddReleases(fakegithub.Release{Tag: "3.8.0"})
			if script != "" {
				env.FakeGH(script)
			}

			res := env.Run("list-remote")

			require.NoError(t, res.Err)
			assert.Empty(t, res.Stderr)
			assert.NotContains(t, res.Stdout, "gho_partial")
			for _, header := range authorizations(t, env) {
				assert.Empty(t, header)
			}
		})
	}
}

func TestATokenFromTheConfigFileWarnsThatItIsDeprecated(t *testing.T) {
	env := clitest.New(t)
	env.GitHub.AddReleases(fakegithub.Release{Tag: "3.8.0"})
	configPath := env.WriteConfig("github:\n  token: " + fileToken + "\n")

	res := env.Run("list-remote")

	require.NoError(t, res.Err)
	assert.Equal(t, "Warning: github.token in "+configPath+" is deprecated, and a later release will stop reading it. "+
		"Set GH_TOKEN or run 'gh auth login' instead, and remove the token from that file.\n", res.Stderr)
	assert.NotContains(t, res.Stdout, fileToken)
	for _, header := range authorizations(t, env) {
		assert.Equal(t, "Bearer "+fileToken, header)
	}
}

func TestATokenInTheConfigFileThatAnotherSourceOverridesDoesNotWarn(t *testing.T) {
	env := clitest.New(t)
	env.GitHub.AddReleases(fakegithub.Release{Tag: "3.8.0"})
	env.WriteConfig("github:\n  token: " + fileToken + "\n")
	env.Setenv("GH_TOKEN", ghEnvToken)

	res := env.Run("list-remote")

	require.NoError(t, res.Err)
	assert.Empty(t, res.Stderr)
}

func TestWithoutATokenRequestsAreUnauthenticated(t *testing.T) {
	env := clitest.New(t)
	env.GitHub.AddReleases(fakegithub.Release{Tag: "3.8.0"})

	res := env.Run("list-remote")

	require.NoError(t, res.Err)
	assert.Empty(t, res.Stderr)
	for _, header := range authorizations(t, env) {
		assert.Empty(t, header)
	}
}
