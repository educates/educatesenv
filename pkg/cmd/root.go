package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/educates/educatesenv/pkg/github"
	"github.com/educates/educatesenv/pkg/token"
	"github.com/educates/educatesenv/pkg/version"
)

// Option changes how NewRootCommand builds the command tree.
type Option func(*options)

type options struct {
	githubBaseURL       string
	downloadIdleTimeout time.Duration
}

// WithGitHubBaseURL points the GitHub client at baseURL instead of
// github.com's API.
func WithGitHubBaseURL(baseURL string) Option {
	return func(o *options) {
		o.githubBaseURL = baseURL
	}
}

// WithDownloadIdleTimeout makes a download fail when no data arrives for
// timeout, instead of the 60 seconds the version manager waits by default.
// A zero timeout keeps the default.
func WithDownloadIdleTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.downloadIdleTimeout = timeout
	}
}

// deps holds what the commands of one tree share. The root command loads
// it before any command runs.
type deps struct {
	opts    options
	cfg     *config.Config
	manager *version.Manager
	errOut  io.Writer

	// token is resolved on first use, by githubToken, and gh is created on
	// first use, by githubClient.
	token *token.Token
	gh    *github.Client
}

// load reads the configuration and creates the version manager. The
// warning about a missing home directory and the manager's progress go to
// out, and warnings about config values, the manager's warnings and token
// warnings to errOut. It neither resolves the GitHub token nor creates the
// GitHub client.
func (d *deps) load(out, errOut io.Writer) error {
	var err error
	d.cfg, err = config.Load(out, errOut)
	if err != nil {
		return err
	}

	d.errOut = errOut
	d.manager = version.New(d.cfg, d.githubClient, out, errOut, version.WithDownloadIdleTimeout(d.opts.downloadIdleTimeout))
	return nil
}

// githubToken returns the GitHub token, resolving it on first use. Only
// config view and the commands that call GitHub use it, so the others
// never look the token up.
func (d *deps) githubToken() token.Token {
	if d.token == nil {
		tok := token.Resolve(d.cfg.File(), d.cfg.Github.Token, d.errOut)
		d.token = &tok
	}
	return *d.token
}

// githubClient returns the GitHub client, creating it on first use.
func (d *deps) githubClient() (*github.Client, error) {
	if d.gh == nil {
		gh, err := github.New(d.cfg, github.WithBaseURL(d.opts.githubBaseURL), github.WithToken(d.githubToken()))
		if err != nil {
			return nil, err
		}
		d.gh = gh
	}
	return d.gh, nil
}

// cleanUpDevelopmentSymlink removes the educates symlink when it points
// at a development binary while development mode is disabled, and warns
// about it. Only the commands that deal with the active binary run it,
// as their PreRun hook.
func (d *deps) cleanUpDevelopmentSymlink(cmd *cobra.Command, _ []string) {
	if err := d.manager.ValidateDevelopmentMode(); err != nil {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: %v\n", err)
	}
}

// NewRootCommand builds the educatesenv command tree. Each call builds a
// new tree with its own flag values and dependencies.
func NewRootCommand(opts ...Option) *cobra.Command {
	d := &deps{}
	for _, opt := range opts {
		opt(&d.opts)
	}

	root := &cobra.Command{
		Use:   "educatesenv",
		Short: "Manage multiple versions of the educates binary",
		Long:  `A version manager for educates, inspired by tfenv.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return d.load(cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}

	root.AddCommand(
		newConfigCommand(d),
		newInitCommand(d),
		newInstallCommand(d),
		newListCommand(d),
		newListLocalCommand(d),
		newListRemoteCommand(d),
		newUninstallCommand(d),
		newUseCommand(d),
		newVersionCommand(),
	)
	return root
}
