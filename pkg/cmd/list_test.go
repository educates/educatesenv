package cmd_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
)

func TestListMarksTheActiveVersion(t *testing.T) {
	env := clitest.New(t)
	env.WriteBinary("3.8.0")
	require.NoError(t, os.Symlink("educates-3.8.0", env.ActiveLink()))

	res := env.Run("list")

	require.NoError(t, res.Err)
	assert.Equal(t, "Installed versions:\n* 3.8.0 (active)\n", res.Stdout)
	assert.Empty(t, res.Stderr)
}

func TestListSaysWhenNoVersionIsInstalled(t *testing.T) {
	env := clitest.New(t)

	res := env.Run("list")

	require.NoError(t, res.Err)
	assert.Equal(t, "Installed versions:\nNo versions installed\n", res.Stdout)
}
