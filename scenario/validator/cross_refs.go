package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidateCrossRefs validates cross-domain references are present in generated files.
func (v *Validator) ValidateCrossRefs() ([]CrossRefResult, error) {
	var results []CrossRefResult

	if len(v.ScenarioConfig.CrossDomain) == 0 {
		return results, nil
	}

	for _, crossDomain := range v.ScenarioConfig.CrossDomain {
		result := CrossRefResult{
			From:         crossDomain.From,
			To:           crossDomain.To,
			Passed:       true,
			RequiredRefs: crossDomain.Validation.RequiredRefs,
			FoundRefs:    []string{},
			MissingRefs:  []string{},
			Locations:    make(map[string][]string),
		}

		if len(crossDomain.Validation.RequiredRefs) == 0 {
			results = append(results, result)
			continue
		}

		// Get files from the target domain
		targetFiles, err := v.getTargetDomainFiles(crossDomain.To)
		if err != nil {
			result.Passed = false
			result.MissingRefs = crossDomain.Validation.RequiredRefs
			results = append(results, result)
			continue
		}

		// Check each required ref
		for _, requiredRef := range crossDomain.Validation.RequiredRefs {
			found, locations := v.findRefInFiles(requiredRef, targetFiles)
			if found {
				result.FoundRefs = append(result.FoundRefs, requiredRef)
				result.Locations[requiredRef] = locations
			} else {
				result.MissingRefs = append(result.MissingRefs, requiredRef)
				result.Passed = false
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// getTargetDomainFiles returns all files belonging to a domain.
func (v *Validator) getTargetDomainFiles(domainName string) ([]string, error) {
	var files []string

	// Check domain subdirectory
	domainDir := filepath.Join(v.ResultsDir, domainName)
	if info, err := os.Stat(domainDir); err == nil && info.IsDir() {
		err := filepath.Walk(domainDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && isContentFile(path) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// Check root directory for domain-related files
	rootFiles, err := v.getRootDomainFiles(domainName)
	if err != nil {
		return nil, err
	}
	files = append(files, rootFiles...)

	return uniqueStrings(files), nil
}

// getRootDomainFiles finds files in the root results dir that belong to a domain.
func (v *Validator) getRootDomainFiles(domainName string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(v.ResultsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !isContentFile(name) {
			continue
		}

		// Check if file belongs to this domain
		if belongsToDomain(strings.ToLower(name), domainName) {
			files = append(files, filepath.Join(v.ResultsDir, name))
		}
	}

	return files, nil
}

// isContentFile returns true if the file is a content file we should search.
func isContentFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml", ".json":
		return true
	default:
		return false
	}
}

// findRefInFiles searches for a reference pattern in the given files.
func (v *Validator) findRefInFiles(ref string, files []string) (bool, []string) {
	var locations []string

	// Create search patterns for the reference
	patterns := createSearchPatterns(ref)

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		contentStr := string(content)

		for _, pattern := range patterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}

			if re.MatchString(contentStr) {
				relPath, _ := filepath.Rel(v.ResultsDir, file)
				if relPath == "" {
					relPath = filepath.Base(file)
				}
				locations = append(locations, relPath)
				break
			}
		}
	}

	return len(locations) > 0, uniqueStrings(locations)
}

// createSearchPatterns creates regex patterns to find a reference.
// References like "${k8s.service_name}" should match various representations.
func createSearchPatterns(ref string) []string {
	var patterns []string

	// Extract the key from ${domain.key} format
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	matches := re.FindStringSubmatch(ref)

	if len(matches) < 2 {
		// Not a ${} pattern, search literally
		patterns = append(patterns, regexp.QuoteMeta(ref))
		return patterns
	}

	refPath := matches[1] // e.g., "k8s.service_name" or "aws.s3.outputs.bucket_name"

	// Split by dots
	parts := strings.Split(refPath, ".")
	if len(parts) < 2 {
		patterns = append(patterns, regexp.QuoteMeta(ref))
		return patterns
	}

	// Get the domain and the key part
	// For "k8s.service_name" -> domain="k8s", key="service_name"
	// For "aws.s3.outputs.bucket_name" -> domain="aws", key="bucket_name"
	key := parts[len(parts)-1]

	// Create patterns that would match this reference:

	// 1. The literal ${} reference
	patterns = append(patterns, regexp.QuoteMeta(ref))

	// 2. Common field names that would contain this value
	// For service_name, look for fields like "service", "serviceName", "service-name"
	fieldVariants := getFieldVariants(key)
	for _, variant := range fieldVariants {
		// Match YAML: field: value
		patterns = append(patterns, fmt.Sprintf(`(?i)%s:\s*\S+`, variant))
		// Match JSON: "field": "value"
		patterns = append(patterns, fmt.Sprintf(`(?i)"%s"\s*:\s*"[^"]+"`, variant))
		// Match shell variable assignment: VAR=value or VAR=$(...)
		patterns = append(patterns, fmt.Sprintf(`(?i)%s=\S+`, variant))
		// Match shell variable reference: $VAR or ${VAR}
		patterns = append(patterns, fmt.Sprintf(`(?i)\$\{?%s\}?`, variant))
		// Match CloudFormation OutputKey reference
		patterns = append(patterns, fmt.Sprintf(`(?i)OutputKey==.%s.`, variant))
	}

	// 3. For k8s references, look for specific k8s field patterns
	if strings.HasPrefix(refPath, "k8s.") {
		switch key {
		case "service_name":
			patterns = append(patterns, `k8s\.service\.name`)
			patterns = append(patterns, `k8s\.service_name`)
			patterns = append(patterns, `service\.name`)
			// Match as JSON value: "k8s.service_name" or "k8s.service.name"
			patterns = append(patterns, `"k8s\.service[_.]name"`)
		case "namespace":
			patterns = append(patterns, `k8s\.namespace\.name`)
			patterns = append(patterns, `k8s\.namespace`)
			patterns = append(patterns, `namespace\.name`)
			// Match as JSON value: "k8s.namespace" or "k8s.namespace.name"
			patterns = append(patterns, `"k8s\.namespace[._]?name?"`)
		}
	}

	// 4. For honeycomb references
	if strings.HasPrefix(refPath, "honeycomb.") {
		switch key {
		case "dataset_name":
			patterns = append(patterns, `dataset:\s*\S+`)
			patterns = append(patterns, `"dataset":\s*"[^"]+"`)
		}
	}

	return patterns
}

// getFieldVariants returns common variants of a field name.
func getFieldVariants(key string) []string {
	variants := []string{key}

	// Convert snake_case to other formats
	if strings.Contains(key, "_") {
		// camelCase
		camel := snakeToCamel(key)
		variants = append(variants, camel)

		// PascalCase
		pascal := snakeToPascal(key)
		variants = append(variants, pascal)

		// kebab-case
		kebab := strings.ReplaceAll(key, "_", "-")
		variants = append(variants, kebab)

		// UPPER_CASE
		upper := strings.ToUpper(key)
		variants = append(variants, upper)

		// Just the last word (e.g., "service_name" -> "service", "name")
		parts := strings.Split(key, "_")
		variants = append(variants, parts...)
	}

	return variants
}

// snakeToCamel converts snake_case to camelCase.
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// snakeToPascal converts snake_case to PascalCase.
func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	for i := 0; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

