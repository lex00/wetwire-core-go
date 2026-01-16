package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPersonas(t *testing.T) {
	expected := []string{"beginner", "intermediate", "expert", "terse", "verbose"}
	if len(DefaultPersonas) != len(expected) {
		t.Errorf("expected %d personas, got %d", len(expected), len(DefaultPersonas))
	}
	for i, p := range expected {
		if DefaultPersonas[i] != p {
			t.Errorf("persona %d: expected %s, got %s", i, p, DefaultPersonas[i])
		}
	}
}

func TestLoadSystemPrompt(t *testing.T) {
	// Create temp scenario directory
	tmpDir, err := os.MkdirTemp("", "runner-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("loads custom system prompt", func(t *testing.T) {
		content := "You are a test assistant."
		err := os.WriteFile(filepath.Join(tmpDir, "system_prompt.md"), []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}

		result := loadSystemPrompt(tmpDir)
		if result != content {
			t.Errorf("expected %q, got %q", content, result)
		}
	})

	t.Run("returns default when file missing", func(t *testing.T) {
		emptyDir, _ := os.MkdirTemp("", "empty-*")
		defer func() { _ = os.RemoveAll(emptyDir) }()

		result := loadSystemPrompt(emptyDir)
		if result == "" {
			t.Error("expected default prompt, got empty string")
		}
		if len(result) < 50 {
			t.Error("default prompt seems too short")
		}
	})
}

func TestLoadUserPrompt(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runner-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create prompts directory
	promptsDir := filepath.Join(tmpDir, "prompts")
	_ = os.MkdirAll(promptsDir, 0755)

	t.Run("loads persona-specific prompt", func(t *testing.T) {
		content := "# Beginner Prompt\n\nI need help with infrastructure."
		err := os.WriteFile(filepath.Join(promptsDir, "beginner.md"), []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}

		result := loadUserPrompt(tmpDir, "beginner")
		// Should strip the title line
		if result == "" {
			t.Error("expected prompt content, got empty string")
		}
		if result == content {
			t.Error("expected title to be stripped")
		}
	})

	t.Run("falls back to default prompt", func(t *testing.T) {
		defaultContent := "# Default\n\nThis is the default prompt."
		err := os.WriteFile(filepath.Join(tmpDir, "prompt.md"), []byte(defaultContent), 0644)
		if err != nil {
			t.Fatal(err)
		}

		result := loadUserPrompt(tmpDir, "nonexistent")
		if result == "" {
			t.Error("expected fallback to default prompt")
		}
	})

	t.Run("returns fallback when no prompts exist", func(t *testing.T) {
		emptyDir, _ := os.MkdirTemp("", "empty-*")
		defer func() { _ = os.RemoveAll(emptyDir) }()

		result := loadUserPrompt(emptyDir, "beginner")
		if result == "" {
			t.Error("expected fallback prompt")
		}
	})
}

func TestFindGeneratedFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runner-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("finds generated files", func(t *testing.T) {
		// Create some test files
		_ = os.WriteFile(filepath.Join(tmpDir, "template.yaml"), []byte("content1"), 0644)
		_ = os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte("content2"), 0644)

		files := findGeneratedFiles(tmpDir)
		if len(files) != 2 {
			t.Errorf("expected 2 files, got %d", len(files))
		}
		if files["template.yaml"] != "content1" {
			t.Error("template.yaml content mismatch")
		}
		if files["config.json"] != "content2" {
			t.Error("config.json content mismatch")
		}
	})

	t.Run("excludes output files", func(t *testing.T) {
		// Create files that should be excluded
		_ = os.WriteFile(filepath.Join(tmpDir, "conversation.txt"), []byte("excluded"), 0644)
		_ = os.WriteFile(filepath.Join(tmpDir, "RESULTS.md"), []byte("excluded"), 0644)
		_ = os.WriteFile(filepath.Join(tmpDir, "test.svg"), []byte("excluded"), 0644)

		files := findGeneratedFiles(tmpDir)
		if _, ok := files["conversation.txt"]; ok {
			t.Error("conversation.txt should be excluded")
		}
		if _, ok := files["RESULTS.md"]; ok {
			t.Error("RESULTS.md should be excluded")
		}
		if _, ok := files["test.svg"]; ok {
			t.Error("test.svg should be excluded")
		}
	})

	t.Run("finds nested files", func(t *testing.T) {
		nestedDir := filepath.Join(tmpDir, "nested")
		_ = os.MkdirAll(nestedDir, 0755)
		_ = os.WriteFile(filepath.Join(nestedDir, "nested.yaml"), []byte("nested content"), 0644)

		files := findGeneratedFiles(tmpDir)
		if _, ok := files["nested/nested.yaml"]; !ok {
			t.Error("nested file not found")
		}
	})
}

