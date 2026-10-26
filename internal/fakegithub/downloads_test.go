package fakegithub_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/internal/fakegithub"
)

func get(t *testing.T, url string) (*http.Response, []byte, error) {
	t.Helper()
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	return resp, body, err
}

func TestChecksumsListsEachAssetsSHA256AsSha256sumDoes(t *testing.T) {
	sums := fakegithub.Checksums(
		fakegithub.Asset{Name: "educates-linux-amd64", Content: []byte("educates 3.8.0 binary")},
		fakegithub.Asset{Name: "educates-darwin-arm64", Content: []byte("educates 3.8.0")},
	)

	assert.Equal(t, "checksums.txt", sums.Name)
	assert.Equal(t, "21163080cb987ff30bc789604b421676c7065e47b4c77415af76459db9bfc4bd  educates-linux-amd64\n"+
		"5e990e9569bdb6a9a588600567e649ae1a63aa23b387ef9010b05dafada7657c  educates-darwin-arm64\n", string(sums.Content))
}

func TestCutsTheNextDownloadMidTransferAndServesLaterOnesWhole(t *testing.T) {
	srv := fakegithub.New(t)
	srv.AddReleases(fakegithub.Release{Tag: "3.8.0", Assets: []fakegithub.Asset{
		{Name: "educates-linux-amd64", Content: []byte("0123456789")},
	}})
	srv.CutNextDownload("3.8.0", "educates-linux-amd64")

	resp, body, err := get(t, srv.DownloadURL("3.8.0", "educates-linux-amd64"))
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int64(10), resp.ContentLength)
	assert.Equal(t, "01234", string(body))

	resp, body, err = get(t, srv.DownloadURL("3.8.0", "educates-linux-amd64"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "0123456789", string(body))
}
