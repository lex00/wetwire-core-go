package lint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// LintFile lints a single file with the given rules and config.
// Returns all issues found that pass the config filters.
func LintFile(path string, rules []Rule, cfg *Config) ([]Issue, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return lintAST(file, fset, path, rules, cfg), nil
}

// LintBytes lints source code from bytes with the given rules and config.
// The filename is used for error messages and issue reporting.
func LintBytes(src []byte, filename string, rules []Rule, cfg *Config) ([]Issue, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return lintAST(file, fset, filename, rules, cfg), nil
}

// LintDir lints all Go files in a directory (non-recursively).
func LintDir(dir string, rules []Rule, cfg *Config) ([]Issue, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var issues []Issue
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		fileIssues, err := LintFile(path, rules, cfg)
		if err != nil {
			return nil, err
		}
		issues = append(issues, fileIssues...)
	}

	return issues, nil
}

// LintDirRecursive lints all Go files in a directory and its subdirectories.
func LintDirRecursive(root string, rules []Rule, cfg *Config) ([]Issue, error) {
	var issues []Issue

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fileIssues, err := LintFile(path, rules, cfg)
		if err != nil {
			return err
		}
		issues = append(issues, fileIssues...)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return issues, nil
}

// lintAST runs all rules on the parsed AST and returns filtered issues.
func lintAST(file *ast.File, fset *token.FileSet, path string, rules []Rule, cfg *Config) []Issue {
	var issues []Issue

	for _, rule := range rules {
		ruleIssues := rule.Check(file, fset)
		for _, issue := range ruleIssues {
			// Set file path if not already set
			if issue.File == "" {
				issue.File = path
			}

			// Filter by config
			if cfg != nil && !cfg.ShouldReport(issue) {
				continue
			}

			issues = append(issues, issue)
		}
	}

	return issues
}
