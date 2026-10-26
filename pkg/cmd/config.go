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
		SilenceErrors: true,
		SilenceUsage:  true,
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
			_, _, configPath, created, err := config.CreateConfigAndFolders()
			if err != nil {
				return err
			}
			if !created {
				return fmt.Errorf("config file already exists at %s", configPath)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Config file created at %s\n", configPath)
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
			// Convert config struct to YAML
			yamlBytes, err := yaml.Marshal(d.cfg)
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
