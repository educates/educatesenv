package version

import (
	"runtime/debug"
)

var (
	// Version can be set via:
	// -ldflags="-X 'github.com/educates/educatesenv/pkg/version.Version=$TAG'"
	defaultVersion = "develop"
	Version        = ""
	moduleName     = "github.com/educates/educatesenv"
)

func init() {
	Version = version()
}

func version() string {
	if Version != "" {
		// Version was set via ldflags, just return it.
		return Version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return defaultVersion
	}

	// Anything else.
	for _, dep := range info.Deps {
		if dep.Path == moduleName {
			return dep.Version
		}
	}

	return defaultVersion
}
