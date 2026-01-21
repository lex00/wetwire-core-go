package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// FormatResult formats a Result based on the requested output format.
// Supported formats: text, json, yaml, raw.
func FormatResult(result *Result, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		return formatJSON(result)
	case "yaml", "yml":
		return formatYAML(result)
	case "text", "":
		return formatText(result), nil
	case "raw":
		return formatRaw(result)
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: text, json, yaml, raw)", format)
	}
}

// formatJSON formats the Result as JSON.
func formatJSON(result *Result) (string, error) {
	bytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(bytes) + "\n", nil
}

// formatYAML formats the Result as YAML.
func formatYAML(result *Result) (string, error) {
	bytes, err := yaml.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return string(bytes), nil
}

// formatText formats the Result as human-readable text.
func formatText(result *Result) string {
	var sb strings.Builder

	// Status line
	if result.Success {
		sb.WriteString("✓ Success")
	} else {
		sb.WriteString("✗ Failed")
	}

	// Message
	if result.Message != "" {
		sb.WriteString(": ")
		sb.WriteString(result.Message)
	}
	sb.WriteString("\n")

	// Errors
	if len(result.Errors) > 0 {
		sb.WriteString("\nErrors:\n")
		for i, err := range result.Errors {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.String()))
		}
	}

	// Data (if present, show as indented JSON)
	if result.Data != nil {
		sb.WriteString("\nData:\n")
		dataBytes, err := json.MarshalIndent(result.Data, "  ", "  ")
		if err == nil {
			sb.WriteString("  ")
			sb.WriteString(string(dataBytes))
			sb.WriteString("\n")
		} else {
			sb.WriteString(fmt.Sprintf("  %v\n", result.Data))
		}
	}

	return sb.String()
}

// formatRaw outputs just the Data field without any wrapper.
// Useful for piping build output directly to files.
func formatRaw(result *Result) (string, error) {
	if result.Data == nil {
		return "", nil
	}
	switch v := result.Data.(type) {
	case string:
		return v, nil
	default:
		bytes, err := json.Marshal(result.Data)
		if err != nil {
			return "", fmt.Errorf("failed to marshal raw data: %w", err)
		}
		return string(bytes), nil
	}
}

// FormatDiffResult formats a DiffResult based on the requested output format.
// Supported formats: json (text is handled separately in outputDiffResult).
func FormatDiffResult(result *DiffResult, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		bytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal JSON: %w", err)
		}
		return string(bytes), nil
	default:
		return "", fmt.Errorf("unsupported format for diff: %s", format)
	}
}
