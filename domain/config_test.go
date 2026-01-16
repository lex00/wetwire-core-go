package domain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFile(t *testing.T) {
	t.Run("loads valid YAML", func(t *testing.T) {
		// Create a temporary file with valid YAML
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "wetwire.yaml")

		validYAML := `domain: test-domain
version: "1.0.0"
build:
  output: "./output"
  format: "json"
lint:
  rules:
    rule1: true
    rule2: false
  exclude:
    - "*.test.go"
`
		if err := os.WriteFile(configPath, []byte(validYAML), 0644); err != nil {
			t.Fatalf("Failed to create test config: %v", err)
		}

		config, err := LoadConfigFile(configPath)
		if err != nil {
			t.Fatalf("LoadConfigFile() error = %v", err)
		}

		if config.Domain != "test-domain" {
			t.Errorf("Domain = %q, want %q", config.Domain, "test-domain")
		}
		if config.Version != "1.0.0" {
			t.Errorf("Version = %q, want %q", config.Version, "1.0.0")
		}
		if config.Build == nil {
			t.Fatal("Build config is nil")
		}
		if config.Build.Output != "./output" {
			t.Errorf("Build.Output = %q, want %q", config.Build.Output, "./output")
		}
		if config.Build.Format != "json" {
			t.Errorf("Build.Format = %q, want %q", config.Build.Format, "json")
		}
		if config.Lint == nil {
			t.Fatal("Lint config is nil")
		}
		if !config.Lint.Rules["rule1"] {
			t.Error("Lint.Rules[rule1] should be true")
		}
		if config.Lint.Rules["rule2"] {
			t.Error("Lint.Rules[rule2] should be false")
		}
		if len(config.Lint.Exclude) != 1 || config.Lint.Exclude[0] != "*.test.go" {
			t.Errorf("Lint.Exclude = %v, want [*.test.go]", config.Lint.Exclude)
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := LoadConfigFile("/nonexistent/path/wetwire.yaml")
		if err == nil {
			t.Error("LoadConfigFile() expected error for non-existent file")
		}
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "wetwire.yaml")

		invalidYAML := `domain: test
  invalid: yaml: structure`
		if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
			t.Fatalf("Failed to create test config: %v", err)
		}

		_, err := LoadConfigFile(configPath)
		if err == nil {
			t.Error("LoadConfigFile() expected error for invalid YAML")
		}
	})
}

func TestLoadConfigFrom(t *testing.T) {
	t.Run("walks up directory tree", func(t *testing.T) {
		// Create a directory structure:
		// tmpDir/
		//   wetwire.yaml
		//   subdir1/
		//     subdir2/
		//       subdir3/
		tmpDir := t.TempDir()
		subdir1 := filepath.Join(tmpDir, "subdir1")
		subdir2 := filepath.Join(subdir1, "subdir2")
		subdir3 := filepath.Join(subdir2, "subdir3")

		if err := os.MkdirAll(subdir3, 0755); err != nil {
			t.Fatalf("Failed to create subdirectories: %v", err)
		}

		// Place config in root
		configPath := filepath.Join(tmpDir, ConfigFilename)
		configYAML := `domain: found-domain
version: "2.0.0"`
		if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		// Start search from deepest directory
		config, foundPath, err := LoadConfigFrom(subdir3)
		if err != nil {
			t.Fatalf("LoadConfigFrom() error = %v", err)
		}

		if config.Domain != "found-domain" {
			t.Errorf("Domain = %q, want %q", config.Domain, "found-domain")
		}
		if foundPath != configPath {
			t.Errorf("foundPath = %q, want %q", foundPath, configPath)
		}
	})

	t.Run("finds config in start directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ConfigFilename)
		configYAML := `domain: local-domain`
		if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		config, foundPath, err := LoadConfigFrom(tmpDir)
		if err != nil {
			t.Fatalf("LoadConfigFrom() error = %v", err)
		}

		if config.Domain != "local-domain" {
			t.Errorf("Domain = %q, want %q", config.Domain, "local-domain")
		}
		if foundPath != configPath {
			t.Errorf("foundPath = %q, want %q", foundPath, configPath)
		}
	})

	t.Run("returns empty config if not found (not error)", func(t *testing.T) {
		tmpDir := t.TempDir()
		subdir := filepath.Join(tmpDir, "subdir")
		if err := os.MkdirAll(subdir, 0755); err != nil {
			t.Fatalf("Failed to create subdir: %v", err)
		}

		config, foundPath, err := LoadConfigFrom(subdir)
		if err != nil {
			t.Errorf("LoadConfigFrom() error = %v, want nil", err)
		}
		if config == nil {
			t.Fatal("config is nil, want empty config")
		}
		if config.Domain != "" {
			t.Errorf("Domain = %q, want empty string", config.Domain)
		}
		if foundPath != "" {
			t.Errorf("foundPath = %q, want empty string", foundPath)
		}
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("loads from current directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		configYAML := `domain: current-dir-domain`
		if err := os.WriteFile(ConfigFilename, []byte(configYAML), 0644); err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		config, foundPath, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		if config.Domain != "current-dir-domain" {
			t.Errorf("Domain = %q, want %q", config.Domain, "current-dir-domain")
		}
		// Resolve symlinks for comparison (macOS /var -> /private/var)
		expectedPath, err := filepath.EvalSymlinks(filepath.Join(tmpDir, ConfigFilename))
		if err != nil {
			t.Fatalf("Failed to resolve expected path: %v", err)
		}
		if foundPath != expectedPath {
			t.Errorf("foundPath = %q, want %q", foundPath, expectedPath)
		}
	})

	t.Run("returns empty config if not found (not error)", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		config, foundPath, err := LoadConfig()
		if err != nil {
			t.Errorf("LoadConfig() error = %v, want nil", err)
		}
		if config == nil {
			t.Fatal("config is nil, want empty config")
		}
		if config.Domain != "" {
			t.Errorf("Domain = %q, want empty string", config.Domain)
		}
		if foundPath != "" {
			t.Errorf("foundPath = %q, want empty string", foundPath)
		}
	})
}

