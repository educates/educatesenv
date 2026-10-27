package version

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildVersion(t *testing.T) {
	mainModule := func(version string) *debug.BuildInfo {
		return &debug.BuildInfo{Main: debug.Module{Path: "github.com/educates/educatesenv", Version: version}}
	}

	tests := []struct {
		name    string
		ldflags string
		info    *debug.BuildInfo
		ok      bool
		want    string
	}{
		{name: "go install at a tag", info: mainModule("v1.1.0"), ok: true, want: "v1.1.0"},
		{name: "go install at a pseudo-version", info: mainModule("v1.0.1-0.20261026101500-0123456789ab"), ok: true, want: "v1.0.1-0.20261026101500-0123456789ab"},
		{name: "devel build", info: mainModule("(devel)"), ok: true, want: "develop"},
		{name: "main module without a version", info: mainModule(""), ok: true, want: "develop"},
		{name: "no build info", info: nil, ok: false, want: "develop"},
		{name: "ldflags win over build info", ldflags: "v1.1.0-rc.1", info: mainModule("v1.0.0"), ok: true, want: "v1.1.0-rc.1"},
		{name: "ldflags without build info", ldflags: "v1.1.0", info: nil, ok: false, want: "v1.1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, buildVersion(tt.ldflags, tt.info, tt.ok))
		})
	}
}
