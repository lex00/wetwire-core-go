package domain

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// generateBuildCmd creates the 'build' command for the CLI.
func generateBuildCmd(builder Builder) *cobra.Command {
	var buildType string
	var output string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "build [path]",
		Short: "Build domain resources from source code",
		Long:  "Build domain resources from source code and output in the specified format.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			verbose, _ := cmd.Flags().GetBool("verbose")
			format, _ := cmd.Flags().GetString("format")

			ctx := NewContextWithVerbose(context.Background(), path, verbose)
			opts := BuildOpts{
				Format: format,
				Type:   buildType,
				Output: output,
				DryRun: dryRun,
			}

			result, err := builder.Build(ctx, path, opts)
			if err != nil {
				return fmt.Errorf("build failed: %w", err)
			}

			return outputResult(result, format)
		},
	}

	cmd.Flags().StringVar(&buildType, "type", "", "Filter build to specific resource type")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output path for generated files")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview output without writing files")

	return cmd
}

// generateLintCmd creates the 'lint' command for the CLI.
func generateLintCmd(linter Linter) *cobra.Command {
	var fix bool
	var disable []string

	cmd := &cobra.Command{
		Use:   "lint [path]",
		Short: "Lint domain resources according to domain rules",
		Long:  "Validate domain resources according to domain-specific rules and output errors/warnings.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			verbose, _ := cmd.Flags().GetBool("verbose")
			format, _ := cmd.Flags().GetString("format")

			ctx := NewContextWithVerbose(context.Background(), path, verbose)
			opts := LintOpts{
				Format:  format,
				Fix:     fix,
				Disable: disable,
			}

			result, err := linter.Lint(ctx, path, opts)
			if err != nil {
				return fmt.Errorf("lint failed: %w", err)
			}

			return outputResult(result, format)
		},
	}

	cmd.Flags().BoolVar(&fix, "fix", false, "Automatically fix fixable issues")
	cmd.Flags().StringSliceVar(&disable, "disable", nil, "Rules to disable (comma-separated)")

	return cmd
}

// generateInitCmd creates the 'init' command for the CLI.
func generateInitCmd(initializer Initializer) *cobra.Command {
	var name string
	var outPath string
	var scenario bool
	var description string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new domain project with example code",
		Long: `Create a new domain project structure with example code and configuration.

Use --scenario to create a full scenario structure with prompts/, expected/,
scenario.yaml, and persona-specific prompt templates.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ := cmd.Flags().GetBool("verbose")
			format, _ := cmd.Flags().GetString("format")

			workDir := "."
			ctx := NewContextWithVerbose(context.Background(), workDir, verbose)
			opts := InitOpts{
				Name:        name,
				Path:        outPath,
				Scenario:    scenario,
				Description: description,
			}

			result, err := initializer.Init(ctx, workDir, opts)
			if err != nil {
				return fmt.Errorf("init failed: %w", err)
			}

			return outputResult(result, format)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Project name")
	cmd.Flags().StringVar(&outPath, "path", ".", "Output directory (default: current directory)")
	cmd.Flags().BoolVar(&scenario, "scenario", false, "Create full scenario structure with prompts and expected outputs")
	cmd.Flags().StringVar(&description, "description", "", "Scenario description (used with --scenario)")

	return cmd
}

// generateValidateCmd creates the 'validate' command for the CLI.
func generateValidateCmd(validator Validator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [path]",
		Short: "Validate that generated output conforms to domain specifications",
		Long:  "Validate that generated output conforms to domain specifications and standards.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			verbose, _ := cmd.Flags().GetBool("verbose")
			format, _ := cmd.Flags().GetString("format")

			ctx := NewContextWithVerbose(context.Background(), path, verbose)
			opts := ValidateOpts{}

			result, err := validator.Validate(ctx, path, opts)
			if err != nil {
				return fmt.Errorf("validate failed: %w", err)
			}

			return outputResult(result, format)
		},
	}

	return cmd
}

// generateImportCmd creates the 'import' command for the CLI (optional).
func generateImportCmd(importer Importer) *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "import [source]",
		Short: "Import external resources or configurations into the domain",
		Long:  "Import external resources or configurations and convert them to domain format.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]

			verbose, _ := cmd.Flags().GetBool("verbose")
			format, _ := cmd.Flags().GetString("format")

			workDir := "."
			ctx := NewContextWithVerbose(context.Background(), workDir, verbose)
			opts := ImportOpts{
				Target: target,
			}

			result, err := importer.Import(ctx, source, opts)
			if err != nil {
				return fmt.Errorf("import failed: %w", err)
			}

			return outputResult(result, format)
		},
	}

	cmd.Flags().StringVar(&target, "target", "", "Target location for imported resources")

	return cmd
}

// generateListCmd creates the 'list' command for the CLI (optional).
func generateListCmd(lister Lister) *cobra.Command {
	var listType string

	cmd := &cobra.Command{
		Use:   "list [path]",
		Short: "List discovered domain resources",
		Long:  "Discover and list domain resources in the specified path.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			verbose, _ := cmd.Flags().GetBool("verbose")
			format, _ := cmd.Flags().GetString("format")

			ctx := NewContextWithVerbose(context.Background(), path, verbose)
			opts := ListOpts{
				Format: format,
				Type:   listType,
			}

			result, err := lister.List(ctx, path, opts)
			if err != nil {
				return fmt.Errorf("list failed: %w", err)
			}

			return outputResult(result, format)
		},
	}

	cmd.Flags().StringVar(&listType, "type", "", "Filter list to specific resource type")

	return cmd
}

// generateGraphCmd creates the 'graph' command for the CLI (optional).
func generateGraphCmd(grapher Grapher) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graph [path]",
		Short: "Visualize relationships between domain resources",
		Long:  "Generate a visualization of relationships between domain resources (DOT, Mermaid, etc).",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			verbose, _ := cmd.Flags().GetBool("verbose")
			format, _ := cmd.Flags().GetString("format")

			ctx := NewContextWithVerbose(context.Background(), path, verbose)
			opts := GraphOpts{
				Format: format,
			}

			result, err := grapher.Graph(ctx, path, opts)
			if err != nil {
				return fmt.Errorf("graph failed: %w", err)
			}

			return outputResult(result, format)
		},
	}

	return cmd
}

// outputResult handles outputting the result based on the format flag.
func outputResult(result *Result, format string) error {
	output, err := FormatResult(result, format)
	if err != nil {
		return fmt.Errorf("failed to format result: %w", err)
	}

	fmt.Fprint(os.Stdout, output)

	// Return error if result indicates failure
	if !result.Success {
		return fmt.Errorf("operation failed")
	}

	return nil
}
