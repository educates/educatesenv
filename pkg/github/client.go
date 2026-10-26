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

// GetReleaseAssetURL gets the download URL for a specific version and platform
func (c *Client) GetReleaseAssetURL(version, assetName string) (string, error) {
	release, resp, err := c.client.Repositories.GetReleaseByTag(context.Background(), c.config.Github.Org, c.config.Github.Repository, version)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return "", fmt.Errorf("version %s not found. Run 'educatesenv list-remote' to see available versions", version)
		}
		return "", fmt.Errorf("failed to fetch release info: %w", err)
	}

	for _, a := range release.Assets {
		if a.GetName() == assetName {
			return a.GetBrowserDownloadURL(), nil
		}
	}
	return "", fmt.Errorf("binary for %s is not available for your platform (%s). Please check supported platforms in the documentation", version, assetName)
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
