package version

import (
	"runtime/debug"
)

var (
	// Version can be set via:
	// -ldflags="-X 'github.com/educates/educatesenv/pkg/version.Version=$TAG'"
	defaultVersion = "develop"
	Version        = ""
)

func init() {
	info, ok := debug.ReadBuildInfo()
	Version = buildVersion(Version, info, ok)
}

// buildVersion returns the version a binary reports. A version set through
// ldflags wins. Otherwise it is the main module's version from the build
// info, which `go install module@version` records, and "develop" when the
// build info has none or records "(devel)".
func buildVersion(ldflagsVersion string, info *debug.BuildInfo, ok bool) string {
	if ldflagsVersion != "" {
		return ldflagsVersion
	}

	if !ok || info == nil {
		return defaultVersion
	}

	switch info.Main.Version {
	case "", "(devel)":
		return defaultVersion
	default:
		return info.Main.Version
	}
}
