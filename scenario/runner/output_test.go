package runner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOutputManifest(t *testing.T) {
	manifest := NewOutputManifest()

	assert.NotNil(t, manifest)
	assert.NotNil(t, manifest.Domains)
	assert.Empty(t, manifest.Domains)
}

func TestOutputManifest_AddDomainOutput(t *testing.T) {
	manifest := NewOutputManifest()

	output := &DomainOutput{
		Resources: map[string]ResourceOutput{
			"s3": {
				Type: "aws_s3",
				Outputs: map[string]interface{}{
					"bucket_name": "my-bucket",
				},
			},
		},
		Files: []string{"bucket.yaml"},
	}

	manifest.AddDomainOutput("aws", output)

	assert.Len(t, manifest.Domains, 1)
	assert.Equal(t, output, manifest.Domains["aws"])
}

func TestOutputManifest_AddDomainOutput_NilDomains(t *testing.T) {
	manifest := &OutputManifest{Domains: nil}

	output := &DomainOutput{
		Resources: map[string]ResourceOutput{},
	}

	manifest.AddDomainOutput("aws", output)

	assert.NotNil(t, manifest.Domains)
	assert.Len(t, manifest.Domains, 1)
}

func TestOutputManifest_GetDomainOutput(t *testing.T) {
	manifest := NewOutputManifest()
	output := &DomainOutput{
		Resources: map[string]ResourceOutput{
			"vpc": {
				Type: "aws_vpc",
				Outputs: map[string]interface{}{
					"vpc_id": "vpc-12345",
				},
			},
		},
	}
	manifest.AddDomainOutput("aws", output)

	t.Run("returns domain output when exists", func(t *testing.T) {
		result := manifest.GetDomainOutput("aws")
		assert.Equal(t, output, result)
	})

	t.Run("returns nil for non-existent domain", func(t *testing.T) {
		result := manifest.GetDomainOutput("nonexistent")
		assert.Nil(t, result)
	})

	t.Run("returns nil when Domains is nil", func(t *testing.T) {
		emptyManifest := &OutputManifest{Domains: nil}
		result := emptyManifest.GetDomainOutput("aws")
		assert.Nil(t, result)
	})
}

func TestOutputManifest_GetResourceOutput(t *testing.T) {
	manifest := NewOutputManifest()
	manifest.AddDomainOutput("aws", &DomainOutput{
		Resources: map[string]ResourceOutput{
			"s3": {
				Type: "aws_s3_bucket",
				Outputs: map[string]interface{}{
					"bucket_name": "my-bucket",
					"bucket_arn":  "arn:aws:s3:::my-bucket",
				},
			},
		},
	})

	t.Run("returns output value when exists", func(t *testing.T) {
		result := manifest.GetResourceOutput("aws", "s3", "bucket_name")
		assert.Equal(t, "my-bucket", result)
	})

	t.Run("returns nil for non-existent domain", func(t *testing.T) {
		result := manifest.GetResourceOutput("gitlab", "s3", "bucket_name")
		assert.Nil(t, result)
	})

	t.Run("returns nil for non-existent resource", func(t *testing.T) {
		result := manifest.GetResourceOutput("aws", "lambda", "bucket_name")
		assert.Nil(t, result)
	})

	t.Run("returns nil for non-existent output key", func(t *testing.T) {
		result := manifest.GetResourceOutput("aws", "s3", "nonexistent")
		assert.Nil(t, result)
	})
}

func TestOutputManifest_SaveAndLoadJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "output-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	manifest := NewOutputManifest()
	manifest.AddDomainOutput("aws", &DomainOutput{
		Resources: map[string]ResourceOutput{
			"s3": {
				Type: "aws_s3_bucket",
				Outputs: map[string]interface{}{
					"bucket_name": "test-bucket",
				},
			},
		},
		Files: []string{"bucket.yaml", "template.yaml"},
	})

	filePath := filepath.Join(tmpDir, "manifest.json")

	t.Run("saves to JSON file", func(t *testing.T) {
		err := manifest.SaveToFile(filePath)
		require.NoError(t, err)

		// Verify file exists and is valid JSON
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)

		var parsed map[string]interface{}
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)
		assert.Contains(t, parsed, "domains")
	})

	t.Run("loads from JSON file", func(t *testing.T) {
		loaded, err := LoadFromFile(filePath)
		require.NoError(t, err)

		assert.NotNil(t, loaded)
		assert.Len(t, loaded.Domains, 1)

		awsOutput := loaded.GetDomainOutput("aws")
		require.NotNil(t, awsOutput)
		assert.Equal(t, "test-bucket", awsOutput.Resources["s3"].Outputs["bucket_name"])
		assert.Equal(t, []string{"bucket.yaml", "template.yaml"}, awsOutput.Files)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := LoadFromFile(filepath.Join(tmpDir, "nonexistent.json"))
		assert.Error(t, err)
	})
}

