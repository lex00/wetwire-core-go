package runner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewOutputManifest(t *testing.T) {
	manifest := NewOutputManifest()
	if manifest == nil {
		t.Fatal("NewOutputManifest returned nil")
	}
	if manifest.Domains == nil {
		t.Error("Domains map is nil")
	}
	if len(manifest.Domains) != 0 {
		t.Error("Expected empty Domains map")
	}
}

func TestAddDomainOutput(t *testing.T) {
	manifest := NewOutputManifest()

	domainOutput := &DomainOutput{
		Resources: map[string]ResourceOutput{
			"bucket1": {
				Type: "aws_s3_bucket",
				Outputs: map[string]interface{}{
					"name": "my-bucket",
					"arn":  "arn:aws:s3:::my-bucket",
				},
			},
		},
	}

	manifest.AddDomainOutput("aws", domainOutput)

	if len(manifest.Domains) != 1 {
		t.Errorf("Expected 1 domain, got %d", len(manifest.Domains))
	}

	retrieved := manifest.GetDomainOutput("aws")
	if retrieved == nil {
		t.Fatal("GetDomainOutput returned nil")
	}
	if len(retrieved.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(retrieved.Resources))
	}
}

func TestGetDomainOutput(t *testing.T) {
	manifest := NewOutputManifest()

	// Test getting non-existent domain
	retrieved := manifest.GetDomainOutput("nonexistent")
	if retrieved != nil {
		t.Error("Expected nil for non-existent domain")
	}

	// Add a domain and retrieve it
	domainOutput := &DomainOutput{
		Resources: make(map[string]ResourceOutput),
	}
	manifest.AddDomainOutput("test", domainOutput)

	retrieved = manifest.GetDomainOutput("test")
	if retrieved == nil {
		t.Error("Expected non-nil for existing domain")
	}
}

func TestSaveToFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "output-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	manifest := NewOutputManifest()
	manifest.AddDomainOutput("aws", &DomainOutput{
		Resources: map[string]ResourceOutput{
			"bucket": {
				Type: "aws_s3_bucket",
				Outputs: map[string]interface{}{
					"name": "test-bucket",
				},
			},
		},
	})

	outputPath := filepath.Join(tmpDir, "outputs.json")
	err = manifest.SaveToFile(outputPath)
	if err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Verify content is valid JSON
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var parsed OutputManifest
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Errorf("Output file is not valid JSON: %v", err)
	}

	// Verify content matches
	if len(parsed.Domains) != 1 {
		t.Errorf("Expected 1 domain in loaded manifest, got %d", len(parsed.Domains))
	}
}

func TestLoadFromFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "output-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test manifest
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
		},
	})

	// Save it
	outputPath := filepath.Join(tmpDir, "outputs.json")
	if err := original.SaveToFile(outputPath); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	// Load it back
	loaded, err := LoadFromFile(outputPath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Verify content
	if len(loaded.Domains) != 1 {
		t.Errorf("Expected 1 domain, got %d", len(loaded.Domains))
	}

	awsDomain := loaded.GetDomainOutput("aws")
	if awsDomain == nil {
		t.Fatal("AWS domain not found in loaded manifest")
	}

	if len(awsDomain.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(awsDomain.Resources))
	}

	bucket, ok := awsDomain.Resources["bucket"]
	if !ok {
		t.Fatal("bucket resource not found")
	}

	if bucket.Type != "aws_s3_bucket" {
		t.Errorf("Expected type aws_s3_bucket, got %s", bucket.Type)
	}

	if bucket.Outputs["name"] != "test-bucket" {
		t.Errorf("Expected name test-bucket, got %v", bucket.Outputs["name"])
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "output-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create invalid JSON file
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	err = os.WriteFile(invalidPath, []byte("not valid json"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Try to load it
	_, err = LoadFromFile(invalidPath)
	if err == nil {
		t.Error("Expected error when loading invalid JSON")
	}
}

func TestLoadFromFile_MissingFile(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/outputs.json")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}
}

func TestCaptureOutputsFromFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "output-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("captures JSON files", func(t *testing.T) {
		// Create a test JSON file
		testData := map[string]interface{}{
			"BucketName": "my-test-bucket",
			"BucketArn":  "arn:aws:s3:::my-test-bucket",
		}
		jsonBytes, _ := json.Marshal(testData)

		outputDir := filepath.Join(tmpDir, "output")
		_ = os.MkdirAll(outputDir, 0755)
		_ = os.WriteFile(filepath.Join(outputDir, "stack.json"), jsonBytes, 0644)

		// Capture outputs
		domainOutput, err := CaptureOutputsFromFiles(tmpDir, "aws", []string{"output/*.json"})
		if err != nil {
			t.Fatalf("CaptureOutputsFromFiles failed: %v", err)
		}

		// Verify captured output
		if len(domainOutput.Resources) != 1 {
			t.Errorf("Expected 1 resource, got %d", len(domainOutput.Resources))
		}

		// Check that the resource was captured
		found := false
		for name, resource := range domainOutput.Resources {
			if resource.Type == "aws_resource" {
				found = true
				if resource.Outputs["BucketName"] != "my-test-bucket" {
					t.Errorf("Expected BucketName my-test-bucket, got %v", resource.Outputs["BucketName"])
				}
			}
			t.Logf("Captured resource: %s with type %s", name, resource.Type)
		}
		if !found {
			t.Error("Expected to find aws_resource type")
		}
	})

	t.Run("handles non-JSON files gracefully", func(t *testing.T) {
		// Create a non-JSON file
		textDir := filepath.Join(tmpDir, "text")
		_ = os.MkdirAll(textDir, 0755)
		_ = os.WriteFile(filepath.Join(textDir, "readme.txt"), []byte("not json"), 0644)

		// This should not error, just not capture the file
		domainOutput, err := CaptureOutputsFromFiles(tmpDir, "test", []string{"text/*.txt"})
		if err != nil {
			t.Fatalf("CaptureOutputsFromFiles should not error on non-JSON: %v", err)
		}

		// Should have empty resources since the file wasn't JSON
		if len(domainOutput.Resources) != 0 {
			t.Errorf("Expected 0 resources for non-JSON file, got %d", len(domainOutput.Resources))
		}
	})

	t.Run("handles no matching files", func(t *testing.T) {
		domainOutput, err := CaptureOutputsFromFiles(tmpDir, "test", []string{"nonexistent/*.json"})
		if err != nil {
			t.Fatalf("CaptureOutputsFromFiles should not error on no matches: %v", err)
		}

		if len(domainOutput.Resources) != 0 {
			t.Errorf("Expected 0 resources when no files match, got %d", len(domainOutput.Resources))
		}
	})
}

func TestDomainOutputJSON(t *testing.T) {
	// Test that DomainOutput can be marshaled and unmarshaled
	original := &DomainOutput{
		Resources: map[string]ResourceOutput{
			"resource1": {
				Type: "test_type",
				Outputs: map[string]interface{}{
					"key1": "value1",
					"key2": 42,
				},
			},
		},
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal DomainOutput: %v", err)
	}

	// Unmarshal
	var parsed DomainOutput
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal DomainOutput: %v", err)
	}

	// Verify
	if len(parsed.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(parsed.Resources))
	}

	resource, ok := parsed.Resources["resource1"]
	if !ok {
		t.Fatal("resource1 not found after unmarshal")
	}

	if resource.Type != "test_type" {
		t.Errorf("Expected type test_type, got %s", resource.Type)
	}
}
