package runner

import (
	"testing"

	"github.com/lex00/wetwire-core-go/scenario"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRefs(t *testing.T) {
	// Create a test manifest
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
			"vpc": {
				Type: "aws_vpc",
				Outputs: map[string]interface{}{
					"vpc_id":     "vpc-12345",
					"cidr_block": "10.0.0.0/16",
				},
			},
		},
	})
	manifest.AddDomainOutput("gitlab", &DomainOutput{
		Resources: map[string]ResourceOutput{
			"pipeline": {
				Type: "gitlab_pipeline",
				Outputs: map[string]interface{}{
					"project_id":  "123",
					"pipeline_id": "456",
				},
			},
		},
	})

	tests := []struct {
		name        string
		config      *scenario.ScenarioConfig
		expectError bool
		errorCount  int
	}{
		{
			name: "valid references",
			config: &scenario.ScenarioConfig{
				Name: "test-scenario",
				CrossDomain: []scenario.CrossDomainSpec{
					{
						From: "aws",
						To:   "gitlab",
						Type: "artifact_reference",
						Validation: scenario.CrossDomainValidation{
							RequiredRefs: []string{
								"${aws.s3.outputs.bucket_name}",
								"${aws.vpc.outputs.vpc_id}",
							},
						},
					},
				},
			},
			expectError: false,
			errorCount:  0,
		},
		{
			name: "invalid: domain not found",
			config: &scenario.ScenarioConfig{
				Name: "test-scenario",
				CrossDomain: []scenario.CrossDomainSpec{
					{
						From: "aws",
						To:   "gitlab",
						Type: "artifact_reference",
						Validation: scenario.CrossDomainValidation{
							RequiredRefs: []string{
								"${nonexistent.s3.outputs.bucket_name}",
							},
						},
					},
				},
			},
			expectError: true,
			errorCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := manifest.ValidateRefs(tt.config)

			if tt.expectError {
				assert.NotEmpty(t, errors)
				assert.Equal(t, tt.errorCount, len(errors))
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}

func TestResolveRef(t *testing.T) {
	// Create a test manifest
	manifest := NewOutputManifest()
	manifest.AddDomainOutput("aws", &DomainOutput{
		Resources: map[string]ResourceOutput{
			"s3": {
				Type: "aws_s3_bucket",
				Outputs: map[string]interface{}{
					"bucket_name": "my-test-bucket",
					"bucket_arn":  "arn:aws:s3:::my-test-bucket",
				},
			},
		},
	})

	tests := []struct {
		name        string
		ref         *scenario.CrossDomainRef
		expected    interface{}
		expectError bool
	}{
		{
			name: "resolve string value",
			ref: &scenario.CrossDomainRef{
				Domain:   "aws",
				Resource: "s3",
				Field:    "bucket_name",
				Raw:      "${aws.s3.outputs.bucket_name}",
			},
			expected:    "my-test-bucket",
			expectError: false,
		},
		{
			name: "invalid domain",
			ref: &scenario.CrossDomainRef{
				Domain:   "nonexistent",
				Resource: "s3",
				Field:    "bucket_name",
				Raw:      "${nonexistent.s3.outputs.bucket_name}",
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := manifest.ResolveRef(tt.ref)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
