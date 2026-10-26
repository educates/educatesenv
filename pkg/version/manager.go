package version

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/educates/educatesenv/pkg/github"
	"github.com/educates/educatesenv/pkg/platform"
)

// Timeouts of the HTTP client that downloads release assets. Each one bounds
// a step that waits on the server before any of the body arrives. Nothing
// bounds the whole transfer, so a slow download that keeps making progress
// finishes.
const (
	connectTimeout        = 30 * time.Second
	tlsHandshakeTimeout   = 10 * time.Second
	responseHeaderTimeout = 30 * time.Second
)

// Manager handles version-related operations
type Manager struct {
	config *config.Config
	github *github.Client
	http   *http.Client
	out    io.Writer
	errOut io.Writer
}

// New creates a new version manager that reports progress to out and
// warnings to errOut
func New(cfg *config.Config, gh *github.Client, out, errOut io.Writer) *Manager {
	return &Manager{
		config: cfg,
		github: gh,
		http:   newDownloadClient(),
		out:    out,
		errOut: errOut,
	}
}

// newDownloadClient returns the HTTP client that downloads release assets.
// It times out on connecting, on the TLS handshake and on waiting for the
// response headers, and sets no limit on the whole transfer.
func newDownloadClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: connectTimeout}).DialContext,
			TLSHandshakeTimeout:   tlsHandshakeTimeout,
			ResponseHeaderTimeout: responseHeaderTimeout,
			ForceAttemptHTTP2:     true,
		},
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

	// A symlink is a development symlink when its target is not directly
	// in the managed bin directory, or lacks the binary prefix. The
	// directories are compared as cleaned paths, so a sibling directory
	// whose name starts with the bin directory's counts as outside.
	// filepath.Dir returns a cleaned path.
	inBinDir := filepath.Dir(target) == filepath.Clean(m.config.Local.Dir)
	return !inBinDir || !strings.HasPrefix(filepath.Base(target), platform.BinaryPrefix), nil
}

// UseVersion sets a version as active. When the version's binary is
// missing, the error wraps fs.ErrNotExist.
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

// InstallVersion installs a specific version of educates. Latest installs
// the highest stable release, and says which version that is.
func (m *Manager) InstallVersion(version string, force bool, activate bool) error {
	if version == Latest {
		resolved, err := m.ResolveLatest()
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(m.out, "Resolved latest to %s\n", resolved)
		version = resolved
	}

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
		_, _ = fmt.Fprintf(m.out, "Version %s is already installed.\n", version)
	} else {
		if versionExists {
			_, _ = fmt.Fprintf(m.out, "Reinstalling version %s...\n", version)
		} else {
			_, _ = fmt.Fprintf(m.out, "Installing version %s...\n", version)
		}

		assetName, err := m.GetPlatformBinaryName()
		if err != nil {
			return fmt.Errorf("failed to determine platform binary name: %w", err)
		}

		assets, err := m.github.GetReleaseAssets(version, assetName)
		if err != nil {
			return err // Pass through the user-friendly error from GitHub client
		}

		releasedSum, err := m.releasedChecksum(version, assetName, assets.ChecksumsURL)
		if err != nil {
			return err
		}

		if err := m.installBinary(assetName, assets.BinaryURL, releasedSum, binaryPath); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(m.out, "educates %s installed successfully.\n", version)
	}

	// Handle activation if requested
	if activate {
		if err := m.UseVersion(version); err != nil {
			return fmt.Errorf("installation succeeded but failed to set version %s as active: %w", version, err)
		}
		_, _ = fmt.Fprintf(m.out, "educates %s is now active.\n", version)
	}

	return nil
}

// createSymlink creates a symlink from source to target
func (m *Manager) createSymlink(source, target string) error {
	// Check if source exists
	if _, err := os.Stat(source); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("binary not found at %s: %w", source, fs.ErrNotExist)
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

// releasedChecksum returns the SHA-256 checksum that the checksums.txt of
// the release of version lists for assetName. When the release has no
// checksums.txt, it warns that the download cannot be verified and returns
// "".
func (m *Manager) releasedChecksum(version, assetName, checksumsURL string) (string, error) {
	if checksumsURL == "" {
		_, _ = fmt.Fprintf(m.errOut, "Warning: release %s has no %s, so %s is installed without verifying its checksum.\n", version, github.ChecksumsAssetName, assetName)
		return "", nil
	}

	var checksums bytes.Buffer
	if err := m.download(checksumsURL, &checksums); err != nil {
		return "", fmt.Errorf("failed to download %s (check your internet connection and try again): %w", github.ChecksumsAssetName, err)
	}
	sum, ok := findChecksum(checksums.String(), assetName)
	if !ok {
		return "", fmt.Errorf("%s has no entry for %s, so the binary cannot be verified as the one that was released", github.ChecksumsAssetName, assetName)
	}
	return sum, nil
}

// findChecksum returns the checksum that checksums, in the format sha256sum
// writes, lists for name.
func findChecksum(checksums, name string) (string, bool) {
	for line := range strings.Lines(checksums) {
		fields := strings.Fields(line)
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == name {
			return strings.ToLower(fields[0]), true
		}
	}
	return "", false
}

// installBinary downloads the binary assetName from url into a temporary
// file next to binaryPath, checks its SHA-256 checksum against releasedSum
// unless releasedSum is empty, makes it executable and only then renames it
// to binaryPath, replacing any binary already there. The temporary file is
// removed when any step fails, so binaryPath holds either the previous
// binary or the complete, verified new one.
func (m *Manager) installBinary(assetName, url, releasedSum, binaryPath string) (err error) {
	dir := filepath.Dir(binaryPath)
	// The leading dot keeps the temporary file out of the installed versions.
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(binaryPath)+"-*.download")
	if err != nil {
		return fmt.Errorf("failed to create a temporary file in %s: %w", dir, err)
	}
	defer func() {
		_ = tmp.Close()
		if err != nil {
			_ = os.Remove(tmp.Name())
		}
	}()

	_, _ = fmt.Fprintf(m.out, "Downloading %s...\n", url)
	hash := sha256.New()
	if err := m.download(url, io.MultiWriter(tmp, hash)); err != nil {
		return fmt.Errorf("failed to download binary (check your internet connection and try again): %w", err)
	}

	if releasedSum != "" {
		if sum := hex.EncodeToString(hash.Sum(nil)); sum != releasedSum {
			return fmt.Errorf("checksum mismatch for %s: %s lists %s, but the download has %s", assetName, github.ChecksumsAssetName, releasedSum, sum)
		}
		_, _ = fmt.Fprintf(m.out, "Checksum verified against %s.\n", github.ChecksumsAssetName)
	}

	if err := tmp.Chmod(0o755); err != nil {
		return fmt.Errorf("failed to set executable permissions on %s: %w", tmp.Name(), err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("failed to write %s: %w", tmp.Name(), err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to write %s: %w", tmp.Name(), err)
	}
	if err := os.Rename(tmp.Name(), binaryPath); err != nil {
		return fmt.Errorf("failed to move the downloaded binary to %s: %w", binaryPath, err)
	}
	return nil
}

// download writes the body of a GET of url to w.
func (m *Manager) download(url string, w io.Writer) error {
	resp, err := m.http.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, resp.Status)
	}

	_, err = io.Copy(w, resp.Body)
	return err
}
