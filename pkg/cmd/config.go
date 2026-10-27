package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"

	"github.com/educates/educatesenv/pkg/config"
)

func newConfigCommand(d *deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "config",
		Short:         "Manage educatesenv configuration",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		// cobra checks the arguments only of a command that can run, so
		// config runs and prints its help.
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newConfigInitCommand(), newConfigViewCommand(d))
	return cmd
}

func newConfigInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Generate a default config file",
		Args:  cobra.NoArgs,
		// Disable the automatic error printing
		SilenceErrors: true,
		// Disable automatic usage printing on error
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, created, err := config.CreateDefault()
			if err != nil {
				return err
			}
			if !created {
				return fmt.Errorf("config file already exists at %s", paths.ConfigFile)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Config file created at %s\n", paths.ConfigFile)
			return nil
		},
	}
}

func newConfigViewCommand(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:           "view",
		Short:         "Show the current configuration",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// The token is shown as its source, never as the token.
			view := *d.cfg
			view.Github.Token = d.githubToken().String()
			yamlBytes, err := yaml.Marshal(view)
			if err != nil {
				return fmt.Errorf("failed to marshal config to YAML: %w", err)
			}

			// Print the configuration
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintln(out, "Current configuration:")
			_, _ = fmt.Fprintln(out, string(yamlBytes))
			return nil
		},
	}
}
