package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/educates/educatesenv/pkg/releases"
	"github.com/educates/educatesenv/pkg/token"
	"github.com/google/go-github/v93/github"
)

// Client wraps the GitHub client with our configuration
type Client struct {
	client *github.Client
	config *config.Config
	token  token.Token
}

// options holds the settings that Option values change.
type options struct {
	baseURL string
	token   token.Token
}

// Option changes how New builds a Client.
type Option func(*options)

// WithBaseURL sets the GitHub API base URL. The default is github.com's API.
func WithBaseURL(baseURL string) Option {
	return func(o *options) {
		o.baseURL = baseURL
	}
}

// WithToken authenticates the client's requests with tok. Without a set
// token, the client calls the API unauthenticated.
func WithToken(tok token.Token) Option {
	return func(o *options) {
		o.token = tok
	}
}

// New creates a new GitHub client for the repository cfg names
func New(cfg *config.Config, opts ...Option) (*Client, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	var clientOpts []github.ClientOptionsFunc
	if o.token.IsSet() {
		clientOpts = append(clientOpts, github.WithAuthToken(o.token.Value()))
	}
	if o.baseURL != "" {
		clientOpts = append(clientOpts, github.WithURLs(&o.baseURL, nil))
	}

	client, err := github.NewClient(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	return &Client{
		client: client,
		config: cfg,
		token:  o.token,
	}, nil
}

// ChecksumsAssetName is the name of the release asset that lists the
// SHA-256 checksum of each of the release's other assets.
const ChecksumsAssetName = "checksums.txt"

// ReleaseAssets holds the download URLs of the release assets that
// installing a version needs.
type ReleaseAssets struct {
	// BinaryURL is the download URL of the platform binary.
	BinaryURL string
	// ChecksumsURL is the download URL of checksums.txt. It is empty when
	// the release has no checksums.txt.
	ChecksumsURL string
}

// GetReleaseAssets looks up the release of a version by its tag, and
// returns the download URLs of its binaryName asset and of its
// checksums.txt.
func (c *Client) GetReleaseAssets(version, binaryName string) (ReleaseAssets, error) {
	release, _, err := c.client.Repositories.GetReleaseByTag(context.Background(), c.config.Github.Org, c.config.Github.Repository, version)
	if err != nil {
		notFound := fmt.Errorf("version %s not found. Run 'educatesenv list-remote' to see available versions", version)
		return ReleaseAssets{}, c.apiError(err, notFound, "failed to fetch release info")
	}

	var assets ReleaseAssets
	for _, a := range release.Assets {
		switch a.GetName() {
		case binaryName:
			assets.BinaryURL = a.GetBrowserDownloadURL()
		case ChecksumsAssetName:
			assets.ChecksumsURL = a.GetBrowserDownloadURL()
		}
	}
	if assets.BinaryURL == "" {
		return ReleaseAssets{}, fmt.Errorf("binary for %s is not available for your platform (%s). Please check supported platforms in the documentation", version, binaryName)
	}
	return assets, nil
}

// ListReleases returns every release of the repository, drafts included,
// in the order GitHub lists them. It reads every page, 100 releases at a
// time.
func (c *Client) ListReleases() ([]releases.Release, error) {
	var all []releases.Release
	opts := &github.ListOptions{PerPage: 100}
	for {
		page, resp, err := c.client.Repositories.ListReleases(context.Background(), c.config.Github.Org, c.config.Github.Repository, opts)
		if err != nil {
			notFound := fmt.Errorf("repository %s/%s not found", c.config.Github.Org, c.config.Github.Repository)
			return nil, c.apiError(err, notFound, "failed to fetch releases")
		}

		for _, rel := range page {
			all = append(all, releases.Release{
				Tag:              rel.GetTagName(),
				PublishedAt:      rel.GetPublishedAt().Time,
				GitHubPrerelease: rel.GetPrerelease(),
				Draft:            rel.GetDraft(),
			})
		}

		if resp.NextPage == 0 {
			return all, nil
		}
		opts.Page = resp.NextPage
	}
}

// apiError returns the error the user sees for err, the error of a GitHub
// API call. A rate limit says when it resets, and a token GitHub rejects
// with a 401 is named by its source. A 404 gives notFound. Any other error
// is wrapped, with failure saying what failed.
func (c *Client) apiError(err, notFound error, failure string) error {
	if rateLimit, ok := errors.AsType[*github.RateLimitError](err); ok {
		return c.rateLimitError(rateLimit.Rate.Reset.Time)
	}
	if response, ok := errors.AsType[*github.ErrorResponse](err); ok && response.Response != nil {
		switch status := response.Response.StatusCode; {
		case status == http.StatusUnauthorized && c.token.IsSet():
			message := response.Message
			if message == "" {
				message = http.StatusText(status)
			}
			return fmt.Errorf("GitHub rejected the token from %s: %d %s. Replace or remove that token and try again", c.token.Source(), status, message)
		case status == http.StatusNotFound:
			return notFound
		}
	}
	return fmt.Errorf("%s: %w", failure, err)
}

// rateLimitError returns the error for a rate limit that resets at reset.
// Without a token, it says how to authenticate, which raises the limit.
func (c *Client) rateLimitError(reset time.Time) error {
	resets := reset.UTC().Format(time.DateTime) + " UTC"
	if c.token.IsSet() {
		return fmt.Errorf("GitHub API rate limit exceeded for the token from %s; it resets at %s", c.token.Source(), resets)
	}
	return fmt.Errorf("GitHub API rate limit exceeded; it resets at %s. %s to authenticate and get a higher limit", resets, token.AuthenticationHint)
}
