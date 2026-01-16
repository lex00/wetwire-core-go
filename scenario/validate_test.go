package scenario

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateStructure(t *testing.T) {
	// Test with the aws_gitlab example scenario
	result := ValidateStructure("../examples/aws_gitlab")
	if !result.IsValid() {
		t.Errorf("aws_gitlab scenario should be valid, got errors:\n%s", result.Error())
	}
}

func TestValidateStructure_MissingDirectory(t *testing.T) {
	result := ValidateStructure("/nonexistent/path")
	if result.IsValid() {
		t.Error("expected validation error for missing directory")
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
	if result.Errors[0].Message != "scenario directory does not exist" {
		t.Errorf("unexpected error message: %s", result.Errors[0].Message)
	}
}

func TestValidateStructure_MissingFiles(t *testing.T) {
	// Create a temporary directory with missing files
	tmpDir := t.TempDir()

	result := ValidateStructure(tmpDir)
	if result.IsValid() {
		t.Error("expected validation errors for missing files")
	}

	// Should have errors for scenario.yaml, system_prompt.md, prompt.md, .gitignore, and 5 persona prompts
	expectedMinErrors := 9
	if len(result.Errors) < expectedMinErrors {
		t.Errorf("expected at least %d errors, got %d", expectedMinErrors, len(result.Errors))
	}
}

func TestValidateStructure_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create required directories and files
	if err := os.MkdirAll(filepath.Join(tmpDir, "prompts"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create invalid YAML
	if err := os.WriteFile(filepath.Join(tmpDir, "scenario.yaml"), []byte("invalid: [yaml:"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "system_prompt.md"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "prompt.md"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("results/\n*.svg"), 0644); err != nil {
		t.Fatal(err)
	}

	for _, p := range RequiredPersonas {
		if err := os.WriteFile(filepath.Join(tmpDir, "prompts", p+".md"), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	result := ValidateStructure(tmpDir)
	if result.IsValid() {
		t.Error("expected validation error for invalid YAML")
	}

	// Check for YAML error
	found := false
	for _, e := range result.Errors {
		if e.Message[:12] == "invalid YAML" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'invalid YAML' error")
	}
}

func TestValidateStructure_MissingGitignoreEntries(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tmpDir, "prompts"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create valid YAML with minimal required fields
	yaml := `name: test
description: test scenario
prompts:
  default: prompt.md
  variants:
    beginner: prompts/beginner.md
    intermediate: prompts/intermediate.md
    expert: prompts/expert.md
    terse: prompts/terse.md
    verbose: prompts/verbose.md
domains:
  - name: test
    outputs:
      - output.yaml
`
	if err := os.WriteFile(filepath.Join(tmpDir, "scenario.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "system_prompt.md"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "prompt.md"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	for _, p := range RequiredPersonas {
		if err := os.WriteFile(filepath.Join(tmpDir, "prompts", p+".md"), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	result := ValidateStructure(tmpDir)
	if result.IsValid() {
		t.Error("expected validation errors for missing gitignore entries")
	}

	// Should have errors for missing results/ and *.svg entries
	gitignoreErrors := 0
	for _, e := range result.Errors {
		if filepath.Base(e.Path) == ".gitignore" {
			gitignoreErrors++
		}
	}
	if gitignoreErrors != 2 {
		t.Errorf("expected 2 gitignore errors, got %d", gitignoreErrors)
	}
}

func TestValidateStructure_EmptyPrompts(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tmpDir, "prompts"), 0755); err != nil {
		t.Fatal(err)
	}

	yaml := `name: test
description: test scenario
prompts:
  default: prompt.md
  variants:
    beginner: prompts/beginner.md
    intermediate: prompts/intermediate.md
    expert: prompts/expert.md
    terse: prompts/terse.md
    verbose: prompts/verbose.md
domains:
  - name: test
    outputs:
      - output.yaml
`
	if err := os.WriteFile(filepath.Join(tmpDir, "scenario.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "system_prompt.md"), []byte("   "), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "prompt.md"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("results/\n*.svg"), 0644); err != nil {
		t.Fatal(err)
	}

	for _, p := range RequiredPersonas {
		if err := os.WriteFile(filepath.Join(tmpDir, "prompts", p+".md"), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	result := ValidateStructure(tmpDir)
	if result.IsValid() {
		t.Error("expected validation errors for empty prompts")
	}

	// Should have errors for empty system_prompt.md and prompt.md
	emptyErrors := 0
	for _, e := range result.Errors {
		if e.Message == "system_prompt.md is empty" || e.Message == "prompt.md is empty" {
			emptyErrors++
		}
	}
	if emptyErrors != 2 {
		t.Errorf("expected 2 empty prompt errors, got %d", emptyErrors)
	}
}

func TestRequiredPersonas(t *testing.T) {
	expected := []string{"beginner", "intermediate", "expert", "terse", "verbose"}
	if len(RequiredPersonas) != len(expected) {
		t.Errorf("expected %d personas, got %d", len(expected), len(RequiredPersonas))
	}
	for i, p := range expected {
		if RequiredPersonas[i] != p {
			t.Errorf("persona %d: expected %s, got %s", i, p, RequiredPersonas[i])
		}
	}
}

func TestStructureError(t *testing.T) {
	e := StructureError{Path: "/path/to/file", Message: "test error"}
	expected := "/path/to/file: test error"
	if e.Error() != expected {
		t.Errorf("expected %q, got %q", expected, e.Error())
	}
}

func TestStructureResult(t *testing.T) {
	t.Run("empty result is valid", func(t *testing.T) {
		r := &StructureResult{}
		if !r.IsValid() {
			t.Error("empty result should be valid")
		}
		if r.Error() != "" {
			t.Error("empty result error should be empty string")
		}
	})

	t.Run("result with errors is invalid", func(t *testing.T) {
		r := &StructureResult{
			Errors: []StructureError{
				{Path: "file1", Message: "error1"},
				{Path: "file2", Message: "error2"},
			},
		}
		if r.IsValid() {
			t.Error("result with errors should be invalid")
		}
		expected := "file1: error1\nfile2: error2"
		if r.Error() != expected {
			t.Errorf("expected %q, got %q", expected, r.Error())
		}
	})
}
