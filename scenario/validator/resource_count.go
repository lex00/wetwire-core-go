package validator

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-core-go/scenario"
)

// ValidateResourceCounts validates resource counts against validation rules.
func (v *Validator) ValidateResourceCounts() (map[string]ResourceCountResult, error) {
	results := make(map[string]ResourceCountResult)

	if v.ScenarioConfig.Validation == nil {
		return results, nil
	}

	for domainName, rules := range v.ScenarioConfig.Validation {
		result := ResourceCountResult{
			Domain: domainName,
			Passed: true,
			Files:  []string{},
		}

		// Determine which constraint to use based on domain type
		var constraint *scenario.CountConstraint
		var resourceType string
		var patterns []string

		if rules.Stacks != nil {
			constraint = rules.Stacks
			resourceType = "stacks"
			patterns = []string{"*.yaml", "*.yml", "*.json"}
		} else if rules.Pipelines != nil {
			constraint = rules.Pipelines
			resourceType = "pipelines"
			patterns = []string{".gitlab-ci.yml", "*.yml"}
		} else if rules.Workflows != nil {
			constraint = rules.Workflows
			resourceType = "workflows"
			patterns = []string{"*.yml", "*.yaml"}
		} else if rules.Manifests != nil {
			constraint = rules.Manifests
			resourceType = "manifests"
			patterns = []string{"*.yaml", "*.yml"}
		} else if rules.Resources != nil {
			constraint = rules.Resources
			resourceType = "resources"
			patterns = getPatternsByDomain(domainName)
		}

		if constraint == nil {
			continue
		}

		result.Min = constraint.Min
		result.Max = constraint.Max
		result.ResourceType = resourceType

		// Count files matching patterns in the results directory
		files, err := v.countDomainFiles(domainName, patterns)
		if err != nil {
			result.Error = err.Error()
			result.Passed = false
			results[domainName] = result
			continue
		}

		result.Files = files
		result.Found = len(files)

		// Check against constraints
		if constraint.Min > 0 && result.Found < constraint.Min {
			result.Passed = false
			result.Error = "insufficient resources"
		}
		if constraint.Max > 0 && result.Found > constraint.Max {
			result.Passed = false
			result.Error = "too many resources"
		}

		results[domainName] = result
	}

	return results, nil
}

// countDomainFiles counts files in the results directory for a domain.
func (v *Validator) countDomainFiles(domainName string, patterns []string) ([]string, error) {
	var files []string

	// Try domain subdirectory first
	domainDir := filepath.Join(v.ResultsDir, domainName)
	if info, err := os.Stat(domainDir); err == nil && info.IsDir() {
		domainFiles, err := findMatchingFilesRecursive(domainDir, patterns)
		if err != nil {
			return nil, err
		}
		files = append(files, domainFiles...)
	}

	// Check for common subdirectory patterns (cfn-templates, .github/workflows, etc.)
	commonSubdirs := getCommonSubdirs(domainName)
	for _, subdir := range commonSubdirs {
		subdirPath := filepath.Join(v.ResultsDir, subdir)
		if info, err := os.Stat(subdirPath); err == nil && info.IsDir() {
			subdirFiles, err := findMatchingFilesRecursive(subdirPath, patterns)
			if err != nil {
				return nil, err
			}
			files = append(files, subdirFiles...)
		}
	}

	// Also check root results directory for files that might belong to this domain
	rootFiles, err := findMatchingFilesForDomain(v.ResultsDir, domainName, patterns)
	if err != nil {
		return nil, err
	}
	files = append(files, rootFiles...)

	// Deduplicate
	return uniqueStrings(files), nil
}

// getCommonSubdirs returns common subdirectory names for a domain.
func getCommonSubdirs(domainName string) []string {
	switch domainName {
	case "aws":
		return []string{"cfn-templates", "cloudformation", "templates"}
	case "github":
		return []string{".github/workflows", ".github"}
	case "gitlab":
		return []string{} // .gitlab-ci.yml is at root
	case "k8s", "kubernetes":
		return []string{"k8s", "kubernetes", "manifests"}
	case "honeycomb":
		return []string{"honeycomb", "observability"}
	default:
		return []string{}
	}
}

// findMatchingFiles finds all files matching the given patterns in a directory (non-recursive).
func findMatchingFiles(dir string, patterns []string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		for _, pattern := range patterns {
			matched, err := filepath.Match(pattern, name)
			if err != nil {
				continue
			}
			if matched {
				files = append(files, filepath.Join(dir, name))
				break
			}
		}
	}

	return files, nil
}

// findMatchingFilesRecursive finds all files matching the given patterns recursively.
func findMatchingFilesRecursive(dir string, patterns []string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}

		name := info.Name()
		for _, pattern := range patterns {
			matched, err := filepath.Match(pattern, name)
			if err != nil {
				continue
			}
			if matched {
				files = append(files, path)
				break
			}
		}
		return nil
	})

	return files, err
}

// findMatchingFilesForDomain finds files in the root that likely belong to a domain.
func findMatchingFilesForDomain(dir, domainName string, patterns []string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip non-matching patterns
		patternMatch := false
		for _, pattern := range patterns {
			matched, err := filepath.Match(pattern, name)
			if err != nil {
				continue
			}
			if matched {
				patternMatch = true
				break
			}
		}
		if !patternMatch {
			continue
		}

		// Check if file belongs to this domain based on naming conventions
		lowerName := strings.ToLower(name)
		if belongsToDomain(lowerName, domainName) {
			files = append(files, filepath.Join(dir, name))
		}
	}

	return files, nil
}

// belongsToDomain checks if a filename likely belongs to a domain.
func belongsToDomain(filename, domainName string) bool {
	switch domainName {
	case "aws":
		return strings.Contains(filename, "cfn") ||
			strings.Contains(filename, "cloudformation") ||
			strings.Contains(filename, "s3") ||
			strings.Contains(filename, "cloudfront") ||
			strings.Contains(filename, "iam")
	case "k8s", "kubernetes":
		return strings.Contains(filename, "namespace") ||
			strings.Contains(filename, "deployment") ||
			strings.Contains(filename, "service") ||
			strings.Contains(filename, "configmap") ||
			strings.HasPrefix(filename, "0") // numbered k8s files
	case "gitlab":
		return strings.Contains(filename, "gitlab-ci") ||
			filename == ".gitlab-ci.yml"
	case "github":
		return strings.Contains(filename, "workflow") ||
			strings.Contains(filename, "build") ||
			strings.Contains(filename, "deploy")
	case "honeycomb":
		return strings.Contains(filename, "query") ||
			strings.Contains(filename, "slo") ||
			strings.Contains(filename, "trigger") ||
			strings.Contains(filename, "board") ||
			strings.HasSuffix(filename, ".json")
	default:
		return false
	}
}

// getPatternsByDomain returns file patterns for a domain.
func getPatternsByDomain(domainName string) []string {
	switch domainName {
	case "aws":
		return []string{"*.yaml", "*.yml", "*.json"}
	case "k8s", "kubernetes":
		return []string{"*.yaml", "*.yml"}
	case "gitlab":
		return []string{".gitlab-ci.yml", "*.yml"}
	case "github":
		return []string{"*.yml", "*.yaml"}
	case "honeycomb":
		return []string{"*.json"}
	default:
		return []string{"*.yaml", "*.yml", "*.json"}
	}
}

// uniqueStrings removes duplicates from a string slice.
func uniqueStrings(strs []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
