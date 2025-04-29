package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/educates/educatesenv/pkg/config"
	"github.com/educates/educatesenv/pkg/github"
	"github.com/educates/educatesenv/pkg/version"
)

var (
	cfg     *config.Config
	gh      *github.Client
	manager *version.Manager
)

var rootCmd = &cobra.Command{
	Use:   "educatesenv",
	Short: "Manage multiple versions of the educates binary",
	Long:  `A version manager for educates, inspired by tfenv.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip validation for help command
		if cmd.Name() == "help" {
			return nil
		}

		// Validate development mode configuration
		if err := manager.ValidateDevelopmentMode(); err != nil {
			fmt.Printf("Warning: %v\n", err)
		}
		return nil
	},
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initDependencies)
}

func initDependencies() {
	// Initialize configuration
	cfg = config.New()
	if err := cfg.Load(); err != nil {
		cobra.CheckErr(err)
	}

	// Initialize GitHub client
	gh = github.New(cfg)

	// Initialize version manager
	manager = version.New(cfg, gh)
}
