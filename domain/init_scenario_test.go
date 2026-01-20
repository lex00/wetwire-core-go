package domain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldScenario(t *testing.T) {
	t.Run("generates all expected files", func(t *testing.T) {
		scenario := ScaffoldScenario("test-scenario", "A test scenario", "testdomain")

		expectedFiles := []string{
			"scenario.yaml",
			"system_prompt.md",
			"prompt.md",
			"prompts/beginner.md",
			"prompts/intermediate.md",
			"prompts/expert.md",
			".gitignore",
			"expected/.gitkeep",
		}

		if len(scenario.Files) != len(expectedFiles) {
			t.Errorf("Expected %d files, got %d", len(expectedFiles), len(scenario.Files))
		}

		for _, filename := range expectedFiles {
			if _, exists := scenario.Files[filename]; !exists {
				t.Errorf("Missing expected file: %s", filename)
			}
		}
	})

	t.Run("scenario.yaml contains correct content", func(t *testing.T) {
		scenario := ScaffoldScenario("my-scenario", "My test scenario", "aws")
		yaml := scenario.Files["scenario.yaml"]

		if !strings.Contains(yaml, "name: my-scenario") {
			t.Error("scenario.yaml missing name field")
		}
		if !strings.Contains(yaml, "description: My test scenario") {
			t.Error("scenario.yaml missing description field")
		}
		if !strings.Contains(yaml, "model: sonnet") {
			t.Error("scenario.yaml missing model field")
		}
		if !strings.Contains(yaml, "- name: aws") {
			t.Error("scenario.yaml missing domain name")
		}
	})

	t.Run("system_prompt.md references domain", func(t *testing.T) {
		scenario := ScaffoldScenario("test", "test", "kubernetes")
		prompt := scenario.Files["system_prompt.md"]

		if !strings.Contains(prompt, "kubernetes") {
			t.Error("system_prompt.md should reference the domain name")
		}
	})
}

func TestScaffoldCrossDomainScenario(t *testing.T) {
	t.Run("generates all expected files for single domain", func(t *testing.T) {
		scenario := ScaffoldCrossDomainScenario("test-scenario", "A test scenario", []string{"aws"})

		expectedFiles := []string{
			"scenario.yaml",
			"system_prompt.md",
			"prompt.md",
			"prompts/beginner.md",
			"prompts/intermediate.md",
			"prompts/expert.md",
			".gitignore",
			"expected/aws/.gitkeep",
		}

		if len(scenario.Files) != len(expectedFiles) {
			t.Errorf("Expected %d files, got %d", len(expectedFiles), len(scenario.Files))
		}

		for _, filename := range expectedFiles {
			if _, exists := scenario.Files[filename]; !exists {
				t.Errorf("Missing expected file: %s", filename)
			}
		}
	})

	t.Run("generates expected directories for multiple domains", func(t *testing.T) {
		domains := []string{"aws", "gitlab", "kubernetes"}
		scenario := ScaffoldCrossDomainScenario("multi-domain", "Multi-domain test", domains)

		for _, domain := range domains {
			expectedFile := "expected/" + domain + "/.gitkeep"
			if _, exists := scenario.Files[expectedFile]; !exists {
				t.Errorf("Missing expected directory marker: %s", expectedFile)
			}
		}
	})

	t.Run("scenario.yaml contains all domains", func(t *testing.T) {
		domains := []string{"aws", "gitlab"}
		scenario := ScaffoldCrossDomainScenario("test", "test", domains)
		yaml := scenario.Files["scenario.yaml"]

		for _, domain := range domains {
			if !strings.Contains(yaml, "name: "+domain) {
				t.Errorf("scenario.yaml missing domain: %s", domain)
			}
		}
	})

	t.Run("scenario.yaml includes cross_domain for multiple domains", func(t *testing.T) {
		domains := []string{"aws", "gitlab", "kubernetes"}
		scenario := ScaffoldCrossDomainScenario("test", "test", domains)
		yaml := scenario.Files["scenario.yaml"]

		if !strings.Contains(yaml, "cross_domain:") {
			t.Error("scenario.yaml should include cross_domain section for multiple domains")
		}

		// Check for relationships between consecutive domains
		if !strings.Contains(yaml, "from: aws") {
			t.Error("scenario.yaml missing cross-domain relationship from aws")
		}
		if !strings.Contains(yaml, "to: gitlab") {
			t.Error("scenario.yaml missing cross-domain relationship to gitlab")
		}
	})

	t.Run("scenario.yaml validation rules for each domain", func(t *testing.T) {
		domains := []string{"aws", "gitlab"}
		scenario := ScaffoldCrossDomainScenario("test", "test", domains)
		yaml := scenario.Files["scenario.yaml"]

		if !strings.Contains(yaml, "validation:") {
			t.Error("scenario.yaml should include validation section")
		}

		for _, domain := range domains {
			if !strings.Contains(yaml, domain+":\n    resources:") {
				t.Errorf("scenario.yaml missing validation for domain: %s", domain)
			}
		}
	})

	t.Run("system_prompt.md references all domains", func(t *testing.T) {
		domains := []string{"aws", "gitlab", "kubernetes"}
		scenario := ScaffoldCrossDomainScenario("test", "test", domains)
		prompt := scenario.Files["system_prompt.md"]

		if !strings.Contains(prompt, "multi-domain") {
			t.Error("system_prompt.md should mention multi-domain")
		}

		for _, domain := range domains {
			if !strings.Contains(prompt, domain) {
				t.Errorf("system_prompt.md should reference domain: %s", domain)
			}
		}
	})

	t.Run("handles single domain", func(t *testing.T) {
		scenario := ScaffoldCrossDomainScenario("single", "Single domain", []string{"aws"})
		yaml := scenario.Files["scenario.yaml"]

		// Single domain should not have cross_domain section
		if strings.Contains(yaml, "cross_domain:") {
			t.Error("single domain scenario should not have cross_domain section")
		}
	})

	t.Run("handles empty domains list", func(t *testing.T) {
		scenario := ScaffoldCrossDomainScenario("empty", "Empty domains", []string{})

		// Should still generate base files
		if _, exists := scenario.Files["scenario.yaml"]; !exists {
			t.Error("Should generate scenario.yaml even with empty domains")
		}
	})
}

