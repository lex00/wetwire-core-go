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
description: Domain A infrastructure with Domain B CI/CD pipeline

domains:
  - name: domain-a
    cli: mock-cli-a
    mcp_tools:
      lint: wetwire_lint
      build: wetwire_build
    outputs:
      - templates/*.json

  - name: domain-b
    cli: mock-cli-b
    mcp_tools:
      lint: wetwire_lint
      build: wetwire_build
    depends_on:
      - domain-a

cross_domain:
  - from: domain-a
    to: domain-b
    type: artifact_reference
    validation:
      required_refs:
        - "${domain-a.resource-1.outputs.resource_id}"

validation:
  domain-a:
    stacks:
      min: 3
  domain-b:
    pipelines:
      min: 1
`

func TestParse(t *testing.T) {
	config, err := Parse([]byte(testYAML))
	require.NoError(t, err)

	assert.Equal(t, "infrastructure_deployment", config.Name)
	assert.Equal(t, "Domain A infrastructure with Domain B CI/CD pipeline", config.Description)

	require.Len(t, config.Domains, 2)

	domainA := config.GetDomain("domain-a")
	require.NotNil(t, domainA)
	assert.Equal(t, "mock-cli-a", domainA.CLI)
	assert.Equal(t, "wetwire_lint", domainA.MCPTools["lint"])
	assert.Equal(t, []string{"templates/*.json"}, domainA.Outputs)

	domainB := config.GetDomain("domain-b")
	require.NotNil(t, domainB)
	assert.Equal(t, []string{"domain-a"}, domainB.DependsOn)

	require.Len(t, config.CrossDomain, 1)
	assert.Equal(t, "domain-a", config.CrossDomain[0].From)
	assert.Equal(t, "domain-b", config.CrossDomain[0].To)
	assert.Equal(t, "artifact_reference", config.CrossDomain[0].Type)

	assert.Equal(t, 3, config.Validation["domain-a"].Stacks.Min)
	assert.Equal(t, 1, config.Validation["domain-b"].Pipelines.Min)
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
				{Name: "domain-b", DependsOn: []string{"domain-a"}},
				{Name: "domain-a"},
			},
		}

		order, err := GetDomainOrder(config)
		require.NoError(t, err)
		assert.Equal(t, []string{"domain-a", "domain-b"}, order)
	})

	t.Run("no dependencies", func(t *testing.T) {
		config := &ScenarioConfig{
			Domains: []DomainSpec{
				{Name: "domain-a"},
				{Name: "domain-b"},
			},
		}

		order, err := GetDomainOrder(config)
		require.NoError(t, err)
		assert.Len(t, order, 2)
		assert.Contains(t, order, "domain-a")
		assert.Contains(t, order, "domain-b")
	})

	t.Run("multi-level dependencies", func(t *testing.T) {
		config := &ScenarioConfig{
			Domains: []DomainSpec{
				{Name: "domain-c", DependsOn: []string{"domain-a"}},
				{Name: "domain-b", DependsOn: []string{"domain-c"}},
				{Name: "domain-a"},
			},
		}

		order, err := GetDomainOrder(config)
		require.NoError(t, err)

		// domain-a must come before domain-c, domain-c before domain-b
		domainAIdx := indexOf(order, "domain-a")
		domainCIdx := indexOf(order, "domain-c")
		domainBIdx := indexOf(order, "domain-b")

		assert.Less(t, domainAIdx, domainCIdx)
		assert.Less(t, domainCIdx, domainBIdx)
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
				{Name: "domain-a", DependsOn: []string{"unknown"}},
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
