package cmd_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/clitest"
	"github.com/educates/educatesenv/internal/fakegithub"
)

// published returns noon UTC on the given day of 2026.
func published(month time.Month, day int) time.Time {
	return time.Date(2026, month, day, 12, 0, 0, 0, time.UTC)
}

// addReleaseHistory adds twelve stable releases, 3.11.0 down to 3.0.0,
// published on the first of each month from December back to January,
// and two pre-releases of 4.0.0 published since.
func addReleaseHistory(env *clitest.Env) {
	env.GitHub.AddReleases(
		fakegithub.Release{Tag: "4.0.0-alpha.13", PublishedAt: published(time.December, 20), Prerelease: true},
		fakegithub.Release{Tag: "4.0.0-alpha.2", PublishedAt: published(time.December, 10), Prerelease: true},
		fakegithub.Release{Tag: "3.11.0", PublishedAt: published(time.December, 1)},
		fakegithub.Release{Tag: "3.10.0", PublishedAt: published(time.November, 1)},
		fakegithub.Release{Tag: "3.9.0", PublishedAt: published(time.October, 1)},
		fakegithub.Release{Tag: "3.8.0", PublishedAt: published(time.September, 1)},
		fakegithub.Release{Tag: "3.7.0", PublishedAt: published(time.August, 1)},
		fakegithub.Release{Tag: "3.6.0", PublishedAt: published(time.July, 1)},
		fakegithub.Release{Tag: "3.5.0", PublishedAt: published(time.June, 1)},
		fakegithub.Release{Tag: "3.4.0", PublishedAt: published(time.May, 1)},
		fakegithub.Release{Tag: "3.3.0", PublishedAt: published(time.April, 1)},
		fakegithub.Release{Tag: "3.2.0", PublishedAt: published(time.March, 1)},
		fakegithub.Release{Tag: "3.1.0", PublishedAt: published(time.February, 1)},
		fakegithub.Release{Tag: "3.0.0", PublishedAt: published(time.January, 1)},
	)
}

func TestListRemoteShowsTheTenNewestStableReleasesBySemver(t *testing.T) {
	env := clitest.New(t)
	addReleaseHistory(env)

	res := env.Run("list-remote")

	require.NoError(t, res.Err)
	assert.Equal(t, ""+
		"VERSION  RELEASED\n"+
		"3.11.0   2026-12-01\n"+
		"3.10.0   2026-11-01\n"+
		"3.9.0    2026-10-01\n"+
		"3.8.0    2026-09-01\n"+
		"3.7.0    2026-08-01\n"+
		"3.6.0    2026-07-01\n"+
		"3.5.0    2026-06-01\n"+
		"3.4.0    2026-05-01\n"+
		"3.3.0    2026-04-01\n"+
		"3.2.0    2026-03-01\n",
		res.Stdout)
	assert.Equal(t, "Showing 10 of 12 stable releases. Use --all to see all, --pre-releases to include pre-releases.\n", res.Stderr)
}

func TestListRemoteRecentSetsHowManyReleasesAreShown(t *testing.T) {
	forms := [][]string{
		{"--recent", "3"},
		{"--recent=3"},
		{"-n", "3"},
		{"-n3"},
	}

	for _, form := range forms {
		t.Run(form[0], func(t *testing.T) {
			env := clitest.New(t)
			addReleaseHistory(env)

			res := env.Run(append([]string{"list-remote"}, form...)...)

			require.NoError(t, res.Err)
			assert.Equal(t, ""+
				"VERSION  RELEASED\n"+
				"3.11.0   2026-12-01\n"+
				"3.10.0   2026-11-01\n"+
				"3.9.0    2026-10-01\n",
				res.Stdout)
			assert.Equal(t, "Showing 3 of 12 stable releases. Use --all to see all, --pre-releases to include pre-releases.\n", res.Stderr)
		})
	}
}

func TestListRemoteAllShowsEveryStableRelease(t *testing.T) {
	env := clitest.New(t)
	addReleaseHistory(env)

	res := env.Run("list-remote", "--all")

	require.NoError(t, res.Err)
	assert.Equal(t, ""+
		"VERSION  RELEASED\n"+
		"3.11.0   2026-12-01\n"+
		"3.10.0   2026-11-01\n"+
		"3.9.0    2026-10-01\n"+
		"3.8.0    2026-09-01\n"+
		"3.7.0    2026-08-01\n"+
		"3.6.0    2026-07-01\n"+
		"3.5.0    2026-06-01\n"+
		"3.4.0    2026-05-01\n"+
		"3.3.0    2026-04-01\n"+
		"3.2.0    2026-03-01\n"+
		"3.1.0    2026-02-01\n"+
		"3.0.0    2026-01-01\n",
		res.Stdout)
	assert.Empty(t, res.Stderr)
}

