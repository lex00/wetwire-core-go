package runner

import (
	"strings"
	"testing"

	"github.com/lex00/wetwire-core-go/domain"
	"github.com/lex00/wetwire-core-go/scenario"
)

func TestOutputManifestToCrossDomainContext(t *testing.T) {
	t.Run("nil manifest", func(t *testing.T) {
		result := OutputManifestToCrossDomainContext(nil)
		if result != nil {
			t.Error("Expected nil for nil manifest")
		}
	})

	t.Run("empty manifest", func(t *testing.T) {
		manifest := NewOutputManifest()
		result := OutputManifestToCrossDomainContext(manifest)
		if result != nil {
			t.Error("Expected nil for empty manifest")
		}
	})

	t.Run("converts single domain", func(t *testing.T) {
		manifest := NewOutputManifest()
		manifest.AddDomainOutput("aws", &DomainOutput{
			Resources: map[string]ResourceOutput{
				"bucket": {
					Type: "aws_s3_bucket",
					Outputs: map[string]interface{}{
						"name": "test-bucket",
						"arn":  "arn:aws:s3:::test-bucket",
					},
				},
			},
		})

		result := OutputManifestToCrossDomainContext(manifest)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		// Verify domain exists
		if !result.HasDependency("aws") {
			t.Error("Expected aws domain to exist")
		}

		// Verify resource output
		value := result.GetResourceOutput("aws", "bucket", "name")
		if value != "test-bucket" {
			t.Errorf("Expected 'test-bucket', got %v", value)
		}
	})

	t.Run("converts multiple domains", func(t *testing.T) {
		manifest := NewOutputManifest()
		manifest.AddDomainOutput("aws", &DomainOutput{
			Resources: map[string]ResourceOutput{
				"bucket": {
					Type: "aws_s3_bucket",
					Outputs: map[string]interface{}{
						"name": "aws-bucket",
					},
				},
			},
		})
		manifest.AddDomainOutput("gitlab", &DomainOutput{
			Resources: map[string]ResourceOutput{
				"pipeline": {
					Type: "gitlab_pipeline",
					Outputs: map[string]interface{}{
						"id": "12345",
					},
				},
			},
		})

		result := OutputManifestToCrossDomainContext(manifest)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		// Verify both domains exist
		if !result.HasDependency("aws") {
			t.Error("Expected aws domain to exist")
		}
		if !result.HasDependency("gitlab") {
			t.Error("Expected gitlab domain to exist")
		}

		// Verify outputs from both domains
		awsValue := result.GetResourceOutput("aws", "bucket", "name")
		if awsValue != "aws-bucket" {
			t.Errorf("Expected 'aws-bucket', got %v", awsValue)
		}

		gitlabValue := result.GetResourceOutput("gitlab", "pipeline", "id")
		if gitlabValue != "12345" {
			t.Errorf("Expected '12345', got %v", gitlabValue)
		}
	})

	t.Run("skips nil domain output", func(t *testing.T) {
		manifest := &OutputManifest{
			Domains: map[string]*DomainOutput{
				"aws":    nil,
				"gitlab": {Resources: map[string]ResourceOutput{}},
			},
		}

		result := OutputManifestToCrossDomainContext(manifest)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		// AWS should not exist since it was nil
		if result.HasDependency("aws") {
			t.Error("Expected aws domain to not exist (was nil)")
		}

		// GitLab should exist
		if !result.HasDependency("gitlab") {
			t.Error("Expected gitlab domain to exist")
		}
	})
}

