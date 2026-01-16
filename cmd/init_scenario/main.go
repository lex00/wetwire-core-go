// init_scenario scaffolds a new wetwire scenario with all required files.
//
// Usage:
//
//	go run ./cmd/init_scenario <path> <description>
//
// Examples:
//
//	go run ./cmd/init_scenario ./examples/eks_cluster "EKS cluster with ArgoCD"
//	go run ./cmd/init_scenario ./examples/lambda_api "Lambda API Gateway"
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: init_scenario <path> <description>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  init_scenario ./examples/eks_cluster \"EKS cluster with ArgoCD\"")
		fmt.Println("  init_scenario ./examples/lambda_api \"Lambda API Gateway\"")
		os.Exit(1)
	}

	scenarioPath := os.Args[1]
	description := os.Args[2]

	// Extract name from path
	name := filepath.Base(scenarioPath)

	fmt.Printf("Creating scenario: %s\n", name)
	fmt.Printf("Path: %s\n", scenarioPath)
	fmt.Printf("Description: %s\n\n", description)

	// Create directories
	dirs := []string{
		scenarioPath,
		filepath.Join(scenarioPath, "prompts"),
		filepath.Join(scenarioPath, "expected"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// Create files
	files := map[string]string{
		"scenario.yaml":           scenarioYAML(name, description),
		"system_prompt.md":        systemPromptMD(),
		"prompt.md":               promptMD(description),
		"prompts/beginner.md":     beginnerMD(description),
		"prompts/intermediate.md": intermediateMD(description),
		"prompts/expert.md":       expertMD(description),
		"prompts/terse.md":        terseMD(description),
		"prompts/verbose.md":      verboseMD(description),
		".gitignore":              gitignore(),
		"expected/.gitkeep":       "",
	}

	for filename, content := range files {
		path := filepath.Join(scenarioPath, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Printf("  Created %s\n", filename)
	}

	fmt.Println()
	fmt.Println("Scenario scaffolded successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit scenario.yaml to define expected outputs")
	fmt.Println("  2. Edit system_prompt.md with domain-specific instructions")
	fmt.Println("  3. Edit prompt.md and prompts/*.md with your requirements")
	fmt.Println("  4. Add expected output files to expected/ (optional)")
	fmt.Println("  5. Run: go run ./cmd/run_scenario " + scenarioPath + " --all")
}

func scenarioYAML(name, description string) string {
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
  - name: TODO_DOMAIN
    outputs:
      - TODO_output.yaml

validation:
  TODO_DOMAIN:
    resources:
      min: 1
`, name, description)
}

func systemPromptMD() string {
	return `You are a helpful infrastructure engineer assistant.

Your task is to help users create infrastructure files based on their requirements.
Use the Write tool to create files. Use mkdir via Bash if needed.

Guidelines:
- Always generate complete, production-quality infrastructure regardless of how brief the request is
- If the user asks questions, answer them
- If the user asks for explanations, provide them
- Include best practices (parameters, outputs, proper configurations) even if not explicitly requested
`
}

func promptMD(description string) string {
	return fmt.Sprintf(`# %s

TODO: Describe what the user wants to create.

## Requirements

- TODO: List the requirements
- TODO: Include expected features

## Expected Outputs

- TODO: List the expected output files
`, description)
}

func beginnerMD(description string) string {
	return fmt.Sprintf(`I'm new to this and need help creating: %s

I think I need:

1. TODO: First output file - I think this is for [reason]?
2. TODO: Second output file

I want:
- TODO: Feature 1 (not sure how this works)
- TODO: Feature 2

Please explain what each part does.

## Questions I have

- TODO: Question about how something works
- TODO: Question about best practices
`, description)
}

func intermediateMD(description string) string {
	return fmt.Sprintf(`Create: %s

## Requirements

1. TODO: First output
   - Feature A
   - Feature B

2. TODO: Second output
   - Feature C
   - Feature D

## Constraints

- TODO: Any constraints or requirements
`, description)
}

func expertMD(description string) string {
	return fmt.Sprintf(`# %s

TODO: Brief technical requirements.

## Outputs

- `+"`TODO_output1.yaml`"+`: [brief spec]
- `+"`TODO_output2.yaml`"+`: [brief spec]

## Config

- TODO: Key configuration points
`, description)
}

func terseMD(description string) string {
	return fmt.Sprintf(`%s.

TODO: Key features, constraints.

Output: TODO_files.
`, description)
}

func verboseMD(description string) string {
	return fmt.Sprintf(`# Comprehensive Request: %s

## Background and Context

TODO: Explain the context and why this is needed. Include relevant background
information that might help understand the requirements better.

## Detailed Requirements

### Primary Output: TODO

TODO: Detailed description of the first output file, including:
- What it should contain
- Why each feature is needed
- Any specific configurations required

### Secondary Output: TODO

TODO: Detailed description of the second output file.

## Technical Specifications

- TODO: Specific technical requirements
- TODO: Version constraints
- TODO: Compatibility requirements

## Expected Behavior

TODO: Describe how the outputs should work together and what the end result
should look like when everything is properly configured.

## Additional Considerations

- TODO: Security considerations
- TODO: Performance considerations
- TODO: Maintenance considerations
`, description)
}

func gitignore() string {
	return `# Scenario run outputs
results/

# SVG recordings
*.svg
`
}