func TestSaveConfig(t *testing.T) {
	t.Run("writes valid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		config := &Config{
			Domain:  "save-test-domain",
			Version: "3.0.0",
			Build: &BuildConfig{
				Output: "./dist",
				Format: "pretty",
			},
			Lint: &LintConfig{
				Rules: map[string]bool{
					"strict": true,
					"warn":   false,
				},
				Exclude: []string{"vendor/*"},
			},
		}

		if err := SaveConfig(config); err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		// Verify the file was created
		configPath := filepath.Join(tmpDir, ConfigFilename)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatal("Config file was not created")
		}

		// Load it back and verify
		loaded, err := LoadConfigFile(configPath)
		if err != nil {
			t.Fatalf("Failed to load saved config: %v", err)
		}

		if loaded.Domain != config.Domain {
			t.Errorf("Domain = %q, want %q", loaded.Domain, config.Domain)
		}
		if loaded.Version != config.Version {
			t.Errorf("Version = %q, want %q", loaded.Version, config.Version)
		}
		if loaded.Build.Output != config.Build.Output {
			t.Errorf("Build.Output = %q, want %q", loaded.Build.Output, config.Build.Output)
		}
		if loaded.Lint.Rules["strict"] != config.Lint.Rules["strict"] {
			t.Error("Lint.Rules[strict] mismatch")
		}
	})
}

func TestSaveConfigTo(t *testing.T) {
	t.Run("writes to specific path", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "custom-config.yaml")

		config := &Config{
			Domain: "custom-domain",
		}

		if err := SaveConfigTo(config, configPath); err != nil {
			t.Fatalf("SaveConfigTo() error = %v", err)
		}

		// Verify the file was created
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatal("Config file was not created")
		}

		// Load it back and verify
		loaded, err := LoadConfigFile(configPath)
		if err != nil {
			t.Fatalf("Failed to load saved config: %v", err)
		}

		if loaded.Domain != config.Domain {
			t.Errorf("Domain = %q, want %q", loaded.Domain, config.Domain)
		}
	})

	t.Run("creates parent directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "nested", "dirs", "config.yaml")

		config := &Config{
			Domain: "nested-domain",
		}

		if err := SaveConfigTo(config, configPath); err != nil {
			t.Fatalf("SaveConfigTo() error = %v", err)
		}

		// Verify the file was created
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatal("Config file was not created in nested directory")
		}

		// Load it back and verify
		loaded, err := LoadConfigFile(configPath)
		if err != nil {
			t.Fatalf("Failed to load saved config: %v", err)
		}

		if loaded.Domain != config.Domain {
			t.Errorf("Domain = %q, want %q", loaded.Domain, config.Domain)
		}
	})
}
