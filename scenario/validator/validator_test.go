package validator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lex00/wetwire-core-go/scenario"
)

func TestValidateResourceCounts(t *testing.T) {
	// Create temp directory with test files
	tempDir := t.TempDir()

	// Create k8s files
	k8sDir := filepath.Join(tempDir, "results", "k8s")
	_ = os.MkdirAll(k8sDir, 0755)
	_ = os.WriteFile(filepath.Join(k8sDir, "namespace.yaml"), []byte("apiVersion: v1\nkind: Namespace"), 0644)
	_ = os.WriteFile(filepath.Join(k8sDir, "deployment.yaml"), []byte("apiVersion: apps/v1\nkind: Deployment"), 0644)

	config := &scenario.ScenarioConfig{
		Validation: map[string]scenario.ValidationRules{
			"k8s": {
				Resources: &scenario.CountConstraint{Min: 2, Max: 5},
			},
		},
	}

	v := New(config, tempDir, filepath.Join(tempDir, "results"))
	results, err := v.ValidateResourceCounts()
	if err != nil {
		t.Fatalf("ValidateResourceCounts failed: %v", err)
	}

	if result, ok := results["k8s"]; ok {
		if !result.Passed {
			t.Errorf("Expected k8s validation to pass, got: %s", result.Error)
		}
		if result.Found != 2 {
			t.Errorf("Expected 2 files found, got %d", result.Found)
		}
	} else {
		t.Error("Expected k8s result in validation results")
	}
}

func TestValidateCrossRefs(t *testing.T) {
	// Create temp directory with test files
	tempDir := t.TempDir()

	// Create honeycomb file with k8s references
	resultsDir := filepath.Join(tempDir, "results")
	_ = os.MkdirAll(resultsDir, 0755)
	honeycombContent := `{
		"dataset": "api-service",
		"filters": [
			{"field": "k8s.service.name", "value": "my-service"},
			{"field": "k8s.namespace.name", "value": "my-namespace"}
		]
	}`
	_ = os.WriteFile(filepath.Join(resultsDir, "query-latency.json"), []byte(honeycombContent), 0644)

	config := &scenario.ScenarioConfig{
		CrossDomain: []scenario.CrossDomainSpec{
			{
				From: "k8s",
				To:   "honeycomb",
				Type: "artifact_reference",
				Validation: scenario.CrossDomainValidation{
					RequiredRefs: []string{"${k8s.service_name}", "${k8s.namespace}"},
				},
			},
		},
	}

	v := New(config, tempDir, resultsDir)
	results, err := v.ValidateCrossRefs()
	if err != nil {
		t.Fatalf("ValidateCrossRefs failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 cross-ref result, got %d", len(results))
	}

	result := results[0]
	if !result.Passed {
		t.Errorf("Expected cross-ref validation to pass, missing: %v", result.MissingRefs)
	}
}

func TestCompareExpected(t *testing.T) {
	// Create temp directory with expected and results
	tempDir := t.TempDir()

	// Create expected file
	expectedDir := filepath.Join(tempDir, "expected", "k8s")
	_ = os.MkdirAll(expectedDir, 0755)
	expectedContent := `apiVersion: v1
kind: Namespace
metadata:
  name: test-ns`
	_ = os.WriteFile(filepath.Join(expectedDir, "namespace.yaml"), []byte(expectedContent), 0644)

	// Create results file
	resultsDir := filepath.Join(tempDir, "results")
	_ = os.MkdirAll(resultsDir, 0755)
	generatedContent := `apiVersion: v1
kind: Namespace
metadata:
  name: generated-ns
  labels:
    app: test`
	_ = os.WriteFile(filepath.Join(resultsDir, "namespace.yaml"), []byte(generatedContent), 0644)

	config := &scenario.ScenarioConfig{}

	v := New(config, tempDir, resultsDir)
	results, err := v.CompareExpected()
	if err != nil {
		t.Fatalf("CompareExpected failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 comparison result, got %d", len(results))
	}

	result := results[0]
	if result.Missing {
		t.Error("Expected file should not be marked as missing")
	}
	if !result.Passed {
		t.Error("Expected comparison to pass (file exists)")
	}
}

func TestFormatReport(t *testing.T) {
	report := &ValidationReport{
		Passed: true,
		ResourceCounts: map[string]ResourceCountResult{
			"k8s": {
				Domain:       "k8s",
				Passed:       true,
				Found:        4,
				Min:          3,
				ResourceType: "resources",
			},
		},
		CrossDomainRefs: []CrossRefResult{
			{
				From:         "k8s",
				To:           "honeycomb",
				Passed:       true,
				RequiredRefs: []string{"${k8s.namespace}"},
				FoundRefs:    []string{"${k8s.namespace}"},
			},
		},
		Score: 12,
	}

	output := FormatReport(report)
	if output == "" {
		t.Error("FormatReport returned empty string")
	}
	if !contains(output, "PASSED") {
		t.Error("Expected report to contain PASSED")
	}
	if !contains(output, "12/12") {
		t.Error("Expected report to contain score 12/12")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
