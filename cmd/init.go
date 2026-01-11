package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewInitCommand creates a new init command that uses the provided Initializer.
func NewInitCommand(initializer Initializer) *cobra.Command {
	var opts InitOptions

	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize a new project",
		Long: `Init creates a new project with the specified name.

This generates the directory structure and starter files needed
to begin defining infrastructure.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			name := args[0]

			if err := initializer.Init(ctx, name, opts); err != nil {
				return fmt.Errorf("init failed: %w", err)
			}

			_, _ = fmt.Fprintf(os.Stdout, "Project %q initialized successfully\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Template, "template", "t", "", "Template to use for initialization")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Overwrite existing files")

	return cmd
}