func TestListRemotePrereleasesAddsPrereleases(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		stdout string
		stderr string
	}{
		{
			name: "the ten newest",
			args: []string{"--pre-releases"},
			stdout: "" +
				"VERSION         RELEASED\n" +
				"4.0.0-alpha.13  2026-12-20\n" +
				"4.0.0-alpha.2   2026-12-10\n" +
				"3.11.0          2026-12-01\n" +
				"3.10.0          2026-11-01\n" +
				"3.9.0           2026-10-01\n" +
				"3.8.0           2026-09-01\n" +
				"3.7.0           2026-08-01\n" +
				"3.6.0           2026-07-01\n" +
				"3.5.0           2026-06-01\n" +
				"3.4.0           2026-05-01\n",
			stderr: "Showing 10 of 14 releases. Use --all to see all.\n",
		},
		{
			name: "every release",
			args: []string{"--pre-releases", "--all"},
			stdout: "" +
				"VERSION         RELEASED\n" +
				"4.0.0-alpha.13  2026-12-20\n" +
				"4.0.0-alpha.2   2026-12-10\n" +
				"3.11.0          2026-12-01\n" +
				"3.10.0          2026-11-01\n" +
				"3.9.0           2026-10-01\n" +
				"3.8.0           2026-09-01\n" +
				"3.7.0           2026-08-01\n" +
				"3.6.0           2026-07-01\n" +
				"3.5.0           2026-06-01\n" +
				"3.4.0           2026-05-01\n" +
				"3.3.0           2026-04-01\n" +
				"3.2.0           2026-03-01\n" +
				"3.1.0           2026-02-01\n" +
				"3.0.0           2026-01-01\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := clitest.New(t)
			addReleaseHistory(env)

			res := env.Run(append([]string{"list-remote"}, tt.args...)...)

			require.NoError(t, res.Err)
			assert.Equal(t, tt.stdout, res.Stdout)
			assert.Equal(t, tt.stderr, res.Stderr)
		})
	}
}

func TestListRemoteByDateOrdersByPublishTime(t *testing.T) {
	at := func(month time.Month, day, hour int) time.Time {
		return time.Date(2026, month, day, hour, 0, 0, 0, time.UTC)
	}

	tests := []struct {
		name   string
		args   []string
		stdout string
		stderr string
	}{
		{
			name: "with pre-releases",
			args: []string{"--by-date", "--pre-releases"},
			stdout: "" +
				"VERSION         RELEASED\n" +
				"4.0.0-alpha.13  2026-09-03\n" +
				"3.8.1-rc.2      2026-09-01\n" +
				"4.0.0-alpha.12  2026-08-27\n" +
				"3.8.0           2026-08-26\n" +
				"3.8.0-rc.13     2026-08-25\n" +
				"3.8.0-rc.12     2026-08-25\n" +
				"4.0.0-alpha.11  2026-08-20\n",
		},
		{
			name: "with pre-releases and a limit",
			args: []string{"--by-date", "--pre-releases", "-n", "3"},
			stdout: "" +
				"VERSION         RELEASED\n" +
				"4.0.0-alpha.13  2026-09-03\n" +
				"3.8.1-rc.2      2026-09-01\n" +
				"4.0.0-alpha.12  2026-08-27\n",
			stderr: "Showing 3 of 7 releases. Use --all to see all.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := clitest.New(t)
			env.GitHub.AddReleases(
				fakegithub.Release{Tag: "4.0.0-alpha.13", PublishedAt: at(time.September, 3, 9), Prerelease: true},
				fakegithub.Release{Tag: "3.8.1-rc.2", PublishedAt: at(time.September, 1, 9), Prerelease: true},
				fakegithub.Release{Tag: "4.0.0-alpha.12", PublishedAt: at(time.August, 27, 9), Prerelease: true},
				fakegithub.Release{Tag: "3.8.0", PublishedAt: at(time.August, 26, 18)},
				fakegithub.Release{Tag: "3.8.0-rc.12", PublishedAt: at(time.August, 25, 9), Prerelease: true},
				fakegithub.Release{Tag: "3.8.0-rc.13", PublishedAt: at(time.August, 25, 17), Prerelease: true},
				fakegithub.Release{Tag: "4.0.0-alpha.11", PublishedAt: at(time.August, 20, 9), Prerelease: true},
			)

			res := env.Run(append([]string{"list-remote"}, tt.args...)...)

			require.NoError(t, res.Err)
			assert.Equal(t, tt.stdout, res.Stdout)
			assert.Equal(t, tt.stderr, res.Stderr)
		})
	}
}

