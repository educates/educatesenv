package fakegithub_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v93/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/fakegithub"
)

const (
	owner = fakegithub.DefaultOwner
	repo  = fakegithub.DefaultRepository
)

func newClient(t *testing.T, srv *fakegithub.Server, opts ...github.ClientOptionsFunc) *github.Client {
	t.Helper()
	baseURL := srv.URL
	client, err := github.NewClient(append([]github.ClientOptionsFunc{github.WithURLs(&baseURL, nil)}, opts...)...)
	require.NoError(t, err)
	return client
}

func TestListsEveryReleaseAcrossPagesInTheOrderAdded(t *testing.T) {
	srv := fakegithub.New(t)
	var want []string
	for i := 250; i > 0; i-- {
		tag := fmt.Sprintf("1.0.%d", i)
		want = append(want, tag)
		srv.AddReleases(fakegithub.Release{Tag: tag})
	}
	client := newClient(t, srv)

	var got []string
	var pages int
	opt := &github.ListOptions{PerPage: 100}
	for {
		releases, resp, err := client.Repositories.ListReleases(context.Background(), owner, repo, opt)
		require.NoError(t, err)
		pages++
		for _, r := range releases {
			got = append(got, r.GetTagName())
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	assert.Equal(t, 3, pages)
	assert.Equal(t, want, got)
}

func TestServesReleaseFieldsAndAssets(t *testing.T) {
	srv := fakegithub.New(t)
	published := time.Date(2026, 8, 26, 14, 30, 0, 0, time.UTC)
	srv.AddReleases(
		fakegithub.Release{Tag: "4.0.0-alpha.13", Prerelease: true, PublishedAt: published.Add(time.Hour)},
		fakegithub.Release{Tag: "3.8.0", PublishedAt: published, Assets: []fakegithub.Asset{
			{Name: "educates-linux-amd64", Content: []byte("binary")},
			{Name: "checksums.txt", Content: []byte("sums")},
		}},
		fakegithub.Release{Tag: "3.9.0", Draft: true},
	)
	client := newClient(t, srv)

	releases, _, err := client.Repositories.ListReleases(context.Background(), owner, repo, nil)
	require.NoError(t, err)
	require.Len(t, releases, 3)
	assert.True(t, releases[0].GetPrerelease())
	assert.True(t, releases[2].GetDraft())
	assert.Nil(t, releases[2].PublishedAt)

	release, _, err := client.Repositories.GetReleaseByTag(context.Background(), owner, repo, "3.8.0")
	require.NoError(t, err)
	assert.Equal(t, "3.8.0", release.GetTagName())
	assert.False(t, release.GetPrerelease())
	assert.Equal(t, published, release.GetPublishedAt().UTC())
	require.Len(t, release.Assets, 2)
	assert.Equal(t, "educates-linux-amd64", release.Assets[0].GetName())
	assert.Equal(t, srv.DownloadURL("3.8.0", "educates-linux-amd64"), release.Assets[0].GetBrowserDownloadURL())

	resp, err := http.Get(release.Assets[0].GetBrowserDownloadURL())
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "binary", string(body))
}

func TestAnswersNotFoundForUnknownTagsAssetsAndRepositories(t *testing.T) {
	srv := fakegithub.New(t)
	srv.AddReleases(fakegithub.Release{Tag: "3.8.0"})
	client := newClient(t, srv)

	_, resp, err := client.Repositories.GetReleaseByTag(context.Background(), owner, repo, "9.9.9")
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	_, resp, err = client.Repositories.ListReleases(context.Background(), "someone", "else", nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	download, err := http.Get(srv.DownloadURL("3.8.0", "missing"))
	require.NoError(t, err)
	_ = download.Body.Close()
	assert.Equal(t, http.StatusNotFound, download.StatusCode)
}

func TestServesTheRepositorySetForIt(t *testing.T) {
	srv := fakegithub.New(t)
	srv.SetRepository("acme", "tools")
	srv.AddReleases(fakegithub.Release{Tag: "1.0.0"})
	client := newClient(t, srv)

	releases, _, err := client.Repositories.ListReleases(context.Background(), "acme", "tools", nil)
	require.NoError(t, err)
	assert.Len(t, releases, 1)

	_, resp, err := client.Repositories.ListReleases(context.Background(), owner, repo, nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandlersRegisteredByATestTakePrecedence(t *testing.T) {
	srv := fakegithub.New(t)
	srv.AddReleases(fakegithub.Release{Tag: "3.8.0"})
	srv.Handle("GET /repos/{owner}/{repo}/releases/tags/3.8.0", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, `{"message":"Bad credentials"}`)
	})
	client := newClient(t, srv)

	_, resp, err := client.Repositories.GetReleaseByTag(context.Background(), owner, repo, "3.8.0")
	require.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	releases, _, err := client.Repositories.ListReleases(context.Background(), owner, repo, nil)
	require.NoError(t, err)
	assert.Len(t, releases, 1)
}

func TestRecordsTheRequestsItReceives(t *testing.T) {
	srv := fakegithub.New(t)
	client := newClient(t, srv, github.WithAuthToken("tok123"))

	_, _, err := client.Repositories.ListReleases(context.Background(), owner, repo, &github.ListOptions{PerPage: 100})
	require.NoError(t, err)

	requests := srv.Requests()
	require.Len(t, requests, 1)
	assert.Equal(t, http.MethodGet, requests[0].Method)
	assert.Equal(t, "/repos/educates/educates-training-platform/releases", requests[0].Path)
	assert.Equal(t, "100", requests[0].Query.Get("per_page"))
	assert.Equal(t, "Bearer tok123", requests[0].Header.Get("Authorization"))
}
