package version

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/educates/educatesenv/pkg/platform"
)

// Installed returns the versions installed in the bin directory, in no
// particular order, and the active one, the version the educates symlink
// points at. The active version is one of versions, the development
// binary's name develop when development mode is enabled and the symlink
// points at that binary, or "" when no version is active. Installed fails
// when the bin directory cannot be read; a missing bin directory gives an
// error that wraps fs.ErrNotExist.
func (m *Manager) Installed() (versions []string, active string, err error) {
	entries, err := os.ReadDir(m.config.Local.Dir)
	if err != nil {
		return nil, "", err
	}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "educates" {
			continue
		}
		if version, ok := strings.CutPrefix(entry.Name(), platform.BinaryPrefix); ok {
			versions = append(versions, version)
		}
	}
	return versions, m.activeVersion(versions), nil
}

// activeVersion returns the version among installed whose binary the
// educates symlink resolves to, or develop when development mode is
// enabled and the symlink resolves to the development binary. It returns
// "" when the symlink resolves to none of them, or does not resolve.
func (m *Manager) activeVersion(installed []string) string {
	active, err := os.Stat(filepath.Join(m.config.Local.Dir, "educates"))
	if err != nil {
		return ""
	}
	isActive := func(path string) bool {
		binary, err := os.Stat(path)
		return err == nil && os.SameFile(active, binary)
	}

	for _, version := range installed {
		if isActive(filepath.Join(m.config.Local.Dir, platform.BinaryPrefix+version)) {
			return version
		}
	}
	if m.config.Development.Enabled && m.config.Development.BinaryLocation != "" && isActive(m.config.Development.BinaryLocation) {
		return "develop"
	}
	return ""
}
