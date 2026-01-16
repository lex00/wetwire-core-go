package scenario_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lex00/wetwire-core-go/mcp"
	"github.com/lex00/wetwire-core-go/scenario"
	skillscenario "github.com/lex00/wetwire-core-go/skills/scenario"
)

// getExamplesDir returns the path to the examples directory.
func getExamplesDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "examples")
}

// TestIntegration_FullScenarioWorkflow tests the complete scenario workflow:
// 1. Load scenario.yaml from examples
// 2. Validate the scenario
// 3. Generate instructions via skill
// 4. Validate cross-domain references
func TestIntegration_FullScenarioWorkflow(t *testing.T) {
	examplesDir := getExamplesDir()
	scenarioDir := filepath.Join(examplesDir, "cross_domain_ab")

	// Skip if examples directory doesn't exist (e.g., in CI without examples)
	if _, err := os.Stat(scenarioDir); os.IsNotExist(err) {
		t.Skip("examples/cross_domain_ab not found, skipping integration test")
	}

	t.Run("load and validate scenario", func(t *testing.T) {
		// Load scenario
		config, err := scenario.Load(scenarioDir)
		require.NoError(t, err, "should load scenario.yaml")

		// Verify basic structure
		assert.Equal(t, "cross_domain_ab_deploy", config.Name)
		assert.Contains(t, config.Description, "bucket")

		// Verify domains
		require.Len(t, config.Domains, 2)
		assert.Equal(t, "domain-a", config.Domains[0].Name)
		assert.Equal(t, "domain-b", config.Domains[1].Name)

		// Verify cross-domain relationships
		require.Len(t, config.CrossDomain, 1)
		assert.Equal(t, "domain-a", config.CrossDomain[0].From)
		assert.Equal(t, "domain-b", config.CrossDomain[0].To)

		// Validate scenario
		result := scenario.Validate(config)
		assert.True(t, result.IsValid(), "scenario should be valid: %s", result.Error())
	})

	t.Run("dependency ordering", func(t *testing.T) {
		config, err := scenario.Load(scenarioDir)
		require.NoError(t, err)

		// Get domain order
		order, err := scenario.GetDomainOrder(config)
		require.NoError(t, err)

		// domain-a should come before domain-b (domain-b depends on domain-a)
		require.Len(t, order, 2)
		assert.Equal(t, "domain-a", order[0], "domain-a should be first (no dependencies)")
		assert.Equal(t, "domain-b", order[1], "domain-b should be second (depends on domain-a)")
	})

	t.Run("skill generates instructions", func(t *testing.T) {
		skill := skillscenario.New(nil, nil) // nil Provider and MCPServer for instruction-only mode
		var buf bytes.Buffer
		skill.SetOutput(&buf)

		err := skill.Run(context.Background(), scenarioDir)
		require.NoError(t, err)

		output := buf.String()

		// Should contain scenario name
		assert.Contains(t, output, "cross_domain_ab_deploy")

		// Should contain domain steps in correct order
		assert.Contains(t, output, "Step 1")
		assert.Contains(t, output, "domain-a")
		assert.Contains(t, output, "Step 2")
		assert.Contains(t, output, "domain-b")

		// Should contain MCP tool references
		assert.Contains(t, output, "wetwire_lint")
		assert.Contains(t, output, "wetwire_build")

		// Should contain cross-domain validation step
		assert.Contains(t, output, "cross-domain")

		// Should contain validation criteria
		assert.Contains(t, output, "Validation")
		assert.Contains(t, output, "Stacks")
	})

	t.Run("cross-domain validation with mock outputs", func(t *testing.T) {
		// Create temporary output directory with mock generated files
		tmpDir := t.TempDir()
		domainADir := filepath.Join(tmpDir, "domain-a", "templates")
		domainBDir := filepath.Join(tmpDir, "domain-b")

		require.NoError(t, os.MkdirAll(domainADir, 0755))
		require.NoError(t, os.MkdirAll(domainBDir, 0755))

		// Create mock domain-a templates
		resource1Template := `{
			"version": "1.0",
			"description": "Resource-1 Stack",
			"outputs": {
				"resource1_id": {"value": "r1-123", "export": {"name": "resource-1-id"}}
			}
		}`
		resource2Template := `{
			"version": "1.0",
			"description": "Resource-2 Stack",
			"outputs": {
				"resource2_name": {"value": "r2-cluster", "export": {"name": "resource-2-name"}}
			}
		}`
		resource3Template := `{
			"version": "1.0",
			"description": "Resource-3 Stack",
			"outputs": {
				"endpoint": {"value": "r3.example.com"}
			}
		}`

		require.NoError(t, os.WriteFile(filepath.Join(domainADir, "resource-1.json"), []byte(resource1Template), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(domainADir, "resource-2.json"), []byte(resource2Template), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(domainADir, "resource-3.json"), []byte(resource3Template), 0644))

		// Create mock domain-b config that references domain-a outputs
		domainBConfig := `
stages:
  - validate
  - deploy
  - test

variables:
  RESOURCE1_ID: "${domain-a.resource-1.outputs.resource1_id}"
  RESOURCE2_NAME: "${domain-a.resource-2.outputs.resource2_name}"
  RESOURCE3_ENDPOINT: "${domain-a.resource-3.outputs.endpoint}"

validate:
  stage: validate
  script:
    - mock-cli-a validate --template-body file://resource-1.json

deploy:
  stage: deploy
  script:
    - mock-cli-a deploy --stack-name resource-1 --template-file resource-1.json
    - mock-cli-a deploy --stack-name resource-2 --template-file resource-2.json
    - mock-cli-a deploy --stack-name resource-3 --template-file resource-3.json
`
		require.NoError(t, os.WriteFile(filepath.Join(domainBDir, "config.yml"), []byte(domainBConfig), 0644))

		// Run cross-domain validation
		ctx := context.Background()
		result, err := mcp.ValidateCrossDomain(ctx, map[string]any{
			"scenario":   filepath.Join(scenarioDir, "scenario.yaml"),
			"output_dir": tmpDir,
		})
		require.NoError(t, err)

		// Parse result
		var validationResult mcp.CrossDomainValidationResult
		err = json.Unmarshal([]byte(result), &validationResult)
		require.NoError(t, err)

		// Verify domains were validated
		assert.Contains(t, validationResult.DomainsValidated, "domain-a")
		assert.Contains(t, validationResult.DomainsValidated, "domain-b")

		// Verify cross-references were checked
		assert.NotEmpty(t, validationResult.CrossReferences)

		// Score should be positive for valid references
		assert.GreaterOrEqual(t, validationResult.Score, 0)
	})

	t.Run("prompt loading", func(t *testing.T) {
		config, err := scenario.Load(scenarioDir)
		require.NoError(t, err)

		// Verify prompt config
		require.NotNil(t, config.Prompts)
		assert.Equal(t, "prompt.md", config.Prompts.Default)
		assert.Contains(t, config.Prompts.Variants, "beginner")
		assert.Contains(t, config.Prompts.Variants, "expert")
		assert.Contains(t, config.Prompts.Variants, "terse")

		// Verify prompt file exists
		promptPath := filepath.Join(scenarioDir, config.Prompts.Default)
		_, err = os.Stat(promptPath)
		assert.NoError(t, err, "default prompt file should exist")

		// Verify variant prompt exists
		variantPath := filepath.Join(scenarioDir, config.Prompts.Variants["beginner"])
		_, err = os.Stat(variantPath)
		assert.NoError(t, err, "variant prompt file should exist")
	})
}

