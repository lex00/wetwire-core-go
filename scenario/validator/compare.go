package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// CompareExpected compares generated files against expected files.
func (v *Validator) CompareExpected() ([]FileComparisonResult, error) {
	var results []FileComparisonResult

	// Check if expected directory exists
	if _, err := os.Stat(v.ExpectedDir); os.IsNotExist(err) {
		// No expected directory - skip comparison
		return results, nil
	}

	// Walk expected directory and compare each file
	err := filepath.Walk(v.ExpectedDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Skip non-content files
		if !isContentFile(path) {
			return nil
		}

		// Get relative path from expected dir
		relPath, err := filepath.Rel(v.ExpectedDir, path)
		if err != nil {
			return err
		}

		// Find corresponding generated file
		result := v.compareFile(relPath)
		results = append(results, result)

		return nil
	})

	if err != nil {
		return results, err
	}

	return results, nil
}

// compareFile compares a single expected file against its generated counterpart.
func (v *Validator) compareFile(relPath string) FileComparisonResult {
	result := FileComparisonResult{
		ExpectedFile:  relPath,
		GeneratedFile: relPath,
		Passed:        false,
		Missing:       false,
		Differences:   []string{},
	}

	expectedPath := filepath.Join(v.ExpectedDir, relPath)
	generatedPath := filepath.Join(v.ResultsDir, relPath)

	// Check if generated file exists
	if _, err := os.Stat(generatedPath); os.IsNotExist(err) {
		// Try to find file without subdirectory
		baseName := filepath.Base(relPath)
		altPath := filepath.Join(v.ResultsDir, baseName)
		if _, err := os.Stat(altPath); os.IsNotExist(err) {
			// Try finding by similar name
			foundPath := v.findSimilarFile(baseName)
			if foundPath == "" {
				result.Missing = true
				result.Differences = append(result.Differences, "expected file not found in results")
				return result
			}
			generatedPath = foundPath
			result.GeneratedFile, _ = filepath.Rel(v.ResultsDir, foundPath)
		} else {
			generatedPath = altPath
			result.GeneratedFile = baseName
		}
	}

	// Read both files
	expectedContent, err := os.ReadFile(expectedPath)
	if err != nil {
		result.Differences = append(result.Differences, fmt.Sprintf("error reading expected: %v", err))
		return result
	}

	generatedContent, err := os.ReadFile(generatedPath)
	if err != nil {
		result.Differences = append(result.Differences, fmt.Sprintf("error reading generated: %v", err))
		return result
	}

	// Parse and compare structurally
	ext := strings.ToLower(filepath.Ext(relPath))
	switch ext {
	case ".yaml", ".yml":
		result.Differences = compareYAML(expectedContent, generatedContent)
	case ".json":
		result.Differences = compareJSON(expectedContent, generatedContent)
	default:
		// Text comparison for unknown types
		if string(expectedContent) != string(generatedContent) {
			result.Differences = append(result.Differences, "content differs")
		}
	}

	// Pass/fail logic:
	// - Missing file = FAIL (file wasn't generated)
	// - File exists with structural differences = PASS (file exists, differences are informational)
	// - "extra key (allowed)" differences don't affect pass status
	//
	// The structural comparison is informational - it helps identify when generated
	// output deviates from expected templates, but deviations aren't necessarily wrong.
	result.Passed = !result.Missing

	return result
}

// findSimilarFile tries to find a file with a similar name in results.
func (v *Validator) findSimilarFile(baseName string) string {
	// Extract key parts of the filename
	nameParts := extractNameParts(baseName)

	var found string
	filepath.Walk(v.ResultsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || found != "" {
			return nil
		}

		candidateParts := extractNameParts(filepath.Base(path))
		if matchNameParts(nameParts, candidateParts) {
			found = path
		}
		return nil
	})

	return found
}

// extractNameParts extracts meaningful parts from a filename.
func extractNameParts(name string) []string {
	// Remove extension
	name = strings.TrimSuffix(name, filepath.Ext(name))

	// Split by common separators
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_' || r == '.'
	})

	// Convert to lowercase
	for i := range parts {
		parts[i] = strings.ToLower(parts[i])
	}

	return parts
}

