package domain

import (
	"github.com/lex00/wetwire-core-go/mcp"
)

// BuildMCPServer creates a fully configured MCP server from a Domain instance.
// It registers all required tools (build, lint, init, validate) and optional tools
// (import, list, graph) if the domain implements the corresponding interfaces.
//
// The server can be started with server.Start() to begin listening for MCP requests,
// or used directly via server.ExecuteTool() for in-process tool execution.
func BuildMCPServer(d Domain) *mcp.Server {
	server := mcp.NewServer(mcp.Config{
		Name:    "wetwire-" + d.Name(),
		Version: d.Version(),
	})

	// Register required tools
	server.RegisterToolWithSchema("wetwire_build", "Build output from domain resources",
		createBuildHandler(d.Builder()), mcp.BuildSchema)

	server.RegisterToolWithSchema("wetwire_lint", "Lint domain resources",
		createLintHandler(d.Linter()), mcp.LintSchema)

	server.RegisterToolWithSchema("wetwire_init", "Initialize new domain project",
		createInitHandler(d.Initializer()), mcp.InitSchema)

	server.RegisterToolWithSchema("wetwire_validate", "Validate generated output",
		createValidateHandler(d.Validator()), mcp.ValidateSchema)

	// Register optional tools via type assertion
	if imp, ok := d.(ImporterDomain); ok {
		server.RegisterToolWithSchema("wetwire_import", "Import external resources",
			createImportHandler(imp.Importer()), mcp.ImportSchema)
	}

	if lst, ok := d.(ListerDomain); ok {
		server.RegisterToolWithSchema("wetwire_list", "List discovered resources",
			createListHandler(lst.Lister()), mcp.ListSchema)
	}

	if gph, ok := d.(GrapherDomain); ok {
		server.RegisterToolWithSchema("wetwire_graph", "Visualize resource relationships",
			createGraphHandler(gph.Grapher()), mcp.GraphSchema)
	}

	return server
}
