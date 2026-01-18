package scenario

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRef(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *CrossDomainRef
		expectError bool
	}{
		{
			name:  "valid AWS S3 reference",
			input: "${aws.s3.outputs.bucket_name}",
			expected: &CrossDomainRef{
				Domain:   "aws",
				Resource: "s3",
				Field:    "bucket_name",
				Raw:      "${aws.s3.outputs.bucket_name}",
			},
			expectError: false,
		},
		{
			name:  "valid GitLab pipeline reference",
			input: "${gitlab.pipeline.outputs.project_id}",
			expected: &CrossDomainRef{
				Domain:   "gitlab",
				Resource: "pipeline",
				Field:    "project_id",
				Raw:      "${gitlab.pipeline.outputs.project_id}",
			},
			expectError: false,
		},
		{
			name:  "valid reference with hyphenated names",
			input: "${domain-a.resource-1.outputs.resource_id}",
			expected: &CrossDomainRef{
				Domain:   "domain-a",
				Resource: "resource-1",
				Field:    "resource_id",
				Raw:      "${domain-a.resource-1.outputs.resource_id}",
			},
			expectError: false,
		},
		{
			name:  "valid reference with underscored field",
			input: "${aws.eks.outputs.cluster_endpoint_url}",
			expected: &CrossDomainRef{
				Domain:   "aws",
				Resource: "eks",
				Field:    "cluster_endpoint_url",
				Raw:      "${aws.eks.outputs.cluster_endpoint_url}",
			},
			expectError: false,
		},
		{
			name:        "invalid: missing outputs keyword",
			input:       "${aws.s3.bucket_name}",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid: missing field",
			input:       "${aws.s3.outputs}",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid: not a reference",
			input:       "just a regular string",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid: empty string",
			input:       "",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseRef(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Domain, result.Domain)
				assert.Equal(t, tt.expected.Resource, result.Resource)
				assert.Equal(t, tt.expected.Field, result.Field)
				assert.Equal(t, tt.expected.Raw, result.Raw)
			}
		})
	}
}

func TestFindRefsInString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*CrossDomainRef
	}{
		{
			name:  "single reference",
			input: "Use bucket: ${aws.s3.outputs.bucket_name}",
			expected: []*CrossDomainRef{
				{
					Domain:   "aws",
					Resource: "s3",
					Field:    "bucket_name",
					Raw:      "${aws.s3.outputs.bucket_name}",
				},
			},
		},
		{
			name:  "multiple references",
			input: "VPC: ${aws.vpc.outputs.vpc_id}, Cluster: ${aws.eks.outputs.cluster_name}",
			expected: []*CrossDomainRef{
				{
					Domain:   "aws",
					Resource: "vpc",
					Field:    "vpc_id",
					Raw:      "${aws.vpc.outputs.vpc_id}",
				},
				{
					Domain:   "aws",
					Resource: "eks",
					Field:    "cluster_name",
					Raw:      "${aws.eks.outputs.cluster_name}",
				},
			},
		},
		{
			name:     "no references",
			input:    "This is just a regular string without any references",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindRefsInString(tt.input)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				require.Equal(t, len(tt.expected), len(result))
				for i, expected := range tt.expected {
					assert.Equal(t, expected.Domain, result[i].Domain)
					assert.Equal(t, expected.Resource, result[i].Resource)
					assert.Equal(t, expected.Field, result[i].Field)
					assert.Equal(t, expected.Raw, result[i].Raw)
				}
			}
		})
	}
}

func TestCrossDomainRefString(t *testing.T) {
	ref := &CrossDomainRef{
		Domain:   "aws",
		Resource: "s3",
		Field:    "bucket_name",
		Raw:      "${aws.s3.outputs.bucket_name}",
	}

	assert.Equal(t, "${aws.s3.outputs.bucket_name}", ref.String())
}
