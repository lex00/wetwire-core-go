package ast

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestExtractImports(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected map[string]string
	}{
		{
			name: "no imports",
			code: `package test`,
			expected: map[string]string{},
		},
		{
			name: "single import",
			code: `package test
import "fmt"`,
			expected: map[string]string{
				"fmt": "fmt",
			},
		},
		{
			name: "multiple imports",
			code: `package test
import (
	"fmt"
	"strings"
)`,
			expected: map[string]string{
				"fmt":     "fmt",
				"strings": "strings",
			},
		},
		{
			name: "aliased import",
			code: `package test
import (
	f "fmt"
	str "strings"
)`,
			expected: map[string]string{
				"f":   "fmt",
				"str": "strings",
			},
		},
		{
			name: "dot import",
			code: `package test
import . "fmt"`,
			expected: map[string]string{
				".": "fmt",
			},
		},
		{
			name: "blank import",
			code: `package test
import _ "fmt"`,
			expected: map[string]string{
				"_": "fmt",
			},
		},
		{
			name: "nested package",
			code: `package test
import "github.com/user/repo/pkg"`,
			expected: map[string]string{
				"pkg": "github.com/user/repo/pkg",
			},
		},
		{
			name: "mixed imports",
			code: `package test
import (
	"fmt"
	str "strings"
	. "testing"
	_ "embed"
	"github.com/user/repo/pkg"
)`,
			expected: map[string]string{
				"fmt": "fmt",
				"str": "strings",
				".":   "testing",
				"_":   "embed",
				"pkg": "github.com/user/repo/pkg",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("failed to parse test code: %v", err)
			}

			result := ExtractImports(file)

			if len(result) != len(tt.expected) {
				t.Errorf("ExtractImports() returned %d imports, want %d", len(result), len(tt.expected))
			}

			for alias, path := range tt.expected {
				if result[alias] != path {
					t.Errorf("ExtractImports()[%q] = %q, want %q", alias, result[alias], path)
				}
			}
		})
	}
}
