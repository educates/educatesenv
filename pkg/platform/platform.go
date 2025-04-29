package platform

// Operating System constants
const (
	// Darwin represents macOS
	Darwin = "darwin"
	// Linux represents Linux OS
	Linux = "linux"
	// Windows represents Windows OS
	Windows = "windows"
)

// Architecture constants
const (
	// AMD64 represents x86_64 architecture
	AMD64 = "amd64"
	// ARM64 represents 64-bit ARM architecture
	ARM64 = "arm64"
)

// BinaryPrefix is the prefix for all binary names
const BinaryPrefix = "educates-"

// GetPlatformBinaryName returns the platform-specific binary name
func GetPlatformBinaryName(os, arch string) string {
	return BinaryPrefix + os + "-" + arch
}

// IsSupportedPlatform checks if the given OS and architecture combination is supported
func IsSupportedPlatform(os, arch string) bool {
	switch os {
	case Darwin:
		return arch == AMD64 || arch == ARM64
	case Linux:
		return arch == AMD64 || arch == ARM64
	default:
		return false
	}
}
