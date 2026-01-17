package ast

import (
	goast "go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestExtractTypeName(t *testing.T) {
	tests := []struct {
		name         string
		code         string
		expectedType string
		expectedPkg  string
	}{
		{
			name:         "simple ident",
			code:         `package test; var x int`,
			expectedType: "int",
			expectedPkg:  "",
		},
		{
			name:         "qualified ident",
			code:         `package test; import "time"; var x time.Time`,
			expectedType: "Time",
			expectedPkg:  "time",
		},
		{
			name:         "pointer type",
			code:         `package test; var x *int`,
			expectedType: "int",
			expectedPkg:  "",
		},
		{
			name:         "qualified pointer",
			code:         `package test; import "time"; var x *time.Time`,
			expectedType: "Time",
			expectedPkg:  "time",
		},
		{
			name:         "slice type",
			code:         `package test; var x []int`,
			expectedType: "int",
			expectedPkg:  "",
		},
		{
			name:         "array type",
			code:         `package test; var x [5]int`,
			expectedType: "int",
			expectedPkg:  "",
		},
		{
			name:         "map type returns empty",
			code:         `package test; var x map[string]int`,
			expectedType: "",
			expectedPkg:  "",
		},
		{
			name:         "channel type",
			code:         `package test; var x chan int`,
			expectedType: "int",
			expectedPkg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			if err != nil {
				t.Fatalf("failed to parse test code: %v", err)
			}

			// Find the var declaration's type
			var typeExpr goast.Expr
			for _, decl := range file.Decls {
				if gd, ok := decl.(*goast.GenDecl); ok && gd.Tok == token.VAR {
					if len(gd.Specs) > 0 {
						if vs, ok := gd.Specs[0].(*goast.ValueSpec); ok {
							typeExpr = vs.Type
						}
					}
				}
			}

			if typeExpr == nil {
				t.Fatal("no type expression found")
			}

			typeName, pkgName := ExtractTypeName(typeExpr)

			if typeName != tt.expectedType {
				t.Errorf("ExtractTypeName() typeName = %q, want %q", typeName, tt.expectedType)
			}
			if pkgName != tt.expectedPkg {
				t.Errorf("ExtractTypeName() pkgName = %q, want %q", pkgName, tt.expectedPkg)
			}
		})
	}
}

func TestInferTypeFromCompositeLit(t *testing.T) {
	tests := []struct {
		name         string
		code         string
		expectedType string
		expectedPkg  string
	}{
		{
			name:         "simple struct literal",
			code:         `package test; type S struct{}; var x = S{}`,
			expectedType: "S",
			expectedPkg:  "",
		},
		{
			name:         "qualified struct literal",
			code:         `package test; import "time"; var x = time.Time{}`,
			expectedType: "Time",
			expectedPkg:  "time",
		},
		{
			name:         "slice literal",
			code:         `package test; var x = []int{1, 2, 3}`,
			expectedType: "int",
			expectedPkg:  "",
		},
		{
			name:         "map literal returns empty",
			code:         `package test; var x = map[string]int{}`,
			expectedType: "",
			expectedPkg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			if err != nil {
				t.Fatalf("failed to parse test code: %v", err)
			}

			// Find the composite literal
			var compLit *goast.CompositeLit
			goast.Inspect(file, func(n goast.Node) bool {
				if cl, ok := n.(*goast.CompositeLit); ok {
					compLit = cl
					return false
				}
				return true
			})

			if compLit == nil {
				t.Fatal("no composite literal found")
			}

			typeName, pkgName := InferTypeFromValue(compLit)

			if typeName != tt.expectedType {
				t.Errorf("InferTypeFromValue() typeName = %q, want %q", typeName, tt.expectedType)
			}
			if pkgName != tt.expectedPkg {
				t.Errorf("InferTypeFromValue() pkgName = %q, want %q", pkgName, tt.expectedPkg)
			}
		})
	}
}
