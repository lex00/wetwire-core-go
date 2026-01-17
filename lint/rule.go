package lint

import (
	"go/ast"
	"go/token"
	"sort"
	"sync"
)

// Rule defines the interface for lint rules.
type Rule interface {
	// ID returns the unique identifier for this rule (e.g., "TEST001").
	ID() string
	// Description returns a brief description of what the rule checks.
	Description() string
	// Check analyzes the given file and returns any issues found.
	Check(file *ast.File, fset *token.FileSet) []Issue
}

// FixableRule is a Rule that can automatically fix the issues it finds.
type FixableRule interface {
	Rule
	// Fix attempts to fix the given issue in the file.
	// Returns the modified source code, or an error if the fix failed.
	Fix(file *ast.File, fset *token.FileSet, issue Issue) ([]byte, error)
}

// PackageAwareRule is a Rule that needs package-level context.
type PackageAwareRule interface {
	Rule
	// CheckPackage analyzes all files in a package together.
	// The files map uses filenames as keys and their parsed AST as values.
	CheckPackage(files map[string]*ast.File, fset *token.FileSet) []Issue
}

// RuleRegistry maintains a collection of rules.
type RuleRegistry struct {
	mu    sync.RWMutex
	rules map[string]Rule
}

// NewRuleRegistry creates a new empty rule registry.
func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{
		rules: make(map[string]Rule),
	}
}

// Register adds a rule to the registry.
// If a rule with the same ID already exists, it will be replaced.
func (r *RuleRegistry) Register(rule Rule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules[rule.ID()] = rule
}

// Get returns the rule with the given ID, or nil if not found.
func (r *RuleRegistry) Get(id string) Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.rules[id]
}

// All returns all registered rules.
func (r *RuleRegistry) All() []Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rules := make([]Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		rules = append(rules, rule)
	}
	return rules
}

// IDs returns all registered rule IDs in sorted order.
func (r *RuleRegistry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.rules))
	for id := range r.rules {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