func TestCrossDomainContextToOutputManifest(t *testing.T) {
	t.Run("nil context", func(t *testing.T) {
		result := CrossDomainContextToOutputManifest(nil)
		if result != nil {
			t.Error("Expected nil for nil context")
		}
	})

	t.Run("empty context", func(t *testing.T) {
		ctx := domain.NewCrossDomainContext()
		result := CrossDomainContextToOutputManifest(ctx)
		if result != nil {
			t.Error("Expected nil for empty context")
		}
	})

	t.Run("converts single domain", func(t *testing.T) {
		ctx := domain.NewCrossDomainContext()
		ctx.AddDomainOutputs("aws", &domain.DomainOutputs{
			Resources: map[string]*domain.ResourceOutputs{
				"bucket": {
					Type: "aws_s3_bucket",
					Outputs: map[string]interface{}{
						"name": "test-bucket",
					},
				},
			},
		})

		result := CrossDomainContextToOutputManifest(ctx)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		// Verify domain exists
		domainOutput := result.GetDomainOutput("aws")
		if domainOutput == nil {
			t.Fatal("Expected aws domain to exist")
		}

		// Verify resource output
		resourceOutput, ok := domainOutput.Resources["bucket"]
		if !ok {
			t.Fatal("Expected bucket resource to exist")
		}
		if resourceOutput.Type != "aws_s3_bucket" {
			t.Errorf("Expected type 'aws_s3_bucket', got '%s'", resourceOutput.Type)
		}
		if resourceOutput.Outputs["name"] != "test-bucket" {
			t.Errorf("Expected name 'test-bucket', got %v", resourceOutput.Outputs["name"])
		}
	})

	t.Run("skips nil domain outputs", func(t *testing.T) {
		ctx := &domain.CrossDomainContext{
			Dependencies: map[string]*domain.DomainOutputs{
				"aws":    nil,
				"gitlab": {Resources: map[string]*domain.ResourceOutputs{}},
			},
		}

		result := CrossDomainContextToOutputManifest(ctx)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		// AWS should not exist since it was nil
		if result.GetDomainOutput("aws") != nil {
			t.Error("Expected aws domain to not exist (was nil)")
		}

		// GitLab should exist
		if result.GetDomainOutput("gitlab") == nil {
			t.Error("Expected gitlab domain to exist")
		}
	})

	t.Run("skips nil resource outputs", func(t *testing.T) {
		ctx := domain.NewCrossDomainContext()
		ctx.AddDomainOutputs("aws", &domain.DomainOutputs{
			Resources: map[string]*domain.ResourceOutputs{
				"bucket": nil,
				"queue": {
					Type: "aws_sqs_queue",
					Outputs: map[string]interface{}{
						"url": "https://sqs.example.com/queue",
					},
				},
			},
		})

		result := CrossDomainContextToOutputManifest(ctx)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		domainOutput := result.GetDomainOutput("aws")
		if domainOutput == nil {
			t.Fatal("Expected aws domain to exist")
		}

		// bucket should not exist since it was nil
		if _, ok := domainOutput.Resources["bucket"]; ok {
			t.Error("Expected bucket resource to not exist (was nil)")
		}

		// queue should exist
		if _, ok := domainOutput.Resources["queue"]; !ok {
			t.Error("Expected queue resource to exist")
		}
	})
}

