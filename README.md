# educatesenv

A version manager for the `educates` binary, inspired by [tfenv](https://github.com/tfutils/tfenv). Easily install, switch, and manage multiple versions of the educates training platform binary.

## Features

- Version management for educates binary
- Cross-platform support
- Development mode for local testing
- Automated releases with goreleaser

## Installation

### Using go install

```bash
go install github.com/educates/educatesenv/cmd/educatesenv@latest
```

### From releases

Download the latest release from the [releases page](https://github.com/educates/educatesenv/releases).

## Usage

The `educatesenv` command provides several subcommands:

```bash
educatesenv version    # Display version information
educatesenv install    # Install a specific version
educatesenv use       # Switch to a specific version
```

## Development

### Prerequisites

- Go 1.21 or later
- Make
- [golangci-lint](https://golangci-lint.run/) for linting
- [goreleaser](https://goreleaser.com/) for releases

### Building and Testing

```bash
# Build the binary
make build

# Run tests
make test

# Run linter
make lint

# Local install
make install

# Test release build
make release-test
```

### Code Quality

We use golangci-lint for strict code quality enforcement. The configuration is in `.golangci.yml` and includes:

- Advanced linters enabled
- Strict formatting rules
- Comprehensive static analysis

Run the linter:
```bash
make lint
```

### GitHub Actions

The project uses GitHub Actions for automated workflows:

#### Build and Test Pipeline (`build.yml`)
- Triggered on pull requests and pushes to main/develop branches
- Matrix builds:
  - Ubuntu and macOS
  - amd64 and arm64 architectures (arm64 currently only on macOS)
- Steps:
  - Go 1.24 setup with dependency caching
  - Unit tests execution
  - Platform-specific binary builds
  - Binary artifacts upload

#### Linting Pipeline (`golangci-lint.yaml`)
- Triggered on pushes to main/develop branches
- Runs golangci-lint v2.1.5
- Performs comprehensive code quality checks
- Runs on Ubuntu latest

#### Release Pipeline (`release.yml`)
- Triggered on version tags (v*) or manual dispatch
- Uses goreleaser for automated releases
- Steps:
  - Go 1.24 setup
  - Creates GitHub releases
  - Builds for all supported platforms
  - Generates release artifacts

The workflow configurations are in `.github/workflows/`:
- `build.yml` - Build and test automation
- `golangci-lint.yaml` - Code quality checks
- `release.yml` - Release automation

### Releases

Releases are automated using goreleaser. The configuration is in `.goreleaser.yml` and handles:

- Cross-platform builds
- Checksums and signatures
- GitHub release automation
- Version information injection

### Project Structure

```
.
├── cmd/            # Command line interface
├── pkg/            # Core packages
│   ├── config/     # Configuration management
│   ├── github/     # GitHub API integration
│   ├── platform/   # Platform-specific code
│   └── version/    # Version management
├── .golangci.yml   # Linter configuration
└── .goreleaser.yml # Release configuration
```

### Contributing

Contributions are welcome! Please ensure you:

1. Follow the code style enforced by golangci-lint
2. Include tests for new functionality
3. Update documentation as needed
4. Run `make lint` and `make test` before submitting

## License

[Apache License 2.0](LICENSE)

---

## Supported OSes

- macOS (amd64, arm64)
- Linux (amd64, arm64)
- Windows (amd64) - *Support in progress*

---

## Installation

### Automatic (Recommended)

1. **Clone the repository:**
   ```sh
   git clone <your-repo-url> educatesenv
   cd educatesenv
   make build
   ```
2. **Initialize configuration and folders:**
   ```sh
   # Basic initialization
   ./educatesenv init
   
   # Initialize and download latest version
   ./educatesenv init --download
   
   # Initialize, download latest version and force reinstall if exists
   ./educatesenv init --download --overwrite
   ```
3. **Add the bin directory to your PATH:**
   - On Linux/macOS:
     ```sh
     echo 'export PATH="$HOME/.educatesenv/bin:$PATH"' >> ~/.bashrc
     source ~/.bashrc
     # or for zsh:
     echo 'export PATH="$HOME/.educatesenv/bin:$PATH"' >> ~/.zprofile
     source ~/.zprofile
     # or for fish:
     set -U fish_user_paths $HOME/.educatesenv/bin $fish_user_paths
     ```
   - On Windows (PowerShell):
     ```powershell
     [Environment]::SetEnvironmentVariable("Path", "$HOME\.educatesenv\bin;" + $Env:Path, [EnvironmentVariableTarget]::User)
     ```

### Manual

1. **Download or build the binary**
2. **Create the config and bin folders:**
   ```sh
   mkdir -p $HOME/.educatesenv/bin
   cp educatesenv $HOME/.educatesenv/bin/
   ```
3. **Add `$HOME/.educatesenv/bin` to your PATH** (see above)

---

## Usage

All commands are run via the `educatesenv` binary:

### Install a version
```sh
# Install a specific version
educatesenv install <version>

# Install and set as active version
educatesenv install <version> --use

# Force reinstall even if version exists
educatesenv install <version> --force

# Install latest version
educatesenv install latest

# Install latest version and set as active
educatesenv install latest --use

# Force reinstall latest version
educatesenv install latest --force
```
Downloads and installs the specified version (or latest version) of `educates` into the bin directory. Use `--use` to automatically set it as the active version after installation. Use `--force` to reinstall even if the version already exists.

### List installed versions
```sh
educatesenv list
```
Lists all installed `educates` binaries. The active version is marked with `*`.

### Use a version
```sh
educatesenv use <version>
```
Switches the active `educates` binary by updating the `educates` symlink in the bin directory.

### List remote versions
```sh
educatesenv list-remote [--skip-pre-releases]
```
Lists all available versions from the [educates GitHub releases](https://github.com/educates/educates-training-platform/releases). Use `--skip-pre-releases` to hide alpha, beta, and rc versions.

### Uninstall a version
```sh
educatesenv uninstall <version>
```