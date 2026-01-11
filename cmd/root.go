package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCommand creates a new root command for a wetwire CLI.
func NewRootCommand(name, description string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: description,
		Long: description + `

This CLI provides commands for building, linting, validating, and managing
infrastructure definitions.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add persistent flags available to all commands
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	return cmd
}
