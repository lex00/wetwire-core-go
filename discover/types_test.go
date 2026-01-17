package discover

import "testing"

func TestDiscoveredResource(t *testing.T) {
	resource := DiscoveredResource{
		Name:         "MyBucket",
		Type:         "aws.S3Bucket",
		File:         "storage.go",
		Line:         10,
		Dependencies: []string{"OtherResource", "AnotherResource"},
	}

	if resource.Name != "MyBucket" {
		t.Errorf("Name = %q, want %q", resource.Name, "MyBucket")
	}
	if resource.Type != "aws.S3Bucket" {
		t.Errorf("Type = %q, want %q", resource.Type, "aws.S3Bucket")
	}
	if resource.File != "storage.go" {
		t.Errorf("File = %q, want %q", resource.File, "storage.go")
	}
	if resource.Line != 10 {
		t.Errorf("Line = %d, want %d", resource.Line, 10)
	}
	if len(resource.Dependencies) != 2 {
		t.Errorf("Dependencies length = %d, want 2", len(resource.Dependencies))
	}
}

func TestDiscoverResult(t *testing.T) {
	result := &DiscoverResult{
		Resources: []DiscoveredResource{
			{Name: "Resource1", Type: "Type1"},
			{Name: "Resource2", Type: "Type2"},
		},
		AllVars: map[string]bool{
			"Resource1": true,
			"Resource2": true,
			"OtherVar":  true,
		},
	}

	if len(result.Resources) != 2 {
		t.Errorf("Resources length = %d, want 2", len(result.Resources))
	}
	if len(result.AllVars) != 3 {
		t.Errorf("AllVars length = %d, want 3", len(result.AllVars))
	}
}

func TestDiscoverResultMerge(t *testing.T) {
	result1 := &DiscoverResult{
		Resources: []DiscoveredResource{
			{Name: "Resource1", Type: "Type1"},
		},
		AllVars: map[string]bool{"Resource1": true},
	}

	result2 := &DiscoverResult{
		Resources: []DiscoveredResource{
			{Name: "Resource2", Type: "Type2"},
		},
		AllVars: map[string]bool{"Resource2": true},
	}

	result1.Merge(result2)

	if len(result1.Resources) != 2 {
		t.Errorf("merged Resources length = %d, want 2", len(result1.Resources))
	}
	if len(result1.AllVars) != 2 {
		t.Errorf("merged AllVars length = %d, want 2", len(result1.AllVars))
	}
	if !result1.AllVars["Resource2"] {
		t.Error("merged AllVars missing Resource2")
	}
}

func TestDiscoverOptions(t *testing.T) {
	matcher := func(pkgName, typeName string, imports map[string]string) (string, bool) {
		if typeName == "S3Bucket" {
			return "aws.S3Bucket", true
		}
		return "", false
	}

	opts := DiscoverOptions{
		Packages:    []string{"./..."},
		Verbose:     true,
		TypeMatcher: matcher,
	}

	if len(opts.Packages) != 1 {
		t.Errorf("Packages length = %d, want 1", len(opts.Packages))
	}
	if !opts.Verbose {
		t.Error("Verbose = false, want true")
	}
	if opts.TypeMatcher == nil {
		t.Error("TypeMatcher = nil, want non-nil")
	}

	// Test matcher
	resourceType, ok := opts.TypeMatcher("", "S3Bucket", nil)
	if !ok || resourceType != "aws.S3Bucket" {
		t.Errorf("TypeMatcher(S3Bucket) = %q, %v, want %q, true", resourceType, ok, "aws.S3Bucket")
	}
}
