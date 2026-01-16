package scenario

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testScenarioYAML = `
name: test_scenario
description: Test scenario for unit tests

prompts:
  default: prompt.md

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
  domain-a:
    stacks:
      min: 1
  domain-b:
    pipelines:
      min: 1
`

const testPrompt = `Generate infrastructure and CI pipeline.

Requirements:
- Resource with multiple components
- Cluster configuration
- Pipeline for deployment
`

func TestSkillName(t *testing.T) {
	skill := New(nil, nil)
	assert.Equal(t, "scenario", skill.Name())
}

func TestSkillDescription(t *testing.T) {
	skill := New(nil, nil)
	assert.Contains(t, skill.Description(), "scenario")
}

func TestSkillRun(t *testing.T) {
	// Create temp directory with scenario files
	tmpDir := t.TempDir()
	scenarioPath := filepath.Join(tmpDir, "scenario.yaml")
	promptPath := filepath.Join(tmpDir, "prompt.md")

	err := os.WriteFile(scenarioPath, []byte(testScenarioYAML), 0644)
	require.NoError(t, err)
	err = os.WriteFile(promptPath, []byte(testPrompt), 0644)
	require.NoError(t, err)

	skill := New(nil, nil)
	var buf bytes.Buffer
	skill.SetOutput(&buf)

	err = skill.Run(context.Background(), tmpDir)
	require.NoError(t, err)

	output := buf.String()

	// Should contain scenario name
	assert.Contains(t, output, "test_scenario")

	// Should contain domain instructions in order (domain-a before domain-b)
	domainAIdx := bytes.Index([]byte(output), []byte("domain-a"))
	domainBIdx := bytes.Index([]byte(output), []byte("domain-b"))
	assert.Less(t, domainAIdx, domainBIdx, "domain-a should come before domain-b")

	// Should contain MCP tool calls
	assert.Contains(t, output, "wetwire_lint")
	assert.Contains(t, output, "wetwire_build")
}

func TestSkillRunSingleDomain(t *testing.T) {
	singleDomainYAML := `
name: simple_scenario
description: Single domain scenario

domains:
  - name: domain-a
    cli: mock-cli-a
    mcp_tools:
      lint: wetwire_lint
`

	tmpDir := t.TempDir()
	scenarioPath := filepath.Join(tmpDir, "scenario.yaml")
	err := os.WriteFile(scenarioPath, []byte(singleDomainYAML), 0644)
	require.NoError(t, err)

	skill := New(nil, nil)
	var buf bytes.Buffer
	skill.SetOutput(&buf)

	err = skill.Run(context.Background(), tmpDir)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "simple_scenario")
	assert.Contains(t, output, "domain-a")
}

func TestSkillRunWithFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	scenarioPath := filepath.Join(tmpDir, "custom.yaml")

	simpleYAML := `
name: custom_scenario
domains:
  - name: domain-c
    cli: mock-cli-c
`
	err := os.WriteFile(scenarioPath, []byte(simpleYAML), 0644)
	require.NoError(t, err)

	skill := New(nil, nil)
	var buf bytes.Buffer
	skill.SetOutput(&buf)

	err = skill.Run(context.Background(), scenarioPath)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "custom_scenario")
}

func TestSkillRunScenarioNotFound(t *testing.T) {
	skill := New(nil, nil)
	var buf bytes.Buffer
	skill.SetOutput(&buf)

	err := skill.Run(context.Background(), "/nonexistent/path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scenario")
}

func TestSkillRunInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	scenarioPath := filepath.Join(tmpDir, "scenario.yaml")
	err := os.WriteFile(scenarioPath, []byte("invalid: yaml: content:"), 0644)
	require.NoError(t, err)

	skill := New(nil, nil)
	var buf bytes.Buffer
	skill.SetOutput(&buf)

	err = skill.Run(context.Background(), tmpDir)
	require.Error(t, err)
}

func TestSkillRunCurrentDirectory(t *testing.T) {
	// Create scenario in temp dir and change to it
	tmpDir := t.TempDir()
	scenarioPath := filepath.Join(tmpDir, "scenario.yaml")

	simpleYAML := `
name: current_dir_scenario
domains:
  - name: test
    cli: wetwire-test
`
	err := os.WriteFile(scenarioPath, []byte(simpleYAML), 0644)
	require.NoError(t, err)

	// Save current dir
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	// Change to temp dir
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	skill := New(nil, nil)
	var buf bytes.Buffer
	skill.SetOutput(&buf)

	// Empty args should use current directory
	err = skill.Run(context.Background(), "")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "current_dir_scenario")
}
