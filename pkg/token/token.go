// Package token finds the GitHub token educatesenv authenticates with. It
// reads the token from the environment, from the gh CLI or from the config
// file, and never stores one.
package token

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

// source is where a token came from.
type source int

const (
	notSet source = iota
	educatesGitHubToken
	ghToken
	githubToken
	ghAuthToken
	configFile
)

// labels name each source in config view and in error messages.
var labels = map[source]string{
	notSet:              "not set",
	educatesGitHubToken: "EDUCATES_GITHUB_TOKEN",
	ghToken:             "GH_TOKEN",
	githubToken:         "GITHUB_TOKEN",
	ghAuthToken:         "gh auth token",
	configFile:          "config file (deprecated)",
}

// environment holds the environment variables that hold a token, in the
// order they are checked.
var environment = []struct {
	variable string
	source   source
}{
	{"EDUCATES_GITHUB_TOKEN", educatesGitHubToken},
	{"GH_TOKEN", ghToken},
	{"GITHUB_TOKEN", githubToken},
}

// Token is a GitHub token and the source it came from. Formatting a Token
// with %s or %v shows its source, never any character of the token.
type Token struct {
	value  string
	source source
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
	return labels[t.source]
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
// A source whose token is empty has none.
func Resolve(file, fileToken string, warnings io.Writer) Token {
	for _, e := range environment {
		if value := os.Getenv(e.variable); value != "" {
			return Token{value: value, source: e.source}
		}
	}
	if value := fromGH(); value != "" {
		return Token{value: value, source: ghAuthToken}
	}
	if fileToken != "" {
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
