package cmd

import (
	"context"
	"fmt"
	"regexp"
	"sort"

	"github.com/google/go-github/v71/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var skipPreReleases bool

var listRemoteCmd = &cobra.Command{
	Use:   "list-remote",
	Short: "List remote educates versions available for install",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var client *github.Client
		if Config.Github.Token != "" {
			ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: Config.Github.Token})
			client = github.NewClient(oauth2.NewClient(ctx, ts))
		} else {
			client = github.NewClient(nil)
		}

		var allVersions []string
		opt := &github.ListOptions{PerPage: 50}
		page := 1
		for {
			opt.Page = page
			releases, resp, err := client.Repositories.ListReleases(ctx, Config.Github.Org, Config.Github.Repository, opt)
			if err != nil {
				return fmt.Errorf("failed to list releases: %w", err)
			}
			for _, rel := range releases {
				if rel.TagName != nil {
					tag := *rel.TagName
					if skipPreReleases {
						if isPreRelease(tag) {
							continue
						}
					}
					allVersions = append(allVersions, tag)
				}
			}
			if resp.NextPage == 0 {
				break
			}
			page = resp.NextPage
		}

		if len(allVersions) == 0 {
			fmt.Println("No educates versions found in remote releases.")
			return nil
		}

		sort.Sort(sort.Reverse(sort.StringSlice(allVersions)))
		for _, v := range allVersions {
			fmt.Println(v)
		}
		return nil
	},
}

func isPreRelease(tag string) bool {
	// Matches -alpha.<number>, -beta.<number>, -rc.<number> (case-insensitive)
	preRelPattern := regexp.MustCompile(`(?i)-(alpha|beta|rc)\.\d+$`)
	return preRelPattern.MatchString(tag)
}

func init() {
	listRemoteCmd.Flags().BoolVar(&skipPreReleases, "skip-pre-releases", false, "Skip pre-release versions (alpha, beta, rc)")
	rootCmd.AddCommand(listRemoteCmd)
}
