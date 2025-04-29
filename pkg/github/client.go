package github

import (
	"context"
	"fmt"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/google/go-github/v71/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub client with our configuration
type Client struct {
	client *github.Client
	config *config.Config
}

// New creates a new GitHub client with optional authentication
func New(cfg *config.Config) *Client {
	ctx := context.Background()
	var client *github.Client

	if cfg.Github.Token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.Github.Token})
		client = github.NewClient(oauth2.NewClient(ctx, ts))
	} else {
		client = github.NewClient(nil)
	}

	return &Client{
		client: client,
		config: cfg,
	}
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
		if rel.TagName != nil && !rel.GetPrerelease() {
			return *rel.TagName, nil
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

// ListReleases returns all releases from the repository
func (c *Client) ListReleases() ([]*github.RepositoryRelease, error) {
	releases, _, err := c.client.Repositories.ListReleases(context.Background(), c.config.Github.Org, c.config.Github.Repository, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	return releases, nil
}
