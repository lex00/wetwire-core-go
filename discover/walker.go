package discover

import (
	"os"
	"path/filepath"
	"strings"
)

// WalkOptions configures directory walking behavior.
type WalkOptions struct {
	// SkipTests skips *_test.go files.
	SkipTests bool
	// SkipVendor skips vendor directories.
	SkipVendor bool
	// SkipHidden skips directories starting with ".".
	SkipHidden bool
	// SkipTestdata skips testdata directories.
	SkipTestdata bool
	// ExcludeDirs lists additional directory names to skip.
	ExcludeDirs []string
}

// WalkDir walks a directory tree and calls fn for each Go source file.
func WalkDir(root string, opts WalkOptions, fn func(path string) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Handle directories
		if info.IsDir() {
			name := info.Name()

			// Skip hidden directories
			if opts.SkipHidden && strings.HasPrefix(name, ".") && path != root {
				return filepath.SkipDir
			}

			// Skip vendor directories
			if opts.SkipVendor && name == "vendor" {
				return filepath.SkipDir
			}

			// Skip testdata directories
			if opts.SkipTestdata && name == "testdata" {
				return filepath.SkipDir
			}

			// Skip excluded directories
			for _, excluded := range opts.ExcludeDirs {
				if name == excluded {
					return filepath.SkipDir
				}
			}

			return nil
		}

		// Handle files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		if opts.SkipTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}

		return fn(path)
	})
}

// CollectGoFiles walks a directory and returns all Go file paths.
func CollectGoFiles(root string, opts WalkOptions) ([]string, error) {
	var files []string
	err := WalkDir(root, opts, func(path string) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
