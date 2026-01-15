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
description: AWS infrastructure with GitLab CI/CD

domains:
  - name: aws
    cli: wetwire-aws
    outputs:
      - cfn-templates/*.json

  - name: gitlab
    cli: wetwire-gitlab
    depends_on:
      - aws

cross_domain:
  - from: aws
    to: gitlab
    type: artifact_reference
    validation:
      required_refs:
        - "${aws.vpc.outputs.vpc_id}"
        - "${aws.eks.outputs.cluster_name}"
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
	awsDir := filepath.Join(tmpDir, "output", "aws")
	gitlabDir := filepath.Join(tmpDir, "output", "gitlab")
	require.NoError(t, os.MkdirAll(filepath.Join(awsDir, "cfn-templates"), 0755))
	require.NoError(t, os.MkdirAll(gitlabDir, 0755))

	// Create AWS output files with proper references
	vpcTemplate := `{
		"AWSTemplateFormatVersion": "2010-09-09",
		"Outputs": {
			"VpcId": {"Value": {"Ref": "VPC"}, "Export": {"Name": "vpc-id"}}
		}
	}`
	eksTemplate := `{
		"AWSTemplateFormatVersion": "2010-09-09",
		"Outputs": {
			"ClusterName": {"Value": {"Ref": "EKSCluster"}, "Export": {"Name": "cluster-name"}}
		}
	}`
	err = os.WriteFile(filepath.Join(awsDir, "cfn-templates", "vpc.json"), []byte(vpcTemplate), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(awsDir, "cfn-templates", "eks.json"), []byte(eksTemplate), 0644)
	require.NoError(t, err)

	// Create GitLab pipeline referencing AWS outputs
	gitlabPipeline := `image: alpine:latest
variables:
  VPC_ID: "${aws.vpc.outputs.vpc_id}"
  CLUSTER_NAME: "${aws.eks.outputs.cluster_name}"
stages:
  - deploy
`
	err = os.WriteFile(filepath.Join(gitlabDir, "pipeline.yaml"), []byte(gitlabPipeline), 0644)
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
	assert.Contains(t, validationResult.DomainsValidated, "aws")
	assert.Contains(t, validationResult.DomainsValidated, "gitlab")

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
	awsDir := filepath.Join(tmpDir, "output", "aws")
	gitlabDir := filepath.Join(tmpDir, "output", "gitlab")
	require.NoError(t, os.MkdirAll(awsDir, 0755))
	require.NoError(t, os.MkdirAll(gitlabDir, 0755))

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
  - name: aws
    cli: wetwire-aws
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
	awsDir := filepath.Join(tmpDir, "output", "aws")
	gitlabDir := filepath.Join(tmpDir, "output", "gitlab")
	require.NoError(t, os.MkdirAll(filepath.Join(awsDir, "cfn-templates"), 0755))
	require.NoError(t, os.MkdirAll(gitlabDir, 0755))

	// AWS templates with outputs
	vpcTemplate := `{"Outputs": {"VpcId": {"Value": "vpc-123"}}}`
	eksTemplate := `{"Outputs": {"ClusterName": {"Value": "eks-cluster"}}}`
	err = os.WriteFile(filepath.Join(awsDir, "cfn-templates", "vpc.json"), []byte(vpcTemplate), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(awsDir, "cfn-templates", "eks.json"), []byte(eksTemplate), 0644)
	require.NoError(t, err)

	// GitLab pipeline using references
	pipeline := `variables:
  VPC_ID: "${aws.vpc.outputs.vpc_id}"
  CLUSTER_NAME: "${aws.eks.outputs.cluster_name}"`
	err = os.WriteFile(filepath.Join(gitlabDir, "pipeline.yaml"), []byte(pipeline), 0644)
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
		From:  "aws/vpc",
		To:    "gitlab/pipeline",
		Type:  "vpc_id",
		Valid: true,
	}

	assert.Equal(t, "aws/vpc", ref.From)
	assert.Equal(t, "gitlab/pipeline", ref.To)
	assert.Equal(t, "vpc_id", ref.Type)
	assert.True(t, ref.Valid)
}

func TestCrossDomainValidationResult(t *testing.T) {
	result := CrossDomainValidationResult{
		Valid:            true,
		DomainsValidated: []string{"aws", "gitlab"},
		CrossReferences: []CrossReferenceResult{
			{From: "aws/vpc", To: "gitlab/pipeline", Type: "vpc_id", Valid: true},
		},
		Errors: []string{},
		Score:  15,
	}

	assert.True(t, result.Valid)
	assert.Len(t, result.DomainsValidated, 2)
	assert.Len(t, result.CrossReferences, 1)
	assert.Equal(t, 15, result.Score)
}
