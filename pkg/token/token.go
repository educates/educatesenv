// Package token finds the GitHub token educatesenv authenticates with. It
// reads the token from the environment, from the gh CLI or from the config
// file, and never stores one.
package token

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// tokenSource is where a token came from.
type tokenSource int

const (
	notSet tokenSource = iota
	educatesGitHubToken
	ghToken
	githubToken
	ghAuthToken
	configFile
)

// sourceLabels name each source in config view and in error messages.
var sourceLabels = map[tokenSource]string{
	notSet:              "not set",
	educatesGitHubToken: "EDUCATES_GITHUB_TOKEN",
	ghToken:             "GH_TOKEN",
	githubToken:         "GITHUB_TOKEN",
	ghAuthToken:         "gh auth token",
	configFile:          "config file (deprecated)",
}

// tokenVariables are the environment variables that hold a token, in the
// order they are checked.
var tokenVariables = []struct {
	variable string
	source   tokenSource
}{
	{"EDUCATES_GITHUB_TOKEN", educatesGitHubToken},
	{"GH_TOKEN", ghToken},
	{"GITHUB_TOKEN", githubToken},
}

// configFileWarning is the warning for a token from the config file. Its
// verb is the file's path.
const configFileWarning = "Warning: github.token in %s is deprecated, and a later release will stop reading it. " +
	"Set GH_TOKEN or run 'gh auth login' instead, and remove the token from that file.\n"

// Token is a GitHub token and the source it came from. Formatting a Token
// with %s or %v shows its source, never any character of the token.
type Token struct {
	value  string
	source tokenSource
}

// Value returns the token, or "" when no source has one.
func (t Token) Value() string {
	return t.value
}

// IsSet reports whether a source had a token.
func (t Token) IsSet() bool {
	return t.value != ""
}

// Source returns the label of the token's source: EDUCATES_GITHUB_TOKEN,
// GH_TOKEN, GITHUB_TOKEN, gh auth token, config file (deprecated), or not
// set.
func (t Token) Source() string {
	return sourceLabels[t.source]
}

// String returns how config view shows the token: <set via GH_TOKEN> and
// the like, <set in config file, deprecated> or <not set>.
func (t Token) String() string {
	switch t.source {
	case notSet:
		return "<not set>"
	case configFile:
		return "<set in config file, deprecated>"
	default:
		return "<set via " + t.Source() + ">"
	}
}

// Resolve returns the token from the first of these sources that has one:
//
//  1. EDUCATES_GITHUB_TOKEN
//  2. GH_TOKEN
//  3. GITHUB_TOKEN
//  4. the output of `gh auth token --hostname github.com`, when gh is on
//     the PATH
//  5. fileToken, the github.token value of the config file at file
//
// A source whose token is empty has none. A token from the config file is
// deprecated, so Resolve writes a warning that names the file to warnings
// when it returns one.
func Resolve(file, fileToken string, warnings io.Writer) Token {
	for _, e := range tokenVariables {
		if value := os.Getenv(e.variable); value != "" {
			return Token{value: value, source: e.source}
		}
	}
	if value := fromGH(); value != "" {
		return Token{value: value, source: ghAuthToken}
	}
	if fileToken != "" {
		_, _ = fmt.Fprintf(warnings, configFileWarning, file)
		return Token{value: fileToken, source: configFile}
	}
	return Token{}
}

// fromGH returns the token gh holds for github.com, or "" when gh is not on
// the PATH, fails, or prints nothing. gh's own output on stderr is
// discarded.
func fromGH() string {
	out, err := exec.Command("gh", "auth", "token", "--hostname", "github.com").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
