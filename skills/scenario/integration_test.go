package scenario

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lex00/wetwire-core-go/mcp"
	"github.com/lex00/wetwire-core-go/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSkill_InstructionMode verifies that skill generates instructions when no provider/mcp is set
func TestSkill_InstructionMode(t *testing.T) {
	tmpDir := t.TempDir()
	scenarioPath := filepath.Join(tmpDir, "scenario.yaml")
	promptPath := filepath.Join(tmpDir, "prompt.md")

	scenarioYAML := `
name: test_scenario
description: Test scenario for instruction generation

prompts:
  default: prompt.md

domains:
  - name: aws
    cli: wetwire-aws
    mcp_tools:
      lint: wetwire_lint
      build: wetwire_build
    outputs:
      - cfn-templates/*.json

validation:
  aws:
    stacks:
      min: 1
      max: 5
`

	promptText := "Create a VPC with public and private subnets"

	err := os.WriteFile(scenarioPath, []byte(scenarioYAML), 0644)
	require.NoError(t, err)
	err = os.WriteFile(promptPath, []byte(promptText), 0644)
	require.NoError(t, err)

	// Create skill WITHOUT provider and MCP server
	skill := New(nil, nil)
	var buf bytes.Buffer
	skill.SetOutput(&buf)

	err = skill.Run(context.Background(), tmpDir)
	require.NoError(t, err)

	output := buf.String()

	// Verify instruction output
	assert.Contains(t, output, "test_scenario")
	assert.Contains(t, output, "aws")
	assert.Contains(t, output, "wetwire_lint")
	assert.Contains(t, output, "wetwire_build")
	assert.Contains(t, output, "VPC with public and private subnets")
	assert.Contains(t, output, "Stacks: min 1, max 5")
}

// TestSkill_ExecutionMode verifies that skill executes scenario when provider/mcp is set
func TestSkill_ExecutionMode(t *testing.T) {
	tmpDir := t.TempDir()
	scenarioPath := filepath.Join(tmpDir, "scenario.yaml")

	scenarioYAML := `
name: execution_test
description: Test scenario execution

domains:
  - name: test
    cli: wetwire-test
    mcp_tools:
      init: test_init
      write: test_write

validation:
  test:
    resources:
      min: 1
`

	err := os.WriteFile(scenarioPath, []byte(scenarioYAML), 0644)
	require.NoError(t, err)

	// Create mock provider that completes after one interaction
	provider := &mockProvider{
		responses: []*providers.MessageResponse{
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "I'll initialize the package"},
					{
						Type:  "tool_use",
						ID:    "tool-1",
						Name:  "test_init",
						Input: json.RawMessage(`{"name": "test-pkg"}`),
					},
				},
				StopReason: providers.StopReasonToolUse,
			},
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "Package initialized, scenario complete"},
				},
				StopReason: providers.StopReasonEndTurn,
			},
		},
	}

	// Create MCP server with test tools
	server := mcp.NewServer(mcp.Config{Name: "test"})
	server.RegisterTool("test_init", "Initialize a test package", func(ctx context.Context, args map[string]any) (string, error) {
		return "Package initialized", nil
	})
	server.RegisterTool("test_write", "Write a test file", func(ctx context.Context, args map[string]any) (string, error) {
		return "File written", nil
	})

	// Create skill WITH provider and MCP server
	skill := New(provider, server)
	var buf bytes.Buffer
	skill.SetOutput(&buf)

	err = skill.Run(context.Background(), tmpDir)
	require.NoError(t, err)

	output := buf.String()

	// Verify execution happened
	assert.Contains(t, output, "test_init")
	assert.Contains(t, output, "Package initialized")
	assert.Contains(t, output, "Scenario completed successfully")

	// Verify provider was called
	assert.Equal(t, 2, provider.callCount)
}

// TestSkill_MultipleDomains tests scenario with multiple domains
func TestSkill_MultipleDomains_InstructionMode(t *testing.T) {
	tmpDir := t.TempDir()
	scenarioPath := filepath.Join(tmpDir, "scenario.yaml")

	scenarioYAML := `
name: multi_domain
description: Multi-domain test

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
    outputs:
      - .gitlab-ci.yml

cross_domain:
  - from: aws
    to: gitlab
    type: artifact_reference

validation:
  aws:
    stacks:
      min: 1
  gitlab:
    pipelines:
      min: 1
`

	err := os.WriteFile(scenarioPath, []byte(scenarioYAML), 0644)
	require.NoError(t, err)

	skill := New(nil, nil)
	var buf bytes.Buffer
	skill.SetOutput(&buf)

	err = skill.Run(context.Background(), tmpDir)
	require.NoError(t, err)

	output := buf.String()

	// Verify both domains are present
	assert.Contains(t, output, "aws")
	assert.Contains(t, output, "gitlab")

	// Verify dependency order (aws before gitlab)
	awsIdx := bytes.Index(buf.Bytes(), []byte("aws"))
	gitlabIdx := bytes.Index(buf.Bytes(), []byte("gitlab"))
	assert.Less(t, awsIdx, gitlabIdx, "aws should come before gitlab")

	// Verify cross-domain section
	assert.Contains(t, output, "cross-domain")
	assert.Contains(t, output, "artifact_reference")

	// Verify validation criteria
	assert.Contains(t, output, "Validation Criteria")
	assert.Contains(t, output, "Stacks")
	assert.Contains(t, output, "Pipelines")
}
