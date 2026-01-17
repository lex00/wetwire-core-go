package discover

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalkDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	subDir := filepath.Join(tmpDir, "sub")
	vendorDir := filepath.Join(tmpDir, "vendor")
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	testDataDir := filepath.Join(tmpDir, "testdata")
	for _, dir := range []string{subDir, vendorDir, hiddenDir, testDataDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Create test files
	files := []string{
		filepath.Join(tmpDir, "root.go"),
		filepath.Join(tmpDir, "root_test.go"),
		filepath.Join(subDir, "sub.go"),
		filepath.Join(vendorDir, "vendor.go"),
		filepath.Join(hiddenDir, "hidden.go"),
		filepath.Join(testDataDir, "testdata.go"),
	}
	for _, f := range files {
		if err := os.WriteFile(f, []byte("package test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("walks all files with default options", func(t *testing.T) {
		var visited []string
		err := WalkDir(tmpDir, WalkOptions{}, func(path string) error {
			visited = append(visited, path)
			return nil
		})
		if err != nil {
			t.Fatalf("WalkDir() error = %v", err)
		}
		if len(visited) != 6 {
			t.Errorf("WalkDir() visited %d files, want 6", len(visited))
		}
	})

	t.Run("skips test files", func(t *testing.T) {
		var visited []string
		err := WalkDir(tmpDir, WalkOptions{SkipTests: true}, func(path string) error {
			visited = append(visited, path)
			return nil
		})
		if err != nil {
			t.Fatalf("WalkDir() error = %v", err)
		}
		for _, v := range visited {
			if filepath.Base(v) == "root_test.go" {
				t.Error("WalkDir() should skip test files")
			}
		}
	})

	t.Run("skips vendor directory", func(t *testing.T) {
		var visited []string
		err := WalkDir(tmpDir, WalkOptions{SkipVendor: true}, func(path string) error {
			visited = append(visited, path)
			return nil
		})
		if err != nil {
			t.Fatalf("WalkDir() error = %v", err)
		}
		for _, v := range visited {
			if filepath.Base(filepath.Dir(v)) == "vendor" {
				t.Error("WalkDir() should skip vendor directory")
			}
		}
	})

	t.Run("skips hidden directories", func(t *testing.T) {
		var visited []string
		err := WalkDir(tmpDir, WalkOptions{SkipHidden: true}, func(path string) error {
			visited = append(visited, path)
			return nil
		})
		if err != nil {
			t.Fatalf("WalkDir() error = %v", err)
		}
		for _, v := range visited {
			if filepath.Base(filepath.Dir(v)) == ".hidden" {
				t.Error("WalkDir() should skip hidden directories")
			}
		}
	})

	t.Run("skips testdata directory", func(t *testing.T) {
		var visited []string
		err := WalkDir(tmpDir, WalkOptions{SkipTestdata: true}, func(path string) error {
			visited = append(visited, path)
			return nil
		})
		if err != nil {
			t.Fatalf("WalkDir() error = %v", err)
		}
		for _, v := range visited {
			if filepath.Base(filepath.Dir(v)) == "testdata" {
				t.Error("WalkDir() should skip testdata directory")
			}
		}
	})

	t.Run("skips excluded directories", func(t *testing.T) {
		var visited []string
		err := WalkDir(tmpDir, WalkOptions{ExcludeDirs: []string{"sub"}}, func(path string) error {
			visited = append(visited, path)
			return nil
		})
		if err != nil {
			t.Fatalf("WalkDir() error = %v", err)
		}
		for _, v := range visited {
			if filepath.Base(filepath.Dir(v)) == "sub" {
				t.Error("WalkDir() should skip excluded directories")
			}
		}
	})

	t.Run("returns error for non-existent directory", func(t *testing.T) {
		err := WalkDir("/nonexistent/dir", WalkOptions{}, func(path string) error {
			return nil
		})
		if err == nil {
			t.Error("WalkDir() expected error for non-existent directory")
		}
	})
}

func TestCollectGoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := []string{
		filepath.Join(tmpDir, "a.go"),
		filepath.Join(tmpDir, "b.go"),
		filepath.Join(tmpDir, "c_test.go"),
		filepath.Join(tmpDir, "d.txt"),
	}
	for _, f := range files {
		if err := os.WriteFile(f, []byte("package test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("collects go files", func(t *testing.T) {
		goFiles, err := CollectGoFiles(tmpDir, WalkOptions{})
		if err != nil {
			t.Fatalf("CollectGoFiles() error = %v", err)
		}
		if len(goFiles) != 3 {
			t.Errorf("CollectGoFiles() returned %d files, want 3", len(goFiles))
		}
	})

	t.Run("skips test files", func(t *testing.T) {
		goFiles, err := CollectGoFiles(tmpDir, WalkOptions{SkipTests: true})
		if err != nil {
			t.Fatalf("CollectGoFiles() error = %v", err)
		}
		if len(goFiles) != 2 {
			t.Errorf("CollectGoFiles() with SkipTests returned %d files, want 2", len(goFiles))
		}
	})
}
