package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lex00/wetwire-core-go/mcp"
)

// Example showing how to create an MCP server with standard tools
func main() {
	// Create a new MCP server
	server := mcp.NewServer(mcp.Config{
		Name:    "wetwire-example",
		Version: "1.0.0",
		Debug:   true,
	})

	// Define handlers for domain-specific tools
	handlers := mcp.StandardToolHandlers{
		// Init handler - domain-specific project initialization
		Init: func(ctx context.Context, args map[string]any) (string, error) {
			name := args["name"].(string)
			return fmt.Sprintf("Initialized project: %s", name), nil
		},

		// Build handler - domain-specific build logic
		Build: func(ctx context.Context, args map[string]any) (string, error) {
			pkg := args["package"].(string)
			return fmt.Sprintf("Built package: %s", pkg), nil
		},

		// Lint handler - domain-specific linting
		Lint: func(ctx context.Context, args map[string]any) (string, error) {
			pkg := args["package"].(string)
			return fmt.Sprintf("Linted package: %s - no issues found", pkg), nil
		},

		// Write and Read handlers are optional - default implementations will be used
		// if not provided. Uncomment below to override defaults:
		//
		// Write: func(ctx context.Context, args map[string]any) (string, error) {
		//     // Custom write logic here
		//     return "Custom write", nil
		// },
		//
		// Read: func(ctx context.Context, args map[string]any) (string, error) {
		//     // Custom read logic here
		//     return "Custom read", nil
		// },
	}

	// Register standard tools with default handlers for file operations
	mcp.RegisterStandardToolsWithDefaults(server, "example", handlers)

	// You can also register additional custom tools
	server.RegisterTool("custom_tool", "A custom domain-specific tool", func(ctx context.Context, args map[string]any) (string, error) {
		return "Custom tool result", nil
	})

	// Start the MCP server (blocks until stdin is closed)
	log.Println("Starting MCP server...")
	if err := server.Start(context.Background()); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
