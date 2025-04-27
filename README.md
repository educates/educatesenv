# educatesenv

A version manager for the `educates` binary, inspired by [tfenv](https://github.com/tfutils/tfenv). Easily install, switch, and manage multiple versions of the educates training platform binary.

---

## Supported OSes

- macOS (amd64, arm64)
- Linux (amd64, arm64)
- Windows (amd64)

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
   ./educatesenv init
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
educatesenv install <version>
```
Downloads and installs the specified version of `educates` into the bin directory.

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
educatesenv init
```
Creates the default config and bin folders, and prints instructions for adding the bin folder to your PATH.

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

Configuration is managed via a YAML file at `$HOME/.educatesenv/config.yaml` (or as set by the `EDUCATES_LOCAL_FOLDER` env var). Example:

```yaml
github:
  org: educates
  repository: educates-training-platform
  token: ""
local:
  folder: /home/youruser/.educatesenv/bin
```

You can override any value with environment variables:
- `EDUCATES_GITHUB_ORG`
- `EDUCATES_GITHUB_REPOSITORY`
- `EDUCATES_GITHUB_TOKEN`
- `EDUCATES_LOCAL_FOLDER`

---

## Environment Variables

- `EDUCATES_GITHUB_ORG`: GitHub org to find educates binaries (default: `educates`)
- `EDUCATES_GITHUB_REPOSITORY`: GitHub repository (default: `educates-training-platform`)
- `EDUCATES_GITHUB_TOKEN`: GitHub token for API access (optional, for higher rate limits)
- `EDUCATES_LOCAL_FOLDER`: Where binaries are downloaded and stored (default: `$HOME/.educatesenv/bin`)

---

## Development

- All CLI logic is implemented using [Cobra](https://github.com/spf13/cobra) and [Viper](https://github.com/spf13/viper).
- Commands are located in the `cmd/` directory.
- To add new commands or functionality, edit or add files in `cmd/`.

### Building

Use the provided Makefile:
```sh
make build
```
This will produce the `educatesenv` binary in the current directory.

---

## License

Apache License 2.0. See [LICENSE](LICENSE).

---

## Contributing

Contributions are welcome! Please open issues or pull requests.

---

## Credits

- Inspired by [tfenv](https://github.com/tfutils/tfenv)
- Uses [Cobra](https://github.com/spf13/cobra), [Viper](https://github.com/spf13/viper), and [go-github](https://github.com/google/go-github) 