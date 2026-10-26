package version

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
)

// UninstallVersions removes the installed versions given, reporting each
// one it removes. It checks every version before it removes any, and
// reports every version it cannot uninstall, so that one bad argument
// leaves all of them in place. The active version is refused unless force
// is set; with force, the educates symlink goes too, and no version is
// active afterwards.
func (m *Manager) UninstallVersions(versions []string, force bool) error {
	// A missing bin directory holds no versions.
	installed, active, err := m.Installed()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to read bin directory %s: %w", m.config.Local.Dir, err)
	}

	var (
		toRemove      []string
		problems      []error
		activeVersion string
	)
	seen := make(map[string]bool, len(versions))
	for _, version := range versions {
		if seen[version] {
			continue
		}
		seen[version] = true

		if version == Develop {
			problems = append(problems, errors.New("cannot uninstall develop: educatesenv does not manage the development binary"))
			continue
		}
		// Looking the version up among the installed ones keeps an
		// argument that holds a path from reaching outside the bin
		// directory.
		if !slices.Contains(installed, version) {
			problems = append(problems, fmt.Errorf("version %s is not installed. Run 'educatesenv list' to see installed versions", version))
			continue
		}
		if version == active {
			if !force {
				problems = append(problems, fmt.Errorf("version %s is active. Switch to another version with 'educatesenv use <version>' first, or pass --force to uninstall it", version))
				continue
			}
			activeVersion = version
		}
		toRemove = append(toRemove, version)
	}
	if len(problems) > 0 {
		return errors.Join(problems...)
	}

	for _, version := range toRemove {
		if version == activeVersion {
			// The symlink goes first, so that a failure leaves no symlink
			// pointing at a removed binary.
			if err := os.Remove(m.activeLink()); err != nil {
				return fmt.Errorf("failed to remove the educates symlink: %w", err)
			}
		}
		if err := os.Remove(m.binaryPath(version)); err != nil {
			return fmt.Errorf("failed to uninstall version %s: %w", version, err)
		}
		_, _ = fmt.Fprintf(m.out, "educates %s uninstalled.\n", version)
	}

	if activeVersion != "" {
		_, _ = fmt.Fprintln(m.out, "No educates version is active now. Run 'educatesenv use <version>' to select one.")
	}
	return nil
}
