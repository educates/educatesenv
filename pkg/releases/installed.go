package releases

import (
	"slices"
	"strings"

	"golang.org/x/mod/semver"
)

// SortInstalled returns the names of installed versions in the order
// list shows them: names that are semver versions, highest first, then
// the names that are not, alphabetically. A name is a semver version by
// the same rule that decides which tags Select considers. Versions that
// semver ranks equal, which differ only in build metadata, are ordered
// alphabetically.
func SortInstalled(names []string) []string {
	sorted := slices.Clone(names)
	slices.SortFunc(sorted, func(a, b string) int {
		aIsSemver, bIsSemver := isSemver(a), isSemver(b)
		switch {
		case aIsSemver && bIsSemver:
			if c := semver.Compare(semverForm(b), semverForm(a)); c != 0 {
				return c
			}
			return strings.Compare(a, b)
		case aIsSemver:
			return -1
		case bIsSemver:
			return 1
		default:
			return strings.Compare(a, b)
		}
	})
	return sorted
}
