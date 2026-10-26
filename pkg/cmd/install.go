package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/educates/educatesenv/pkg/platform"
)

func newInstallCommand(d *deps) *cobra.Command {
	var (
		useAfterInstall bool
		forceOverwrite  bool
	)

	cmd := &cobra.Command{
		Use:           "install <version>",
		Short:         "Install a specific version of educates",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun:        d.cleanUpDevelopmentSymlink,
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]

			if !platform.IsSupportedPlatform(runtime.GOOS, runtime.GOARCH) {
				return fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
			}

			if err := d.manager.InstallVersion(version, forceOverwrite, useAfterInstall); err != nil {
				return fmt.Errorf("failed to install version %s: %w", version, err)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&useAfterInstall, "use", false, "Set the installed version as active")
	cmd.Flags().BoolVar(&forceOverwrite, "overwrite", false, "Force download even if the version already exists")
	return cmd
}
