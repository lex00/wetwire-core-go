package scenario

import (
	"fmt"
	"io"
	"sort"
	"strings"

	scenarioPkg "github.com/lex00/wetwire-core-go/scenario"
)

// GenerateInstructions generates Claude Code instructions for a scenario.
func GenerateInstructions(w io.Writer, config *scenarioPkg.ScenarioConfig, prompt string, variants map[string]string) error {
	// Validate config
	if err := scenarioPkg.ValidateRequired(config); err != nil {
		return fmt.Errorf("invalid scenario config: %w", err)
	}

	// Get domain order respecting dependencies
	order, err := scenarioPkg.GetDomainOrder(config)
	if err != nil {
		return fmt.Errorf("failed to determine domain order: %w", err)
	}

	// Write header
	_, _ = fmt.Fprintf(w, "# Scenario: %s\n\n", config.Name)
	if config.Description != "" {
		_, _ = fmt.Fprintf(w, "%s\n\n", config.Description)
	}

	// Write prompt if provided
	if prompt != "" {
		_, _ = fmt.Fprintf(w, "## Instructions\n\n%s\n\n", prompt)
	}

	// Write domain steps
	_, _ = fmt.Fprintln(w, "## Execution Steps")
	_, _ = fmt.Fprintln(w)

	for i, domainName := range order {
		domain := config.GetDomain(domainName)
		if domain == nil {
			continue
		}
		FormatDomainStep(w, i+1, domain)
	}

	// Write cross-domain validation step if there are cross-domain relationships
	if len(config.CrossDomain) > 0 {
		stepNum := len(order) + 1
		_, _ = fmt.Fprintf(w, "### Step %d: Validate cross-domain relationships\n\n", stepNum)
		for _, cd := range config.CrossDomain {
			_, _ = fmt.Fprintf(w, "- Verify %s → %s (%s)\n", cd.From, cd.To, cd.Type)
			if len(cd.Validation.RequiredRefs) > 0 {
				_, _ = fmt.Fprintf(w, "  - Required refs: %s\n", strings.Join(cd.Validation.RequiredRefs, ", "))
			}
		}
		_, _ = fmt.Fprintln(w)
	}

	// Write validation criteria
	if len(config.Validation) > 0 {
		_, _ = fmt.Fprintln(w, "## Validation Criteria")
		_, _ = fmt.Fprintln(w)
		FormatValidationCriteria(w, config.Validation)
	}

	return nil
}

// FormatDomainStep formats a single domain step.
func FormatDomainStep(w io.Writer, stepNum int, domain *scenarioPkg.DomainSpec) {
	_, _ = fmt.Fprintf(w, "### Step %d: Generate %s domain\n\n", stepNum, domain.Name)
	_, _ = fmt.Fprintf(w, "CLI: `%s`\n\n", domain.CLI)

	if len(domain.DependsOn) > 0 {
		_, _ = fmt.Fprintf(w, "Dependencies: %s\n\n", strings.Join(domain.DependsOn, ", "))
	}

	// MCP tools to call
	if len(domain.MCPTools) > 0 {
		_, _ = fmt.Fprintln(w, "MCP Tools:")
		// Sort keys for consistent output
		keys := make([]string, 0, len(domain.MCPTools))
		for k := range domain.MCPTools {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, purpose := range keys {
			toolName := domain.MCPTools[purpose]
			_, _ = fmt.Fprintf(w, "- Call `%s` for %s\n", toolName, purpose)
		}
		_, _ = fmt.Fprintln(w)
	}

	// Expected outputs
	if len(domain.Outputs) > 0 {
		_, _ = fmt.Fprintln(w, "Expected outputs:")
		for _, output := range domain.Outputs {
			_, _ = fmt.Fprintf(w, "- %s\n", output)
		}
		_, _ = fmt.Fprintln(w)
	}
}

// FormatValidationCriteria formats validation criteria for all domains.
func FormatValidationCriteria(w io.Writer, validation map[string]scenarioPkg.ValidationRules) {
	// Sort domain names for consistent output
	domains := make([]string, 0, len(validation))
	for d := range validation {
		domains = append(domains, d)
	}
	sort.Strings(domains)

	for _, domainName := range domains {
		rules := validation[domainName]
		_, _ = fmt.Fprintf(w, "### %s\n\n", domainName)

		if rules.Stacks != nil {
			formatConstraint(w, "Stacks", rules.Stacks)
		}
		if rules.Pipelines != nil {
			formatConstraint(w, "Pipelines", rules.Pipelines)
		}
		if rules.Workflows != nil {
			formatConstraint(w, "Workflows", rules.Workflows)
		}
		if rules.Manifests != nil {
			formatConstraint(w, "Manifests", rules.Manifests)
		}
		if rules.Resources != nil {
			formatConstraint(w, "Resources", rules.Resources)
		}
	}
}

// formatConstraint formats a count constraint.
func formatConstraint(w io.Writer, name string, c *scenarioPkg.CountConstraint) {
	if c.Max > 0 {
		_, _ = fmt.Fprintf(w, "- %s: min %d, max %d\n", name, c.Min, c.Max)
	} else {
		_, _ = fmt.Fprintf(w, "- %s: min %d\n", name, c.Min)
	}
}
