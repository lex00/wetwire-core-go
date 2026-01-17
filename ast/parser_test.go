package ast

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	// Create a temp file with valid Go code
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

var X = 1
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("parses valid file", func(t *testing.T) {
		file, fset, err := ParseFile(testFile)
		if err != nil {
			t.Fatalf("ParseFile() error = %v", err)
		}
		if file == nil {
			t.Fatal("ParseFile() returned nil file")
		}
		if fset == nil {
			t.Fatal("ParseFile() returned nil fset")
		}
		if file.Name.Name != "test" {
			t.Errorf("ParseFile() package name = %q, want %q", file.Name.Name, "test")
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, _, err := ParseFile("/nonexistent/file.go")
		if err == nil {
			t.Error("ParseFile() expected error for non-existent file")
		}
	})

	t.Run("returns error for invalid Go code", func(t *testing.T) {
		invalidFile := filepath.Join(tmpDir, "invalid.go")
		if err := os.WriteFile(invalidFile, []byte("not valid go code"), 0644); err != nil {
			t.Fatal(err)
		}
		_, _, err := ParseFile(invalidFile)
		if err == nil {
			t.Error("ParseFile() expected error for invalid Go code")
		}
	})
}

func TestParseDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"a.go":      "package test\nvar A = 1",
		"b.go":      "package test\nvar B = 2",
		"a_test.go": "package test\nvar T = 3",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("parses all files by default", func(t *testing.T) {
		files, fset, err := ParseDir(tmpDir, ParseOptions{})
		if err != nil {
			t.Fatalf("ParseDir() error = %v", err)
		}
		if fset == nil {
			t.Fatal("ParseDir() returned nil fset")
		}
		if len(files) != 3 {
			t.Errorf("ParseDir() returned %d files, want 3", len(files))
		}
	})

	t.Run("skips test files when SkipTests is true", func(t *testing.T) {
		files, _, err := ParseDir(tmpDir, ParseOptions{SkipTests: true})
		if err != nil {
			t.Fatalf("ParseDir() error = %v", err)
		}
		if len(files) != 2 {
			t.Errorf("ParseDir() with SkipTests returned %d files, want 2", len(files))
		}
	})

	t.Run("returns error for non-existent directory", func(t *testing.T) {
		_, _, err := ParseDir("/nonexistent/dir", ParseOptions{})
		if err == nil {
			t.Error("ParseDir() expected error for non-existent directory")
		}
	})
}

func TestWalkGoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	subDir := filepath.Join(tmpDir, "sub")
	vendorDir := filepath.Join(tmpDir, "vendor")
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	for _, dir := range []string{subDir, vendorDir, hiddenDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Create test files
	testFiles := []string{
		filepath.Join(tmpDir, "root.go"),
		filepath.Join(tmpDir, "root_test.go"),
		filepath.Join(subDir, "sub.go"),
		filepath.Join(vendorDir, "vendor.go"),
		filepath.Join(hiddenDir, "hidden.go"),
	}
	for _, f := range testFiles {
		if err := os.WriteFile(f, []byte("package test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("walks all files by default", func(t *testing.T) {
		var visited []string
		err := WalkGoFiles(tmpDir, ParseOptions{}, func(path string) error {
			visited = append(visited, path)
			return nil
		})
		if err != nil {
			t.Fatalf("WalkGoFiles() error = %v", err)
		}
		if len(visited) != 5 {
			t.Errorf("WalkGoFiles() visited %d files, want 5", len(visited))
		}
	})

	t.Run("skips test files when SkipTests is true", func(t *testing.T) {
		var visited []string
		err := WalkGoFiles(tmpDir, ParseOptions{SkipTests: true}, func(path string) error {
			visited = append(visited, path)
			return nil
		})
		if err != nil {
			t.Fatalf("WalkGoFiles() error = %v", err)
		}
		for _, v := range visited {
			if filepath.Base(v) == "root_test.go" {
				t.Error("WalkGoFiles() should skip test files")
			}
		}
	})

	t.Run("skips vendor when SkipVendor is true", func(t *testing.T) {
		var visited []string
		err := WalkGoFiles(tmpDir, ParseOptions{SkipVendor: true}, func(path string) error {
			visited = append(visited, path)
			return nil
		})
		if err != nil {
			t.Fatalf("WalkGoFiles() error = %v", err)
		}
		for _, v := range visited {
			if filepath.Base(filepath.Dir(v)) == "vendor" {
				t.Error("WalkGoFiles() should skip vendor directory")
			}
		}
	})

	t.Run("skips hidden dirs when SkipHidden is true", func(t *testing.T) {
		var visited []string
		err := WalkGoFiles(tmpDir, ParseOptions{SkipHidden: true}, func(path string) error {
			visited = append(visited, path)
			return nil
		})
		if err != nil {
			t.Fatalf("WalkGoFiles() error = %v", err)
		}
		for _, v := range visited {
			if filepath.Base(filepath.Dir(v)) == ".hidden" {
				t.Error("WalkGoFiles() should skip hidden directories")
			}
		}
	})

	t.Run("skips excluded dirs", func(t *testing.T) {
		var visited []string
		err := WalkGoFiles(tmpDir, ParseOptions{ExcludeDirs: []string{"sub"}}, func(path string) error {
			visited = append(visited, path)
			return nil
		})
		if err != nil {
			t.Fatalf("WalkGoFiles() error = %v", err)
		}
		for _, v := range visited {
			if filepath.Base(filepath.Dir(v)) == "sub" {
				t.Error("WalkGoFiles() should skip excluded directories")
			}
		}
	})
}
