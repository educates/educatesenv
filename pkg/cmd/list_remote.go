package cmd

import (
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/educates/educatesenv/pkg/releases"
)

func newListRemoteCommand(d *deps) *cobra.Command {
	var (
		recent      int
		all         bool
		preReleases bool
		byDate      bool
		recents     bool
	)

	cmd := &cobra.Command{
		Use:   "list-remote",
		Short: "List the newest versions available from GitHub releases",
		Long: `List the versions available from GitHub releases, with the date each
was published. By default it shows the 10 newest stable releases,
highest version first.`,
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if recents {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Flag --recents has been deprecated, the 10 most recent releases are now shown by default")
			}
			if recent < 1 {
				return fmt.Errorf("--recent must be at least 1, got %d", recent)
			}

			published, err := d.gh.ListReleases()
			if err != nil {
				return err
			}

			opts := releases.Options{IncludePrereleases: preReleases, Limit: recent}
			if all {
				opts.Limit = releases.All
			}
			if byDate {
				opts.Order = releases.ByPublishTime
			}
			chosen, candidates := releases.Select(published, opts)
			if len(chosen) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No versions found")
				return nil
			}

			table := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(table, "VERSION\tRELEASED")
			for _, r := range chosen {
				_, _ = fmt.Fprintf(table, "%s\t%s\n", r.Tag, r.PublishedAt.UTC().Format(time.DateOnly))
			}
			_ = table.Flush()

			if len(chosen) < candidates {
				if preReleases {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Showing %d of %d releases. Use --all to see all.\n", len(chosen), candidates)
				} else {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Showing %d of %d stable releases. Use --all to see all, --pre-releases to include pre-releases.\n", len(chosen), candidates)
				}
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&recent, "recent", "n", 10, "Show the N newest releases")
	cmd.Flags().BoolVar(&all, "all", false, "Show every release instead of the newest")
	cmd.Flags().BoolVar(&preReleases, "pre-releases", false, "Include pre-releases")
	cmd.Flags().BoolVar(&byDate, "by-date", false, "Order by publish date instead of by version")
	// --recents is deprecated. pflag's MarkDeprecated would print its
	// warning through cobra's output writer, which is stdout once SetOut is
	// called, so the flag is hidden and RunE prints the warning to stderr.
	cmd.Flags().BoolVar(&recents, "recents", false, "Show only the 10 most recent versions")
	_ = cmd.Flags().MarkHidden("recents")
	cmd.MarkFlagsMutuallyExclusive("all", "recent")
	return cmd
}
