package releases

// Latest returns the highest stable release in all by semver. It follows
// the rules Select follows without pre-releases: drafts, pre-releases and
// tags that are not semver never count. Neither the order of all nor
// publish times matter. ok is false when all holds no stable release.
func Latest(all []Release) (latest Release, ok bool) {
	chosen, _ := Select(all, Options{Limit: 1, Order: BySemver})
	if len(chosen) == 0 {
		return Release{}, false
	}
	return chosen[0], true
}
