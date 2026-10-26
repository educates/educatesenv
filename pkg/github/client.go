package github

import (
	"context"
	"fmt"

	"github.com/educates/educatesenv/pkg/config"
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

// GetLatestReleaseVersion returns the latest stable release version
func (c *Client) GetLatestReleaseVersion() (string, error) {
	releases, resp, err := c.client.Repositories.ListReleases(context.Background(), c.config.Github.Org, c.config.Github.Repository, &github.ListOptions{PerPage: 10})
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return "", fmt.Errorf("repository %s/%s not found", c.config.Github.Org, c.config.Github.Repository)
		}
		return "", fmt.Errorf("failed to fetch releases: %w", err)
	}

	if len(releases) == 0 {
		return "", fmt.Errorf("no releases found in %s/%s", c.config.Github.Org, c.config.Github.Repository)
	}

	for _, rel := range releases {
		if rel.GetTagName() != "" && !rel.GetPrerelease() {
			return rel.GetTagName(), nil
		}
	}
	return "", fmt.Errorf("no stable releases found in %s/%s. Try 'educatesenv list-remote --all' to see pre-releases", c.config.Github.Org, c.config.Github.Repository)
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

// ListReleases returns all releases from the repository
func (c *Client) ListReleases() ([]*github.RepositoryRelease, error) {
	releases, _, err := c.client.Repositories.ListReleases(context.Background(), c.config.Github.Org, c.config.Github.Repository, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	return releases, nil
}
