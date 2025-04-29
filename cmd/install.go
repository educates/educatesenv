package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	useAfterInstall bool
	forceInstall    bool
)

var installCmd = &cobra.Command{
	Use:   "install [version|latest]",
	Short: "Install a specific educates version or the latest version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		requestedVersion := args[0]
		version := requestedVersion

		if requestedVersion == "latest" {
			fmt.Println("Fetching latest educates version...")
			latest, err := getLatestReleaseVersion()
			if err != nil {
				return fmt.Errorf("failed to get latest release version: %w", err)
			}
			fmt.Printf("Latest version: %s\n", latest)
			version = latest
		}

		if err := installAndOptionallyUse(version, forceInstall, useAfterInstall); err != nil {
			return fmt.Errorf("failed to install version %s: %w", version, err)
		}
		return nil
	},
}

func downloadFile(url, outPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func init() {
	installCmd.Flags().BoolVar(&useAfterInstall, "use", false, "Set the installed version as active")
	installCmd.Flags().BoolVar(&forceInstall, "force", false, "Force reinstall even if the version already exists")
	rootCmd.AddCommand(installCmd)
}
