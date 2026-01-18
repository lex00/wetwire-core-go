package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// OutputManifest stores outputs captured from domain executions.
// It maps domain names to their resource outputs.
type OutputManifest struct {
	// Domains maps domain name to domain outputs
	Domains map[string]*DomainOutput `json:"domains"`
}

// DomainOutput contains outputs for a single domain.
type DomainOutput struct {
	// Resources maps resource names to their outputs
	Resources map[string]ResourceOutput `json:"resources"`
}

// ResourceOutput contains the output data for a single resource.
type ResourceOutput struct {
	// Type is the resource type (e.g., "aws_s3_bucket", "gitlab_pipeline")
	Type string `json:"type,omitempty"`

	// Outputs is a map of output names to values
	Outputs map[string]interface{} `json:"outputs,omitempty"`
}

// NewOutputManifest creates a new empty OutputManifest.
func NewOutputManifest() *OutputManifest {
	return &OutputManifest{
		Domains: make(map[string]*DomainOutput),
	}
}

// AddDomainOutput adds outputs for a domain.
func (m *OutputManifest) AddDomainOutput(domainName string, output *DomainOutput) {
	if m.Domains == nil {
		m.Domains = make(map[string]*DomainOutput)
	}
	m.Domains[domainName] = output
}

// GetDomainOutput retrieves outputs for a domain.
func (m *OutputManifest) GetDomainOutput(domainName string) *DomainOutput {
	return m.Domains[domainName]
}

// SaveToFile writes the manifest to a JSON file.
func (m *OutputManifest) SaveToFile(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// LoadFromFile reads a manifest from a JSON file.
func LoadFromFile(path string) (*OutputManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest OutputManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	return &manifest, nil
}

// CaptureOutputsFromFiles discovers and captures outputs from generated files.
// This is a helper function that looks for common output patterns in domain files.
func CaptureOutputsFromFiles(workDir string, domainName string, outputPatterns []string) (*DomainOutput, error) {
	domainOutput := &DomainOutput{
		Resources: make(map[string]ResourceOutput),
	}

	// For each output pattern, try to find and parse files
	for _, pattern := range outputPatterns {
		matches, err := filepath.Glob(filepath.Join(workDir, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			// Extract resource name from file path
			relPath, err := filepath.Rel(workDir, match)
			if err != nil {
				continue
			}

			// Read file content
			content, err := os.ReadFile(match)
			if err != nil {
				continue
			}

			// Try to parse as JSON (common for CloudFormation, Terraform, etc.)
			var outputs map[string]interface{}
			if err := json.Unmarshal(content, &outputs); err == nil {
				// Successfully parsed as JSON
				resourceName := strings.TrimSuffix(relPath, filepath.Ext(relPath))
				domainOutput.Resources[resourceName] = ResourceOutput{
					Type:    domainName + "_resource",
					Outputs: outputs,
				}
			}
		}
	}

	return domainOutput, nil
}
