package releases_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/educates/educatesenv/pkg/releases"
)

func TestLatestIsTheHighestStableReleaseBySemver(t *testing.T) {
	// Twelve release candidates of 3.8.0, newest first, as GitHub lists
	// them during a release-candidate cycle.
	var candidates []releases.Release
	for rc := 12; rc >= 1; rc-- {
		candidates = append(candidates, releases.Release{
			Tag:              fmt.Sprintf("3.8.0-rc.%d", rc),
			PublishedAt:      day(time.August, 14+rc),
			GitHubPrerelease: true,
		})
	}

	tests := []struct {
		name string
		all  []releases.Release
		want string
	}{
		{
			name: "ordered by semver, not as strings",
			all:  []releases.Release{stable("3.9.0"), stable("3.10.0"), stable("3.8.0")},
			want: "3.10.0",
		},
		{
			name: "after more than ten consecutive pre-releases",
			all:  slices.Concat(candidates, []releases.Release{stable("3.7.3"), stable("3.7.2")}),
			want: "3.7.3",
		},
		{
			name: "a backport published after a higher release",
			all: []releases.Release{
				{Tag: "3.7.4", PublishedAt: day(time.September, 10)},
				{Tag: "3.8.0", PublishedAt: day(time.August, 26)},
				{Tag: "3.7.3", PublishedAt: day(time.July, 1)},
			},
			want: "3.8.0",
		},
		{
			name: "pre-release tags that GitHub never flagged",
			all:  []releases.Release{stable("3.9.0-rc.1"), stable("2.6.0-alpha.30"), stable("2.5.0")},
			want: "2.5.0",
		},
		{
			name: "a release GitHub flags as a pre-release",
			all: []releases.Release{
				{Tag: "3.9.0", PublishedAt: day(time.September, 1), GitHubPrerelease: true},
				stable("3.8.0"),
			},
			want: "3.8.0",
		},
		{
			name: "drafts and tags that are not semver",
			all: []releases.Release{
				{Tag: "3.9.0", Draft: true},
				stable("latest"),
				stable("v4.0.0"),
				stable("4.0"),
				stable("3.8.0"),
			},
			want: "3.8.0",
		},
		{
			name: "no stable release",
			all: slices.Concat(candidates, []releases.Release{
				{Tag: "3.7.0", Draft: true},
				stable("3.6.0-beta.1"),
			}),
		},
		{
			name: "no release at all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			latest, ok := releases.Latest(tt.all)

			assert.Equal(t, tt.want != "", ok)
			assert.Equal(t, tt.want, latest.Tag)
		})
	}
}
