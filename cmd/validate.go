package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewValidateCommand creates a new validate command that uses the provided Validator.
func NewValidateCommand(validator Validator) *cobra.Command {
	var opts ValidateOptions
	var path string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate generated output",
		Long: `Validate checks that generated output is valid.

This includes syntax validation, schema validation, and any domain-specific
validation rules.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			opts.Verbose, _ = cmd.Flags().GetBool("verbose")

			errors, err := validator.Validate(ctx, path, opts)
			if err != nil {
				return fmt.Errorf("validate failed: %w", err)
			}

			if len(errors) == 0 {
				_, _ = fmt.Fprintln(os.Stdout, "Validation passed")
				return nil
			}

			for _, ve := range errors {
				_, _ = fmt.Fprintf(os.Stdout, "%s: %s (%s)\n", ve.Path, ve.Message, ve.Code)
			}

			return fmt.Errorf("validation found %d error(s)", len(errors))
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", ".", "Path to validate")
	cmd.Flags().BoolVarP(&opts.Strict, "strict", "s", false, "Enable strict validation")

	return cmd
}
