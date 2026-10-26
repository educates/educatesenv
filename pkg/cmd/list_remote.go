package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// isPreRelease checks if a version string contains pre-release indicators
func isPreRelease(version string) bool {
	preReleaseIndicators := []string{
		"-alpha", "-beta", "-rc",
		".alpha.", ".beta.", ".rc.",
	}
	version = strings.ToLower(version)
	for _, indicator := range preReleaseIndicators {
		if strings.Contains(version, indicator) {
			return true
		}
	}
	return false
}

func newListRemoteCommand(d *deps) *cobra.Command {
	var (
		showAll     bool
		showRecents bool
	)

	cmd := &cobra.Command{
		Use:           "list-remote",
		Short:         "List all available versions from GitHub releases",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			releases, err := d.gh.ListReleases()
			if err != nil {
				return fmt.Errorf("failed to fetch releases: %w", err)
			}

			// Filter and collect versions
			var versions []string
			for _, rel := range releases {
				version := rel.GetTagName()
				if !showAll && (rel.GetPrerelease() || isPreRelease(version)) {
					continue
				}
				versions = append(versions, version)
			}

			// Sort versions in reverse order (newest first)
			sort.Sort(sort.Reverse(sort.StringSlice(versions)))

			// Limit to recent versions if requested
			if showRecents && len(versions) > 10 {
				versions = versions[:10]
			}

			// Print versions
			_, _ = fmt.Fprintln(out, "Available versions:")
			for _, version := range versions {
				_, _ = fmt.Fprintf(out, "- %s\n", version)
			}

			if len(versions) == 0 {
				_, _ = fmt.Fprintln(out, "No versions found")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&showAll, "all", false, "Show all versions including pre-releases")
	cmd.Flags().BoolVar(&showRecents, "recents", false, "Show only the 10 most recent versions")
	return cmd
}
