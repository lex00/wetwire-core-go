package ast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ParseOptions configures file and directory parsing behavior.
type ParseOptions struct {
	SkipTests   bool     // Skip *_test.go files
	SkipVendor  bool     // Skip vendor directories
	SkipHidden  bool     // Skip directories starting with .
	ExcludeDirs []string // Additional directories to exclude
}

// ParseFile parses a single Go source file and returns the AST and FileSet.
func ParseFile(path string) (*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return file, fset, nil
}

// ParseDir parses all Go source files in a directory and returns a map of
// filename to AST, along with the FileSet.
func ParseDir(dir string, opts ParseOptions) (map[string]*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	files := make(map[string]*ast.File)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}

		if opts.SkipTests && strings.HasSuffix(name, "_test.go") {
			continue
		}

		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, nil, err
		}
		files[name] = file
	}

	return files, fset, nil
}

// WalkGoFiles walks a directory tree and calls fn for each Go source file,
// respecting the provided options.
func WalkGoFiles(root string, opts ParseOptions, fn func(path string) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Handle directories
		if info.IsDir() {
			name := info.Name()

			// Skip hidden directories
			if opts.SkipHidden && strings.HasPrefix(name, ".") && path != root {
				return filepath.SkipDir
			}

			// Skip vendor directories
			if opts.SkipVendor && name == "vendor" {
				return filepath.SkipDir
			}

			// Skip excluded directories
			for _, excluded := range opts.ExcludeDirs {
				if name == excluded {
					return filepath.SkipDir
				}
			}

			return nil
		}

		// Handle files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		if opts.SkipTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}

		return fn(path)
	})
}
