package lint

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

// testRule creates issues for files with a specific pattern
type testRule struct {
	id      string
	trigger string // if file contains this string, create issue
}

func (r *testRule) ID() string          { return r.id }
func (r *testRule) Description() string { return "Test rule: " + r.id }
func (r *testRule) Check(file *ast.File, fset *token.FileSet) []Issue {
	// Simple rule: report if package name matches trigger
	if file.Name.Name == r.trigger {
		return []Issue{{
			Rule:     r.id,
			Message:  "package name matches trigger",
			Line:     1,
			Severity: SeverityError,
		}}
	}
	return nil
}

func TestLintFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package trigger

var X = 1
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	rules := []Rule{
		&testRule{id: "TEST001", trigger: "trigger"},
		&testRule{id: "TEST002", trigger: "other"},
	}

	t.Run("returns issues matching rules", func(t *testing.T) {
		issues, err := LintFile(testFile, rules, nil)
		if err != nil {
			t.Fatalf("LintFile() error = %v", err)
		}
		if len(issues) != 1 {
			t.Errorf("LintFile() returned %d issues, want 1", len(issues))
		}
		if len(issues) > 0 && issues[0].Rule != "TEST001" {
			t.Errorf("Issue.Rule = %q, want %q", issues[0].Rule, "TEST001")
		}
	})

	t.Run("sets file path on issues", func(t *testing.T) {
		issues, _ := LintFile(testFile, rules, nil)
		if len(issues) > 0 && issues[0].File != testFile {
			t.Errorf("Issue.File = %q, want %q", issues[0].File, testFile)
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := LintFile("/nonexistent/file.go", rules, nil)
		if err == nil {
			t.Error("LintFile() expected error for non-existent file")
		}
	})

	t.Run("respects disabled rules in config", func(t *testing.T) {
		cfg := &Config{
			DisabledRules: []string{"TEST001"},
		}
		issues, err := LintFile(testFile, rules, cfg)
		if err != nil {
			t.Fatalf("LintFile() error = %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("LintFile() returned %d issues, want 0 (rule disabled)", len(issues))
		}
	})

	t.Run("respects min severity in config", func(t *testing.T) {
		// Write a file that triggers a warning-level rule
		warningFile := filepath.Join(tmpDir, "warning.go")
		if err := os.WriteFile(warningFile, []byte("package warning\n"), 0644); err != nil {
			t.Fatal(err)
		}

		warningRule := &warningTestRule{id: "WARN001", trigger: "warning"}
		rules := []Rule{warningRule}

		// With error min severity, warnings should be filtered
		cfg := &Config{MinSeverity: SeverityError}
		issues, _ := LintFile(warningFile, rules, cfg)
		if len(issues) != 0 {
			t.Errorf("LintFile() returned %d issues, want 0 (filtered by severity)", len(issues))
		}

		// With warning min severity, warnings should be included
		cfg = &Config{MinSeverity: SeverityWarning}
		issues, _ = LintFile(warningFile, rules, cfg)
		if len(issues) != 1 {
			t.Errorf("LintFile() returned %d issues, want 1", len(issues))
		}
	})
}

type warningTestRule struct {
	id      string
	trigger string
}

func (r *warningTestRule) ID() string          { return r.id }
func (r *warningTestRule) Description() string { return "Warning rule: " + r.id }
func (r *warningTestRule) Check(file *ast.File, fset *token.FileSet) []Issue {
	if file.Name.Name == r.trigger {
		return []Issue{{
			Rule:     r.id,
			Message:  "warning triggered",
			Severity: SeverityWarning,
		}}
	}
	return nil
}

func TestLintDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"trigger.go": "package trigger\nvar A = 1",
		"other.go":   "package other\nvar B = 2",
		"test.go":    "package trigger\nvar C = 3", // also triggers
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	rules := []Rule{
		&testRule{id: "TEST001", trigger: "trigger"},
	}

	t.Run("lints all files in directory", func(t *testing.T) {
		issues, err := LintDir(tmpDir, rules, nil)
		if err != nil {
			t.Fatalf("LintDir() error = %v", err)
		}
		// Should find 2 files with "trigger" package
		if len(issues) != 2 {
			t.Errorf("LintDir() returned %d issues, want 2", len(issues))
		}
	})

	t.Run("returns error for non-existent directory", func(t *testing.T) {
		_, err := LintDir("/nonexistent/dir", rules, nil)
		if err == nil {
			t.Error("LintDir() expected error for non-existent directory")
		}
	})
}

func TestLintDirRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test files at different levels
	files := map[string]string{
		filepath.Join(tmpDir, "root.go"):   "package trigger",
		filepath.Join(subDir, "nested.go"): "package trigger",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	rules := []Rule{
		&testRule{id: "TEST001", trigger: "trigger"},
	}

	issues, err := LintDirRecursive(tmpDir, rules, nil)
	if err != nil {
		t.Fatalf("LintDirRecursive() error = %v", err)
	}
	if len(issues) != 2 {
		t.Errorf("LintDirRecursive() returned %d issues, want 2", len(issues))
	}
}

func TestLintBytes(t *testing.T) {
	rules := []Rule{
		&testRule{id: "TEST001", trigger: "trigger"},
	}

	t.Run("lints source code bytes", func(t *testing.T) {
		source := []byte("package trigger\nvar X = 1")
		issues, err := LintBytes(source, "test.go", rules, nil)
		if err != nil {
			t.Fatalf("LintBytes() error = %v", err)
		}
		if len(issues) != 1 {
			t.Errorf("LintBytes() returned %d issues, want 1", len(issues))
		}
	})

	t.Run("returns error for invalid source", func(t *testing.T) {
		source := []byte("not valid go code")
		_, err := LintBytes(source, "test.go", rules, nil)
		if err == nil {
			t.Error("LintBytes() expected error for invalid source")
		}
	})
}
