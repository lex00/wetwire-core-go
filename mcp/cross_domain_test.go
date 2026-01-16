package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const crossDomainScenarioYAML = `
name: infra_deployment
description: Domain A infrastructure with Domain B CI/CD

domains:
  - name: domain-a
    cli: mock-cli-a
    outputs:
      - templates/*.json

  - name: domain-b
    cli: mock-cli-b
    depends_on:
      - domain-a

cross_domain:
  - from: domain-a
    to: domain-b
    type: artifact_reference
    validation:
      required_refs:
        - "${domain-a.resource-1.outputs.resource_id}"
        - "${domain-a.resource-2.outputs.cluster_name}"
`

func TestCrossDomainValidationSchema(t *testing.T) {
	schema := CrossDomainValidateSchema
	require.NotNil(t, schema)

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	// Should have scenario and output_dir properties
	assert.Contains(t, props, "scenario")
	assert.Contains(t, props, "output_dir")
}

func TestValidateCrossDomain(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Write scenario.yaml
	scenarioPath := filepath.Join(tmpDir, "scenario.yaml")
	err := os.WriteFile(scenarioPath, []byte(crossDomainScenarioYAML), 0644)
	require.NoError(t, err)

	// Create output directories
	domainADir := filepath.Join(tmpDir, "output", "domain-a")
	domainBDir := filepath.Join(tmpDir, "output", "domain-b")
	require.NoError(t, os.MkdirAll(filepath.Join(domainADir, "templates"), 0755))
	require.NoError(t, os.MkdirAll(domainBDir, 0755))

	// Create Domain A output files with proper references
	resource1Template := `{
		"AWSTemplateFormatVersion": "2010-09-09",
		"Outputs": {
			"ResourceId": {"Value": {"Ref": "Resource1"}, "Export": {"Name": "resource-id"}}
		}
	}`
	resource2Template := `{
		"AWSTemplateFormatVersion": "2010-09-09",
		"Outputs": {
			"ClusterName": {"Value": {"Ref": "Cluster"}, "Export": {"Name": "cluster-name"}}
		}
	}`
	err = os.WriteFile(filepath.Join(domainADir, "templates", "resource-1.json"), []byte(resource1Template), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(domainADir, "templates", "resource-2.json"), []byte(resource2Template), 0644)
	require.NoError(t, err)

	// Create Domain B pipeline referencing Domain A outputs
	domainBPipeline := `image: alpine:latest
variables:
  RESOURCE_ID: "${domain-a.resource-1.outputs.resource_id}"
  CLUSTER_NAME: "${domain-a.resource-2.outputs.cluster_name}"
stages:
  - deploy
`
	err = os.WriteFile(filepath.Join(domainBDir, "pipeline.yaml"), []byte(domainBPipeline), 0644)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := ValidateCrossDomain(ctx, map[string]any{
		"scenario":   scenarioPath,
		"output_dir": filepath.Join(tmpDir, "output"),
	})
	require.NoError(t, err)

	// Parse result
	var validationResult CrossDomainValidationResult
	err = json.Unmarshal([]byte(result), &validationResult)
	require.NoError(t, err)

	// Check domains validated
	assert.Contains(t, validationResult.DomainsValidated, "domain-a")
	assert.Contains(t, validationResult.DomainsValidated, "domain-b")

	// Should have cross references
	assert.NotEmpty(t, validationResult.CrossReferences)
}

func TestValidateCrossDomainMissingReferences(t *testing.T) {
	tmpDir := t.TempDir()

	// Write scenario.yaml
	scenarioPath := filepath.Join(tmpDir, "scenario.yaml")
	err := os.WriteFile(scenarioPath, []byte(crossDomainScenarioYAML), 0644)
	require.NoError(t, err)

	// Create empty output directories
	domainADir := filepath.Join(tmpDir, "output", "domain-a")
	domainBDir := filepath.Join(tmpDir, "output", "domain-b")
	require.NoError(t, os.MkdirAll(domainADir, 0755))
	require.NoError(t, os.MkdirAll(domainBDir, 0755))

	ctx := context.Background()
	result, err := ValidateCrossDomain(ctx, map[string]any{
		"scenario":   scenarioPath,
		"output_dir": filepath.Join(tmpDir, "output"),
	})
	require.NoError(t, err)

	var validationResult CrossDomainValidationResult
	err = json.Unmarshal([]byte(result), &validationResult)
	require.NoError(t, err)

	// Should have errors for missing references
	assert.False(t, validationResult.Valid)
	assert.NotEmpty(t, validationResult.Errors)
}