func TestOutputManifest_SaveAndLoadYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "output-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	manifest := NewOutputManifest()
	manifest.AddDomainOutput("gitlab", &DomainOutput{
		Resources: map[string]ResourceOutput{
			"pipeline": {
				Type: "gitlab_pipeline",
				Outputs: map[string]interface{}{
					"pipeline_id": "456",
				},
			},
		},
		Files: []string{".gitlab-ci.yml"},
	})

	filePath := filepath.Join(tmpDir, "manifest.yaml")

	t.Run("saves to YAML file", func(t *testing.T) {
		err := manifest.SaveToYAML(filePath)
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(filePath)
		require.NoError(t, err)
	})

	t.Run("loads from YAML file", func(t *testing.T) {
		loaded, err := LoadFromYAML(filePath)
		require.NoError(t, err)

		assert.NotNil(t, loaded)
		gitlabOutput := loaded.GetDomainOutput("gitlab")
		require.NotNil(t, gitlabOutput)
		assert.Equal(t, "456", gitlabOutput.Resources["pipeline"].Outputs["pipeline_id"])
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := LoadFromYAML(filepath.Join(tmpDir, "nonexistent.yaml"))
		assert.Error(t, err)
	})
}

func TestOutputExtractor_ExtractFromDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "extractor-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	cfnYAML := `AWSTemplateFormatVersion: '2010-09-09'
Resources:
  MyBucket:
    Type: AWS::S3::Bucket
Outputs:
  BucketName:
    Value: !Ref MyBucket
  BucketArn:
    Value: !GetAtt MyBucket.Arn
`
	err = os.WriteFile(filepath.Join(tmpDir, "bucket.yaml"), []byte(cfnYAML), 0644)
	require.NoError(t, err)

	goFile := `package main

func createBucket() {
	bucket := NewBucket("my-bucket")
	bucket.Output("bucket_name", bucket.Ref())
	bucket.AddOutput("bucket_arn", bucket.Arn())
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "bucket.go"), []byte(goFile), 0644)
	require.NoError(t, err)

	extractor := NewOutputExtractor()

	t.Run("extracts outputs from YAML files", func(t *testing.T) {
		output, err := extractor.ExtractFromDir(tmpDir, "aws", []string{"*.yaml"})
		require.NoError(t, err)

		assert.NotNil(t, output)
		assert.Contains(t, output.Files, "bucket.yaml")

		// Check that outputs were extracted
		s3Resource := output.Resources["s3"]
		assert.NotNil(t, s3Resource.Outputs)
	})

	t.Run("extracts outputs from Go files", func(t *testing.T) {
		output, err := extractor.ExtractFromDir(tmpDir, "aws", []string{"*.go"})
		require.NoError(t, err)

		assert.NotNil(t, output)
		assert.Contains(t, output.Files, "bucket.go")

		// Check for Go DSL output patterns
		s3Resource := output.Resources["s3"]
		assert.NotNil(t, s3Resource.Outputs)
		assert.Contains(t, s3Resource.Outputs, "bucket_name")
		assert.Contains(t, s3Resource.Outputs, "bucket_arn")
	})

	t.Run("extracts from all files when no patterns specified", func(t *testing.T) {
		output, err := extractor.ExtractFromDir(tmpDir, "aws", nil)
		require.NoError(t, err)

		assert.NotNil(t, output)
		assert.GreaterOrEqual(t, len(output.Files), 2)
	})
}

func TestOutputExtractor_ExtractFromJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "extractor-json-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// CloudFormation JSON format
	cfnJSON := `{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Outputs": {
    "BucketName": {
      "Value": {"Ref": "MyBucket"},
      "Description": "Name of the bucket"
    },
    "BucketArn": {
      "Value": {"Fn::GetAtt": ["MyBucket", "Arn"]}
    }
  }
}`
	err = os.WriteFile(filepath.Join(tmpDir, "template.json"), []byte(cfnJSON), 0644)
	require.NoError(t, err)

	extractor := NewOutputExtractor()
	output, err := extractor.ExtractFromDir(tmpDir, "aws", []string{"*.json"})
	require.NoError(t, err)

	assert.NotNil(t, output)
	assert.Contains(t, output.Files, "template.json")

	// Check Outputs section was parsed
	cfnResource := output.Resources["cloudformation"]
	assert.NotNil(t, cfnResource.Outputs)
	assert.Contains(t, cfnResource.Outputs, "BucketName")
	assert.Contains(t, cfnResource.Outputs, "BucketArn")
}

func TestCaptureOutputsFromFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "capture-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a simple YAML file
	yamlContent := `Outputs:
  VpcId:
    Value: !Ref VPC
`
	err = os.WriteFile(filepath.Join(tmpDir, "vpc.yaml"), []byte(yamlContent), 0644)
	require.NoError(t, err)

	output, err := CaptureOutputsFromFiles(tmpDir, "aws", []string{"*.yaml"})
	require.NoError(t, err)

	assert.NotNil(t, output)
	assert.Contains(t, output.Files, "vpc.yaml")
}

func TestInferResourceName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"bucket.yaml", "s3"},
		{"s3_bucket.yaml", "s3"},
		{"lambda_function.go", "lambda"},
		{"my_vpc.yaml", "vpc"},
		{"iam_role.json", "iam"},
		{"deployment.yaml", "kubernetes"},
		{"pipeline.yaml", "pipeline"},
		{"workflow.yaml", "workflow"},
		{"template.yaml", "cloudformation"},
		{"custom_resource.yaml", "custom_resource"},
		{"my_service.yaml", "kubernetes"},
		{"rds_database.yaml", "rds"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := inferResourceName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchesPatterns(t *testing.T) {
	tests := []struct {
		path     string
		patterns []string
		expected bool
	}{
		{"bucket.yaml", []string{"*.yaml"}, true},
		{"bucket.yaml", []string{"*.json"}, false},
		{"nested/bucket.yaml", []string{"*.yaml"}, true},
		{"bucket.yaml", []string{"bucket.*"}, true},
		{"bucket.yaml", []string{"*.yaml", "*.json"}, true},
		{"bucket.yaml", []string{}, false},
		{"template.json", []string{"template.*"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := matchesPatterns(tt.path, tt.patterns)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewOutputExtractor(t *testing.T) {
	extractor := NewOutputExtractor()

	assert.NotNil(t, extractor)
	assert.NotNil(t, extractor.Patterns)
	assert.Contains(t, extractor.Patterns, ".yaml")
	assert.Contains(t, extractor.Patterns, ".yml")
	assert.Contains(t, extractor.Patterns, ".json")
	assert.Contains(t, extractor.Patterns, ".go")
}

func TestDefaultOutputPatterns(t *testing.T) {
	patterns := defaultOutputPatterns()

	t.Run("YAML patterns exist", func(t *testing.T) {
		yamlPatterns := patterns[".yaml"]
		assert.NotEmpty(t, yamlPatterns)
	})

	t.Run("JSON patterns exist", func(t *testing.T) {
		jsonPatterns := patterns[".json"]
		assert.NotEmpty(t, jsonPatterns)
	})

	t.Run("Go patterns exist", func(t *testing.T) {
		goPatterns := patterns[".go"]
		assert.NotEmpty(t, goPatterns)

		// Test Go pattern matches Output() calls
		goPattern := goPatterns[0]
		matches := goPattern.Regex.FindStringSubmatch(`Output("bucket_name", val)`)
		assert.NotEmpty(t, matches)
		assert.Equal(t, "bucket_name", matches[1])
	})
}

func TestOutputManifest_SaveCreatesDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "save-dir-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	manifest := NewOutputManifest()
	manifest.AddDomainOutput("aws", &DomainOutput{
		Resources: map[string]ResourceOutput{},
	})

	// Save to nested directory that doesn't exist
	nestedPath := filepath.Join(tmpDir, "nested", "deep", "manifest.json")
	err = manifest.SaveToFile(nestedPath)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(nestedPath)
	require.NoError(t, err)
}

func TestOutputExtractor_HandlesMalformedFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "malformed-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create malformed JSON
	err = os.WriteFile(filepath.Join(tmpDir, "bad.json"), []byte("not valid json"), 0644)
	require.NoError(t, err)

	// Create malformed YAML
	err = os.WriteFile(filepath.Join(tmpDir, "bad.yaml"), []byte(":::invalid:yaml:::"), 0644)
	require.NoError(t, err)

	extractor := NewOutputExtractor()
	output, err := extractor.ExtractFromDir(tmpDir, "test", nil)

	// Should not error - just skip malformed files
	require.NoError(t, err)
	assert.NotNil(t, output)
	// Files should still be tracked even if parsing failed
	assert.Contains(t, output.Files, "bad.json")
	assert.Contains(t, output.Files, "bad.yaml")
}

func TestDomainOutput_EmptyResources(t *testing.T) {
	output := &DomainOutput{
		Resources: map[string]ResourceOutput{},
		Files:     []string{},
	}

	assert.Empty(t, output.Resources)
	assert.Empty(t, output.Files)
}

func TestResourceOutput_NilOutputs(t *testing.T) {
	manifest := NewOutputManifest()
	manifest.AddDomainOutput("aws", &DomainOutput{
		Resources: map[string]ResourceOutput{
			"s3": {
				Type:    "aws_s3",
				Outputs: nil,
			},
		},
	})

	result := manifest.GetResourceOutput("aws", "s3", "any_key")
	assert.Nil(t, result)
}
