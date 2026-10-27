package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
)

func TestVersionPrintsTheBuildVersion(t *testing.T) {
	env := clitest.New(t)

	res := env.Run("version")

	require.NoError(t, res.Err)
	assert.Equal(t, "develop\n", res.Stdout)
	assert.Empty(t, res.Stderr)
}
