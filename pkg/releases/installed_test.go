package releases_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/educates/educatesenv/pkg/releases"
)

func TestSortInstalledOrdersSemverNamesHighestFirst(t *testing.T) {
	tests := []struct {
		name  string
		names []string
		want  []string
	}{
		{
			name:  "numeric minor versions",
			names: []string{"3.9.0", "3.10.0", "3.8.1"},
			want:  []string{"3.10.0", "3.9.0", "3.8.1"},
		},
		{
			name:  "numeric pre-release identifiers",
			names: []string{"4.0.0-alpha.2", "4.0.0-alpha.13", "4.0.0-alpha.9"},
			want:  []string{"4.0.0-alpha.13", "4.0.0-alpha.9", "4.0.0-alpha.2"},
		},
		{
			name:  "a release above its pre-releases",
			names: []string{"3.8.0-rc.13", "3.7.4", "3.8.0", "3.8.0-rc.2"},
			want:  []string{"3.8.0", "3.8.0-rc.13", "3.8.0-rc.2", "3.7.4"},
		},
		{
			name:  "versions that differ only in build metadata, alphabetically",
			names: []string{"3.8.0+build.2", "3.9.0", "3.8.0+build.1"},
			want:  []string{"3.9.0", "3.8.0+build.1", "3.8.0+build.2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, releases.SortInstalled(tt.names))
		})
	}
}

func TestSortInstalledPutsNamesThatAreNotSemverLastAlphabetically(t *testing.T) {
	tests := []struct {
		name  string
		names []string
		want  []string
	}{
		{
			name:  "after the semver names",
			names: []string{"latest", "3.9.0", "my-build", "3.10.0"},
			want:  []string{"3.10.0", "3.9.0", "latest", "my-build"},
		},
		{
			name: "shorthands, prefixes and leading zeros are not semver",
			names: []string{
				"v3.7.0", "3.8", "3.08.0", "3.7.0", "3",
			},
			want: []string{"3.7.0", "3", "3.08.0", "3.8", "v3.7.0"},
		},
		{
			name:  "only names that are not semver",
			names: []string{"zeta", "alpha", "Beta"},
			want:  []string{"Beta", "alpha", "zeta"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, releases.SortInstalled(tt.names))
		})
	}
}

func TestSortInstalledLeavesItsInputAlone(t *testing.T) {
	names := []string{"3.7.0", "latest", "3.8.0"}

	_ = releases.SortInstalled(names)

	assert.Equal(t, []string{"3.7.0", "latest", "3.8.0"}, names)
}
