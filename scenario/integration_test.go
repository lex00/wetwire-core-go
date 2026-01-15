package scenario_test

import (
	"bytes"
	"context"
	"encoding/json"
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
	scenarioDir := filepath.Join(examplesDir, "aws_gitlab")

	// Skip if examples directory doesn't exist (e.g., in CI without examples)
	if _, err := os.Stat(scenarioDir); os.IsNotExist(err) {
		t.Skip("examples/aws_gitlab not found, skipping integration test")
	}

	t.Run("load and validate scenario", func(t *testing.T) {
		// Load scenario
		config, err := scenario.Load(scenarioDir)
		require.NoError(t, err, "should load scenario.yaml")

		// Verify basic structure
		assert.Equal(t, "aws_gitlab_deployment", config.Name)
		assert.Contains(t, config.Description, "AWS infrastructure")

		// Verify domains
		require.Len(t, config.Domains, 2)
		assert.Equal(t, "aws", config.Domains[0].Name)
		assert.Equal(t, "gitlab", config.Domains[1].Name)

		// Verify cross-domain relationships
		require.Len(t, config.CrossDomain, 1)
		assert.Equal(t, "aws", config.CrossDomain[0].From)
		assert.Equal(t, "gitlab", config.CrossDomain[0].To)

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

		// AWS should come before GitLab (gitlab depends on aws)
		require.Len(t, order, 2)
		assert.Equal(t, "aws", order[0], "aws should be first (no dependencies)")
		assert.Equal(t, "gitlab", order[1], "gitlab should be second (depends on aws)")
	})

	t.Run("skill generates instructions", func(t *testing.T) {
		skill := skillscenario.New()
		var buf bytes.Buffer
		skill.SetOutput(&buf)

		err := skill.Run(context.Background(), scenarioDir)
		require.NoError(t, err)

		output := buf.String()

		// Should contain scenario name
		assert.Contains(t, output, "aws_gitlab_deployment")

		// Should contain domain steps in correct order
		assert.Contains(t, output, "Step 1")
		assert.Contains(t, output, "aws")
		assert.Contains(t, output, "Step 2")
		assert.Contains(t, output, "gitlab")

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
		awsDir := filepath.Join(tmpDir, "aws", "cfn-templates")
		gitlabDir := filepath.Join(tmpDir, "gitlab")

		require.NoError(t, os.MkdirAll(awsDir, 0755))
		require.NoError(t, os.MkdirAll(gitlabDir, 0755))

		// Create mock AWS CloudFormation templates
		vpcTemplate := `{
			"AWSTemplateFormatVersion": "2010-09-09",
			"Description": "VPC Stack",
			"Outputs": {
				"VpcId": {"Value": {"Ref": "VPC"}, "Export": {"Name": "vpc-id"}}
			}
		}`
		eksTemplate := `{
			"AWSTemplateFormatVersion": "2010-09-09",
			"Description": "EKS Stack",
			"Outputs": {
				"ClusterName": {"Value": {"Ref": "EKSCluster"}, "Export": {"Name": "cluster-name"}}
			}
		}`
		rdsTemplate := `{
			"AWSTemplateFormatVersion": "2010-09-09",
			"Description": "RDS Stack",
			"Outputs": {
				"Endpoint": {"Value": {"Fn::GetAtt": ["Database", "Endpoint.Address"]}}
			}
		}`

		require.NoError(t, os.WriteFile(filepath.Join(awsDir, "vpc.json"), []byte(vpcTemplate), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(awsDir, "eks.json"), []byte(eksTemplate), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(awsDir, "rds.json"), []byte(rdsTemplate), 0644))

		// Create mock GitLab pipeline that references AWS outputs
		gitlabPipeline := `
stages:
  - validate
  - deploy
  - test

variables:
  VPC_ID: "${aws.vpc.outputs.vpc_id}"
  CLUSTER_NAME: "${aws.eks.outputs.cluster_name}"
  DB_ENDPOINT: "${aws.rds.outputs.endpoint}"

validate:
  stage: validate
  script:
    - aws cloudformation validate-template --template-body file://vpc.json

deploy:
  stage: deploy
  script:
    - aws cloudformation deploy --stack-name vpc --template-file vpc.json
    - aws cloudformation deploy --stack-name eks --template-file eks.json
    - aws cloudformation deploy --stack-name rds --template-file rds.json
`
		require.NoError(t, os.WriteFile(filepath.Join(gitlabDir, ".gitlab-ci.yml"), []byte(gitlabPipeline), 0644))

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
		assert.Contains(t, validationResult.DomainsValidated, "aws")
		assert.Contains(t, validationResult.DomainsValidated, "gitlab")

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
		assert.Contains(t, config.Prompts.Variants, "minimal")

		// Verify prompt file exists
		promptPath := filepath.Join(scenarioDir, config.Prompts.Default)
		_, err = os.Stat(promptPath)
		assert.NoError(t, err, "default prompt file should exist")

		// Verify variant prompt exists
		variantPath := filepath.Join(scenarioDir, config.Prompts.Variants["minimal"])
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
				{Name: "aws", CLI: "wetwire-aws"},
			},
			CrossDomain: []scenario.CrossDomainSpec{
				{From: "aws", To: "nonexistent", Type: "artifact_reference"},
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
			{Name: "app", CLI: "cli", DependsOn: []string{"k8s", "db"}},
			{Name: "k8s", CLI: "cli", DependsOn: []string{"aws"}},
			{Name: "db", CLI: "cli", DependsOn: []string{"aws"}},
			{Name: "aws", CLI: "cli"},
			{Name: "monitoring", CLI: "cli", DependsOn: []string{"app", "k8s"}},
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
	assert.Less(t, indexOf("aws"), indexOf("k8s"), "aws must come before k8s")
	assert.Less(t, indexOf("aws"), indexOf("db"), "aws must come before db")
	assert.Less(t, indexOf("k8s"), indexOf("app"), "k8s must come before app")
	assert.Less(t, indexOf("db"), indexOf("app"), "db must come before app")
	assert.Less(t, indexOf("app"), indexOf("monitoring"), "app must come before monitoring")
	assert.Less(t, indexOf("k8s"), indexOf("monitoring"), "k8s must come before monitoring")
}

// TestIntegration_RecordingWithTermsvg tests the termsvg recording functionality.
func TestIntegration_RecordingWithTermsvg(t *testing.T) {
	examplesDir := getExamplesDir()
	scenarioDir := filepath.Join(examplesDir, "aws_gitlab")

	if _, err := os.Stat(scenarioDir); os.IsNotExist(err) {
		t.Skip("examples/aws_gitlab not found, skipping integration test")
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
		skill := skillscenario.New()
		var buf bytes.Buffer
		skill.SetOutput(&buf)

		// Record skill execution to SVG
		err := scenario.RunWithRecording("aws_gitlab_skill", scenario.RecordOptions{
			Enabled:   true,
			OutputDir: tmpDir,
		}, func() error {
			return skill.Run(context.Background(), scenarioDir)
		})

		require.NoError(t, err)

		// Verify SVG was created
		svgPath := filepath.Join(tmpDir, "aws_gitlab_skill.svg")
		info, err := os.Stat(svgPath)
		require.NoError(t, err, "SVG file should exist")
		assert.Greater(t, info.Size(), int64(0), "SVG should have content")

		// Verify skill output was generated
		assert.Contains(t, buf.String(), "aws_gitlab_deployment")
	})

	t.Run("RecordToSVG convenience function", func(t *testing.T) {
		tmpDir := t.TempDir()
		svgPath := filepath.Join(tmpDir, "recordings", "convenience_test.svg")

		executed := false
		err := scenario.RecordToSVG(svgPath, func() error {
			executed = true
			return nil
		})

		// Either succeeds or returns ErrTermsvgNotFound
		if err != nil {
			assert.ErrorIs(t, err, scenario.ErrTermsvgNotFound)
		} else {
			assert.True(t, executed)
			// Verify SVG exists
			_, err := os.Stat(svgPath)
			assert.NoError(t, err)
		}
	})

	t.Run("recorder config and paths", func(t *testing.T) {
		config := scenario.RecorderConfig{
			OutputDir:    "/tmp/recordings",
			ScenarioName: "aws_gitlab_deployment",
		}

		recorder := scenario.NewRecorder(config)

		assert.Equal(t, "/tmp/recordings/aws_gitlab_deployment.svg", recorder.OutputPath())
		assert.Equal(t, "/tmp/recordings/aws_gitlab_deployment.cast", recorder.CastPath())
	})
}
