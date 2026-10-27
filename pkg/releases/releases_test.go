package releases_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/educates/educatesenv/pkg/releases"
)

// day returns noon UTC on the given day of 2026, so tests can order
// releases by publish time without spelling out timestamps.
func day(month time.Month, d int) time.Time {
	return time.Date(2026, month, d, 12, 0, 0, 0, time.UTC)
}

// stable returns a published release that GitHub does not flag.
func stable(tag string) releases.Release {
	return releases.Release{Tag: tag, PublishedAt: day(time.January, 1)}
}

// tags returns the tags of rs, in order.
func tags(rs []releases.Release) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, r.Tag)
	}
	return out
}

func TestSelectLeavesOutPrereleasesUnlessIncluded(t *testing.T) {
	all := []releases.Release{
		stable("3.8.0"),
		{Tag: "3.8.0-rc.2", PublishedAt: day(time.August, 14), GitHubPrerelease: true},
		// Older pre-releases that GitHub never flagged.
		stable("2.6.0-alpha.30"),
		stable("2.6.0-beta.1"),
		// GitHub flags it, though the tag has no pre-release part.
		{Tag: "2.5.9", PublishedAt: day(time.January, 2), GitHubPrerelease: true},
		stable("2.5.0"),
	}

	tests := []struct {
		name               string
		includePrereleases bool
		want               []string
	}{
		{
			name: "stable only",
			want: []string{"3.8.0", "2.5.0"},
		},
		{
			name:               "with pre-releases",
			includePrereleases: true,
			want:               []string{"3.8.0", "3.8.0-rc.2", "2.6.0-beta.1", "2.6.0-alpha.30", "2.5.9", "2.5.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chosen, candidates := releases.Select(all, releases.SelectOptions{
				IncludePrereleases: tt.includePrereleases,
				Limit:              releases.All,
			})

			assert.Equal(t, tt.want, tags(chosen))
			assert.Equal(t, len(tt.want), candidates)
		})
	}
}

func TestSelectNeverChoosesDraftsOrTagsThatAreNotSemver(t *testing.T) {
	all := []releases.Release{
		{Tag: "3.9.0", Draft: true},
		{Tag: "3.9.0-rc.1", Draft: true, GitHubPrerelease: true},
		stable("3.8.0"),
		stable("latest"),
		stable("v3.7.0"),
		stable("3.8"),
		stable("3.08.0"),
		stable(""),
		stable("3.7.0-rc.1"),
		stable("3.6.0"),
	}

	chosen, candidates := releases.Select(all, releases.SelectOptions{
		IncludePrereleases: true,
		Limit:              releases.All,
	})

	assert.Equal(t, []string{"3.8.0", "3.7.0-rc.1", "3.6.0"}, tags(chosen))
	assert.Equal(t, 3, candidates)
}

func TestSelectByPublishTimeOrdersNewestFirst(t *testing.T) {
	at := func(d, hour int) time.Time {
		return time.Date(2026, time.August, d, hour, 0, 0, 0, time.UTC)
	}
	all := []releases.Release{
		{Tag: "4.0.0-alpha.12", PublishedAt: at(10, 9)},
		{Tag: "3.8.0-rc.12", PublishedAt: at(25, 9)},
		{Tag: "3.8.1-rc.2", PublishedAt: at(28, 9)},
		{Tag: "3.8.0", PublishedAt: at(26, 18)},
		{Tag: "3.8.0-rc.13", PublishedAt: at(25, 17)},
		{Tag: "4.0.0-alpha.13", PublishedAt: at(27, 9)},
	}

	chosen, _ := releases.Select(all, releases.SelectOptions{
		IncludePrereleases: true,
		Limit:              releases.All,
		Order:              releases.ByPublishTime,
	})

	assert.Equal(t, []string{
		"3.8.1-rc.2",
		"4.0.0-alpha.13",
		"3.8.0",
		"3.8.0-rc.13",
		"3.8.0-rc.12",
		"4.0.0-alpha.12",
	}, tags(chosen))
}

