package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScenarioYAML(t *testing.T) {
	result := scenarioYAML("test_scenario", "Test description")

	if !strings.Contains(result, "name: test_scenario") {
		t.Error("expected name field")
	}
	if !strings.Contains(result, "description: Test description") {
		t.Error("expected description field")
	}
	if !strings.Contains(result, "prompts:") {
		t.Error("expected prompts section")
	}
	if !strings.Contains(result, "default: prompt.md") {
		t.Error("expected default prompt")
	}
	if !strings.Contains(result, "beginner: prompts/beginner.md") {
		t.Error("expected beginner variant")
	}
	if !strings.Contains(result, "domains:") {
		t.Error("expected domains section")
	}
}

func TestSystemPromptMD(t *testing.T) {
	result := systemPromptMD()

	if !strings.Contains(result, "infrastructure engineer") {
		t.Error("expected infrastructure engineer mention")
	}
	if !strings.Contains(result, "Write tool") {
		t.Error("expected Write tool mention")
	}
	if !strings.Contains(result, "production-quality") {
		t.Error("expected production-quality guideline")
	}
}

func TestPromptMD(t *testing.T) {
	result := promptMD("Test description")

	if !strings.Contains(result, "Test Description") {
		t.Error("expected title with description")
	}
	if !strings.Contains(result, "Requirements") {
		t.Error("expected Requirements section")
	}
	if !strings.Contains(result, "Expected Outputs") {
		t.Error("expected Expected Outputs section")
	}
}

func TestBeginnerMD(t *testing.T) {
	result := beginnerMD("Test description")

	if !strings.Contains(result, "I'm new") {
		t.Error("expected beginner language")
	}
	if !strings.Contains(result, "Test description") {
		t.Error("expected description")
	}
	if !strings.Contains(result, "Please explain") {
		t.Error("expected request for explanation")
	}
	if !strings.Contains(result, "Questions I have") {
		t.Error("expected questions section")
	}
}

func TestIntermediateMD(t *testing.T) {
	result := intermediateMD("Test description")

	if !strings.Contains(result, "Create: Test description") {
		t.Error("expected create directive")
	}
	if !strings.Contains(result, "Requirements") {
		t.Error("expected Requirements section")
	}
	if !strings.Contains(result, "Constraints") {
		t.Error("expected Constraints section")
	}
}

func TestExpertMD(t *testing.T) {
	result := expertMD("Test description")

	if !strings.Contains(result, "Test description") {
		t.Error("expected description")
	}
	if !strings.Contains(result, "Outputs") {
		t.Error("expected Outputs section")
	}
	if !strings.Contains(result, "Config") {
		t.Error("expected Config section")
	}
}

func TestTerseMD(t *testing.T) {
	result := terseMD("Test description")

	if !strings.Contains(result, "Test description") {
		t.Error("expected description")
	}
	// Terse should be short
	if len(result) > 200 {
		t.Errorf("terse prompt too long: %d chars", len(result))
	}
	if !strings.Contains(result, "Output:") {
		t.Error("expected Output directive")
	}
}

func TestVerboseMD(t *testing.T) {
	result := verboseMD("Test description")

	if !strings.Contains(result, "Comprehensive Request") {
		t.Error("expected Comprehensive Request title")
	}
	if !strings.Contains(result, "Background and Context") {
		t.Error("expected Background section")
	}
	if !strings.Contains(result, "Detailed Requirements") {
		t.Error("expected Detailed Requirements section")
	}
	if !strings.Contains(result, "Technical Specifications") {
		t.Error("expected Technical Specifications section")
	}
	if !strings.Contains(result, "Additional Considerations") {
		t.Error("expected Additional Considerations section")
	}
	// Verbose should be long
	if len(result) < 500 {
		t.Errorf("verbose prompt too short: %d chars", len(result))
	}
}

func TestGitignore(t *testing.T) {
	result := gitignore()

	if !strings.Contains(result, "results/") {
		t.Error("expected results/ entry")
	}
	if !strings.Contains(result, "*.svg") {
		t.Error("expected *.svg entry")
	}
}

func TestScenarioCreation(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "init-scenario-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	scenarioPath := filepath.Join(tmpDir, "test_scenario")
	description := "Test scenario description"
	name := filepath.Base(scenarioPath)

	// Create directories
	dirs := []string{
		scenarioPath,
		filepath.Join(scenarioPath, "prompts"),
		filepath.Join(scenarioPath, "expected"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
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
			t.Fatalf("failed to write %s: %v", filename, err)
		}
	}

	// Verify all files exist
	for filename := range files {
		path := filepath.Join(scenarioPath, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("file not created: %s", filename)
		}
	}

	// Verify directories exist
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("directory not created: %s", dir)
		}
	}
}

func TestAllPersonasHaveContent(t *testing.T) {
	desc := "test"
	personas := map[string]func(string) string{
		"beginner":     beginnerMD,
		"intermediate": intermediateMD,
		"expert":       expertMD,
		"terse":        terseMD,
		"verbose":      verboseMD,
	}

	for name, fn := range personas {
		result := fn(desc)
		if result == "" {
			t.Errorf("persona %s returned empty content", name)
		}
		if !strings.Contains(result, desc) {
			t.Errorf("persona %s doesn't include description", name)
		}
	}
}
