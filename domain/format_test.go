package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestFormatResult(t *testing.T) {
	t.Run("format as JSON", func(t *testing.T) {
		result := NewResult("test message")
		output, err := FormatResult(result, "json")

		require.NoError(t, err)
		assert.Contains(t, output, "{")
		assert.Contains(t, output, "}")
		assert.Contains(t, output, "\"success\"")
		assert.Contains(t, output, "\"message\"")

		// Verify it's valid JSON
		var parsed map[string]interface{}
		err = json.Unmarshal([]byte(output), &parsed)
		require.NoError(t, err)
		assert.Equal(t, true, parsed["success"])
		assert.Equal(t, "test message", parsed["message"])
	})

	t.Run("format as YAML", func(t *testing.T) {
		result := NewResult("test message")
		output, err := FormatResult(result, "yaml")

		require.NoError(t, err)
		assert.Contains(t, output, "success:")
		assert.Contains(t, output, "message:")
		assert.NotContains(t, output, "{")

		// Verify it's valid YAML
		var parsed map[string]interface{}
		err = yaml.Unmarshal([]byte(output), &parsed)
		require.NoError(t, err)
		assert.Equal(t, true, parsed["success"])
		assert.Equal(t, "test message", parsed["message"])
	})

	t.Run("format as text", func(t *testing.T) {
		result := NewResult("test message")
		output, err := FormatResult(result, "text")

		require.NoError(t, err)
		assert.Contains(t, output, "Success")
		assert.Contains(t, output, "test message")
	})

	t.Run("unsupported format returns error", func(t *testing.T) {
		result := NewResult("test message")
		_, err := FormatResult(result, "xml")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})
}

func TestFormatText(t *testing.T) {
	t.Run("simple success result", func(t *testing.T) {
		result := NewResult("operation successful")
		output := formatText(result)

		assert.Contains(t, output, "✓")
		assert.Contains(t, output, "Success")
		assert.Contains(t, output, "operation successful")
	})

	t.Run("simple error result", func(t *testing.T) {
		domainErr := Error{Message: "something went wrong"}
		result := NewErrorResult("operation failed", domainErr)
		output := formatText(result)

		assert.Contains(t, output, "✗")
		assert.Contains(t, output, "Failed")
		assert.Contains(t, output, "operation failed")
	})

	t.Run("result with data", func(t *testing.T) {
		data := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}
		result := NewResultWithData("data result", data)
		output := formatText(result)

		assert.Contains(t, output, "Success")
		assert.Contains(t, output, "data result")
		assert.Contains(t, output, "Data:")
		assert.Contains(t, output, "key")
	})

	t.Run("result with errors shows error details", func(t *testing.T) {
		errors := []Error{
			{
				Path:     "file1.go",
				Line:     10,
				Column:   5,
				Severity: "error",
				Message:  "syntax error",
				Code:     "E001",
			},
			{
				Path:     "file2.go",
				Line:     20,
				Severity: "warning",
				Message:  "unused variable",
			},
		}
		result := NewErrorResultMultiple("validation failed", errors)
		output := formatText(result)

		assert.Contains(t, output, "Failed")
		assert.Contains(t, output, "validation failed")
		assert.Contains(t, output, "Errors:")
		assert.Contains(t, output, "file1.go")
		assert.Contains(t, output, "file2.go")
		assert.Contains(t, output, "syntax error")
		assert.Contains(t, output, "unused variable")
		assert.Contains(t, output, "10:5")
		assert.Contains(t, output, "E001")
	})
}
