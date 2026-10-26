package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
	"github.com/educates/educatesenv/internal/fakegithub"
)

func TestListRemoteListsStableReleasesFromGitHub(t *testing.T) {
	env := clitest.New(t)
	env.GitHub.AddReleases(
		fakegithub.Release{Tag: "4.0.0-alpha.1", Prerelease: true},
		fakegithub.Release{Tag: "3.8.0"},
		fakegithub.Release{Tag: "3.7.0"},
	)

	res := env.Run("list-remote")

	require.NoError(t, res.Err)
	assert.Equal(t, "Available versions:\n- 3.8.0\n- 3.7.0\n", res.Stdout)
	assert.Empty(t, res.Stderr)
}
