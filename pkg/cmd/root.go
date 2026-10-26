package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/educates/educatesenv/pkg/github"
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
// version manager. Warnings and the manager's progress go to out.
func (d *deps) load(out io.Writer) error {
	d.cfg = config.New()
	if err := d.cfg.Load(out); err != nil {
		return err
	}

	var err error
	d.gh, err = github.New(d.cfg, github.WithBaseURL(d.opts.githubBaseURL))
	if err != nil {
		return err
	}

	d.manager = version.New(d.cfg, d.gh, out)
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
			return d.load(cmd.OutOrStdout())
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
