package version

import (
	"fmt"

	"github.com/educates/educatesenv/pkg/releases"
)

// Latest is the version name that InstallVersion resolves to the highest
// stable release.
const Latest = "latest"

// ResolveLatest returns the tag of the highest stable release by semver,
// among every release of the repository.
func (m *Manager) ResolveLatest() (string, error) {
	gh, err := m.github()
	if err != nil {
		return "", err
	}
	published, err := gh.ListReleases()
	if err != nil {
		return "", err
	}

	latest, ok := releases.Latest(published)
	if !ok {
		return "", fmt.Errorf("no stable releases found in %s/%s. Run 'educatesenv list-remote --pre-releases' to see pre-releases", m.config.Github.Org, m.config.Github.Repository)
	}
	return latest.Tag, nil
}
