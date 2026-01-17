package lint

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// FixResult represents the result of attempting to fix an issue.
type FixResult struct {
	// Issue is the original lint issue.
	Issue Issue
	// Fixed indicates whether the issue was successfully fixed.
	Fixed bool
	// NewCode contains the fixed source code (if Fixed is true).
	NewCode []byte
	// Error contains any error that occurred during fixing.
	Error error
}

// Fix attempts to fix issues in a file without writing the changes.
// Returns a slice of FixResults indicating what was fixed.
func Fix(path string, rules []Rule, cfg *Config) ([]FixResult, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// First, find all issues
	var results []FixResult
	for _, rule := range rules {
		issues := rule.Check(file, fset)
		for _, issue := range issues {
			if issue.File == "" {
				issue.File = path
			}

			// Filter by config
			if cfg != nil && !cfg.ShouldReport(issue) {
				continue
			}

			result := FixResult{Issue: issue}

			// Check if rule is fixable
			if fixable, ok := rule.(FixableRule); ok && issue.Fixable {
				newCode, fixErr := fixable.Fix(file, fset, issue)
				if fixErr != nil {
					result.Error = fixErr
				} else {
					result.Fixed = true
					result.NewCode = newCode
				}
			}

			results = append(results, result)
		}
	}

	return results, nil
}

// FixFile fixes issues in a file and writes the changes back.
// Returns a slice of FixResults indicating what was fixed.
func FixFile(path string, rules []Rule, cfg *Config) ([]FixResult, error) {
	results, err := Fix(path, rules, cfg)
	if err != nil {
		return nil, err
	}

	// Write fixed content for each successful fix
	for _, result := range results {
		if result.Fixed && len(result.NewCode) > 0 {
			if err := os.WriteFile(path, result.NewCode, 0644); err != nil {
				return nil, err
			}
		}
	}

	return results, nil
}

// FixDir fixes issues in all Go files in a directory (non-recursively).
func FixDir(dir string, rules []Rule, cfg *Config) ([]FixResult, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var results []FixResult
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		fileResults, err := FixFile(path, rules, cfg)
		if err != nil {
			return nil, err
		}
		results = append(results, fileResults...)
	}

	return results, nil
}
