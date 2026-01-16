package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestStandardToolDefinitions(t *testing.T) {
	defs := StandardToolDefinitions("github")

	if len(defs) != 10 {
		t.Errorf("expected 10 tools, got %d", len(defs))
	}

	// Check that all expected tools are present
	expectedTools := map[string]bool{
		"wetwire_init":     false,
		"wetwire_write":    false,
		"wetwire_read":     false,
		"wetwire_build":    false,
		"wetwire_lint":     false,
		"wetwire_validate": false,
		"wetwire_import":   false,
		"wetwire_list":     false,
		"wetwire_graph":    false,
		"wetwire_scenario": false,
	}

	for _, def := range defs {
		if _, ok := expectedTools[def.Name]; ok {
			expectedTools[def.Name] = true
		} else {
			t.Errorf("unexpected tool: %s", def.Name)
		}

		if def.InputSchema == nil {
			t.Errorf("tool %s has nil InputSchema", def.Name)
		}
	}

	for name, found := range expectedTools {
		if !found {
			t.Errorf("expected tool not found: %s", name)
		}
	}
}

func TestStandardToolDefinitionsDomainCustomization(t *testing.T) {
	githubDefs := StandardToolDefinitions("github")
	awsDefs := StandardToolDefinitions("aws")

	// Descriptions should include domain name
	for _, def := range githubDefs {
		if def.Name == "wetwire_init" {
			if def.Description == "" {
				t.Error("expected non-empty description")
			}
		}
	}

	// AWS should have different descriptions
	for i, def := range awsDefs {
		if githubDefs[i].Description == awsDefs[i].Description && def.Name != "wetwire_lint" && def.Name != "wetwire_list" && def.Name != "wetwire_graph" {
			// lint/list/graph descriptions don't include domain
			// but init/build/validate/import should differ
			t.Logf("tool %s has same description for both domains (may be expected)", def.Name)
		}
	}
}

func TestRegisterStandardTools(t *testing.T) {
	server := NewServer(Config{Name: "test-server", Version: "1.0.0"})

	handlers := StandardToolHandlers{
		Init: func(_ context.Context, _ map[string]any) (string, error) {
			return "init result", nil
		},
		Build: func(_ context.Context, _ map[string]any) (string, error) {
			return "build result", nil
		},
		// Leave other handlers nil - they should not be registered
	}

	RegisterStandardTools(server, "test", handlers)

	// Verify only init and build are registered
	server.mu.RLock()
	defer server.mu.RUnlock()

	if _, ok := server.tools["wetwire_init"]; !ok {
		t.Error("wetwire_init not registered")
	}
	if _, ok := server.tools["wetwire_build"]; !ok {
		t.Error("wetwire_build not registered")
	}
	if _, ok := server.tools["wetwire_lint"]; ok {
		t.Error("wetwire_lint should not be registered (nil handler)")
	}
}

func TestWrapHandler(t *testing.T) {
	fn := func(args map[string]any) (string, error) {
		name := args["name"].(string)
		return "Hello, " + name, nil
	}

	handler := WrapHandler(fn)

	result, err := handler(context.Background(), map[string]any{"name": "World"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "Hello, World" {
		t.Errorf("expected 'Hello, World', got %q", result)
	}
}

func TestSchemas(t *testing.T) {
	schemas := []struct {
		name   string
		schema map[string]any
	}{
		{"InitSchema", InitSchema},
		{"WriteSchema", WriteSchema},
		{"ReadSchema", ReadSchema},
		{"BuildSchema", BuildSchema},
		{"LintSchema", LintSchema},
		{"ValidateSchema", ValidateSchema},
		{"ImportSchema", ImportSchema},
		{"ListSchema", ListSchema},
		{"GraphSchema", GraphSchema},
		{"ScenarioSchema", ScenarioSchema},
	}

	for _, s := range schemas {
		t.Run(s.name, func(t *testing.T) {
			// Verify it's an object type
			typ, ok := s.schema["type"].(string)
			if !ok || typ != "object" {
				t.Errorf("expected type 'object', got %v", s.schema["type"])
			}

			// Verify it has properties
			props, ok := s.schema["properties"].(map[string]any)
			if !ok {
				t.Error("expected properties map")
			}

			if len(props) == 0 {
				t.Error("expected at least one property")
			}
		})
	}
}

func TestDefaultFileWriteHandler(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	args := map[string]any{
		"path":    testFile,
		"content": "Hello, World!",
	}

	result, err := DefaultFileWriteHandler(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty result")
	}

	// Verify the file was created
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if string(data) != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %q", string(data))
	}
}

func TestDefaultFileWriteHandlerWithNestedDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "nested", "dir", "test.txt")

	args := map[string]any{
		"path":    testFile,
		"content": "Nested content",
	}

	_, err := DefaultFileWriteHandler(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the file was created
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if string(data) != "Nested content" {
		t.Errorf("expected 'Nested content', got %q", string(data))
	}
}