func TestSelectCombinesPrereleasesLimitAndOrder(t *testing.T) {
	// Listed in creation order, newest first, as GitHub lists them.
	all := []releases.Release{
		{Tag: "3.7.4", PublishedAt: day(time.September, 2)},
		{Tag: "3.8.1-rc.2", PublishedAt: day(time.August, 28), GitHubPrerelease: true},
		{Tag: "4.0.0-alpha.13", PublishedAt: day(time.August, 27), GitHubPrerelease: true},
		{Tag: "3.8.0", PublishedAt: day(time.August, 26)},
		{Tag: "3.8.0-rc.13", PublishedAt: day(time.August, 25), GitHubPrerelease: true},
		{Tag: "3.7.3", PublishedAt: day(time.July, 15)},
		{Tag: "4.0.0-alpha.2", PublishedAt: day(time.July, 1), GitHubPrerelease: true},
		{Tag: "3.6.0", PublishedAt: day(time.May, 1)},
	}

	tests := []struct {
		name       string
		opts       releases.SelectOptions
		want       []string
		candidates int
	}{
		{
			name:       "stable, 3, by semver",
			opts:       releases.SelectOptions{Limit: 3, Order: releases.BySemver},
			want:       []string{"3.8.0", "3.7.4", "3.7.3"},
			candidates: 4,
		},
		{
			name:       "stable, all, by semver",
			opts:       releases.SelectOptions{Limit: releases.All, Order: releases.BySemver},
			want:       []string{"3.8.0", "3.7.4", "3.7.3", "3.6.0"},
			candidates: 4,
		},
		{
			name:       "stable, 3, by publish time",
			opts:       releases.SelectOptions{Limit: 3, Order: releases.ByPublishTime},
			want:       []string{"3.7.4", "3.8.0", "3.7.3"},
			candidates: 4,
		},
		{
			name:       "stable, all, by publish time",
			opts:       releases.SelectOptions{Limit: releases.All, Order: releases.ByPublishTime},
			want:       []string{"3.7.4", "3.8.0", "3.7.3", "3.6.0"},
			candidates: 4,
		},
		{
			name:       "with pre-releases, 3, by semver",
			opts:       releases.SelectOptions{IncludePrereleases: true, Limit: 3, Order: releases.BySemver},
			want:       []string{"4.0.0-alpha.13", "4.0.0-alpha.2", "3.8.1-rc.2"},
			candidates: 8,
		},
		{
			name: "with pre-releases, all, by semver",
			opts: releases.SelectOptions{IncludePrereleases: true, Limit: releases.All, Order: releases.BySemver},
			want: []string{
				"4.0.0-alpha.13", "4.0.0-alpha.2", "3.8.1-rc.2", "3.8.0",
				"3.8.0-rc.13", "3.7.4", "3.7.3", "3.6.0",
			},
			candidates: 8,
		},
		{
			name:       "with pre-releases, 3, by publish time",
			opts:       releases.SelectOptions{IncludePrereleases: true, Limit: 3, Order: releases.ByPublishTime},
			want:       []string{"3.7.4", "3.8.1-rc.2", "4.0.0-alpha.13"},
			candidates: 8,
		},
		{
			name: "with pre-releases, all, by publish time",
			opts: releases.SelectOptions{IncludePrereleases: true, Limit: releases.All, Order: releases.ByPublishTime},
			want: []string{
				"3.7.4", "3.8.1-rc.2", "4.0.0-alpha.13", "3.8.0",
				"3.8.0-rc.13", "3.7.3", "4.0.0-alpha.2", "3.6.0",
			},
			candidates: 8,
		},
		{
			name:       "a limit above the number of candidates",
			opts:       releases.SelectOptions{Limit: 10},
			want:       []string{"3.8.0", "3.7.4", "3.7.3", "3.6.0"},
			candidates: 4,
		},
		{
			name:       "a limit equal to the number of candidates",
			opts:       releases.SelectOptions{Limit: 4},
			want:       []string{"3.8.0", "3.7.4", "3.7.3", "3.6.0"},
			candidates: 4,
		},
		{
			name:       "a negative limit",
			opts:       releases.SelectOptions{Limit: -1},
			want:       []string{"3.8.0", "3.7.4", "3.7.3", "3.6.0"},
			candidates: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chosen, candidates := releases.Select(all, tt.opts)

			assert.Equal(t, tt.want, tags(chosen))
			assert.Equal(t, tt.candidates, candidates)
		})
	}
}

func TestSelectLeavesItsInputAlone(t *testing.T) {
	all := []releases.Release{stable("3.7.0"), stable("3.9.0-rc.1"), stable("3.8.0")}

	_, _ = releases.Select(all, releases.SelectOptions{Limit: 1})

	assert.Equal(t, []string{"3.7.0", "3.9.0-rc.1", "3.8.0"}, tags(all))
}

func TestSelectOrdersBySemverHighestFirst(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want []string
	}{
		{
			name: "numeric minor versions",
			tags: []string{"3.9.0", "3.10.0", "3.8.1"},
			want: []string{"3.10.0", "3.9.0", "3.8.1"},
		},
		{
			name: "numeric pre-release identifiers",
			tags: []string{"4.0.0-alpha.2", "4.0.0-alpha.13", "4.0.0-alpha.9"},
			want: []string{"4.0.0-alpha.13", "4.0.0-alpha.9", "4.0.0-alpha.2"},
		},
		{
			name: "a release above its pre-releases",
			tags: []string{"3.8.0-rc.13", "3.8.0", "3.8.0-rc.2", "3.7.4"},
			want: []string{"3.8.0", "3.8.0-rc.13", "3.8.0-rc.2", "3.7.4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var all []releases.Release
			for _, tag := range tt.tags {
				all = append(all, stable(tag))
			}

			chosen, _ := releases.Select(all, releases.SelectOptions{
				IncludePrereleases: true,
				Limit:              releases.All,
				Order:              releases.BySemver,
			})

			assert.Equal(t, tt.want, tags(chosen))
		})
	}
}
