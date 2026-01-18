package domain

import (
	"context"
	"testing"
)

func TestNewCrossDomainContext(t *testing.T) {
	ctx := NewCrossDomainContext()
	if ctx == nil {
		t.Fatal("NewCrossDomainContext returned nil")
	}
	if ctx.Dependencies == nil {
		t.Error("Dependencies map is nil")
	}
	if len(ctx.Dependencies) != 0 {
		t.Error("Expected empty Dependencies map")
	}
}

func TestAddDomainOutputs(t *testing.T) {
	ctx := NewCrossDomainContext()

	outputs := &DomainOutputs{
		Resources: map[string]*ResourceOutputs{
			"bucket1": {
				Type: "aws_s3_bucket",
				Outputs: map[string]interface{}{
					"name": "my-bucket",
					"arn":  "arn:aws:s3:::my-bucket",
				},
			},
		},
	}

	ctx.AddDomainOutputs("aws", outputs)

	if len(ctx.Dependencies) != 1 {
		t.Errorf("Expected 1 domain, got %d", len(ctx.Dependencies))
	}

	retrieved := ctx.GetDomainOutputs("aws")
	if retrieved == nil {
		t.Fatal("GetDomainOutputs returned nil")
	}
	if len(retrieved.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(retrieved.Resources))
	}
}

func TestGetDomainOutputs(t *testing.T) {
	ctx := NewCrossDomainContext()

	// Test getting non-existent domain
	retrieved := ctx.GetDomainOutputs("nonexistent")
	if retrieved != nil {
		t.Error("Expected nil for non-existent domain")
	}

	// Add a domain and retrieve it
	outputs := &DomainOutputs{
		Resources: make(map[string]*ResourceOutputs),
	}
	ctx.AddDomainOutputs("test", outputs)

	retrieved = ctx.GetDomainOutputs("test")
	if retrieved == nil {
		t.Error("Expected non-nil for existing domain")
	}
}

func TestGetResourceOutput(t *testing.T) {
	ctx := NewCrossDomainContext()

	outputs := &DomainOutputs{
		Resources: map[string]*ResourceOutputs{
			"bucket": {
				Type: "aws_s3_bucket",
				Outputs: map[string]interface{}{
					"name": "test-bucket",
					"arn":  "arn:aws:s3:::test-bucket",
				},
			},
		},
	}
	ctx.AddDomainOutputs("aws", outputs)

	// Test successful retrieval
	value := ctx.GetResourceOutput("aws", "bucket", "name")
	if value != "test-bucket" {
		t.Errorf("Expected 'test-bucket', got %v", value)
	}

	// Test non-existent domain
	value = ctx.GetResourceOutput("gitlab", "bucket", "name")
	if value != nil {
		t.Error("Expected nil for non-existent domain")
	}

	// Test non-existent resource
	value = ctx.GetResourceOutput("aws", "nonexistent", "name")
	if value != nil {
		t.Error("Expected nil for non-existent resource")
	}

	// Test non-existent output key
	value = ctx.GetResourceOutput("aws", "bucket", "nonexistent")
	if value != nil {
		t.Error("Expected nil for non-existent output key")
	}
}

func TestHasDependency(t *testing.T) {
	ctx := NewCrossDomainContext()

	// Test before adding
	if ctx.HasDependency("aws") {
		t.Error("Expected false for non-existent domain")
	}

	// Add a domain
	ctx.AddDomainOutputs("aws", &DomainOutputs{
		Resources: make(map[string]*ResourceOutputs),
	})

	// Test after adding
	if !ctx.HasDependency("aws") {
		t.Error("Expected true for existing domain")
	}
}

func TestDomainNames(t *testing.T) {
	ctx := NewCrossDomainContext()

	// Test empty
	names := ctx.DomainNames()
	if len(names) != 0 {
		t.Errorf("Expected 0 names, got %d", len(names))
	}

	// Add domains
	ctx.AddDomainOutputs("aws", &DomainOutputs{Resources: make(map[string]*ResourceOutputs)})
	ctx.AddDomainOutputs("gitlab", &DomainOutputs{Resources: make(map[string]*ResourceOutputs)})

	names = ctx.DomainNames()
	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}

	// Check both names are present (order not guaranteed)
	hasAWS := false
	hasGitlab := false
	for _, name := range names {
		if name == "aws" {
			hasAWS = true
		}
		if name == "gitlab" {
			hasGitlab = true
		}
	}
	if !hasAWS || !hasGitlab {
		t.Errorf("Expected both 'aws' and 'gitlab' in names, got %v", names)
	}
}