func TestWriteScenario(t *testing.T) {
	t.Run("writes single domain scenario files", func(t *testing.T) {
		tmpDir := t.TempDir()
		scenario := ScaffoldScenario("test", "test scenario", "aws")

		created, err := WriteScenario(tmpDir, scenario)
		if err != nil {
			t.Fatalf("WriteScenario() error = %v", err)
		}

		if len(created) != len(scenario.Files) {
			t.Errorf("Created %d files, expected %d", len(created), len(scenario.Files))
		}

		// Verify files exist
		for filename := range scenario.Files {
			path := filepath.Join(tmpDir, filename)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("File not created: %s", filename)
			}
		}

		// Verify directories exist
		promptsDir := filepath.Join(tmpDir, "prompts")
		if _, err := os.Stat(promptsDir); os.IsNotExist(err) {
			t.Error("prompts/ directory not created")
		}

		expectedDir := filepath.Join(tmpDir, "expected")
		if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
			t.Error("expected/ directory not created")
		}
	})

	t.Run("writes multi-domain scenario files", func(t *testing.T) {
		tmpDir := t.TempDir()
		domains := []string{"aws", "gitlab", "kubernetes"}
		scenario := ScaffoldCrossDomainScenario("multi", "multi-domain test", domains)

		created, err := WriteScenario(tmpDir, scenario)
		if err != nil {
			t.Fatalf("WriteScenario() error = %v", err)
		}

		if len(created) != len(scenario.Files) {
			t.Errorf("Created %d files, expected %d", len(created), len(scenario.Files))
		}

		// Verify domain subdirectories exist
		for _, domain := range domains {
			domainDir := filepath.Join(tmpDir, "expected", domain)
			if _, err := os.Stat(domainDir); os.IsNotExist(err) {
				t.Errorf("Domain directory not created: %s", domainDir)
			}

			gitkeepPath := filepath.Join(domainDir, ".gitkeep")
			if _, err := os.Stat(gitkeepPath); os.IsNotExist(err) {
				t.Errorf(".gitkeep not created in: %s", domainDir)
			}
		}
	})

	t.Run("creates nested directories as needed", func(t *testing.T) {
		tmpDir := t.TempDir()
		scenario := &ScenarioFiles{
			Files: map[string]string{
				"deep/nested/path/file.txt": "content",
			},
		}

		_, err := WriteScenario(tmpDir, scenario)
		if err != nil {
			t.Fatalf("WriteScenario() error = %v", err)
		}

		filePath := filepath.Join(tmpDir, "deep/nested/path/file.txt")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("Nested file not created")
		}
	})

	t.Run("overwrites existing files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Write initial scenario
		scenario1 := &ScenarioFiles{
			Files: map[string]string{
				"test.txt": "initial content",
			},
		}
		_, err := WriteScenario(tmpDir, scenario1)
		if err != nil {
			t.Fatalf("First WriteScenario() error = %v", err)
		}

		// Write updated scenario
		scenario2 := &ScenarioFiles{
			Files: map[string]string{
				"test.txt": "updated content",
			},
		}
		_, err = WriteScenario(tmpDir, scenario2)
		if err != nil {
			t.Fatalf("Second WriteScenario() error = %v", err)
		}

		// Verify content was updated
		content, err := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(content) != "updated content" {
			t.Errorf("Content = %q, want %q", string(content), "updated content")
		}
	})
}

func TestFormatDomainList(t *testing.T) {
	t.Run("formats empty list", func(t *testing.T) {
		result := formatDomainList([]string{})
		if result != "" {
			t.Errorf("Expected empty string, got %q", result)
		}
	})

	t.Run("formats single domain", func(t *testing.T) {
		result := formatDomainList([]string{"aws"})
		expected := "- aws\n"
		if result != expected {
			t.Errorf("Result = %q, want %q", result, expected)
		}
	})

	t.Run("formats multiple domains", func(t *testing.T) {
		result := formatDomainList([]string{"aws", "gitlab", "kubernetes"})
		if !strings.Contains(result, "- aws\n") {
			t.Error("Missing aws in domain list")
		}
		if !strings.Contains(result, "- gitlab\n") {
			t.Error("Missing gitlab in domain list")
		}
		if !strings.Contains(result, "- kubernetes\n") {
			t.Error("Missing kubernetes in domain list")
		}
	})
}
