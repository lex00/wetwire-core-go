package discover

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Discover finds resources in the specified packages/directories.
func Discover(opts DiscoverOptions) (*DiscoverResult, error) {
	result := NewDiscoverResult()

	for _, pkg := range opts.Packages {
		// Check if it's a directory
		info, err := os.Stat(pkg)
		if err != nil {
			result.AddError(err)
			continue
		}

		if info.IsDir() {
			// Recursively discover in directory
			walkOpts := WalkOptions{
				SkipTests:    true,
				SkipVendor:   true,
				SkipHidden:   true,
				SkipTestdata: true,
			}
			err := WalkDir(pkg, walkOpts, func(path string) error {
				fileResult, err := DiscoverFile(path, opts.TypeMatcher)
				if err != nil {
					result.AddError(err)
					return nil // Continue walking
				}
				result.Merge(fileResult)
				return nil
			})
			if err != nil {
				result.AddError(err)
			}
		} else {
			// Single file
			fileResult, err := DiscoverFile(pkg, opts.TypeMatcher)
			if err != nil {
				result.AddError(err)
				continue
			}
			result.Merge(fileResult)
		}
	}

	return result, nil
}

// DiscoverFile finds resources in a single Go file.
func DiscoverFile(filePath string, matcher TypeMatcher) (*DiscoverResult, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return DiscoverAST(fset, file, filePath, matcher), nil
}

// DiscoverAST finds resources in a parsed AST.
func DiscoverAST(fset *token.FileSet, file *ast.File, filePath string, matcher TypeMatcher) *DiscoverResult {
	result := NewDiscoverResult()

	// Extract imports for type resolution
	imports := extractImports(file)

	// Walk all declarations
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, name := range valueSpec.Names {
				// Track all variables
				result.AddVar(name.Name)

				// Try to match the type
				if matcher == nil {
					continue
				}

				// Get the type from either explicit type or value
				var pkgName, typeName string
				if valueSpec.Type != nil {
					pkgName, typeName = extractTypeInfo(valueSpec.Type)
				} else if i < len(valueSpec.Values) {
					pkgName, typeName = inferTypeFromValue(valueSpec.Values[i])
				}

				if typeName == "" {
					continue
				}

				// Check if this is a resource type
				resourceType, ok := matcher(pkgName, typeName, imports)
				if !ok {
					continue
				}

				// Get position
				pos := fset.Position(name.Pos())

				// Extract dependencies from the value
				var deps []string
				if i < len(valueSpec.Values) {
					deps = extractDependencies(valueSpec.Values[i], result.AllVars)
				}

				result.AddResource(DiscoveredResource{
					Name:         name.Name,
					Type:         resourceType,
					File:         filePath,
					Line:         pos.Line,
					Dependencies: deps,
				})
			}
		}
	}

	return result
}

// DiscoverDir discovers resources in all Go files in a directory (non-recursively).
func DiscoverDir(dir string, matcher TypeMatcher) (*DiscoverResult, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	result := NewDiscoverResult()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		fileResult, err := DiscoverFile(filePath, matcher)
		if err != nil {
			result.AddError(err)
			continue
		}
		result.Merge(fileResult)
	}

	return result, nil
}

// extractImports extracts imports from an AST file into a map of alias -> path.
func extractImports(file *ast.File) map[string]string {
	imports := make(map[string]string)
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		var alias string
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			alias = path.Base(importPath)
		}
		imports[alias] = importPath
	}
	return imports
}

// extractTypeInfo extracts package name and type name from a type expression.
func extractTypeInfo(expr ast.Expr) (pkgName, typeName string) {
	switch t := expr.(type) {
	case *ast.Ident:
		return "", t.Name
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name, t.Sel.Name
		}
	case *ast.StarExpr:
		return extractTypeInfo(t.X)
	case *ast.ArrayType:
		return extractTypeInfo(t.Elt)
	}
	return "", ""
}

// inferTypeFromValue attempts to infer the type from a value expression.
func inferTypeFromValue(expr ast.Expr) (pkgName, typeName string) {
	switch v := expr.(type) {
	case *ast.UnaryExpr:
		// &schema.NodeType{} -> look at the composite lit
		return inferTypeFromValue(v.X)
	case *ast.CompositeLit:
		if v.Type != nil {
			return extractTypeInfo(v.Type)
		}
	case *ast.CallExpr:
		// Type conversion: schema.NodeType(...)
		return extractTypeInfo(v.Fun)
	}
	return "", ""
}

// extractDependencies finds identifier references in a value expression.
func extractDependencies(expr ast.Expr, knownVars map[string]bool) []string {
	var deps []string
	seen := make(map[string]bool)

	ast.Inspect(expr, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			name := ident.Name
			// Only include if it's a known variable and not already seen
			if knownVars[name] && !seen[name] {
				deps = append(deps, name)
				seen[name] = true
			}
		}
		return true
	})

	return deps
}
