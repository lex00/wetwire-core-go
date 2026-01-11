package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewLintCommand creates a new lint command that uses the provided Linter.
func NewLintCommand(linter Linter) *cobra.Command {
	var opts LintOptions
	var path string

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Check source files for issues",
		Long: `Lint analyzes source definition files for common issues.

Issues are categorized by severity (error, warning, info) and include
file location and rule information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			opts.Verbose, _ = cmd.Flags().GetBool("verbose")

			issues, err := linter.Lint(ctx, path, opts)
			if err != nil {
				return fmt.Errorf("lint failed: %w", err)
			}

			if len(issues) == 0 {
				_, _ = fmt.Fprintln(os.Stdout, "No issues found")
				return nil
			}

			for _, issue := range issues {
				_, _ = fmt.Fprintf(os.Stdout, "%s:%d:%d: %s: %s (%s)\n",
					issue.File, issue.Line, issue.Column,
					issue.Severity, issue.Message, issue.Rule)
			}

			// Count errors vs warnings
			errorCount := 0
			for _, issue := range issues {
				if issue.Severity == "error" {
					errorCount++
				}
			}

			if errorCount > 0 {
				return fmt.Errorf("lint found %d error(s)", errorCount)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", ".", "Path to lint")
	cmd.Flags().BoolVarP(&opts.Fix, "fix", "f", false, "Automatically fix issues where possible")

	return cmd
}
