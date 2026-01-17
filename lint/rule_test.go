package lint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// mockRule is a simple rule implementation for testing
type mockRule struct {
	id          string
	description string
	issues      []Issue
}

func (r *mockRule) ID() string {
	return r.id
}

func (r *mockRule) Description() string {
	return r.description
}

func (r *mockRule) Check(file *ast.File, fset *token.FileSet) []Issue {
	return r.issues
}

// mockFixableRule is a rule that can fix issues
type mockFixableRule struct {
	mockRule
	fixResult []byte
	fixErr    error
}

func (r *mockFixableRule) Fix(file *ast.File, fset *token.FileSet, issue Issue) ([]byte, error) {
	return r.fixResult, r.fixErr
}

func TestRuleInterface(t *testing.T) {
	rule := &mockRule{
		id:          "TEST001",
		description: "Test rule description",
		issues: []Issue{
			{Rule: "TEST001", Message: "test issue", Line: 1},
		},
	}

	// Verify it implements Rule interface
	var _ Rule = rule

	if rule.ID() != "TEST001" {
		t.Errorf("ID() = %q, want %q", rule.ID(), "TEST001")
	}
	if rule.Description() != "Test rule description" {
		t.Errorf("Description() = %q, want %q", rule.Description(), "Test rule description")
	}

	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "test.go", "package test", 0)

	issues := rule.Check(file, fset)
	if len(issues) != 1 {
		t.Errorf("Check() returned %d issues, want 1", len(issues))
	}
}

func TestFixableRuleInterface(t *testing.T) {
	rule := &mockFixableRule{
		mockRule: mockRule{
			id:          "TEST002",
			description: "Fixable test rule",
			issues: []Issue{
				{Rule: "TEST002", Message: "fixable issue", Fixable: true},
			},
		},
		fixResult: []byte("package test\n// fixed"),
	}

	// Verify it implements both Rule and FixableRule interfaces
	var _ Rule = rule
	var _ FixableRule = rule

	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "test.go", "package test", 0)

	issues := rule.Check(file, fset)
	if len(issues) != 1 {
		t.Fatalf("Check() returned %d issues, want 1", len(issues))
	}

	result, err := rule.Fix(file, fset, issues[0])
	if err != nil {
		t.Errorf("Fix() error = %v", err)
	}
	if string(result) != "package test\n// fixed" {
		t.Errorf("Fix() = %q, want %q", string(result), "package test\n// fixed")
	}
}

func TestRuleRegistration(t *testing.T) {
	registry := NewRuleRegistry()

	rule1 := &mockRule{id: "TEST001", description: "Rule 1"}
	rule2 := &mockRule{id: "TEST002", description: "Rule 2"}

	registry.Register(rule1)
	registry.Register(rule2)

	if got := registry.Get("TEST001"); got == nil {
		t.Error("Get(TEST001) = nil, want rule")
	}
	if got := registry.Get("TEST002"); got == nil {
		t.Error("Get(TEST002) = nil, want rule")
	}
	if got := registry.Get("NONEXISTENT"); got != nil {
		t.Error("Get(NONEXISTENT) = non-nil, want nil")
	}

	all := registry.All()
	if len(all) != 2 {
		t.Errorf("All() returned %d rules, want 2", len(all))
	}
}

func TestRuleRegistryIDs(t *testing.T) {
	registry := NewRuleRegistry()
	registry.Register(&mockRule{id: "B001", description: "Rule B"})
	registry.Register(&mockRule{id: "A001", description: "Rule A"})
	registry.Register(&mockRule{id: "C001", description: "Rule C"})

	ids := registry.IDs()
	if len(ids) != 3 {
		t.Errorf("IDs() returned %d, want 3", len(ids))
	}
	// IDs should be sorted
	if ids[0] != "A001" || ids[1] != "B001" || ids[2] != "C001" {
		t.Errorf("IDs() = %v, want sorted [A001 B001 C001]", ids)
	}
}
