// Package releases holds the rules about Educates releases: which ones
// take part in a listing, which ones are pre-releases, and how they are
// ordered. It does no I/O.
package releases

import (
	"slices"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

// Release is a GitHub release of Educates.
type Release struct {
	// Tag is the release's tag as published, such as "3.8.0".
	Tag string
	// PublishedAt is when the release was published.
	PublishedAt time.Time
	// GitHubPrerelease is GitHub's pre-release flag.
	GitHubPrerelease bool
	// Draft is GitHub's draft flag.
	Draft bool
}

// Order is the order Select returns releases in.
type Order int

const (
	// BySemver orders releases by version, highest first.
	BySemver Order = iota
	// ByPublishTime orders releases by publish time, newest first.
	ByPublishTime
)

// All is the Limit that chooses every candidate.
const All = 0

// Options are the choices Select makes.
type Options struct {
	// IncludePrereleases adds pre-releases to the candidates.
	IncludePrereleases bool
	// Limit is the most releases to choose. All, or any value below 1,
	// chooses every candidate.
	Limit int
	// Order is the order of the chosen releases.
	Order Order
}

// Select chooses releases from all. It returns the chosen releases, in
// the order opts names, and the number of candidates they were chosen
// from.
func Select(all []Release, opts Options) (chosen []Release, candidates int) {
	for _, r := range all {
		if r.Draft || !isSemver(r.Tag) {
			continue
		}
		if r.isPrerelease() && !opts.IncludePrereleases {
			continue
		}
		chosen = append(chosen, r)
	}
	slices.SortStableFunc(chosen, func(a, b Release) int {
		if opts.Order == ByPublishTime {
			return b.PublishedAt.Compare(a.PublishedAt)
		}
		return semver.Compare(version(b.Tag), version(a.Tag))
	})

	candidates = len(chosen)
	if opts.Limit > 0 && opts.Limit < candidates {
		chosen = chosen[:opts.Limit]
	}
	return chosen, candidates
}

// isPrerelease reports whether r is a pre-release: its tag has a semver
// pre-release part, or GitHub flags it.
func (r Release) isPrerelease() bool {
	return r.GitHubPrerelease || semver.Prerelease(version(r.Tag)) != ""
}

// version returns tag in the form golang.org/x/mod/semver reads, which
// needs a "v" prefix that Educates tags do not have.
func version(tag string) string {
	return "v" + tag
}

// isSemver reports whether tag is a full semver version, such as "3.8.0"
// or "4.0.0-alpha.13". The shorthands golang.org/x/mod/semver accepts,
// such as "3.8", are not.
func isSemver(tag string) bool {
	v := version(tag)
	return semver.Canonical(v) == strings.TrimSuffix(v, semver.Build(v))
}
