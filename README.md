# educatesenv

A version manager for the `educates` binary, inspired by [tfenv](https://github.com/tfutils/tfenv). Easily install, switch, and manage multiple versions of the educates training platform binary.

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
Removes the specified version from the bin directory. If it was the active version, the symlink is removed and you are prompted to select a new one.

### Initialize configuration and folders
```sh
# Basic initialization
educatesenv init

# Initialize and download latest version
educatesenv init --download

# Initialize and force download latest version even if exists
educatesenv init --download --overwrite
```
Creates the default config and bin folders, and prints instructions for adding the bin folder to your PATH. Use `--download` to automatically download and set the latest version as active. Use `--overwrite` to force download even if the version already exists.

### Configuration management
- View current config:
  ```sh
  educatesenv config view
  ```
- Generate a default config file:
  ```sh
  educatesenv config init
  ```

---

## Configuration

Configuration is managed via a YAML file at `$HOME/.educatesenv/config.yaml` (or as set by the `EDUCATES_LOCAL_DIR` env var). Example:

```yaml
github:
  org: educates
  repository: educates-training-platform
  token: ""
local:
  dir: /home/youruser/.educatesenv/bin
development:
  enabled: false
  binaryLocation: ""
```

You can override any value with environment variables:
- `EDUCATES_GITHUB_ORG`
- `EDUCATES_GITHUB_REPOSITORY`
- `EDUCATES_GITHUB_TOKEN`
- `EDUCATES_LOCAL_DIR`
- `EDUCATES_DEVELOPMENT_ENABLED`
- `EDUCATES_DEVELOPMENT_BINARY_LOCATION`

---

## Environment Variables

- `EDUCATES_GITHUB_ORG`: GitHub org to find educates binaries (default: `educates`)
- `EDUCATES_GITHUB_REPOSITORY`: GitHub repository (default: `educates-training-platform`)
- `EDUCATES_GITHUB_TOKEN`: GitHub token for API access (optional, for higher rate limits)
- `EDUCATES_LOCAL_DIR`: Where binaries are downloaded and stored (default: `$HOME/.educatesenv/bin`)
- `EDUCATES_DEVELOPMENT_ENABLED`: Enable development mode (default: `false`)
- `EDUCATES_DEVELOPMENT_BINARY_LOCATION`: Path to development binary when in development mode

---

## Project Structure

```
educatesenv/
├── cmd/
│   └── educatesenv/      # Main entry point
│       └── main.go
├── pkg/
│   ├── cmd/              # Command implementations
│   │   └── root.go
│   ├── config/           # Configuration management
│   │   └── config.go
│   ├── github/           # GitHub API interactions
│   │   └── client.go
│   └── version/          # Version management
│       └── manager.go
├── .github/              # GitHub Actions workflows
├── Makefile             # Build automation
├── go.mod               # Go module definition
└── README.md            # This file
```

## Development

### Prerequisites

- Go 1.21 or later
- Make
- golangci-lint (for linting)

### Building and Testing

```sh
# Build the binary
make build

# Run tests
make test

# Run linter
make lint

# Install locally
make install
```

### Running Tests

The project includes tests for core functionality:
- Configuration management (`pkg/config`)
- Platform detection (`pkg/platform`)
- Version management (`pkg/version`)

Run all tests with:
```sh
make test
```

### Code Quality

We use golangci-lint for code quality checks. Run the linter with:
```sh
make lint
```

The configuration for golangci-lint is in `.golangci.yml`.

### Project TODOs

1. Platform Support:
   - [ ] Complete Windows support including `.exe` extension handling
   - [ ] Add tests for Windows-specific functionality

2. Testing:
   - [ ] Implement proper mocking for GitHub client in version manager tests
   - [ ] Add tests for the `cmd` package
   - [ ] Add tests for the `github` package
   - [ ] Improve test coverage for error cases

3. Features:
   - [ ] Add version validation before installation
   - [ ] Support for version ranges (e.g., ">=1.0.0")
   - [ ] Add offline mode support

4. Documentation:
   - [ ] Add godoc comments for all exported functions
   - [ ] Add examples in documentation
   - [ ] Add contribution guidelines

### Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to:
1. Update tests as appropriate
2. Update documentation
3. Follow the existing code style
4. Run tests and linting before submitting

---

## License

Apache License 2.0. See [LICENSE](LICENSE).

---

## Credits

- Inspired by [tfenv](https://github.com/tfutils/tfenv)
- Uses [Cobra](https://github.com/spf13/cobra), [Viper](https://github.com/spf13/viper), and [go-github](https://github.com/google/go-github)