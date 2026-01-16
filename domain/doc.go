// Package domain provides the core interface for wetwire domain implementations.
//
// Domain packages implement the Domain interface to get automatic CLI command
// and MCP tool generation. This eliminates boilerplate and ensures consistency
// across all wetwire domains.
//
// # Required Interface
//
// All domains must implement the Domain interface:
//
//   - Name() string - domain identifier (e.g., "aws", "gitlab")
//   - Version() string - domain version
//   - Builder() Builder - generates output from source definitions
//   - Linter() Linter - checks code quality
//   - Initializer() Initializer - creates project scaffolding
//   - Validator() Validator - validates generated output
//
// # Optional Interfaces
//
// Domains may also implement these for additional capabilities:
//
//   - ImporterDomain - for importing external configs (adds wetwire_import)
//   - ListerDomain - for listing discovered resources (adds wetwire_list)
//   - GrapherDomain - for visualizing dependencies (adds wetwire_graph)
//
// # CLI Generation
//
// Use Run() to generate a complete CLI from a Domain:
//
//	func main() {
//	    cli := domain.Run(&MyDomain{})
//	    cli.Execute()
//	}
//
// This creates CLI commands: build, lint, init, validate (plus import, list, graph
// if the optional interfaces are implemented).
//
// # MCP Server Generation
//
// Use BuildMCPServer() to generate an MCP server:
//
//	server := domain.BuildMCPServer(&MyDomain{})
//	server.Start()
//
// This registers MCP tools: wetwire_build, wetwire_lint, wetwire_init, wetwire_validate
// (plus wetwire_import, wetwire_list, wetwire_graph if implemented).
//
// # Compile-Time Enforcement
//
// Use interface assertions to catch missing methods at compile time:
//
//	var _ domain.Domain = (*MyDomain)(nil)         // Required
//	var _ domain.ImporterDomain = (*MyDomain)(nil) // Optional
//
// # Adding Custom Commands
//
// Domain-specific commands are added after Run():
//
//	func main() {
//	    cli := domain.Run(&MyDomain{})
//	    cli.AddCommand(newDesignCmd())  // AI-assisted design
//	    cli.AddCommand(newTestCmd())    // Persona testing
//	    cli.Execute()
//	}
//
// # Example Implementation
//
// See domain/mock_domain_test.go for a complete example implementation.
package domain
