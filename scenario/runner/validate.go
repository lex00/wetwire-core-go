package runner

import (
	"fmt"

	"github.com/lex00/wetwire-core-go/scenario"
)

// ValidateRefs validates that all required references in the scenario configuration
// can be resolved against the provided output manifest.
//
// It checks:
//  1. All required_refs in CrossDomainSpec validation rules exist in the manifest
//  2. The referenced domain exists in the manifest
//  3. The referenced resource exists in the domain
//  4. The referenced field exists in the resource outputs
//
// Returns a slice of errors, one for each validation failure. An empty slice
// indicates all references are valid.
func (m *OutputManifest) ValidateRefs(config *scenario.ScenarioConfig) []error {
	var errors []error

	if config == nil {
		return []error{fmt.Errorf("scenario config is nil")}
	}

	// Check all cross-domain relationships for required references
	for _, crossDomain := range config.CrossDomain {
		for _, refStr := range crossDomain.Validation.RequiredRefs {
			// Parse the reference
			ref, err := scenario.ParseRef(refStr)
			if err != nil {
				errors = append(errors, fmt.Errorf("invalid reference in cross-domain '%s->%s': %w",
					crossDomain.From, crossDomain.To, err))
				continue
			}

			// Validate against manifest
			if err := m.validateRefAgainstManifest(ref); err != nil {
				errors = append(errors, fmt.Errorf("reference validation failed in cross-domain '%s->%s': %w",
					crossDomain.From, crossDomain.To, err))
			}
		}
	}

	return errors
}

// validateRefAgainstManifest checks if a reference can be resolved in the manifest.
func (m *OutputManifest) validateRefAgainstManifest(ref *scenario.CrossDomainRef) error {
	// Check domain exists
	domainOutput := m.GetDomainOutput(ref.Domain)
	if domainOutput == nil {
		return fmt.Errorf("domain '%s' not found in manifest", ref.Domain)
	}

	// Check resource exists
	resourceOutput, exists := domainOutput.Resources[ref.Resource]
	if !exists {
		return fmt.Errorf("resource '%s' not found in domain '%s'", ref.Resource, ref.Domain)
	}

	// Check field exists in outputs
	if resourceOutput.Outputs == nil {
		return fmt.Errorf("resource '%s.%s' has no outputs", ref.Domain, ref.Resource)
	}

	if _, exists := resourceOutput.Outputs[ref.Field]; !exists {
		return fmt.Errorf("field '%s' not found in resource '%s.%s' outputs", ref.Field, ref.Domain, ref.Resource)
	}

	return nil
}

// ResolveRef resolves a reference against a manifest and returns the output value.
// Returns an error if the reference cannot be resolved.
func (m *OutputManifest) ResolveRef(ref *scenario.CrossDomainRef) (interface{}, error) {
	if err := m.validateRefAgainstManifest(ref); err != nil {
		return nil, err
	}

	domainOutput := m.GetDomainOutput(ref.Domain)
	resourceOutput := domainOutput.Resources[ref.Resource]
	return resourceOutput.Outputs[ref.Field], nil
}