func TestRoundTripConversion(t *testing.T) {
	// Create an OutputManifest
	original := NewOutputManifest()
	original.AddDomainOutput("aws", &DomainOutput{
		Resources: map[string]ResourceOutput{
			"bucket": {
				Type: "aws_s3_bucket",
				Outputs: map[string]interface{}{
					"name": "test-bucket",
					"arn":  "arn:aws:s3:::test-bucket",
				},
			},
			"queue": {
				Type: "aws_sqs_queue",
				Outputs: map[string]interface{}{
					"url": "https://sqs.example.com/queue",
				},
			},
		},
	})
	original.AddDomainOutput("gitlab", &DomainOutput{
		Resources: map[string]ResourceOutput{
			"pipeline": {
				Type: "gitlab_pipeline",
				Outputs: map[string]interface{}{
					"id": "12345",
				},
			},
		},
	})

	// Convert to CrossDomainContext
	crossDomain := OutputManifestToCrossDomainContext(original)
	if crossDomain == nil {
		t.Fatal("Expected non-nil crossDomain")
	}

	// Convert back to OutputManifest
	result := CrossDomainContextToOutputManifest(crossDomain)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Verify domains
	if len(result.Domains) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(result.Domains))
	}

	// Verify AWS resources
	awsDomain := result.GetDomainOutput("aws")
	if awsDomain == nil {
		t.Fatal("Expected aws domain to exist")
	}
	if len(awsDomain.Resources) != 2 {
		t.Errorf("Expected 2 AWS resources, got %d", len(awsDomain.Resources))
	}

	bucket := awsDomain.Resources["bucket"]
	if bucket.Outputs["name"] != "test-bucket" {
		t.Errorf("Expected bucket name 'test-bucket', got %v", bucket.Outputs["name"])
	}

	// Verify GitLab resources
	gitlabDomain := result.GetDomainOutput("gitlab")
	if gitlabDomain == nil {
		t.Fatal("Expected gitlab domain to exist")
	}
	if len(gitlabDomain.Resources) != 1 {
		t.Errorf("Expected 1 GitLab resource, got %d", len(gitlabDomain.Resources))
	}
}

func TestBuildDomainSystemPromptWithDependencies(t *testing.T) {
	config := &scenario.ScenarioConfig{
		Name: "test-scenario",
		Domains: []scenario.DomainSpec{
			{
				Name: "aws",
				CLI:  "wetwire-aws",
			},
			{
				Name:      "gitlab",
				CLI:       "wetwire-gitlab",
				DependsOn: []string{"aws"},
			},
		},
	}

	t.Run("without dependency outputs", func(t *testing.T) {
		prompt := buildDomainSystemPrompt(config, nil)

		// Should contain basic sections
		if !strings.Contains(prompt, "ABSOLUTE REQUIREMENTS") {
			t.Error("Expected prompt to contain ABSOLUTE REQUIREMENTS section")
		}
		if !strings.Contains(prompt, "Available MCP Tools") {
			t.Error("Expected prompt to contain Available MCP Tools section")
		}

		// Should not contain dependency outputs section
		if strings.Contains(prompt, "Available Dependency Outputs") {
			t.Error("Expected prompt to NOT contain Available Dependency Outputs section")
		}
	})

	t.Run("with dependency outputs", func(t *testing.T) {
		dependencyOutputs := NewOutputManifest()
		dependencyOutputs.AddDomainOutput("aws", &DomainOutput{
			Resources: map[string]ResourceOutput{
				"bucket": {
					Type: "aws_s3_bucket",
					Outputs: map[string]interface{}{
						"name": "my-artifacts-bucket",
						"arn":  "arn:aws:s3:::my-artifacts-bucket",
					},
				},
			},
		})

		prompt := buildDomainSystemPrompt(config, dependencyOutputs)

		// Should contain dependency outputs section
		if !strings.Contains(prompt, "Available Dependency Outputs") {
			t.Error("Expected prompt to contain Available Dependency Outputs section")
		}

		// Should contain the actual output values
		if !strings.Contains(prompt, "my-artifacts-bucket") {
			t.Error("Expected prompt to contain bucket name")
		}
		if !strings.Contains(prompt, "aws_s3_bucket") {
			t.Error("Expected prompt to contain resource type")
		}

		// Should contain reference syntax hint
		if !strings.Contains(prompt, "${domain.resource.outputs.field}") {
			t.Error("Expected prompt to contain reference syntax")
		}
	})

	t.Run("with empty dependency outputs", func(t *testing.T) {
		dependencyOutputs := NewOutputManifest()

		prompt := buildDomainSystemPrompt(config, dependencyOutputs)

		// Should not contain dependency outputs section when manifest is empty
		if strings.Contains(prompt, "Available Dependency Outputs") {
			t.Error("Expected prompt to NOT contain Available Dependency Outputs section for empty manifest")
		}
	})
}