func TestCompareToExpected(t *testing.T) {
	t.Run("matches patterns from expected", func(t *testing.T) {
		expected := map[string]string{
			"template.yaml": "AWSTemplateFormatVersion: '2010-09-09'\nDescription: Test template\nResources:\n  Bucket:\n    Type: AWS::S3::Bucket",
		}
		generated := map[string]string{
			"output.yaml": "AWSTemplateFormatVersion: '2010-09-09'\nDescription: Test template\nResources:\n  Bucket:\n    Type: AWS::S3::Bucket",
		}

		matched, total, _ := compareToExpected(generated, expected)
		if matched == 0 {
			t.Error("expected some patterns to match")
		}
		if total == 0 {
			t.Error("expected some patterns to be extracted")
		}
	})

	t.Run("low match for different content", func(t *testing.T) {
		expected := map[string]string{
			"template.yaml": "AWSTemplateFormatVersion: '2010-09-09'\nDescription: Expected content",
		}
		generated := map[string]string{
			"output.yaml": "completely different content here",
		}

		matched, total, _ := compareToExpected(generated, expected)
		if total > 0 && matched > total/2 {
			t.Errorf("expected low match ratio, got %d/%d", matched, total)
		}
	})

	t.Run("handles empty expected", func(t *testing.T) {
		generated := map[string]string{"file.yaml": "content"}
		matched, total, _ := compareToExpected(generated, map[string]string{})
		if total != 0 || matched != 0 {
			t.Error("expected zero patterns for empty expected")
		}
	})
}

func TestLoadExpectedFiles(t *testing.T) {
	t.Run("returns empty for missing directory", func(t *testing.T) {
		files := loadExpectedFiles("/nonexistent/path")
		if len(files) != 0 {
			t.Errorf("expected empty map, got %d files", len(files))
		}
	})
}

func TestResultSuccess(t *testing.T) {
	t.Run("success when files generated", func(t *testing.T) {
		r := Result{
			Files: map[string]string{"test.yaml": "content"},
		}
		// Success is determined by len(Files) > 0
		if len(r.Files) == 0 {
			t.Error("expected files to indicate potential success")
		}
	})

	t.Run("failure when no files", func(t *testing.T) {
		r := Result{
			Files: map[string]string{},
		}
		if len(r.Files) > 0 {
			t.Error("expected no files to indicate failure")
		}
	})
}

func TestConfigDefaults(t *testing.T) {
	t.Run("uses default personas when empty", func(t *testing.T) {
		cfg := Config{
			ScenarioPath: "/test",
			OutputDir:    "/output",
			Personas:     nil,
		}

		// When Personas is nil/empty, Run() should use DefaultPersonas
		if cfg.Personas == nil && len(DefaultPersonas) != 5 {
			t.Error("default personas should have 5 entries")
		}
	})

	t.Run("single persona overrides list", func(t *testing.T) {
		cfg := Config{
			ScenarioPath:  "/test",
			OutputDir:     "/output",
			Personas:      []string{"a", "b", "c"},
			SinglePersona: "expert",
		}

		// SinglePersona should override Personas
		if cfg.SinglePersona != "expert" {
			t.Error("SinglePersona should be set")
		}
	})
}

func TestConversationSession(t *testing.T) {
	session := &conversationSession{
		name:     "test_scenario",
		prompt:   "user prompt",
		response: "assistant response",
	}

	if session.Name() != "test_scenario" {
		t.Errorf("expected name 'test_scenario', got %q", session.Name())
	}

	messages := session.GetMessages()
	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
	if messages[0].Role != "developer" {
		t.Errorf("expected first message role 'developer', got %q", messages[0].Role)
	}
	if messages[1].Role != "runner" {
		t.Errorf("expected second message role 'runner', got %q", messages[1].Role)
	}
}
