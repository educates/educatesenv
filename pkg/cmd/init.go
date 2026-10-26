package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/educates/educatesenv/pkg/platform"
)

func newInitCommand(d *deps) *cobra.Command {
	var (
		downloadLatest bool
		overwrite      bool
	)

	cmd := &cobra.Command{
		Use:           "init",
		Short:         "Initialize educatesenv: create default config and folders, and show PATH instructions",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			_, defaultBin, configPath, configCreated, err := config.CreateConfigAndFolders()
			if err != nil {
				return err
			}

			if configCreated {
				_, _ = fmt.Fprintf(out, "Config file created at %s\n", configPath)
			} else {
				_, _ = fmt.Fprintf(out, "Config file already exists at %s\n", configPath)
			}
			_, _ = fmt.Fprintf(out, "Bin directory ensured at %s\n", defaultBin)

			// Print PATH instructions
			_, _ = fmt.Fprintln(out, "\nTo use educatesenv, add the bin directory to your PATH:")
			var pathCmd string
			switch runtime.GOOS {
			case platform.Darwin, platform.Linux:
				pathCmd = fmt.Sprintf("echo 'export PATH=\"%s:$PATH\"' >> ~/.bashrc\nsource ~/.bashrc", defaultBin)
				_, _ = fmt.Fprintf(out, "\nFor bash/zsh, run:\n  %s\n", pathCmd)
				_, _ = fmt.Fprintf(out, "\nFor fish shell, run:\n  set -U fish_user_paths %s $fish_user_paths\n", defaultBin)
			case platform.Windows:
				_, _ = fmt.Fprintf(out, "\nFor Windows (PowerShell), run:\n  [Environment]::SetEnvironmentVariable(\"Path\", \"%s;\" + $Env:Path, [EnvironmentVariableTarget]::User)\n", defaultBin)
			default:
				_, _ = fmt.Fprintf(out, "\nAdd %s to your PATH manually.\n", defaultBin)
			}
			_, _ = fmt.Fprintln(out, "\nRestart your terminal or source your profile to apply the changes.")

			if downloadLatest {
				_, _ = fmt.Fprintln(out, "\nFetching latest educates version...")
				latest, err := d.gh.GetLatestReleaseVersion()
				if err != nil {
					return fmt.Errorf("failed to get latest release version: %w", err)
				}
				_, _ = fmt.Fprintf(out, "Latest version: %s\n", latest)

				if err := d.manager.InstallVersion(latest, overwrite, true); err != nil {
					return err
				}
			} else {
				_, _ = fmt.Fprintln(out, "\nRun 'educatesenv install <version>' to install a specific version")
				_, _ = fmt.Fprintln(out, "Run 'educatesenv list-remote' to see available versions")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&downloadLatest, "download", false, "Download and set as active the latest stable version")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Force download even if the version already exists")
	return cmd
}
