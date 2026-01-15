package scenario

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultFilename is the default name for scenario configuration files.
const DefaultFilename = "scenario.yaml"

// Load loads a scenario configuration from the specified path.
// If path is a directory, it looks for scenario.yaml in that directory.
// If path is empty, it looks in the current directory.
func Load(path string) (*ScenarioConfig, error) {
	if path == "" {
		path = "."
	}

	// Check if path is a directory
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		path = filepath.Join(path, DefaultFilename)
	}

	return LoadFile(path)
}

// LoadFile loads a scenario configuration from a specific file.
func LoadFile(path string) (*ScenarioConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file: %w", err)
	}

	return Parse(data)
}

// Parse parses scenario configuration from YAML bytes.
func Parse(data []byte) (*ScenarioConfig, error) {
	var config ScenarioConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse scenario YAML: %w", err)
	}

	return &config, nil
}

// GetDomainOrder returns domains in dependency order (dependencies first).
// Returns an error if there are circular dependencies.
func GetDomainOrder(config *ScenarioConfig) ([]string, error) {
	// Build dependency graph
	deps := make(map[string][]string)
	all := make(map[string]bool)

	for _, d := range config.Domains {
		all[d.Name] = true
		deps[d.Name] = d.DependsOn
	}

	// Validate all dependencies exist
	for name, dependencies := range deps {
		for _, dep := range dependencies {
			if !all[dep] {
				return nil, fmt.Errorf("domain %q depends on unknown domain %q", name, dep)
			}
		}
	}

	// Topological sort using Kahn's algorithm
	// Count incoming edges (how many dependencies each domain has)
	inDegree := make(map[string]int)
	for name := range all {
		inDegree[name] = 0
	}
	for _, d := range config.Domains {
		inDegree[d.Name] = len(d.DependsOn)
	}

	// Find domains with no dependencies
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	var result []string
	for len(queue) > 0 {
		// Take first item
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Find domains that depend on current
		for _, d := range config.Domains {
			for _, dep := range d.DependsOn {
				if dep == current {
					inDegree[d.Name]--
					if inDegree[d.Name] == 0 {
						queue = append(queue, d.Name)
					}
				}
			}
		}
	}

	// Check for cycles
	if len(result) != len(all) {
		return nil, fmt.Errorf("circular dependency detected in domains")
	}

	return result, nil
}
