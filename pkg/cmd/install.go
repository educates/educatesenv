package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/educates/educatesenv/pkg/platform"
)

var (
	useAfterInstall bool
	forceOverwrite  bool
)

var installCmd = &cobra.Command{
	Use:           "install <version>",
	Short:         "Install a specific version of educates",
	Args:          cobra.ExactArgs(1),
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		version := args[0]

		if !platform.IsSupportedPlatform(runtime.GOOS, runtime.GOARCH) {
			return fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
		}

		if err := manager.InstallVersion(version, forceOverwrite, useAfterInstall); err != nil {
			return fmt.Errorf("failed to install version %s: %w", version, err)
		}

		return nil
	},
}

func init() {
	installCmd.Flags().BoolVar(&useAfterInstall, "use", false, "Set the installed version as active")
	installCmd.Flags().BoolVar(&forceOverwrite, "overwrite", false, "Force download even if the version already exists")
	rootCmd.AddCommand(installCmd)
}
