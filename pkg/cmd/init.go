package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/educates/educatesenv/pkg/platform"
	"github.com/educates/educatesenv/pkg/version"
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
		PreRun:        d.cleanUpDevelopmentSymlink,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			paths, configCreated, err := config.CreateDefault()
			if err != nil {
				return err
			}
			defaultBin := paths.BinDir

			if configCreated {
				_, _ = fmt.Fprintf(out, "Config file created at %s\n", paths.File)
			} else {
				_, _ = fmt.Fprintf(out, "Config file already exists at %s\n", paths.File)
			}
			_, _ = fmt.Fprintf(out, "Bin directory ensured at %s\n", defaultBin)

			// Print PATH instructions
			_, _ = fmt.Fprintln(out, "\nTo use educatesenv, add the bin directory to your PATH:")
			switch runtime.GOOS {
			case platform.Windows:
				_, _ = fmt.Fprintf(out, "\nFor Windows (PowerShell), run:\n  [Environment]::SetEnvironmentVariable(\"Path\", \"%s;\" + $Env:Path, [EnvironmentVariableTarget]::User)\n", defaultBin)
			default:
				writeShellPathInstructions(out, os.Getenv("SHELL"), defaultBin)
			}
			_, _ = fmt.Fprintln(out, "\nRestart your terminal or source your profile to apply the changes.")

			if downloadLatest {
				_, _ = fmt.Fprintln(out, "\nFetching latest educates version...")
				if err := d.manager.InstallVersion(version.Latest, overwrite, true); err != nil {
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

// writeShellPathInstructions writes how to add binDir to the PATH in the
// shell at the path shell, as $SHELL gives it: zsh and bash get a line for
// their rc file, fish gets fish_user_paths, and any other shell gets
// generic instructions.
func writeShellPathInstructions(out io.Writer, shell, binDir string) {
	switch name := filepath.Base(shell); name {
	case "zsh", "bash":
		rcFile := "~/." + name + "rc"
		_, _ = fmt.Fprintf(out, "\nFor %s, run:\n  echo 'export PATH=\"%s:$PATH\"' >> %s\n  source %s\n", name, binDir, rcFile, rcFile)
	case "fish":
		_, _ = fmt.Fprintf(out, "\nFor fish, run:\n  set -U fish_user_paths %s $fish_user_paths\n", binDir)
	default:
		_, _ = fmt.Fprintf(out, "\nAdd %s to your PATH in your shell's startup file.\n", binDir)
	}
}
