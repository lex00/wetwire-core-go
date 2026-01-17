package ast

import (
	"go/ast"
	"path"
	"strings"
)

// ExtractImports extracts all imports from an AST file and returns a map
// of alias/name to import path. For imports without an explicit alias,
// the last component of the path is used as the key.
func ExtractImports(file *ast.File) map[string]string {
	imports := make(map[string]string)

	for _, imp := range file.Imports {
		// Get the import path (strip quotes)
		importPath := strings.Trim(imp.Path.Value, `"`)

		var alias string
		if imp.Name != nil {
			// Explicit alias (including "." and "_")
			alias = imp.Name.Name
		} else {
			// Use last component of path as implicit alias
			alias = path.Base(importPath)
		}

		imports[alias] = importPath
	}

	return imports
}
