package domain

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigFilename is the standard name for wetwire configuration files
const ConfigFilename = "wetwire.yaml"

// Config represents the wetwire domain configuration
type Config struct {
	Domain  string                 `yaml:"domain"`
	Version string                 `yaml:"version,omitempty"`
	Build   *BuildConfig           `yaml:"build,omitempty"`
	Lint    *LintConfig            `yaml:"lint,omitempty"`
	Extra   map[string]interface{} `yaml:",inline"`
}

// BuildConfig represents build-related configuration
type BuildConfig struct {
	Output string `yaml:"output,omitempty"`
	Format string `yaml:"format,omitempty"`
}

// LintConfig represents linting-related configuration
type LintConfig struct {
	Rules   map[string]bool `yaml:"rules,omitempty"`
	Exclude []string        `yaml:"exclude,omitempty"`
}

// LoadConfig loads from current directory, walking up to find wetwire.yaml
func LoadConfig() (*Config, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return LoadConfigFrom(cwd)
}

// LoadConfigFrom loads starting from specified directory, walking up the tree
func LoadConfigFrom(startDir string) (*Config, string, error) {
	// Normalize the start directory
	absDir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	currentDir := absDir
	for {
		configPath := filepath.Join(currentDir, ConfigFilename)

		// Check if config file exists
		if _, err := os.Stat(configPath); err == nil {
			// File exists, load it
			config, err := LoadConfigFile(configPath)
			if err != nil {
				return nil, "", err
			}
			return config, configPath, nil
		}

		// Move to parent directory
		parentDir := filepath.Dir(currentDir)

		// Check if we've reached the root
		if parentDir == currentDir {
			// Not found, return empty config (not an error)
			return &Config{}, "", nil
		}

		currentDir = parentDir
	}
}

// LoadConfigFile loads from specific path
func LoadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	return &config, nil
}

// SaveConfig saves to wetwire.yaml in current directory
func SaveConfig(config *Config) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(cwd, ConfigFilename)
	return SaveConfigTo(config, configPath)
}

// SaveConfigTo saves to specific path
func SaveConfigTo(config *Config, path string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
