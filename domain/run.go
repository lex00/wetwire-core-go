package domain

import (
	"github.com/spf13/cobra"
)

// Run creates a CLI from the Domain interface. It returns the root command
// so domains can add custom commands before executing.
//
// Example usage:
//
//	root := domain.Run(myDomain)
//	root.AddCommand(myCustomCommand)
//	root.Execute()
func Run(d Domain) *cobra.Command {
	return buildCLI(d)
}

// buildCLI constructs the complete command tree from a Domain.
// It adds all required commands (build, lint, init, validate) and
// conditionally adds optional commands if the domain implements
// the corresponding interfaces (ImporterDomain, ListerDomain, GrapherDomain).
func buildCLI(d Domain) *cobra.Command {
	root := &cobra.Command{
		Use:     "wetwire-" + d.Name(),
		Version: d.Version(),
		Short:   "wetwire " + d.Name() + " domain CLI",
	}

	// Add persistent flags that apply to all commands
	root.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
	root.PersistentFlags().StringP("format", "f", "text", "Output format (text, json, yaml)")

	// Add required commands
	root.AddCommand(generateBuildCmd(d.Builder()))
	root.AddCommand(generateLintCmd(d.Linter()))
	root.AddCommand(generateInitCmd(d.Initializer()))
	root.AddCommand(generateValidateCmd(d.Validator()))

	// Add optional commands via type assertions
	if imp, ok := d.(ImporterDomain); ok {
		root.AddCommand(generateImportCmd(imp.Importer()))
	}
	if lst, ok := d.(ListerDomain); ok {
		root.AddCommand(generateListCmd(lst.Lister()))
	}
	if gph, ok := d.(GrapherDomain); ok {
		root.AddCommand(generateGraphCmd(gph.Grapher()))
	}

	return root
}
