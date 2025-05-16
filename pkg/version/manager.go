package version

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/educates/educatesenv/pkg/github"
	"github.com/educates/educatesenv/pkg/platform"
)

// Manager handles version-related operations
type Manager struct {
	config *config.Config
	github *github.Client
}

// New creates a new version manager
func New(cfg *config.Config, gh *github.Client) *Manager {
	return &Manager{
		config: cfg,
		github: gh,
	}
}

// ValidateDevelopmentMode checks and cleans up development symlinks when development mode is disabled
func (m *Manager) ValidateDevelopmentMode() error {
	if m.config.Development.Enabled {
		return nil
	}

	symlinkPath := filepath.Join(m.config.Local.Dir, "educates")
	isDev, err := m.isDevSymlink(symlinkPath)
	if err != nil {
		return fmt.Errorf("failed to validate development mode: %w", err)
	}

	if isDev {
		if err := os.Remove(symlinkPath); err != nil {
			return fmt.Errorf("failed to remove development symlink: %w", err)
		}
		return fmt.Errorf("development mode is disabled; removed symlink to development binary. Please use 'educatesenv use <version>' to select a version")
	}

	return nil
}

// isDevSymlink checks if the symlink points to a development binary
func (m *Manager) isDevSymlink(symlinkPath string) (bool, error) {
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read symlink: %w", err)
	}

	// Resolve relative symlink if needed
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(symlinkPath), target)
	}

	// A symlink is considered a development symlink if:
	// 1. Development mode is disabled AND
	// 2. The target is outside the managed bin directory OR doesn't have the expected binary prefix
	return !strings.HasPrefix(target, m.config.Local.Dir) || !strings.HasPrefix(filepath.Base(target), platform.BinaryPrefix), nil
}

// UseVersion sets a version as active
func (m *Manager) UseVersion(version string) error {
	symlinkPath := filepath.Join(m.config.Local.Dir, "educates")

	// Handle development version
	if version == "develop" {
		if !m.config.Development.Enabled {
			return fmt.Errorf("development mode is not enabled. Enable it in the config file by setting development.enabled to true")
		}
		if m.config.Development.BinaryLocation == "" {
			return fmt.Errorf("development binary location is not set. Set development.binaryLocation in the config file")
		}
		return m.createSymlink(m.config.Development.BinaryLocation, symlinkPath)
	}

	// Handle regular version
	binaryPath := filepath.Join(m.config.Local.Dir, fmt.Sprintf("%s%s", platform.BinaryPrefix, version))
	err := m.createSymlink(binaryPath, symlinkPath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// GetPlatformBinaryName returns the platform-specific binary name
func (m *Manager) GetPlatformBinaryName() (string, error) {
	os, arch := runtime.GOOS, runtime.GOARCH

	if !platform.IsSupportedPlatform(os, arch) {
		return "", fmt.Errorf("unsupported platform: %s-%s", os, arch)
	}

	return platform.GetPlatformBinaryName(os, arch), nil
}

// InstallVersion installs a specific version of educates
func (m *Manager) InstallVersion(version string, force bool, activate bool) error {
	binDir := m.config.Local.Dir
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("failed to create bin directory %s: %w", binDir, err)
	}

	// Check if version already exists
	binaryPath := filepath.Join(binDir, fmt.Sprintf("%s%s", platform.BinaryPrefix, version))
	_, err := os.Stat(binaryPath)
	versionExists := err == nil

	// Handle installation
	if versionExists && !force {
		fmt.Printf("Version %s is already installed.\n", version)
	} else {
		if versionExists {
			fmt.Printf("Reinstalling version %s...\n", version)
		} else {
			fmt.Printf("Installing version %s...\n", version)
		}

		assetName, err := m.GetPlatformBinaryName()
		if err != nil {
			return fmt.Errorf("failed to determine platform binary name: %w", err)
		}

		downloadURL, err := m.github.GetReleaseAssetURL(version, assetName)
		if err != nil {
			return err // Pass through the user-friendly error from GitHub client
		}

		fmt.Printf("Downloading %s...\n", downloadURL)
		if err := m.downloadFile(downloadURL, binaryPath); err != nil {
			return fmt.Errorf("failed to download binary (check your internet connection and try again): %w", err)
		}
		if err := os.Chmod(binaryPath, 0o755); err != nil {
			return fmt.Errorf("failed to set executable permissions on %s: %w", binaryPath, err)
		}
		fmt.Printf("educates %s installed successfully.\n", version)
	}

	// Handle activation if requested
	if activate {
		if err := m.UseVersion(version); err != nil {
			return fmt.Errorf("installation succeeded but failed to set version %s as active: %w", version, err)
		}
		fmt.Printf("educates %s is now active.\n", version)
	}

	return nil
}

// createSymlink creates a symlink from source to target
func (m *Manager) createSymlink(source, target string) error {
	// Check if source exists
	if _, err := os.Stat(source); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("binary not found at %s", source)
		}
		return fmt.Errorf("failed to check binary: %w", err)
	}

	// Remove existing symlink if it exists
	if fi, err := os.Lstat(target); err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(target); err != nil {
				return fmt.Errorf("failed to remove existing symlink: %w", err)
			}
		} else {
			return fmt.Errorf("%s exists and is not a symlink", target)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check symlink: %w", err)
	}

	// Create new symlink
	relTarget, err := filepath.Rel(filepath.Dir(target), source)
	if err != nil {
		relTarget = source // fallback to absolute path
	}
	if err := os.Symlink(relTarget, target); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// downloadFile downloads a file from a URL to a local path
func (m *Manager) downloadFile(url, outPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("error closing response body: %w", cerr)
		}
	}()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("error closing output file: %w", cerr)
		}
	}()

	_, err = io.Copy(out, resp.Body)
	return err
}
