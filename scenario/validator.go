package scenario

import (
	"fmt"
	"strings"
)

// ValidationError represents a validation error with context.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// ValidationResult contains all validation errors found.
type ValidationResult struct {
	Errors []ValidationError
}

// IsValid returns true if there are no validation errors.
func (r *ValidationResult) IsValid() bool {
	return len(r.Errors) == 0
}

// AddError adds a validation error.
func (r *ValidationResult) AddError(field, message string) {
	r.Errors = append(r.Errors, ValidationError{Field: field, Message: message})
}

// Error returns a combined error message.
func (r *ValidationResult) Error() string {
	if r.IsValid() {
		return ""
	}

	var msgs []string
	for _, e := range r.Errors {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "; ")
}

// Validate validates a scenario configuration.
func Validate(config *ScenarioConfig) *ValidationResult {
	result := &ValidationResult{}

	// Validate required fields
	if config.Name == "" {
		result.AddError("name", "scenario name is required")
	}

	if len(config.Domains) == 0 {
		result.AddError("domains", "at least one domain is required")
	}

	// Validate domains
	domainNames := make(map[string]bool)
	for i, d := range config.Domains {
		field := fmt.Sprintf("domains[%d]", i)

		if d.Name == "" {
			result.AddError(field+".name", "domain name is required")
		} else {
			if domainNames[d.Name] {
				result.AddError(field+".name", fmt.Sprintf("duplicate domain name: %s", d.Name))
			}
			domainNames[d.Name] = true
		}

		if d.CLI == "" {
			result.AddError(field+".cli", "domain CLI is required")
		}
	}

	// Validate domain dependencies reference existing domains
	for i, d := range config.Domains {
		for _, dep := range d.DependsOn {
			if !domainNames[dep] {
				result.AddError(
					fmt.Sprintf("domains[%d].depends_on", i),
					fmt.Sprintf("unknown dependency: %s", dep),
				)
			}
			if dep == d.Name {
				result.AddError(
					fmt.Sprintf("domains[%d].depends_on", i),
					"domain cannot depend on itself",
				)
			}
		}
	}

	// Validate cross-domain references
	for i, cd := range config.CrossDomain {
		field := fmt.Sprintf("cross_domain[%d]", i)

		if cd.From == "" {
			result.AddError(field+".from", "source domain is required")
		} else if !domainNames[cd.From] {
			result.AddError(field+".from", fmt.Sprintf("unknown domain: %s", cd.From))
		}

		if cd.To == "" {
			result.AddError(field+".to", "target domain is required")
		} else if !domainNames[cd.To] {
			result.AddError(field+".to", fmt.Sprintf("unknown domain: %s", cd.To))
		}

		if cd.Type == "" {
			result.AddError(field+".type", "relationship type is required")
		}
	}

	// Check for circular dependencies
	_, err := GetDomainOrder(config)
	if err != nil {
		result.AddError("domains", err.Error())
	}

	return result
}

// ValidateRequired validates that required fields are present.
// This is a lighter validation than Validate().
func ValidateRequired(config *ScenarioConfig) error {
	if config.Name == "" {
		return &ValidationError{Field: "name", Message: "scenario name is required"}
	}
	if len(config.Domains) == 0 {
		return &ValidationError{Field: "domains", Message: "at least one domain is required"}
	}
	return nil
}