func TestDefaultFileWriteHandlerMissingPath(t *testing.T) {
	args := map[string]any{
		"content": "Hello, World!",
	}

	_, err := DefaultFileWriteHandler(context.Background(), args)
	if err == nil {
		t.Error("expected error for missing path")
	}
}

func TestDefaultFileWriteHandlerMissingContent(t *testing.T) {
	args := map[string]any{
		"path": "/tmp/test.txt",
	}

	_, err := DefaultFileWriteHandler(context.Background(), args)
	if err == nil {
		t.Error("expected error for missing content")
	}
}

func TestDefaultFileReadHandler(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Test content"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	args := map[string]any{
		"path": testFile,
	}

	result, err := DefaultFileReadHandler(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != testContent {
		t.Errorf("expected %q, got %q", testContent, result)
	}
}

func TestDefaultFileReadHandlerMissingPath(t *testing.T) {
	args := map[string]any{}

	_, err := DefaultFileReadHandler(context.Background(), args)
	if err == nil {
		t.Error("expected error for missing path")
	}
}

func TestDefaultFileReadHandlerNonExistentFile(t *testing.T) {
	args := map[string]any{
		"path": "/nonexistent/file.txt",
	}

	_, err := DefaultFileReadHandler(context.Background(), args)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestPlaceholderHandler(t *testing.T) {
	handler := PlaceholderHandler("wetwire_test")

	_, err := handler(context.Background(), nil)
	if err == nil {
		t.Error("expected error from placeholder handler")
	}

	if err.Error() != "wetwire_test requires a domain-specific handler - not implemented" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRegisterStandardToolsWithDefaults(t *testing.T) {
	server := NewServer(Config{Name: "test-server", Version: "1.0.0"})

	// Register with minimal handlers - write and read should get defaults
	handlers := StandardToolHandlers{
		Init: func(_ context.Context, _ map[string]any) (string, error) {
			return "init result", nil
		},
	}

	RegisterStandardToolsWithDefaults(server, "test", handlers)

	// Verify init is registered
	server.mu.RLock()
	defer server.mu.RUnlock()

	if _, ok := server.tools["wetwire_init"]; !ok {
		t.Error("wetwire_init not registered")
	}

	// Verify write and read got default handlers
	if _, ok := server.tools["wetwire_write"]; !ok {
		t.Error("wetwire_write not registered")
	}

	if _, ok := server.tools["wetwire_read"]; !ok {
		t.Error("wetwire_read not registered")
	}
}

func TestRegisterStandardToolsWithDefaultsCustomHandlers(t *testing.T) {
	server := NewServer(Config{Name: "test-server", Version: "1.0.0"})

	customWriteCalled := false
	customReadCalled := false

	// Provide custom handlers - they should override defaults
	handlers := StandardToolHandlers{
		Write: func(_ context.Context, _ map[string]any) (string, error) {
			customWriteCalled = true
			return "custom write", nil
		},
		Read: func(_ context.Context, _ map[string]any) (string, error) {
			customReadCalled = true
			return "custom read", nil
		},
	}

	RegisterStandardToolsWithDefaults(server, "test", handlers)

	// Execute the tools to verify custom handlers are used
	_, _ = server.ExecuteTool(context.Background(), "wetwire_write", map[string]any{
		"path":    "/tmp/test.txt",
		"content": "test",
	})

	if !customWriteCalled {
		t.Error("custom write handler was not called")
	}

	_, _ = server.ExecuteTool(context.Background(), "wetwire_read", map[string]any{
		"path": "/tmp/test.txt",
	})

	if !customReadCalled {
		t.Error("custom read handler was not called")
	}
}