func TestContextWithCrossDomain(t *testing.T) {
	crossDomain := NewCrossDomainContext()
	crossDomain.AddDomainOutputs("aws", &DomainOutputs{
		Resources: map[string]*ResourceOutputs{
			"bucket": {
				Type: "aws_s3_bucket",
				Outputs: map[string]interface{}{
					"name": "test-bucket",
				},
			},
		},
	})

	// Test NewContextWithCrossDomain
	ctx := NewContextWithCrossDomain(context.Background(), "/tmp/work", true, crossDomain)
	if ctx.CrossDomain == nil {
		t.Fatal("CrossDomain is nil")
	}
	if ctx.WorkDir != "/tmp/work" {
		t.Errorf("Expected WorkDir '/tmp/work', got '%s'", ctx.WorkDir)
	}
	if !ctx.Verbose {
		t.Error("Expected Verbose to be true")
	}

	// Verify cross-domain data is accessible
	value := ctx.CrossDomain.GetResourceOutput("aws", "bucket", "name")
	if value != "test-bucket" {
		t.Errorf("Expected 'test-bucket', got %v", value)
	}
}

func TestContext_WithCrossDomain(t *testing.T) {
	// Create context without cross-domain
	ctx := NewContext(context.Background(), "/tmp/work")
	if ctx.CrossDomain != nil {
		t.Error("Expected CrossDomain to be nil initially")
	}

	// Add cross-domain context
	crossDomain := NewCrossDomainContext()
	crossDomain.AddDomainOutputs("aws", &DomainOutputs{
		Resources: map[string]*ResourceOutputs{
			"bucket": {
				Type: "aws_s3_bucket",
				Outputs: map[string]interface{}{
					"name": "test-bucket",
				},
			},
		},
	})

	newCtx := ctx.WithCrossDomain(crossDomain)

	// Verify original context is unchanged
	if ctx.CrossDomain != nil {
		t.Error("Original context should not be modified")
	}

	// Verify new context has cross-domain data
	if newCtx.CrossDomain == nil {
		t.Fatal("New context should have CrossDomain")
	}

	// Verify other fields are preserved
	if newCtx.WorkDir != ctx.WorkDir {
		t.Error("WorkDir should be preserved")
	}
	if newCtx.Verbose != ctx.Verbose {
		t.Error("Verbose should be preserved")
	}

	// Verify cross-domain data is accessible
	value := newCtx.CrossDomain.GetResourceOutput("aws", "bucket", "name")
	if value != "test-bucket" {
		t.Errorf("Expected 'test-bucket', got %v", value)
	}
}

func TestCrossDomainContext_NilSafety(t *testing.T) {
	// Test nil Dependencies map
	ctx := &CrossDomainContext{}

	// These should not panic
	_ = ctx.GetDomainOutputs("aws")
	_ = ctx.GetResourceOutput("aws", "bucket", "name")
	_ = ctx.HasDependency("aws")
	_ = ctx.DomainNames()

	// AddDomainOutputs should work on nil map
	ctx.AddDomainOutputs("aws", &DomainOutputs{})
	if ctx.Dependencies == nil {
		t.Error("AddDomainOutputs should initialize Dependencies map")
	}
}

func TestCrossDomainContext_NilResourceOutputs(t *testing.T) {
	ctx := NewCrossDomainContext()

	// Add domain with nil resource outputs
	ctx.AddDomainOutputs("aws", &DomainOutputs{
		Resources: map[string]*ResourceOutputs{
			"bucket": nil, // nil resource outputs
		},
	})

	// Should return nil without panicking
	value := ctx.GetResourceOutput("aws", "bucket", "name")
	if value != nil {
		t.Error("Expected nil for nil resource outputs")
	}
}

func TestCrossDomainContext_NilOutputsMap(t *testing.T) {
	ctx := NewCrossDomainContext()

	// Add domain with nil Outputs map in resource
	ctx.AddDomainOutputs("aws", &DomainOutputs{
		Resources: map[string]*ResourceOutputs{
			"bucket": {
				Type:    "aws_s3_bucket",
				Outputs: nil, // nil outputs map
			},
		},
	})

	// Should return nil without panicking
	value := ctx.GetResourceOutput("aws", "bucket", "name")
	if value != nil {
		t.Error("Expected nil for nil outputs map")
	}
}
