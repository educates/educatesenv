package platform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSupportedPlatform(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		arch     string
		expected bool
	}{
		{"darwin-amd64", "darwin", "amd64", true},
		{"darwin-arm64", "darwin", "arm64", true},
		{"linux-amd64", "linux", "amd64", true},
		{"linux-arm64", "linux", "arm64", true},
		// TODO: Implement Windows support
		// {"windows-amd64", "windows", "amd64", true},
		{"unsupported-os", "freebsd", "amd64", false},
		{"unsupported-arch", "linux", "386", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSupportedPlatform(tt.os, tt.arch)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPlatformBinaryName(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		arch     string
		expected string
	}{
		{"darwin-amd64", "darwin", "amd64", "educates-darwin-amd64"},
		{"darwin-arm64", "darwin", "arm64", "educates-darwin-arm64"},
		{"linux-amd64", "linux", "amd64", "educates-linux-amd64"},
		{"linux-arm64", "linux", "arm64", "educates-linux-arm64"},
		// TODO: Implement Windows support with .exe extension
		// {"windows-amd64", "windows", "amd64", "educates-windows-amd64.exe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPlatformBinaryName(tt.os, tt.arch)
			assert.Equal(t, tt.expected, result)
		})
	}
}
