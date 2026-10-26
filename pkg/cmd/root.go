package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/educates/educatesenv/pkg/github"
	"github.com/educates/educatesenv/pkg/token"
	"github.com/educates/educatesenv/pkg/version"
)

// Option changes how NewRootCommand builds the command tree.
type Option func(*options)

type options struct {
	githubBaseURL string
}

// WithGitHubBaseURL points the GitHub client at baseURL instead of
// github.com's API.
func WithGitHubBaseURL(baseURL string) Option {
	return func(o *options) {
		o.githubBaseURL = baseURL
	}
}

// deps holds what the commands of one tree share. The root command loads
// it before any command runs.
type deps struct {
	opts    options
	cfg     *config.Config
	gh      *github.Client
	manager *version.Manager
}

// load reads the configuration and creates the GitHub client and the
// version manager. Config warnings and the manager's progress go to out,
// and the manager's warnings to errOut.
func (d *deps) load(out, errOut io.Writer) error {
	var err error
	d.cfg, err = config.Load(out)
	if err != nil {
		return err
	}

	tok := token.Resolve(d.cfg.File(), d.cfg.Github.Token, errOut)
	d.gh, err = github.New(d.cfg, github.WithBaseURL(d.opts.githubBaseURL), github.WithToken(tok))
	if err != nil {
		return err
	}

	d.manager = version.New(d.cfg, d.gh, out, errOut)
	return nil
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