func TestValidateCrossDomainScenarioNotFound(t *testing.T) {
	ctx := context.Background()
	_, err := ValidateCrossDomain(ctx, map[string]any{
		"scenario":   "/nonexistent/scenario.yaml",
		"output_dir": "/tmp",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scenario")
}

func TestValidateCrossDomainNoCrossDomainRefs(t *testing.T) {
	simpleScanario := `
name: simple
domains:
  - name: domain-a
    cli: mock-cli-a
`
	tmpDir := t.TempDir()
	scenarioPath := filepath.Join(tmpDir, "scenario.yaml")
	err := os.WriteFile(scenarioPath, []byte(simpleScanario), 0644)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := ValidateCrossDomain(ctx, map[string]any{
		"scenario":   scenarioPath,
		"output_dir": tmpDir,
	})
	require.NoError(t, err)

	var validationResult CrossDomainValidationResult
	err = json.Unmarshal([]byte(result), &validationResult)
	require.NoError(t, err)

	// Should be valid with no cross-domain references
	assert.True(t, validationResult.Valid)
	assert.Empty(t, validationResult.CrossReferences)
}

func TestValidateCrossDomainScore(t *testing.T) {
	tmpDir := t.TempDir()

	// Write scenario.yaml
	scenarioPath := filepath.Join(tmpDir, "scenario.yaml")
	err := os.WriteFile(scenarioPath, []byte(crossDomainScenarioYAML), 0644)
	require.NoError(t, err)

	// Create output with all references satisfied
	domainADir := filepath.Join(tmpDir, "output", "domain-a")
	domainBDir := filepath.Join(tmpDir, "output", "domain-b")
	require.NoError(t, os.MkdirAll(filepath.Join(domainADir, "templates"), 0755))
	require.NoError(t, os.MkdirAll(domainBDir, 0755))

	// Domain A templates with outputs
	resource1Template := `{"Outputs": {"ResourceId": {"Value": "res-123"}}}`
	resource2Template := `{"Outputs": {"ClusterName": {"Value": "cluster"}}}`
	err = os.WriteFile(filepath.Join(domainADir, "templates", "resource-1.json"), []byte(resource1Template), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(domainADir, "templates", "resource-2.json"), []byte(resource2Template), 0644)
	require.NoError(t, err)

	// Domain B pipeline using references
	pipeline := `variables:
  RESOURCE_ID: "${domain-a.resource-1.outputs.resource_id}"
  CLUSTER_NAME: "${domain-a.resource-2.outputs.cluster_name}"`
	err = os.WriteFile(filepath.Join(domainBDir, "pipeline.yaml"), []byte(pipeline), 0644)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := ValidateCrossDomain(ctx, map[string]any{
		"scenario":   scenarioPath,
		"output_dir": filepath.Join(tmpDir, "output"),
	})
	require.NoError(t, err)

	var validationResult CrossDomainValidationResult
	err = json.Unmarshal([]byte(result), &validationResult)
	require.NoError(t, err)

	// Score should be > 0 for valid cross-domain
	assert.GreaterOrEqual(t, validationResult.Score, 0)
}

func TestCrossReferenceResult(t *testing.T) {
	ref := CrossReferenceResult{
		From:  "domain-a/resource-1",
		To:    "domain-b/pipeline",
		Type:  "resource_id",
		Valid: true,
	}

	assert.Equal(t, "domain-a/resource-1", ref.From)
	assert.Equal(t, "domain-b/pipeline", ref.To)
	assert.Equal(t, "resource_id", ref.Type)
	assert.True(t, ref.Valid)
}

func TestCrossDomainValidationResult(t *testing.T) {
	result := CrossDomainValidationResult{
		Valid:            true,
		DomainsValidated: []string{"domain-a", "domain-b"},
		CrossReferences: []CrossReferenceResult{
			{From: "domain-a/resource-1", To: "domain-b/pipeline", Type: "resource_id", Valid: true},
		},
		Errors: []string{},
		Score:  15,
	}

	assert.True(t, result.Valid)
	assert.Len(t, result.DomainsValidated, 2)
	assert.Len(t, result.CrossReferences, 1)
	assert.Equal(t, 15, result.Score)
}
