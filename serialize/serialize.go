// Package serialize provides utilities for converting Go structs to maps
// and serializing to JSON/YAML with configurable naming conventions.
package serialize

import (
	"encoding/json"
	"reflect"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

// Option configures serialization behavior.
type Option func(*options)

type options struct {
	namingConvention func(string) string
	omitEmpty        bool
}

// SnakeCase converts field names to snake_case (e.g., FirstName -> first_name).
var SnakeCase Option = func(o *options) {
	o.namingConvention = toSnakeCase
}

// CamelCase converts field names to camelCase (e.g., FirstName -> firstName).
var CamelCase Option = func(o *options) {
	o.namingConvention = toCamelCase
}

// PascalCase keeps field names as PascalCase (e.g., FirstName -> FirstName).
var PascalCase Option = func(o *options) {
	o.namingConvention = toPascalCase
}

// OmitEmpty omits fields with zero values from output.
var OmitEmpty Option = func(o *options) {
	o.omitEmpty = true
}

// ToMap converts a struct to a map with the given options.
func ToMap(v any, opts ...Option) map[string]any {
	o := &options{
		namingConvention: toPascalCase, // default
		omitEmpty:        false,
	}
	for _, opt := range opts {
		opt(o)
	}

	return structToMap(reflect.ValueOf(v), o)
}

// ToYAML serializes a struct to YAML bytes with the given options.
func ToYAML(v any, opts ...Option) ([]byte, error) {
	m := ToMap(v, opts...)
	return yaml.Marshal(m)
}

// ToJSON serializes a struct to JSON bytes with the given options.
func ToJSON(v any, opts ...Option) ([]byte, error) {
	m := ToMap(v, opts...)
	return json.Marshal(m)
}

func structToMap(v reflect.Value, o *options) map[string]any {
	// Handle pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	result := make(map[string]any)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Check for zero value if omitEmpty is set
		if o.omitEmpty && isZeroValue(fieldValue) {
			continue
		}

		key := o.namingConvention(field.Name)
		value := convertValue(fieldValue, o)
		result[key] = value
	}

	return result
}

func convertValue(v reflect.Value, o *options) any {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return nil
		}
		return convertValue(v.Elem(), o)
	case reflect.Struct:
		return structToMap(v, o)
	case reflect.Slice, reflect.Array:
		if v.IsNil() {
			return nil
		}
		// Return slice values directly for simple types
		if v.Type().Elem().Kind() == reflect.String {
			result := make([]string, v.Len())
			for i := 0; i < v.Len(); i++ {
				result[i] = v.Index(i).String()
			}
			return result
		}
		result := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = convertValue(v.Index(i), o)
		}
		return result
	case reflect.Map:
		if v.IsNil() {
			return nil
		}
		result := make(map[string]any)
		for _, key := range v.MapKeys() {
			result[o.namingConvention(key.String())] = convertValue(v.MapIndex(key), o)
		}
		return result
	case reflect.Interface:
		if v.IsNil() {
			return nil
		}
		return convertValue(v.Elem(), o)
	default:
		return v.Interface()
	}
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.IsNil() || v.Len() == 0
	case reflect.Struct:
		return v.IsZero()
	default:
		return false
	}
}

// toSnakeCase converts PascalCase to snake_case.
// Handles consecutive capitals (e.g., "ID" -> "id", "APIKey" -> "api_key").
func toSnakeCase(s string) string {
	if len(s) == 0 {
		return s
	}

	runes := []rune(s)
	var result strings.Builder

	for i, r := range runes {
		if unicode.IsUpper(r) {
			// Add underscore before uppercase if:
			// - Not at start AND
			// - Either previous char is lowercase OR next char is lowercase
			if i > 0 {
				prevLower := unicode.IsLower(runes[i-1])
				nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
				if prevLower || nextLower {
					result.WriteRune('_')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// toCamelCase converts PascalCase to camelCase.
func toCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// toPascalCase returns the string as-is (already PascalCase).
func toPascalCase(s string) string {
	return s
}
