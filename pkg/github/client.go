package github

import (
	"context"
	"fmt"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/educates/educatesenv/pkg/releases"
	"github.com/google/go-github/v93/github"
)

// Client wraps the GitHub client with our configuration
type Client struct {
	client *github.Client
	config *config.Config
}

// options holds the settings that Option values change.
type options struct {
	baseURL string
}

// Option changes how New builds a Client.
type Option func(*options)

// WithBaseURL sets the GitHub API base URL. The default is github.com's API.
func WithBaseURL(baseURL string) Option {
	return func(o *options) {
		o.baseURL = baseURL
	}
}

// New creates a new GitHub client, authenticated when the config holds a token
func New(cfg *config.Config, opts ...Option) (*Client, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	var clientOpts []github.ClientOptionsFunc
	if cfg.Github.Token != "" {
		clientOpts = append(clientOpts, github.WithAuthToken(cfg.Github.Token))
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
	release, resp, err := c.client.Repositories.GetReleaseByTag(context.Background(), c.config.Github.Org, c.config.Github.Repository, version)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return ReleaseAssets{}, fmt.Errorf("version %s not found. Run 'educatesenv list-remote' to see available versions", version)
		}
		return ReleaseAssets{}, fmt.Errorf("failed to fetch release info: %w", err)
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
			if resp != nil && resp.StatusCode == 404 {
				return nil, fmt.Errorf("repository %s/%s not found", c.config.Github.Org, c.config.Github.Repository)
			}
			return nil, fmt.Errorf("failed to fetch releases: %w", err)
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
