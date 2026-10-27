# educatesenv

A version manager for the `educates` CLI of the
[Educates training platform](https://github.com/educates/educates-training-platform),
inspired by [tfenv](https://github.com/tfutils/tfenv). It installs
`educates` binaries from the platform's GitHub releases, keeps several
versions side by side, and switches between them.

educatesenv keeps everything under `~/.educatesenv`. Each installed
version is a binary named `educates-<version>` in the bin directory, and
an `educates` symlink in the same directory points at the active one:

```
~/.educatesenv/
├── config.yaml
└── bin/
    ├── educates -> educates-3.8.0
    ├── educates-3.7.3
    └── educates-3.8.0
```

With the bin directory on your `PATH`, `educates` runs whichever version
you selected last.

educatesenv runs on macOS and Linux, on amd64 and arm64.

## Installation

### From a release

Each release publishes one binary per platform, named
`educatesenv-<os>-<arch>`, and a `checksums.txt` with their SHA-256
checksums. Download the binary for your platform into the bin directory,
for example on macOS on Apple silicon:

```sh
mkdir -p ~/.educatesenv/bin
cd ~/.educatesenv/bin
curl -fLO https://github.com/educates/educatesenv/releases/latest/download/educatesenv-darwin-arm64
curl -fLO https://github.com/educates/educatesenv/releases/latest/download/checksums.txt
shasum -a 256 --ignore-missing -c checksums.txt
mv educatesenv-darwin-arm64 educatesenv
chmod +x educatesenv
rm checksums.txt
```

The other binaries are `educatesenv-darwin-amd64`,
`educatesenv-linux-amd64` and `educatesenv-linux-arm64`. On Linux,
`sha256sum --ignore-missing -c checksums.txt` checks the download.

Any directory on your `PATH` works. `~/.educatesenv/bin` has the
advantage that `educatesenv` and `educates` share one `PATH` entry.

#### macOS and the quarantine attribute

The macOS binaries are signed and notarized, but a release can ship them
unsigned when signing is unavailable. macOS refuses to run an unsigned
binary that a browser downloaded, because the browser marks the file with
the quarantine attribute. Clear the attribute to run it:

```sh
xattr -d com.apple.quarantine ~/.educatesenv/bin/educatesenv
```

Downloads made with `curl`, as above, are not quarantined and need no
such step, and neither do the `educates` binaries that educatesenv
downloads.

### With go install

With Go 1.26 or later:

```sh
go install github.com/educates/educatesenv/cmd/educatesenv@latest
```

This puts `educatesenv` in `$(go env GOPATH)/bin`, and
`educatesenv version` reports the module version you installed.

### Setting up

Run `init` once. It creates `~/.educatesenv/config.yaml`, readable only
by you, and the bin directory, and prints how to add the bin directory to
your `PATH` in your shell:

```sh
~/.educatesenv/bin/educatesenv init
```

Run it as plain `educatesenv init` when `educatesenv` is on your `PATH`
already, as it is after `go install`.

The instructions follow `$SHELL`: zsh users are told to edit `~/.zshrc`,
bash users `~/.bashrc`, and fish users get a `fish_user_paths` command.
For zsh, they read:

```sh
echo 'export PATH="/Users/you/.educatesenv/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

`init` never replaces a config file that exists. With `--download`, it
also installs the latest stable release of `educates` and makes it
active, and `--overwrite` makes it reinstall that release when it is
already installed:

```sh
educatesenv init --download
```

## Usage

```
educatesenv init [--download] [--overwrite]
educatesenv list-remote [--recent N | --all] [--pre-releases] [--by-date]
educatesenv install <version|latest> [--use] [--overwrite]
educatesenv use <version|develop>
educatesenv list
educatesenv uninstall <version>... [--force]
educatesenv config init
educatesenv config view
educatesenv version
educatesenv completion <bash|zsh|fish|powershell>
```

Every command takes `--help`. Commands that take no arguments fail when
given one, so a mistyped flag never does something else silently.

### Listing available versions

`list-remote` shows the 10 newest stable releases, highest version first,
with the date each was published, in UTC:

```console
$ educatesenv list-remote
VERSION  RELEASED
3.8.0    2026-08-26
3.7.3    2026-07-17
3.7.2    2026-06-30
3.7.1    2026-04-12
3.7.0    2026-03-30
3.6.1    2026-03-19
3.6.0    2026-03-04
3.5.1    2025-12-12
3.5.0    2025-12-01
3.4.0    2025-09-29
Showing 10 of 66 stable releases. Use --all to see all, --pre-releases to include pre-releases.
```

Two independent choices shape the list. Which releases: stable ones by
default, or stable ones and pre-releases with `--pre-releases`. How
many: 10 by default, `--recent N` (or `-n N`) for the N newest, or
`--all` for every one. `--all` and `--recent` conflict, and `--recent`
takes 1 or more. `--by-date` orders the list by publish date instead of
by version, and combines with every other flag:

```sh
# The 20 newest stable releases
educatesenv list-remote --recent 20

# Every stable release
educatesenv list-remote --all

# The 10 highest versions, pre-releases included
educatesenv list-remote --pre-releases

# Every release
educatesenv list-remote --pre-releases --all

# The 5 most recently published releases, pre-releases included
educatesenv list-remote --pre-releases --by-date -n 5
```

The table goes to stdout. The line that says how many releases are shown
out of how many goes to stderr, so piping `list-remote` into another tool
passes only the table. When the list holds every release that matches,
there is no such line.

educatesenv reads every release of the repository. A release counts as a
pre-release when its tag has a semver pre-release part, such as `-rc.1`
or `-alpha.13`, or when GitHub flags it as one. Draft releases are never
listed, and neither are tags that are not full semver versions without a
`v` prefix, the form Educates tags take.

The `--recents` flag is deprecated. It still parses, prints a
deprecation warning on stderr and changes nothing, since the 10 newest
releases are the default. A later release removes it.

### Installing a version

```sh
educatesenv install 3.8.0
educatesenv install latest --use
```

`install` downloads the release's `educates` binary for your platform and
saves it in the bin directory as `educates-<version>`. `latest` is the
highest stable release by semver, the same release `init --download`
picks, and `install` prints which version it resolved to. `--use` makes
the version active once it is installed, and `--overwrite` downloads it
again when it is already installed.

Every download is checked against the SHA-256 checksum the release lists
in its `checksums.txt`:

- A mismatch fails the install.
- A `checksums.txt` that does not list the binary fails the install too,
  since the binary cannot be verified.
- A release without a `checksums.txt` prints a warning on stderr and
  installs the binary unverified.

The binary downloads into a temporary file in the bin directory and is
renamed into place only once it is complete and verified. A failed or
interrupted download leaves nothing behind, and a version that was
already installed stays as it was. A download times out when the server
takes more than 30 seconds to accept the connection, 10 seconds to
complete the TLS handshake or 30 seconds to start its response, and it
fails as stalled when no data arrives for 60 seconds. Nothing caps the
transfer as a whole, so a slow download that keeps making progress
finishes.

### Switching versions

```sh
educatesenv use 3.7.3
```

`use` points the `educates` symlink at an installed version. When the
version is not installed, it says so and how to install it; any other
failure, such as a bin directory it cannot write to, is reported with its
cause. `use` does not resolve `latest`: to install the latest stable
release and switch to it, run `educatesenv install latest --use`.
`use develop` selects the development binary, as described under
[Development mode](#development-mode).

### Listing installed versions

```console
$ educatesenv list
Installed versions:
* 3.8.0 (active)
  3.7.3
```

`list` shows the installed versions newest first by semver, and marks
the active one with `*`. Installed binaries whose names are not semver
versions come last, in alphabetical order. With development mode
enabled, the `develop` entry comes first.

### Uninstalling versions

```sh
educatesenv uninstall 3.7.0 3.7.1
```

`uninstall` removes one or more installed versions. It checks every
version before it removes any, so a typo in one argument removes
nothing. It refuses:

- a version that is not installed;
- `develop`, since educatesenv does not manage the development binary;
- the active version, unless you pass `--force`. With `--force`, the
  `educates` symlink is removed with the binary, and no version is active
  until you run `educatesenv use` again.

### Other commands

`config init` creates the default config file, as `init` does, and fails
when one exists. `config view` prints the configuration in effect, with
the GitHub token shown as where it comes from, never as the token.

`version` prints the version of educatesenv itself, and `completion`
generates a shell completion script.

## Configuration

educatesenv reads its configuration from `~/.educatesenv/config.yaml`
and nowhere else: a `config.yaml` in the working directory is ignored.
The default file holds:

```yaml
github:
    org: educates
    repository: educates-training-platform
local:
    dir: /Users/you/.educatesenv/bin
development:
    enabled: false
    binaryLocation: ""
```

- `github.org` and `github.repository` name the repository whose releases
  educatesenv lists and installs.
- `local.dir` is the bin directory, where installed versions and the
  `educates` symlink live. `init` always creates `~/.educatesenv/bin` and
  prints `PATH` instructions for it, so when you set `local.dir`
  elsewhere, put that directory on your `PATH` instead.
- `development.enabled` and `development.binaryLocation` set up
  [development mode](#development-mode).

Keys match regardless of case, so `binarylocation` works as well as
`binaryLocation`. These environment variables override the file, and an
empty one counts as not set:

| Variable | Overrides |
|---|---|
| `EDUCATES_GITHUB_ORG` | `github.org` |
| `EDUCATES_GITHUB_REPOSITORY` | `github.repository` |
| `EDUCATES_LOCAL_DIR` | `local.dir` |
| `EDUCATES_DEVELOPMENT_ENABLED` | `development.enabled` |
| `EDUCATES_DEVELOPMENT_BINARY_LOCATION` | `development.binaryLocation` |

Booleans take `true` or `false`, and the other forms Go's
`strconv.ParseBool` accepts, such as `1` and `0`. In the config file, a
number other than `0` is true as well. A value in any other form, such
as `yes`, reads as false, and educatesenv prints a warning on stderr that
names the key and file, or the variable.

### GitHub token

educatesenv works without a token, but GitHub allows unauthenticated
clients only 60 API requests an hour, so a shared network address can
run out. With a token, the limit is far higher. educatesenv takes the
token from the first of these sources that has one:

1. `EDUCATES_GITHUB_TOKEN`, to give educatesenv a token of its own;
2. `GH_TOKEN`;
3. `GITHUB_TOKEN`;
4. the GitHub CLI: when `gh` is on your `PATH`, the output of
   `gh auth token --hostname github.com`;
5. `github.token` in the config file, which is deprecated.

If you use the GitHub CLI, `gh auth login` is all the setup educatesenv
needs. It always asks `gh` for its github.com token, so a `GH_HOST` set
for another GitHub server does not change which token it gets. A `gh`
that is not logged in, fails or prints nothing counts as no token.

educatesenv never writes a token to disk. Only the commands that call
GitHub, `list-remote`, `install` and `init --download`, look the token
up, and `config view`, which looks it up to show where it comes from.
The other commands never run `gh`. `config view` shows the source in
place of the token, as one of `<set via EDUCATES_GITHUB_TOKEN>`,
`<set via GH_TOKEN>`, `<set via GITHUB_TOKEN>`,
`<set via gh auth token>`, `<set in config file, deprecated>` or
`<not set>`.

When GitHub rate-limits a request, the error says when the limit resets,
and without a token, how to authenticate. When GitHub rejects a token,
the error names the source it came from, so you know which one to fix.

#### The config-file token is deprecated

A token in `github.token` still works, but only as the last source, and
using it prints a warning on stderr that names the file. A later release
stops reading it. To move off it, set `GH_TOKEN` or run `gh auth login`,
then delete the `token` line from `~/.educatesenv/config.yaml`.
educatesenv no longer writes a `token` key into the files it creates.

### Development mode

Development mode lets the `educates` symlink point at a binary you built
yourself:

```yaml
development:
    enabled: true
    binaryLocation: /path/to/your/build/educates
```

With it enabled, `educatesenv use develop` selects that binary, and
`list` shows it first as `develop -> <path>`. When you disable
development mode while the symlink still points at a development binary,
the next `list`, `use`, `install` or `init` removes the symlink and warns
you to select a version with `educatesenv use <version>`. Other commands
never touch the symlink.

## Development

### Prerequisites

- Go 1.26 or later. go.mod sets the language version to Go 1.26.0 and
  the toolchain to go1.26.10, which CI and release builds use. A local Go
  older than 1.26.10 downloads that toolchain, and a newer one builds
  with itself.
- Make.
- [golangci-lint](https://golangci-lint.run/) v2.15.0, the version CI
  runs.
- [GoReleaser](https://goreleaser.com/) v2.19.0, only to try the release
  build locally.

### Building and testing

```sh
make build    # build bin/educatesenv, with the version from git describe
make test     # run the tests
make lint     # run golangci-lint
make install  # build and copy the binary to ~/.educatesenv/bin
make clean    # remove bin/
make help     # list the targets
```

The tests need no network access, no real `gh` and no real home
directory. They run each command in-process against a fake GitHub
server, `internal/fakegithub`, with a temporary home directory and a
`PATH` that holds at most a fake `gh` script, through the harness in
`internal/clitest`.

### Project structure

```
.
├── cmd/educatesenv/   # main: builds the command tree and runs it
├── internal/
│   ├── clitest/       # runs commands in-process for tests
│   └── fakegithub/    # fake GitHub API and release downloads for tests
├── pkg/
│   ├── cmd/           # the cobra commands
│   ├── config/        # config file, environment overrides and paths
│   ├── github/        # GitHub API client and error messages
│   ├── platform/      # supported platforms and binary names
│   ├── releases/      # release selection and ordering rules, no I/O
│   ├── token/         # GitHub token resolution
│   └── version/       # install, use, uninstall and build version
├── .golangci.yml      # linter configuration
└── .goreleaser.yml    # release configuration
```

### CI

The workflows in `.github/workflows` take the Go version from go.mod:

- `build.yml` runs the tests and builds the binary on Ubuntu and macOS,
  on pushes and pull requests to `main` and `develop`.
- `golangci-lint.yaml` runs golangci-lint v2.15.0 on the same pushes and
  pull requests.
- `release.yml` builds and publishes releases, as described below.

Every action is pinned by the commit SHA of a release, with the release
tag in a comment, and Dependabot updates the pins weekly. The
golangci-lint and GoReleaser versions are action inputs, which Dependabot
does not update, so they change by hand.

## Releasing

Pushing a tag such as `v1.1.0` runs the release workflow:

```sh
git tag v1.1.0
git push origin v1.1.0
```

GoReleaser v2.19.0 builds the binaries for macOS and Linux on amd64 and
arm64, names them `educatesenv-<os>-<arch>`, writes `checksums.txt`,
signs and notarizes the macOS binaries, and creates a draft release.

### Signing and notarization

The macOS binaries are signed with the "Developer ID Application: Jorge
Morales Pou (XNSN5MMC5C)" certificate, which was issued by Apple's
Developer ID Certification Authority G2 intermediate and expires on
2030-05-16. Renew it before then. Apple's retirement of the original
Developer ID intermediate on 2027-02-01 does not affect it.

Signing a binary needs a valid certificate, and notarizing it needs an
active Apple Developer Program membership: when the membership lapses,
Apple's notary service stops accepting submissions. Binaries signed
before the certificate or the membership expired keep working; a revoked
certificate stops them from launching.

The workflow reads the certificate and the notary key from these
repository secrets:

| Secret | Holds |
|---|---|
| `MACOS_SIGN_P12` | the base64 of the `.p12` certificate |
| `MACOS_SIGN_PASSWORD` | the password of the `.p12` certificate |
| `MACOS_NOTARY_KEY` | the base64 of the App Store Connect `.p8` key |
| `MACOS_NOTARY_KEY_ID` | the ID of the `.p8` key |
| `MACOS_NOTARY_ISSUER_ID` | the issuer UUID of the `.p8` key |

Every release signs and notarizes the macOS binaries, except a manual run
with the `skip-signing` input described below. Its first step checks the
five secrets and fails, naming the ones that are empty or missing. The
run also fails when signing fails or when the notarization submission
fails. GoReleaser signs and notarizes before it creates the draft
release, so a failed run leaves the tag but no release behind, and a
manual run with `skip-signing` can release that tag. A run that fails
later, while uploading, leaves a draft, which the next run for the same
tag deletes and creates again. A snapshot build,
such as `goreleaser release --snapshot --clean`, signs only when
`MACOS_SIGN_P12` is set, so it builds without the certificate.

### Publishing the draft

The workflow submits the macOS binaries for notarization without waiting
for Apple's verdict, so a release stays a draft until Apple's
notarization email arrives. Once the email confirms that notarization
succeeded, edit the draft's notes and publish it.

### Releasing unsigned

When the membership or the certificate is unavailable, release without
signing from the workflow's manual trigger. In the Actions tab, run the
Release workflow with `release_tag` set to the tag and `skip-signing`
ticked, or with the GitHub CLI:

```sh
gh workflow run release.yml -f release_tag=v1.1.0 -f skip-signing=true
```

GoReleaser then runs with `--skip=notarize`, and every platform ships,
the macOS binaries unsigned. No secret needs to change. There is no
notarization email to wait for, so publish the draft once its notes are
ready, and point macOS users who download in a browser to the quarantine
note under [Installation](#macos-and-the-quarantine-attribute). Only the
manual trigger offers `skip-signing`.

## Contributing

Contributions are welcome. Before you open a pull request, run
`make lint` and `make test`, and add tests for new behavior.

## License

[Apache License 2.0](LICENSE)