// TestIntegration_ValidationErrors tests that validation catches errors correctly.
func TestIntegration_ValidationErrors(t *testing.T) {
	t.Run("circular dependency detection", func(t *testing.T) {
		config := &scenario.ScenarioConfig{
			Name: "circular_test",
			Domains: []scenario.DomainSpec{
				{Name: "a", CLI: "cli-a", DependsOn: []string{"b"}},
				{Name: "b", CLI: "cli-b", DependsOn: []string{"c"}},
				{Name: "c", CLI: "cli-c", DependsOn: []string{"a"}}, // circular!
			},
		}

		_, err := scenario.GetDomainOrder(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "circular")
	})

	t.Run("cross-domain unknown domain", func(t *testing.T) {
		config := &scenario.ScenarioConfig{
			Name: "unknown_domain_test",
			Domains: []scenario.DomainSpec{
				{Name: "domain-a", CLI: "mock-cli-a"},
			},
			CrossDomain: []scenario.CrossDomainSpec{
				{From: "domain-a", To: "nonexistent", Type: "artifact_reference"},
			},
		}

		result := scenario.Validate(config)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Error(), "nonexistent")
	})

	t.Run("missing required fields", func(t *testing.T) {
		config := &scenario.ScenarioConfig{
			// Missing name and domains
		}

		result := scenario.Validate(config)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Error(), "name")
		assert.Contains(t, result.Error(), "domains")
	})
}

// TestIntegration_ComplexDependencyGraph tests multi-level dependency resolution.
func TestIntegration_ComplexDependencyGraph(t *testing.T) {
	config := &scenario.ScenarioConfig{
		Name: "complex_deps",
		Domains: []scenario.DomainSpec{
			{Name: "app", CLI: "cli", DependsOn: []string{"domain-c", "domain-d"}},
			{Name: "domain-c", CLI: "cli", DependsOn: []string{"domain-a"}},
			{Name: "domain-d", CLI: "cli", DependsOn: []string{"domain-a"}},
			{Name: "domain-a", CLI: "cli"},
			{Name: "monitoring", CLI: "cli", DependsOn: []string{"app", "domain-c"}},
		},
	}

	order, err := scenario.GetDomainOrder(config)
	require.NoError(t, err)
	require.Len(t, order, 5)

	// Build index map
	indexOf := func(name string) int {
		for i, n := range order {
			if n == name {
				return i
			}
		}
		return -1
	}

	// Verify ordering constraints
	assert.Less(t, indexOf("domain-a"), indexOf("domain-c"), "domain-a must come before domain-c")
	assert.Less(t, indexOf("domain-a"), indexOf("domain-d"), "domain-a must come before domain-d")
	assert.Less(t, indexOf("domain-c"), indexOf("app"), "domain-c must come before app")
	assert.Less(t, indexOf("domain-d"), indexOf("app"), "domain-d must come before app")
	assert.Less(t, indexOf("app"), indexOf("monitoring"), "app must come before monitoring")
	assert.Less(t, indexOf("domain-c"), indexOf("monitoring"), "domain-c must come before monitoring")
}

