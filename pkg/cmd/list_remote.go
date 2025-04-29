package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
	showAll     bool
	showRecents bool
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

var listRemoteCmd = &cobra.Command{
	Use:           "list-remote",
	Short:         "List all available versions from GitHub releases",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		releases, err := gh.ListReleases()
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
		fmt.Println("Available versions:")
		for _, version := range versions {
			fmt.Printf("- %s\n", version)
		}

		if len(versions) == 0 {
			fmt.Println("No versions found")
		}

		return nil
	},
}

func init() {
	listRemoteCmd.Flags().BoolVar(&showAll, "all", false, "Show all versions including pre-releases")
	listRemoteCmd.Flags().BoolVar(&showRecents, "recents", false, "Show only the 10 most recent versions")
	rootCmd.AddCommand(listRemoteCmd)
}
