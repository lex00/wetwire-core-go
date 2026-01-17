package ast

// builtinTypes contains all Go builtin types
var builtinTypes = map[string]bool{
	"string":     true,
	"int":        true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"uint":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"uintptr":    true,
	"float32":    true,
	"float64":    true,
	"complex64":  true,
	"complex128": true,
	"bool":       true,
	"byte":       true,
	"rune":       true,
	"error":      true,
	"any":        true,
}

// builtinFuncs contains all Go builtin functions
var builtinFuncs = map[string]bool{
	"len":     true,
	"cap":     true,
	"make":    true,
	"new":     true,
	"append":  true,
	"copy":    true,
	"delete":  true,
	"close":   true,
	"panic":   true,
	"recover": true,
	"print":   true,
	"println": true,
	"complex": true,
	"real":    true,
	"imag":    true,
	"clear":   true,
	"min":     true,
	"max":     true,
}

// builtinConsts contains all Go builtin constants
var builtinConsts = map[string]bool{
	"true":  true,
	"false": true,
	"nil":   true,
	"iota":  true,
}

// keywords contains all Go keywords
var keywords = map[string]bool{
	"break":       true,
	"case":        true,
	"chan":        true,
	"const":       true,
	"continue":    true,
	"default":     true,
	"defer":       true,
	"else":        true,
	"fallthrough": true,
	"for":         true,
	"func":        true,
	"go":          true,
	"goto":        true,
	"if":          true,
	"import":      true,
	"interface":   true,
	"map":         true,
	"package":     true,
	"range":       true,
	"return":      true,
	"select":      true,
	"struct":      true,
	"switch":      true,
	"type":        true,
	"var":         true,
}

// IsBuiltinType returns true if the given name is a Go builtin type.
func IsBuiltinType(name string) bool {
	return builtinTypes[name]
}

// IsBuiltinIdent returns true if the given name is a Go builtin identifier
// (type, function, or constant).
func IsBuiltinIdent(name string) bool {
	return builtinTypes[name] || builtinFuncs[name] || builtinConsts[name]
}

// IsKeyword returns true if the given name is a Go keyword.
func IsKeyword(name string) bool {
	return keywords[name]
}
