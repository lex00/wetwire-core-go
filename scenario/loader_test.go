package scenario

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testYAML = `
name: infrastructure_deployment
description: AWS infrastructure with GitLab CI/CD pipeline

domains:
  - name: aws
    cli: wetwire-aws
    mcp_tools:
      lint: wetwire_lint
      build: wetwire_build
    outputs:
      - cfn-templates/*.json

  - name: gitlab
    cli: wetwire-gitlab
    mcp_tools:
      lint: wetwire_lint
      build: wetwire_build
    depends_on:
      - aws

cross_domain:
  - from: aws
    to: gitlab
    type: artifact_reference
    validation:
      required_refs:
        - "${aws.vpc.outputs.vpc_id}"

validation:
  aws:
    stacks:
      min: 3
  gitlab:
    pipelines:
      min: 1
`

func TestParse(t *testing.T) {
	config, err := Parse([]byte(testYAML))
	require.NoError(t, err)

	assert.Equal(t, "infrastructure_deployment", config.Name)
	assert.Equal(t, "AWS infrastructure with GitLab CI/CD pipeline", config.Description)

	require.Len(t, config.Domains, 2)

	aws := config.GetDomain("aws")
	require.NotNil(t, aws)
	assert.Equal(t, "wetwire-aws", aws.CLI)
	assert.Equal(t, "wetwire_lint", aws.MCPTools["lint"])
	assert.Equal(t, []string{"cfn-templates/*.json"}, aws.Outputs)

	gitlab := config.GetDomain("gitlab")
	require.NotNil(t, gitlab)
	assert.Equal(t, []string{"aws"}, gitlab.DependsOn)

	require.Len(t, config.CrossDomain, 1)
	assert.Equal(t, "aws", config.CrossDomain[0].From)
	assert.Equal(t, "gitlab", config.CrossDomain[0].To)
	assert.Equal(t, "artifact_reference", config.CrossDomain[0].Type)

	assert.Equal(t, 3, config.Validation["aws"].Stacks.Min)
	assert.Equal(t, 1, config.Validation["gitlab"].Pipelines.Min)
}

func TestParseInvalidYAML(t *testing.T) {
	_, err := Parse([]byte("invalid: yaml: content:"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestLoadFile(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "scenario.yaml")
	err := os.WriteFile(path, []byte(testYAML), 0644)
	require.NoError(t, err)

	config, err := LoadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "infrastructure_deployment", config.Name)
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := LoadFile("/nonexistent/scenario.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read")
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "scenario.yaml")
	err := os.WriteFile(path, []byte(testYAML), 0644)
	require.NoError(t, err)

	t.Run("from directory", func(t *testing.T) {
		config, err := Load(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "infrastructure_deployment", config.Name)
	})

	t.Run("from file", func(t *testing.T) {
		config, err := Load(path)
		require.NoError(t, err)
		assert.Equal(t, "infrastructure_deployment", config.Name)
	})
}

func TestGetDomainOrder(t *testing.T) {
	t.Run("simple dependency", func(t *testing.T) {
		config := &ScenarioConfig{
			Domains: []DomainSpec{
				{Name: "gitlab", DependsOn: []string{"aws"}},
				{Name: "aws"},
			},
		}

		order, err := GetDomainOrder(config)
		require.NoError(t, err)
		assert.Equal(t, []string{"aws", "gitlab"}, order)
	})

	t.Run("no dependencies", func(t *testing.T) {
		config := &ScenarioConfig{
			Domains: []DomainSpec{
				{Name: "aws"},
				{Name: "gitlab"},
			},
		}

		order, err := GetDomainOrder(config)
		require.NoError(t, err)
		assert.Len(t, order, 2)
		assert.Contains(t, order, "aws")
		assert.Contains(t, order, "gitlab")
	})

	t.Run("multi-level dependencies", func(t *testing.T) {
		config := &ScenarioConfig{
			Domains: []DomainSpec{
				{Name: "k8s", DependsOn: []string{"aws"}},
				{Name: "gitlab", DependsOn: []string{"k8s"}},
				{Name: "aws"},
			},
		}

		order, err := GetDomainOrder(config)
		require.NoError(t, err)

		// aws must come before k8s, k8s before gitlab
		awsIdx := indexOf(order, "aws")
		k8sIdx := indexOf(order, "k8s")
		gitlabIdx := indexOf(order, "gitlab")

		assert.Less(t, awsIdx, k8sIdx)
		assert.Less(t, k8sIdx, gitlabIdx)
	})

	t.Run("circular dependency", func(t *testing.T) {
		config := &ScenarioConfig{
			Domains: []DomainSpec{
				{Name: "a", DependsOn: []string{"b"}},
				{Name: "b", DependsOn: []string{"a"}},
			},
		}

		_, err := GetDomainOrder(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "circular")
	})

	t.Run("unknown dependency", func(t *testing.T) {
		config := &ScenarioConfig{
			Domains: []DomainSpec{
				{Name: "aws", DependsOn: []string{"unknown"}},
			},
		}

		_, err := GetDomainOrder(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown")
	})
}

func indexOf(slice []string, s string) int {
	for i, v := range slice {
		if v == s {
			return i
		}
	}
	return -1
}
