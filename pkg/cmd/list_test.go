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

func TestListShowsInstalledVersionsNewestFirstBySemver(t *testing.T) {
	tests := []struct {
		name   string
		active string
		want   string
	}{
		{
			name: "no version active",
			want: "Installed versions:\n" +
				"  4.0.0-alpha.13\n" +
				"  4.0.0-alpha.2\n" +
				"  3.10.0\n" +
				"  3.9.0\n" +
				"  3.8\n" +
				"  my-build\n",
		},
		{
			name:   "a semver version active",
			active: "3.9.0",
			want: "Installed versions:\n" +
				"  4.0.0-alpha.13\n" +
				"  4.0.0-alpha.2\n" +
				"  3.10.0\n" +
				"* 3.9.0 (active)\n" +
				"  3.8\n" +
				"  my-build\n",
		},
		{
			name:   "a name that is not semver active",
			active: "my-build",
			want: "Installed versions:\n" +
				"  4.0.0-alpha.13\n" +
				"  4.0.0-alpha.2\n" +
				"  3.10.0\n" +
				"  3.9.0\n" +
				"  3.8\n" +
				"* my-build (active)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := clitest.New(t)
			for _, v := range []string{"3.9.0", "my-build", "4.0.0-alpha.2", "3.8", "3.10.0", "4.0.0-alpha.13"} {
				env.WriteBinary(v)
			}
			if tt.active != "" {
				require.NoError(t, os.Symlink("educates-"+tt.active, env.ActiveLink()))
			}

			res := env.Run("list")

			require.NoError(t, res.Err)
			assert.Equal(t, tt.want, res.Stdout)
			assert.Empty(t, res.Stderr)
		})
	}
}

func TestListShowsDevelopFirstWhenDevelopmentModeIsEnabled(t *testing.T) {
	tests := []struct {
		name          string
		activeVersion string
		want          func(devBinary string) string
	}{
		{
			name: "develop active",
			want: func(devBinary string) string {
				return "Installed versions:\n" +
					"* develop (active) -> " + devBinary + "\n" +
					"  4.0.0-alpha.2\n" +
					"  3.10.0\n" +
					"  3.9.0\n" +
					"  my-build\n"
			},
		},
		{
			name:          "an installed version active",
			activeVersion: "3.9.0",
			want: func(devBinary string) string {
				return "Installed versions:\n" +
					"  develop -> " + devBinary + "\n" +
					"  4.0.0-alpha.2\n" +
					"  3.10.0\n" +
					"* 3.9.0 (active)\n" +
					"  my-build\n"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := clitest.New(t)
			devBinary := writeDevelopmentBinary(t, env)
			env.WriteConfig("development:\n  enabled: true\n  binaryLocation: " + devBinary + "\n")
			for _, v := range []string{"3.9.0", "my-build", "4.0.0-alpha.2", "3.10.0"} {
				env.WriteBinary(v)
			}
			target := devBinary
			if tt.activeVersion != "" {
				target = "educates-" + tt.activeVersion
			}
			require.NoError(t, os.Symlink(target, env.ActiveLink()))

			res := env.Run("list")

			require.NoError(t, res.Err)
			assert.Equal(t, tt.want(devBinary), res.Stdout)
			assert.Empty(t, res.Stderr)
		})
	}
}

func TestListSaysWhenNoVersionIsInstalled(t *testing.T) {
	env := clitest.New(t)

	res := env.Run("list")

	require.NoError(t, res.Err)
	assert.Equal(t, "Installed versions:\nNo versions installed\n", res.Stdout)
}