func TestListRemoteRecentsIsDeprecated(t *testing.T) {
	env := clitest.New(t)
	addReleaseHistory(env)

	res := env.Run("list-remote", "--recents")

	require.NoError(t, res.Err)
	assert.Equal(t, ""+
		"VERSION  RELEASED\n"+
		"3.11.0   2026-12-01\n"+
		"3.10.0   2026-11-01\n"+
		"3.9.0    2026-10-01\n"+
		"3.8.0    2026-09-01\n"+
		"3.7.0    2026-08-01\n"+
		"3.6.0    2026-07-01\n"+
		"3.5.0    2026-06-01\n"+
		"3.4.0    2026-05-01\n"+
		"3.3.0    2026-04-01\n"+
		"3.2.0    2026-03-01\n",
		res.Stdout)
	assert.Equal(t, ""+
		"Flag --recents has been deprecated, the 10 most recent releases are now shown by default\n"+
		"Showing 10 of 12 stable releases. Use --all to see all, --pre-releases to include pre-releases.\n",
		res.Stderr)
}

func TestListRemoteHelpLeavesOutRecents(t *testing.T) {
	env := clitest.New(t)

	res := env.Run("list-remote", "--help")

	require.NoError(t, res.Err)
	assert.Contains(t, res.Stdout, "--recent int")
	assert.NotContains(t, res.Stdout, "--recents")
}

func TestListRemoteConsidersReleasesOnEveryPage(t *testing.T) {
	env := clitest.New(t)
	// 250 releases, 2.249.0 down to 2.0.0, span three pages of 100.
	first := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	for i := 249; i >= 0; i-- {
		env.GitHub.AddReleases(fakegithub.Release{
			Tag:         fmt.Sprintf("2.%d.0", i),
			PublishedAt: first.AddDate(0, 0, i),
		})
	}

	res := env.Run("list-remote", "--all")

	require.NoError(t, res.Err)
	lines := strings.Split(strings.TrimSuffix(res.Stdout, "\n"), "\n")
	require.Len(t, lines, 251)
	assert.Equal(t, "VERSION  RELEASED", lines[0])
	assert.Equal(t, "2.249.0  2020-09-06", lines[1])
	assert.Equal(t, "2.0.0    2020-01-01", lines[250])

	res = env.Run("list-remote")

	require.NoError(t, res.Err)
	assert.Equal(t, "Showing 10 of 250 stable releases. Use --all to see all, --pre-releases to include pre-releases.\n", res.Stderr)

	for _, req := range env.GitHub.Requests() {
		assert.Equal(t, "100", req.Query.Get("per_page"), "per_page of %s?%s", req.Path, req.Query.Encode())
	}
}

func TestListRemoteShowsPublishDatesInUTC(t *testing.T) {
	env := clitest.New(t)
	env.GitHub.Handle("GET /repos/educates/educates-training-platform/releases", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"id":1,"tag_name":"3.8.0","draft":false,"prerelease":false,"published_at":"2026-08-27T01:30:00+02:00"}]`)
	})

	res := env.Run("list-remote")

	require.NoError(t, res.Err)
	assert.Equal(t, "VERSION  RELEASED\n3.8.0    2026-08-26\n", res.Stdout)
}

func TestListRemoteSaysWhenNoVersionsAreFound(t *testing.T) {
	tests := []struct {
		name     string
		releases []fakegithub.Release
	}{
		{name: "no releases"},
		{
			name: "only drafts and pre-releases",
			releases: []fakegithub.Release{
				{Tag: "3.9.0", Draft: true},
				{Tag: "3.9.0-rc.1", PublishedAt: published(time.August, 1), Prerelease: true},
				{Tag: "2.6.0-alpha.30", PublishedAt: published(time.July, 1)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := clitest.New(t)
			env.GitHub.AddReleases(tt.releases...)

			res := env.Run("list-remote")

			require.NoError(t, res.Err)
			assert.Equal(t, "No versions found\n", res.Stdout)
			assert.Empty(t, res.Stderr)
		})
	}
}

func TestListRemoteRejectsAllTogetherWithRecent(t *testing.T) {
	forms := [][]string{
		{"--all", "--recent", "5"},
		{"--recent=3", "--all"},
		{"--all", "-n", "10"},
	}

	for _, form := range forms {
		t.Run(form[0]+" "+form[1], func(t *testing.T) {
			env := clitest.New(t)
			addReleaseHistory(env)

			res := env.Run(append([]string{"list-remote"}, form...)...)

			require.EqualError(t, res.Err, "if any flags in the group [all recent] are set none of the others can be; [all recent] were all set")
			assert.Empty(t, res.Stdout)
			assert.Empty(t, res.Stderr)
		})
	}
}

func TestListRemoteRejectsRecentBelowOne(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{args: []string{"--recent", "0"}, want: "--recent must be at least 1, got 0"},
		{args: []string{"--recent=-5"}, want: "--recent must be at least 1, got -5"},
		{args: []string{"-n", "-1"}, want: "--recent must be at least 1, got -1"},
	}

	for _, tt := range tests {
		t.Run(tt.args[0], func(t *testing.T) {
			env := clitest.New(t)
			addReleaseHistory(env)

			res := env.Run(append([]string{"list-remote"}, tt.args...)...)

			require.EqualError(t, res.Err, tt.want)
			assert.Empty(t, res.Stdout)
			assert.Empty(t, res.Stderr)
		})
	}
}
