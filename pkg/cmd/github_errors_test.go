package cmd_test

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
)

// The API routes the commands call first.
const (
	releasesRoute     = "GET /repos/educates/educates-training-platform/releases"
	releaseByTagRoute = "GET /repos/educates/educates-training-platform/releases/tags/3.8.0"
)

// githubCall is a command and the API route it calls first.
type githubCall struct {
	args  []string
	route string
}

// githubCalls are the commands that call GitHub, each with the route a
// failing response is served on.
var githubCalls = []githubCall{
	{args: []string{"list-remote"}, route: releasesRoute},
	{args: []string{"install", "3.8.0"}, route: releaseByTagRoute},
	{args: []string{"install", "latest"}, route: releasesRoute},
	{args: []string{"init", "--download"}, route: releasesRoute},
}

// rateLimitReset is when the rate limit the fake server reports resets.
var rateLimitReset = time.Date(2030, time.January, 2, 15, 4, 5, 0, time.UTC)

// rateLimited answers as GitHub does once the rate limit is used up. Only
// this response carries rate limit headers, since a client that saw no
// remaining requests refuses to make more.
func rateLimited(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-RateLimit-Limit", "60")
	w.Header().Set("X-RateLimit-Remaining", "0")
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(rateLimitReset.Unix(), 10))
	w.WriteHeader(http.StatusForbidden)
	_, _ = fmt.Fprint(w, `{"message":"API rate limit exceeded for 127.0.0.1."}`)
}

// badCredentials answers as GitHub does when it rejects the token.
func badCredentials(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = fmt.Fprint(w, `{"message":"Bad credentials"}`)
}

func TestARateLimitSaysWhenItResetsAndHowToAuthenticate(t *testing.T) {
	for _, call := range githubCalls {
		t.Run(strings.Join(call.args, " "), func(t *testing.T) {
			env := clitest.New(t)
			addRelease(env, "3.8.0", []byte("educates 3.8.0 binary"))
			env.GitHub.Handle(call.route, rateLimited)

			res := env.Run(call.args...)

			require.Error(t, res.Err)
			assert.Contains(t, res.Err.Error(), "GitHub API rate limit exceeded; it resets at 2030-01-02 15:04:05 UTC. "+
				"Set GH_TOKEN or run 'gh auth login' to authenticate and get a higher limit")
		})
	}
}

func TestARateLimitOnATokenNamesTheTokensSource(t *testing.T) {
	env := clitest.New(t)
	env.Setenv("GH_TOKEN", ghEnvToken)
	env.GitHub.Handle(releasesRoute, rateLimited)

	res := env.Run("list-remote")

	require.EqualError(t, res.Err, "GitHub API rate limit exceeded for the token from GH_TOKEN; it resets at 2030-01-02 15:04:05 UTC")
}

func TestARejectedTokenNamesItsSource(t *testing.T) {
	for _, source := range tokenSources {
		for _, call := range githubCalls {
			t.Run(source.name+"/"+strings.Join(call.args, " "), func(t *testing.T) {
				env := clitest.New(t)
				addRelease(env, "3.8.0", []byte("educates 3.8.0 binary"))
				source.set(env)
				env.GitHub.Handle(call.route, badCredentials)

				res := env.Run(call.args...)

				require.Error(t, res.Err)
				assert.Contains(t, res.Err.Error(), "GitHub rejected the token from "+source.name+": 401 Bad credentials. "+
					"Replace or remove that token and try again")
				assert.NotContains(t, res.Err.Error(), source.token)
				assert.NotContains(t, res.Stdout, source.token)
				assert.NotContains(t, res.Stderr, source.token)
			})
		}
	}
}
