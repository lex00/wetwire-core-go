package domain

import (
	"fmt"
	"os"
	"path/filepath"
)

// ScenarioFiles holds the generated file contents for a scenario.
type ScenarioFiles struct {
	// Files maps relative paths to file contents
	Files map[string]string
}

// ScaffoldScenario generates the file structure for a scenario.
// Domains can use this to implement scenario initialization when InitOpts.Scenario is true.
func ScaffoldScenario(name, description, domainName string) *ScenarioFiles {
	files := make(map[string]string)

	files["scenario.yaml"] = scenarioYAML(name, description, domainName)
	files["system_prompt.md"] = systemPromptMD(domainName)
	files["prompt.md"] = promptMD(description)
	files["prompts/beginner.md"] = beginnerMD(description)
	files["prompts/intermediate.md"] = intermediateMD(description)
	files["prompts/expert.md"] = expertMD(description)
	files["prompts/terse.md"] = terseMD(description)
	files["prompts/verbose.md"] = verboseMD(description)
	files[".gitignore"] = gitignore()
	files["expected/.gitkeep"] = ""

	return &ScenarioFiles{Files: files}
}

// WriteScenario writes the scenario files to the specified path.
func WriteScenario(path string, scenario *ScenarioFiles) ([]string, error) {
	// Create directories
	dirs := []string{
		path,
		filepath.Join(path, "prompts"),
		filepath.Join(path, "expected"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	// Write files
	var created []string
	for filename, content := range scenario.Files {
		filePath := filepath.Join(path, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("write %s: %w", filename, err)
		}
		created = append(created, filename)
	}

	return created, nil
}

func scenarioYAML(name, description, domainName string) string {
	return fmt.Sprintf(`name: %s
description: %s

# Model to use: haiku (fast), sonnet (balanced), opus (best quality)
model: sonnet

prompts:
  default: prompt.md
  variants:
    beginner: prompts/beginner.md
    intermediate: prompts/intermediate.md
    expert: prompts/expert.md
    terse: prompts/terse.md
    verbose: prompts/verbose.md

domains:
  - name: %s
    outputs:
      - "**/*.go"

validation:
  %s:
    resources:
      min: 1
`, name, description, domainName, domainName)
}

func systemPromptMD(domainName string) string {
	return fmt.Sprintf(`You are a helpful %s assistant.

Your task is to help users create %s resources based on their requirements.
Use the available tools to create files and validate your work.

Guidelines:
- Always generate complete, production-quality resources regardless of how brief the request is
- If the user asks questions, answer them
- If the user asks for explanations, provide them
- Include best practices even if not explicitly requested
- Use the lint tool to validate your output before finishing
`, domainName, domainName)
}

func promptMD(description string) string {
	return fmt.Sprintf(`# %s

Describe what you want to create.

## Requirements

- List the requirements
- Include expected features

## Expected Outputs

- List the expected output files
`, description)
}

func beginnerMD(description string) string {
	return fmt.Sprintf(`I'm new to this and need help creating: %s

I think I need:

1. First output file - I think this is for [reason]?
2. Second output file

I want:
- Feature 1 (not sure how this works)
- Feature 2

Please explain what each part does.

## Questions I have

- Question about how something works
- Question about best practices
`, description)
}

func intermediateMD(description string) string {
	return fmt.Sprintf(`Create: %s

## Requirements

1. First output
   - Feature A
   - Feature B

2. Second output
   - Feature C
   - Feature D

## Constraints

- Any constraints or requirements
`, description)
}

func expertMD(description string) string {
	return fmt.Sprintf(`# %s

Brief technical requirements.

## Outputs

- output1: [brief spec]
- output2: [brief spec]

## Config

- Key configuration points
`, description)
}

func terseMD(description string) string {
	return fmt.Sprintf(`%s.

Key features, constraints.

Output: files.
`, description)
}

func verboseMD(description string) string {
	return fmt.Sprintf(`# Comprehensive Request: %s

## Background and Context

Explain the context and why this is needed. Include relevant background
information that might help understand the requirements better.

## Detailed Requirements

### Primary Output

Detailed description of the first output file, including:
- What it should contain
- Why each feature is needed
- Any specific configurations required

### Secondary Output

Detailed description of the second output file.

## Technical Specifications

- Specific technical requirements
- Version constraints
- Compatibility requirements

## Expected Behavior

Describe how the outputs should work together and what the end result
should look like when everything is properly configured.

## Additional Considerations

- Security considerations
- Performance considerations
- Maintenance considerations
`, description)
}

func gitignore() string {
	return `# Scenario run outputs
results/

# SVG recordings
*.svg
`
}
