package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/go-github/v71/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var installCmd = &cobra.Command{
	Use:   "install [version]",
	Short: "Install a specific educates version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version := args[0]
		binDir := Config.Local.Folder
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			return fmt.Errorf("failed to create bin directory: %w", err)
		}

		platform, arch := runtime.GOOS, runtime.GOARCH
		var assetName string
		switch platform {
		case "darwin":
			if arch == "arm64" {
				assetName = "educates-darwin-arm64"
			} else {
				assetName = "educates-darwin-amd64"
			}
		case "linux":
			if arch == "arm64" {
				assetName = "educates-linux-arm64"
			} else {
				assetName = "educates-linux-amd64"
			}
		default:
			return fmt.Errorf("unsupported platform: %s-%s", platform, arch)
		}

		ctx := context.Background()
		var client *github.Client
		if Config.Github.Token != "" {
			ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: Config.Github.Token})
			client = github.NewClient(oauth2.NewClient(ctx, ts))
		} else {
			client = github.NewClient(nil)
		}
		tag := version
		release, _, err := client.Repositories.GetReleaseByTag(ctx, Config.Github.Org, Config.Github.Repository, tag)
		if err != nil {
			return fmt.Errorf("failed to fetch release info: %w", err)
		}

		var downloadURL string
		for _, a := range release.Assets {
			if a.GetName() == assetName {
				downloadURL = a.GetBrowserDownloadURL()
				break
			}
		}
		if downloadURL == "" {
			return fmt.Errorf("could not find asset %s in release %s", assetName, version)
		}

		fmt.Printf("Downloading %s...\n", downloadURL)
		outPath := filepath.Join(binDir, fmt.Sprintf("educates-%s", version))
		if err := downloadFile(downloadURL, outPath); err != nil {
			return fmt.Errorf("failed to download binary: %w", err)
		}
		if err := os.Chmod(outPath, 0o755); err != nil {
			return fmt.Errorf("failed to set executable permission: %w", err)
		}
		fmt.Printf("Installed educates version %s as %s\n", version, outPath)
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
	rootCmd.AddCommand(installCmd)
}
