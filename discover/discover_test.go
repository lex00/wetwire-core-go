package discover

import (
	"os"
	"path/filepath"
	"testing"
)

// mockMatcher returns resource type for known wetwire patterns
func mockMatcher(pkgName, typeName string, imports map[string]string) (string, bool) {
	// Check for qualified names (e.g., schema.NodeType)
	if pkgName == "schema" {
		switch typeName {
		case "NodeType":
			return "schema.NodeType", true
		case "RelationshipType":
			return "schema.RelationshipType", true
		}
	}
	// Check imports to resolve packages
	for alias, path := range imports {
		if alias == pkgName || filepath.Base(path) == pkgName {
			if typeName == "S3Bucket" {
				return "aws.S3Bucket", true
			}
		}
	}
	return "", false
}

func TestDiscoverFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package test

import "github.com/example/schema"

var Person = &schema.NodeType{
	Label: "Person",
}

var WorksFor = &schema.RelationshipType{
	Label: "WORKS_FOR",
}

var normalVar = "not a resource"
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("discovers resources in file", func(t *testing.T) {
		result, err := DiscoverFile(testFile, mockMatcher)
		if err != nil {
			t.Fatalf("DiscoverFile() error = %v", err)
		}
		if len(result.Resources) != 2 {
			t.Errorf("DiscoverFile() found %d resources, want 2", len(result.Resources))
		}
		// Verify resource details
		found := map[string]bool{}
		for _, r := range result.Resources {
			found[r.Name] = true
		}
		if !found["Person"] {
			t.Error("DiscoverFile() missing Person resource")
		}
		if !found["WorksFor"] {
			t.Error("DiscoverFile() missing WorksFor resource")
		}
	})

	t.Run("tracks all variables", func(t *testing.T) {
		result, _ := DiscoverFile(testFile, mockMatcher)
		if len(result.AllVars) != 3 {
			t.Errorf("AllVars length = %d, want 3", len(result.AllVars))
		}
		if !result.AllVars["normalVar"] {
			t.Error("AllVars missing normalVar")
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := DiscoverFile("/nonexistent/file.go", mockMatcher)
		if err == nil {
			t.Error("DiscoverFile() expected error for non-existent file")
		}
	})
}

func TestDiscoverDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"nodes.go": `package test

import "github.com/example/schema"

var Person = &schema.NodeType{Label: "Person"}
`,
		"relationships.go": `package test

import "github.com/example/schema"

var WorksFor = &schema.RelationshipType{Label: "WORKS_FOR"}
`,
		"other.go": `package test

var config = "not a resource"
`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("discovers resources across directory", func(t *testing.T) {
		result, err := DiscoverDir(tmpDir, mockMatcher)
		if err != nil {
			t.Fatalf("DiscoverDir() error = %v", err)
		}
		if len(result.Resources) != 2 {
			t.Errorf("DiscoverDir() found %d resources, want 2", len(result.Resources))
		}
	})

	t.Run("returns error for non-existent directory", func(t *testing.T) {
		_, err := DiscoverDir("/nonexistent/dir", mockMatcher)
		if err == nil {
			t.Error("DiscoverDir() expected error for non-existent directory")
		}
	})
}

func TestDiscover(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test files at different levels
	files := map[string]string{
		filepath.Join(tmpDir, "root.go"): `package test

import "github.com/example/schema"

var RootNode = &schema.NodeType{Label: "Root"}
`,
		filepath.Join(subDir, "nested.go"): `package test

import "github.com/example/schema"

var NestedNode = &schema.NodeType{Label: "Nested"}
`,
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	opts := DiscoverOptions{
		Packages:    []string{tmpDir},
		TypeMatcher: mockMatcher,
	}

	t.Run("discovers resources recursively", func(t *testing.T) {
		result, err := Discover(opts)
		if err != nil {
			t.Fatalf("Discover() error = %v", err)
		}
		if len(result.Resources) != 2 {
			t.Errorf("Discover() found %d resources, want 2", len(result.Resources))
		}
	})
}

func TestDiscoverWithNilMatcher(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test
var x = 1
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// With nil matcher, should find no resources but still track vars
	result, err := DiscoverFile(testFile, nil)
	if err != nil {
		t.Fatalf("DiscoverFile() with nil matcher error = %v", err)
	}
	if len(result.Resources) != 0 {
		t.Errorf("DiscoverFile() with nil matcher found %d resources, want 0", len(result.Resources))
	}
	if !result.AllVars["x"] {
		t.Error("AllVars should still track x")
	}
}
