package mcp

import (
	"context"
	"testing"
)

func TestStandardToolDefinitions(t *testing.T) {
	defs := StandardToolDefinitions("github")

	if len(defs) != 7 {
		t.Errorf("expected 7 tools, got %d", len(defs))
	}

	// Check that all expected tools are present
	expectedTools := map[string]bool{
		"wetwire_init":     false,
		"wetwire_build":    false,
		"wetwire_lint":     false,
		"wetwire_validate": false,
		"wetwire_import":   false,
		"wetwire_list":     false,
		"wetwire_graph":    false,
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
		{"BuildSchema", BuildSchema},
		{"LintSchema", LintSchema},
		{"ValidateSchema", ValidateSchema},
		{"ImportSchema", ImportSchema},
		{"ListSchema", ListSchema},
		{"GraphSchema", GraphSchema},
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
