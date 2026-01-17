package lint

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

// fixableTestRule is a rule that can fix issues
type fixableTestRule struct {
	id      string
	trigger string
}

func (r *fixableTestRule) ID() string          { return r.id }
func (r *fixableTestRule) Description() string { return "Fixable rule: " + r.id }
func (r *fixableTestRule) Check(file *ast.File, fset *token.FileSet) []Issue {
	if file.Name.Name == r.trigger {
		return []Issue{{
			Rule:     r.id,
			Message:  "fixable issue",
			Fixable:  true,
			Severity: SeverityError,
		}}
	}
	return nil
}

func (r *fixableTestRule) Fix(file *ast.File, fset *token.FileSet, issue Issue) ([]byte, error) {
	// Simple fix: change package name to "fixed"
	return []byte("package fixed\n"), nil
}

func TestFix(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package trigger

var X = 1
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	fixableRule := &fixableTestRule{id: "FIX001", trigger: "trigger"}
	rules := []Rule{fixableRule}

	t.Run("fixes issues and returns results", func(t *testing.T) {
		results, err := Fix(testFile, rules, nil)
		if err != nil {
			t.Fatalf("Fix() error = %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Fix() returned %d results, want 1", len(results))
		}
		if len(results) > 0 {
			if !results[0].Fixed {
				t.Error("FixResult.Fixed = false, want true")
			}
			if results[0].Issue.Rule != "FIX001" {
				t.Errorf("FixResult.Issue.Rule = %q, want %q", results[0].Issue.Rule, "FIX001")
			}
		}
	})

	t.Run("skips non-fixable issues", func(t *testing.T) {
		nonFixableRule := &testRule{id: "NOFIX001", trigger: "trigger"}
		rules := []Rule{nonFixableRule}
		results, err := Fix(testFile, rules, nil)
		if err != nil {
			t.Fatalf("Fix() error = %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Fix() returned %d results, want 1", len(results))
		}
		if len(results) > 0 && results[0].Fixed {
			t.Error("FixResult.Fixed = true, want false (rule not fixable)")
		}
	})
}

func TestFixFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package trigger

var X = 1
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	fixableRule := &fixableTestRule{id: "FIX001", trigger: "trigger"}
	rules := []Rule{fixableRule}

	t.Run("writes fixed content to file", func(t *testing.T) {
		results, err := FixFile(testFile, rules, nil)
		if err != nil {
			t.Fatalf("FixFile() error = %v", err)
		}
		if len(results) == 0 {
			t.Fatal("FixFile() returned no results")
		}
		if !results[0].Fixed {
			t.Error("FixResult.Fixed = false, want true")
		}

		// Verify file was modified
		newContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read fixed file: %v", err)
		}
		if string(newContent) != "package fixed\n" {
			t.Errorf("fixed content = %q, want %q", string(newContent), "package fixed\n")
		}
	})
}

func TestFixResult(t *testing.T) {
	result := FixResult{
		Issue: Issue{
			Rule:    "TEST001",
			Message: "test issue",
		},
		Fixed:   true,
		NewCode: []byte("fixed code"),
	}

	if result.Issue.Rule != "TEST001" {
		t.Errorf("FixResult.Issue.Rule = %q, want %q", result.Issue.Rule, "TEST001")
	}
	if !result.Fixed {
		t.Error("FixResult.Fixed = false, want true")
	}
	if string(result.NewCode) != "fixed code" {
		t.Errorf("FixResult.NewCode = %q, want %q", string(result.NewCode), "fixed code")
	}
}

func TestFixDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"trigger1.go": "package trigger",
		"trigger2.go": "package trigger",
		"other.go":    "package other",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	fixableRule := &fixableTestRule{id: "FIX001", trigger: "trigger"}
	rules := []Rule{fixableRule}

	t.Run("fixes all files in directory", func(t *testing.T) {
		results, err := FixDir(tmpDir, rules, nil)
		if err != nil {
			t.Fatalf("FixDir() error = %v", err)
		}
		// Should fix 2 files
		fixedCount := 0
		for _, r := range results {
			if r.Fixed {
				fixedCount++
			}
		}
		if fixedCount != 2 {
			t.Errorf("FixDir() fixed %d files, want 2", fixedCount)
		}
	})
}
