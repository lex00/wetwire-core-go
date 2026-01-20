package scenario

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// StructureError represents a scenario directory structure validation error.
type StructureError struct {
	Path    string
	Message string
}

func (e StructureError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

// StructureResult contains all structure validation errors for a scenario.
type StructureResult struct {
	ScenarioPath string
	Errors       []StructureError
}

// IsValid returns true if there are no structure validation errors.
func (r StructureResult) IsValid() bool {
	return len(r.Errors) == 0
}

// Error returns a combined error message.
func (r StructureResult) Error() string {
	if r.IsValid() {
		return ""
	}
	var msgs []string
	for _, e := range r.Errors {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "\n")
}

// RequiredPersonas is the list of persona prompts that must exist.
var RequiredPersonas = []string{"beginner", "intermediate", "expert"}

// ValidateStructure checks that a scenario directory has all required files
// and follows the specification.
func ValidateStructure(scenarioPath string) *StructureResult {
	result := &StructureResult{ScenarioPath: scenarioPath}

	// Check scenario directory exists
	if _, err := os.Stat(scenarioPath); os.IsNotExist(err) {
		result.Errors = append(result.Errors, StructureError{
			Path:    scenarioPath,
			Message: "scenario directory does not exist",
		})
		return result
	}

	// Required files
	requiredFiles := []string{
		"scenario.yaml",
		"system_prompt.md",
		"prompt.md",
		".gitignore",
	}

	for _, file := range requiredFiles {
		path := filepath.Join(scenarioPath, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			result.Errors = append(result.Errors, StructureError{
				Path:    path,
				Message: "required file missing",
			})
		}
	}

	// Required persona prompts
	for _, persona := range RequiredPersonas {
		path := filepath.Join(scenarioPath, "prompts", persona+".md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			result.Errors = append(result.Errors, StructureError{
				Path:    path,
				Message: fmt.Sprintf("persona prompt missing: %s", persona),
			})
		}
	}

	// Validate scenario.yaml content
	scenarioYAMLPath := filepath.Join(scenarioPath, "scenario.yaml")
	if content, err := os.ReadFile(scenarioYAMLPath); err == nil {
		validateScenarioYAML(content, scenarioYAMLPath, result)
	}

	// Validate .gitignore contains required entries
	gitignorePath := filepath.Join(scenarioPath, ".gitignore")
	if content, err := os.ReadFile(gitignorePath); err == nil {
		validateGitignore(string(content), gitignorePath, result)
	}

	// Validate system_prompt.md is not empty
	systemPromptPath := filepath.Join(scenarioPath, "system_prompt.md")
	if content, err := os.ReadFile(systemPromptPath); err == nil {
		if len(strings.TrimSpace(string(content))) == 0 {
			result.Errors = append(result.Errors, StructureError{
				Path:    systemPromptPath,
				Message: "system_prompt.md is empty",
			})
		}
	}

	// Validate prompt.md is not empty
	promptPath := filepath.Join(scenarioPath, "prompt.md")
	if content, err := os.ReadFile(promptPath); err == nil {
		if len(strings.TrimSpace(string(content))) == 0 {
			result.Errors = append(result.Errors, StructureError{
				Path:    promptPath,
				Message: "prompt.md is empty",
			})
		}
	}

	return result
}

// scenarioYAMLConfig is a simplified structure for validating scenario.yaml files.
// It differs from ScenarioConfig in schema.go which is for runtime configuration.
type scenarioYAMLConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Prompts     struct {
		Default  string            `yaml:"default"`
		Variants map[string]string `yaml:"variants"`
	} `yaml:"prompts"`
	Domains []struct {
		Name    string   `yaml:"name"`
		Outputs []string `yaml:"outputs"`
	} `yaml:"domains"`
}

func validateScenarioYAML(content []byte, path string, result *StructureResult) {
	var config scenarioYAMLConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		result.Errors = append(result.Errors, StructureError{
			Path:    path,
			Message: fmt.Sprintf("invalid YAML: %v", err),
		})
		return
	}

	if config.Name == "" {
		result.Errors = append(result.Errors, StructureError{
			Path:    path,
			Message: "missing 'name' field",
		})
	}

	if config.Description == "" {
		result.Errors = append(result.Errors, StructureError{
			Path:    path,
			Message: "missing 'description' field",
		})
	}

	if config.Prompts.Default == "" {
		result.Errors = append(result.Errors, StructureError{
			Path:    path,
			Message: "missing 'prompts.default' field",
		})
	}

	// Check all required persona variants are defined
	for _, persona := range RequiredPersonas {
		if _, ok := config.Prompts.Variants[persona]; !ok {
			result.Errors = append(result.Errors, StructureError{
				Path:    path,
				Message: fmt.Sprintf("missing prompt variant: %s", persona),
			})
		}
	}

	if len(config.Domains) == 0 {
		result.Errors = append(result.Errors, StructureError{
			Path:    path,
			Message: "no domains defined",
		})
	}

	for _, domain := range config.Domains {
		if len(domain.Outputs) == 0 {
			result.Errors = append(result.Errors, StructureError{
				Path:    path,
				Message: fmt.Sprintf("domain '%s' has no outputs defined", domain.Name),
			})
		}
	}
}

func validateGitignore(content, path string, result *StructureResult) {
	requiredEntries := []string{"results/", "*.svg"}

	for _, entry := range requiredEntries {
		if !strings.Contains(content, entry) {
			result.Errors = append(result.Errors, StructureError{
				Path:    path,
				Message: fmt.Sprintf("missing entry: %s", entry),
			})
		}
	}
}
