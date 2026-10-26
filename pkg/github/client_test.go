package github_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/educates/educatesenv/pkg/github"
)

func TestRequestsCarryTheConfiguredToken(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		authorization string
	}{
		{name: "with a token", token: "testtoken", authorization: "Bearer testtoken"},
		{name: "without a token", token: "", authorization: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var received []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				received = append(received, r.Header.Get("Authorization"))
				if r.URL.Path != "/repos/testorg/testrepo/releases" {
					http.NotFound(w, r)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[]`))
			}))
			defer server.Close()

			cfg := &config.Config{
				Github: config.GithubConfig{Org: "testorg", Repository: "testrepo", Token: tt.token},
			}
			client, err := github.New(cfg, github.WithBaseURL(server.URL))
			require.NoError(t, err)

			_, err = client.ListReleases()
			require.NoError(t, err)

			assert.Equal(t, []string{tt.authorization}, received)
		})
	}
}