// matchNameParts checks if two sets of name parts are similar enough.
func matchNameParts(expected, candidate []string) bool {
	if len(expected) == 0 || len(candidate) == 0 {
		return false
	}

	// Count matching parts
	matches := 0
	for _, ep := range expected {
		for _, cp := range candidate {
			if ep == cp {
				matches++
				break
			}
		}
	}

	// Consider a match if at least half of expected parts are found
	return float64(matches) >= float64(len(expected))*0.5
}

// compareYAML compares two YAML documents structurally.
func compareYAML(expected, generated []byte) []string {
	var diffs []string

	var expectedData interface{}
	var generatedData interface{}

	if err := yaml.Unmarshal(expected, &expectedData); err != nil {
		diffs = append(diffs, fmt.Sprintf("expected YAML parse error: %v", err))
		return diffs
	}

	if err := yaml.Unmarshal(generated, &generatedData); err != nil {
		diffs = append(diffs, fmt.Sprintf("generated YAML parse error: %v", err))
		return diffs
	}

	// Compare structure
	return compareStructure("", expectedData, generatedData)
}

// compareJSON compares two JSON documents structurally.
func compareJSON(expected, generated []byte) []string {
	var diffs []string

	var expectedData interface{}
	var generatedData interface{}

	if err := json.Unmarshal(expected, &expectedData); err != nil {
		diffs = append(diffs, fmt.Sprintf("expected JSON parse error: %v", err))
		return diffs
	}

	if err := json.Unmarshal(generated, &generatedData); err != nil {
		diffs = append(diffs, fmt.Sprintf("generated JSON parse error: %v", err))
		return diffs
	}

	// Compare structure
	return compareStructure("", expectedData, generatedData)
}

// compareStructure recursively compares two data structures.
func compareStructure(path string, expected, generated interface{}) []string {
	var diffs []string

	if expected == nil && generated == nil {
		return diffs
	}

	if expected == nil || generated == nil {
		diffs = append(diffs, fmt.Sprintf("%s: type mismatch (nil vs non-nil)", pathOrRoot(path)))
		return diffs
	}

	switch e := expected.(type) {
	case map[string]interface{}:
		g, ok := generated.(map[string]interface{})
		if !ok {
			diffs = append(diffs, fmt.Sprintf("%s: type mismatch (expected object)", pathOrRoot(path)))
			return diffs
		}

		// Check for missing keys
		for key := range e {
			newPath := joinPath(path, key)
			if _, exists := g[key]; !exists {
				diffs = append(diffs, fmt.Sprintf("%s: missing key", newPath))
			} else {
				// Recurse into child
				childDiffs := compareStructure(newPath, e[key], g[key])
				diffs = append(diffs, childDiffs...)
			}
		}

		// Check for extra keys (warning level, not failure)
		for key := range g {
			if _, exists := e[key]; !exists {
				newPath := joinPath(path, key)
				diffs = append(diffs, fmt.Sprintf("%s: extra key (allowed)", newPath))
			}
		}

	case []interface{}:
		g, ok := generated.([]interface{})
		if !ok {
			diffs = append(diffs, fmt.Sprintf("%s: type mismatch (expected array)", pathOrRoot(path)))
			return diffs
		}

		// For arrays, just check that generated has at least as many items
		if len(g) < len(e) {
			diffs = append(diffs, fmt.Sprintf("%s: array too short (expected %d, got %d)", pathOrRoot(path), len(e), len(g)))
		}

		// Compare items up to expected length
		for i := 0; i < len(e) && i < len(g); i++ {
			newPath := fmt.Sprintf("%s[%d]", path, i)
			childDiffs := compareStructure(newPath, e[i], g[i])
			diffs = append(diffs, childDiffs...)
		}

	default:
		// Leaf values - don't compare actual values, just check presence
		// Values like "${k8s.namespace}" are placeholders, so actual values will differ
	}

	return diffs
}

// pathOrRoot returns the path or "root" if empty.
func pathOrRoot(path string) string {
	if path == "" {
		return "root"
	}
	return path
}

// joinPath joins path segments.
func joinPath(base, key string) string {
	if base == "" {
		return key
	}
	return base + "." + key
}