// TestIntegration_RecordingWithTermsvg tests the termsvg recording functionality.
func TestIntegration_RecordingWithTermsvg(t *testing.T) {
	examplesDir := getExamplesDir()
	scenarioDir := filepath.Join(examplesDir, "cross_domain_ab")

	if _, err := os.Stat(scenarioDir); os.IsNotExist(err) {
		t.Skip("examples/cross_domain_ab not found, skipping integration test")
	}

	t.Run("CanRecord returns bool", func(t *testing.T) {
		// CanRecord should return true if termsvg is installed, false otherwise
		canRecord := scenario.CanRecord()
		assert.IsType(t, true, canRecord)

		if canRecord {
			t.Log("termsvg is available - recording tests will execute")
		} else {
			t.Log("termsvg not installed - recording tests will skip gracefully")
		}
	})

	t.Run("RunWithRecording graceful fallback", func(t *testing.T) {
		tmpDir := t.TempDir()
		executed := false

		// With GracefulFallback, should always succeed
		err := scenario.RunWithRecording("test_scenario", scenario.RecordOptions{
			Enabled:          true,
			OutputDir:        tmpDir,
			GracefulFallback: true, // Key: don't fail if termsvg missing
		}, func() error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed, "function should have been executed")

		// If termsvg was available, SVG should exist
		if scenario.CanRecord() {
			svgPath := filepath.Join(tmpDir, "test_scenario.svg")
			_, err := os.Stat(svgPath)
			assert.NoError(t, err, "SVG file should exist when termsvg available")
		}
	})

	t.Run("recorder with skill execution", func(t *testing.T) {
		if !scenario.CanRecord() {
			t.Skip("termsvg not installed, skipping recording test")
		}

		tmpDir := t.TempDir()

		// Record skill execution to SVG
		// Note: skill must write to os.Stdout INSIDE the closure to be captured
		err := scenario.RunWithRecording("cross_domain_ab_skill", scenario.RecordOptions{
			Enabled:   true,
			OutputDir: tmpDir,
		}, func() error {
			// Create skill inside closure so it gets the redirected stdout
			skill := skillscenario.New(nil, nil) // nil Provider and MCPServer for instruction-only mode
			skill.SetOutput(os.Stdout)
			return skill.Run(context.Background(), scenarioDir)
		})

		require.NoError(t, err)

		// Verify SVG was created
		svgPath := filepath.Join(tmpDir, "cross_domain_ab_skill.svg")
		info, err := os.Stat(svgPath)
		require.NoError(t, err, "SVG file should exist")
		assert.Greater(t, info.Size(), int64(0), "SVG should have content")

		// Read SVG content to verify it contains scenario output
		svgContent, err := os.ReadFile(svgPath)
		require.NoError(t, err)
		assert.Contains(t, string(svgContent), "cross_domain_ab_deploy")
	})

	t.Run("RecordToSVG convenience function", func(t *testing.T) {
		tmpDir := t.TempDir()
		svgPath := filepath.Join(tmpDir, "recordings", "convenience_test.svg")

		executed := false
		err := scenario.RecordToSVG(svgPath, func() error {
			executed = true
			fmt.Println("Recording test output")
			fmt.Println("This should appear in the SVG")
			return nil
		})

		// Either succeeds or returns ErrTermsvgNotFound
		if err != nil {
			assert.ErrorIs(t, err, scenario.ErrTermsvgNotFound)
		} else {
			assert.True(t, executed)
			// Verify SVG exists and has content
			info, err := os.Stat(svgPath)
			assert.NoError(t, err)
			assert.Greater(t, info.Size(), int64(0), "SVG should have content")
		}
	})

	t.Run("recorder config and paths", func(t *testing.T) {
		config := scenario.RecorderConfig{
			OutputDir:    "/tmp/recordings",
			ScenarioName: "cross_domain_ab_deployment",
		}

		recorder := scenario.NewRecorder(config)

		assert.Equal(t, "/tmp/recordings/cross_domain_ab_deployment.svg", recorder.OutputPath())
		assert.Equal(t, "/tmp/recordings/cross_domain_ab_deployment.cast", recorder.CastPath())
	})
}
