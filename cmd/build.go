package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewBuildCommand creates a new build command that uses the provided Builder.
func NewBuildCommand(builder Builder) *cobra.Command {
	var opts BuildOptions
	var path string

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build infrastructure from source definitions",
		Long: `Build synthesizes infrastructure output from source definitions.

The build process reads definition files and generates the target output
format (e.g., CloudFormation, Terraform).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			opts.Verbose, _ = cmd.Flags().GetBool("verbose")

			if err := builder.Build(ctx, path, opts); err != nil {
				return fmt.Errorf("build failed: %w", err)
			}

			if !opts.DryRun {
				_, _ = fmt.Fprintln(os.Stdout, "Build completed successfully")
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", ".", "Path to source definitions")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output path")
	cmd.Flags().BoolVarP(&opts.DryRun, "dry-run", "n", false, "Show what would be built without writing files")

	return cmd
}
