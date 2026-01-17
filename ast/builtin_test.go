package ast

import "testing"

func TestIsBuiltinType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Primitive types
		{"string is builtin", "string", true},
		{"int is builtin", "int", true},
		{"int8 is builtin", "int8", true},
		{"int16 is builtin", "int16", true},
		{"int32 is builtin", "int32", true},
		{"int64 is builtin", "int64", true},
		{"uint is builtin", "uint", true},
		{"uint8 is builtin", "uint8", true},
		{"uint16 is builtin", "uint16", true},
		{"uint32 is builtin", "uint32", true},
		{"uint64 is builtin", "uint64", true},
		{"uintptr is builtin", "uintptr", true},
		{"float32 is builtin", "float32", true},
		{"float64 is builtin", "float64", true},
		{"complex64 is builtin", "complex64", true},
		{"complex128 is builtin", "complex128", true},
		{"bool is builtin", "bool", true},
		{"byte is builtin", "byte", true},
		{"rune is builtin", "rune", true},
		{"error is builtin", "error", true},
		{"any is builtin", "any", true},

		// Non-builtin types
		{"custom type not builtin", "MyType", false},
		{"empty string not builtin", "", false},
		{"package.Type not builtin", "time.Time", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBuiltinType(tt.input)
			if result != tt.expected {
				t.Errorf("IsBuiltinType(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsBuiltinIdent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Builtin types (also identifiers)
		{"string is builtin ident", "string", true},
		{"int is builtin ident", "int", true},
		{"bool is builtin ident", "bool", true},
		{"error is builtin ident", "error", true},

		// Builtin functions
		{"len is builtin ident", "len", true},
		{"cap is builtin ident", "cap", true},
		{"make is builtin ident", "make", true},
		{"new is builtin ident", "new", true},
		{"append is builtin ident", "append", true},
		{"copy is builtin ident", "copy", true},
		{"delete is builtin ident", "delete", true},
		{"close is builtin ident", "close", true},
		{"panic is builtin ident", "panic", true},
		{"recover is builtin ident", "recover", true},
		{"print is builtin ident", "print", true},
		{"println is builtin ident", "println", true},
		{"complex is builtin ident", "complex", true},
		{"real is builtin ident", "real", true},
		{"imag is builtin ident", "imag", true},
		{"clear is builtin ident", "clear", true},
		{"min is builtin ident", "min", true},
		{"max is builtin ident", "max", true},

		// Builtin constants
		{"true is builtin ident", "true", true},
		{"false is builtin ident", "false", true},
		{"nil is builtin ident", "nil", true},
		{"iota is builtin ident", "iota", true},

		// Non-builtin
		{"custom ident not builtin", "myFunc", false},
		{"empty string not builtin", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBuiltinIdent(tt.input)
			if result != tt.expected {
				t.Errorf("IsBuiltinIdent(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsKeyword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Keywords
		{"break is keyword", "break", true},
		{"case is keyword", "case", true},
		{"chan is keyword", "chan", true},
		{"const is keyword", "const", true},
		{"continue is keyword", "continue", true},
		{"default is keyword", "default", true},
		{"defer is keyword", "defer", true},
		{"else is keyword", "else", true},
		{"fallthrough is keyword", "fallthrough", true},
		{"for is keyword", "for", true},
		{"func is keyword", "func", true},
		{"go is keyword", "go", true},
		{"goto is keyword", "goto", true},
		{"if is keyword", "if", true},
		{"import is keyword", "import", true},
		{"interface is keyword", "interface", true},
		{"map is keyword", "map", true},
		{"package is keyword", "package", true},
		{"range is keyword", "range", true},
		{"return is keyword", "return", true},
		{"select is keyword", "select", true},
		{"struct is keyword", "struct", true},
		{"switch is keyword", "switch", true},
		{"type is keyword", "type", true},
		{"var is keyword", "var", true},

		// Non-keywords
		{"string is not keyword", "string", false},
		{"myVar is not keyword", "myVar", false},
		{"empty is not keyword", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsKeyword(tt.input)
			if result != tt.expected {
				t.Errorf("IsKeyword(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
