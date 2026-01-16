// Package scenario provides types and utilities for multi-domain scenario definitions.
//
// A scenario defines one or more domains (e.g., aws, gitlab) and their relationships,
// enabling cross-domain infrastructure generation and validation.
package scenario

// ScenarioConfig represents a multi-domain scenario configuration.
type ScenarioConfig struct {
	// Name is the scenario identifier
	Name string `yaml:"name"`

	// Description explains what this scenario produces
	Description string `yaml:"description,omitempty"`

	// Model specifies the Claude model to use (e.g., "haiku", "sonnet", "opus")
	// Defaults to the Claude CLI default if not specified
	Model string `yaml:"model,omitempty"`

	// Prompts contains prompt configuration for design mode
	Prompts *PromptConfig `yaml:"prompts,omitempty"`

	// Domains lists the domains involved in this scenario
	Domains []DomainSpec `yaml:"domains"`

	// CrossDomain defines relationships between domains
	CrossDomain []CrossDomainSpec `yaml:"cross_domain,omitempty"`

	// Validation defines validation rules for the scenario
	Validation map[string]ValidationRules `yaml:"validation,omitempty"`
}

// PromptConfig contains prompt configuration for design mode.
type PromptConfig struct {
	// Default is the path to the default prompt file
	Default string `yaml:"default,omitempty"`

	// Variants maps variant names to prompt file paths
	Variants map[string]string `yaml:"variants,omitempty"`
}

// DomainSpec defines a domain within a scenario.
type DomainSpec struct {
	// Name is the domain identifier (e.g., "aws", "gitlab")
	Name string `yaml:"name"`

	// CLI is the wetwire CLI command for this domain (e.g., "wetwire-aws")
	CLI string `yaml:"cli"`

	// MCPTools maps tool purposes to MCP tool names
	MCPTools map[string]string `yaml:"mcp_tools,omitempty"`

	// DependsOn lists domains that must be generated before this one
	DependsOn []string `yaml:"depends_on,omitempty"`

	// Outputs lists output file patterns for this domain
	Outputs []string `yaml:"outputs,omitempty"`
}

// CrossDomainSpec defines a relationship between two domains.
type CrossDomainSpec struct {
	// From is the source domain name
	From string `yaml:"from"`

	// To is the target domain name
	To string `yaml:"to"`

	// Type describes the relationship type (e.g., "artifact_reference", "output_mapping")
	Type string `yaml:"type"`

	// Validation contains specific validation rules for this relationship
	Validation CrossDomainValidation `yaml:"validation,omitempty"`
}

// CrossDomainValidation contains validation rules for cross-domain relationships.
type CrossDomainValidation struct {
	// RequiredRefs lists references that must exist in the target domain
	RequiredRefs []string `yaml:"required_refs,omitempty"`
}

// ValidationRules defines validation constraints for a domain.
type ValidationRules struct {
	// Stacks validation for CloudFormation stacks (AWS domain)
	Stacks *CountConstraint `yaml:"stacks,omitempty"`

	// Pipelines validation for CI pipelines (GitLab/GitHub domain)
	Pipelines *CountConstraint `yaml:"pipelines,omitempty"`

	// Workflows validation for GitHub Actions workflows
	Workflows *CountConstraint `yaml:"workflows,omitempty"`

	// Manifests validation for Kubernetes manifests
	Manifests *CountConstraint `yaml:"manifests,omitempty"`

	// Resources is a generic resource count constraint
	Resources *CountConstraint `yaml:"resources,omitempty"`
}

// CountConstraint specifies min/max count requirements.
type CountConstraint struct {
	// Min is the minimum required count
	Min int `yaml:"min,omitempty"`

	// Max is the maximum allowed count (0 means no limit)
	Max int `yaml:"max,omitempty"`
}

// GetDomain returns the domain with the given name, or nil if not found.
func (s *ScenarioConfig) GetDomain(name string) *DomainSpec {
	for i := range s.Domains {
		if s.Domains[i].Name == name {
			return &s.Domains[i]
		}
	}
	return nil
}

// DomainNames returns a list of all domain names in the scenario.
func (s *ScenarioConfig) DomainNames() []string {
	names := make([]string, len(s.Domains))
	for i, d := range s.Domains {
		names[i] = d.Name
	}
	return names
}
