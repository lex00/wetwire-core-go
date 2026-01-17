package ast

import (
	"go/ast"
)

// ExtractTypeName extracts the type name and package name from a type expression.
// For pointer types, slice types, array types, and channel types, it unwraps to
// find the underlying type. Map types return empty strings.
// Returns (typeName, packageName).
func ExtractTypeName(expr ast.Expr) (string, string) {
	switch t := expr.(type) {
	case *ast.Ident:
		// Simple identifier: int, string, MyType
		return t.Name, ""

	case *ast.SelectorExpr:
		// Qualified identifier: time.Time, pkg.Type
		if x, ok := t.X.(*ast.Ident); ok {
			return t.Sel.Name, x.Name
		}
		return "", ""

	case *ast.StarExpr:
		// Pointer type: *int, *time.Time
		return ExtractTypeName(t.X)

	case *ast.ArrayType:
		// Slice or array type: []int, [5]int
		return ExtractTypeName(t.Elt)

	case *ast.ChanType:
		// Channel type: chan int
		return ExtractTypeName(t.Value)

	case *ast.MapType:
		// Map type: map[string]int - return empty
		return "", ""

	default:
		return "", ""
	}
}

// InferTypeFromValue attempts to infer the type name and package from an
// expression value (typically a composite literal or other value expression).
// Returns (typeName, packageName).
func InferTypeFromValue(expr ast.Expr) (string, string) {
	switch v := expr.(type) {
	case *ast.CompositeLit:
		// Composite literal: S{}, time.Time{}, []int{1,2,3}
		if v.Type != nil {
			return ExtractTypeName(v.Type)
		}
		return "", ""

	case *ast.UnaryExpr:
		// Address-of expression: &S{}
		return InferTypeFromValue(v.X)

	case *ast.CallExpr:
		// Function call - check if it's a type conversion
		if len(v.Args) == 1 {
			return ExtractTypeName(v.Fun)
		}
		return "", ""

	default:
		return "", ""
	}
}
