package cmd_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
)

func TestCommandsWithoutArgumentsRejectStrayOnes(t *testing.T) {
	tests := [][]string{
		{"list-remote"},
		{"list"},
		{"list-local"},
		{"init"},
		{"config"},
		{"config", "init"},
		{"config", "view"},
		{"version"},
	}

	for _, command := range tests {
		name := strings.Join(command, " ")
		t.Run(name, func(t *testing.T) {
			env := clitest.New(t)

			res := env.Run(append(command, "stray")...)

			require.EqualError(t, res.Err, `unknown command "stray" for "educatesenv `+name+`"`)
			assert.Empty(t, res.Stdout)
			assert.NoDirExists(t, env.ConfigDir())
			assert.Empty(t, env.GitHub.Requests())
		})
	}
}
