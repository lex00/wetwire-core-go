# Changelog

## [Unreleased]

### Added
- Output capture for scenario runner
  - `OutputManifest` type to store domain outputs in structured format
  - `DomainOutput` and `ResourceOutput` types for per-domain, per-resource output storage
  - `NewOutputManifest()`, `AddDomainOutput()`, `GetDomainOutput()`, `GetResourceOutput()` for manifest management
  - `SaveToFile()` and `LoadFromFile()` for JSON persistence
  - `SaveToYAML()` and `LoadFromYAML()` for YAML persistence
  - `OutputExtractor` with configurable patterns for YAML, JSON, and Go DSL formats
  - `CaptureOutputsFromFiles()` to discover and parse output files matching domain output patterns
  - Comprehensive test coverage for output capture functionality
  - Closes #91
- Cross-domain reference resolver for scenario package
  - `CrossDomainRef` type representing parsed references with Domain, Resource, and Field
  - `ParseRef()` function to parse `${domain.resource.outputs.field}` syntax
  - `FindRefsInString()` to extract all references from a string
  - `OutputManifest.ValidateRefs()` to validate references against output manifest
  - `OutputManifest.ResolveRef()` to resolve references to their values
  - Closes #92
- `domain.ScaffoldCrossDomainScenario()` for multi-domain scenario scaffolding
  - Generates scenario.yaml with multiple domains, cross_domain relationships, and validation rules
  - Creates system_prompt.md with cross-domain instructions
  - Generates persona prompts (beginner, intermediate, expert, terse, verbose)
  - Creates expected/{domain}/ subdirectories for each domain
  - Closes #94
- Enhanced `domain.WriteScenario()` to handle nested directory creation for multi-domain scenarios
- `ast/` package for shared Go AST parsing utilities
  - `ParseFile`, `ParseDir`, `WalkGoFiles` with `ParseOptions` for configurable skipping
  - `ExtractImports` for extracting import map from ast.File
  - `ExtractTypeName`, `InferTypeFromValue` for type analysis
  - `IsBuiltinType`, `IsBuiltinIdent`, `IsKeyword` for identifier classification
  - Closes #85
- `lint/` package for shared linting infrastructure
  - `Rule`, `FixableRule`, `PackageAwareRule` interfaces for lint rules
  - `Issue`, `Severity`, `Config` types for lint results and configuration
  - `LintFile`, `LintDir`, `LintDirRecursive`, `LintBytes` for analysis
  - `Fix`, `FixFile`, `FixDir` for automatic fixing
  - `RuleRegistry` for managing rule collections
  - Closes #86
- `discover/` package for shared resource discovery infrastructure
  - `DiscoveredResource`, `DiscoverResult`, `TypeMatcher` types
  - `Discover`, `DiscoverFile`, `DiscoverDir`, `DiscoverAST` for resource discovery
  - `WalkDir`, `CollectGoFiles` for directory traversal
  - Pluggable `TypeMatcher` for domain-specific resource identification
  - Closes #87
- Extended `BuildOpts` with `Output` and `DryRun` fields for common build options
- Extended `LintOpts` with `Fix` and `Disable` fields for common lint options
- CLI flags: `--output`, `--dry-run` for build; `--fix`, `--disable` for lint
- MCP schema updated with `disable` field for lint tool
- Closes #77
- Codecov integration for coverage reporting in CI
- `providers/kiro` package implementing the Provider interface for Kiro CLI
  - Wraps `kiro.RunTest()` for non-interactive AI execution
  - Supports `CreateMessage` and `StreamMessage` methods
  - Closes #23
- `mcp/client.go` MCP client for connecting to MCP servers
  - `NewClient()` spawns MCP server process and initializes connection
  - `ListTools()` discovers available tools from MCP server
  - `CallTool()` executes tools via MCP server
- MCP integration for Anthropic provider
  - `MCPConfig` option in `anthropic.Config` for MCP server settings
  - `NewWithMCP()` constructor that starts MCP server automatically
  - `GetMCPTools()` returns tools discovered from MCP server
  - `CallMCPTool()` executes tools via MCP server
  - `HasMCP()` checks if MCP is configured
  - Closes #29
- `scenario/` package for multi-domain scenario definitions
  - `ScenarioConfig` struct with domains, cross-domain relationships, validation
  - `DomainSpec` for domain configuration (CLI, MCP tools, dependencies, outputs)
  - `CrossDomainSpec` for relationships between domains
  - `Load()` and `Parse()` for YAML scenario file loading
  - `GetDomainOrder()` for topological sorting by dependencies
  - `Validate()` for comprehensive scenario validation
  - Closes #41

## [1.2.0] - 2026-01-10

### Added
- `version/` package for exposing version info to dependent packages via `runtime/debug`
- `cmd/` package for shared CLI command framework with cobra
  - `NewRootCommand`, `NewBuildCommand`, `NewLintCommand`, `NewInitCommand`, `NewValidateCommand`
  - `Builder`, `Linter`, `Initializer`, `Validator` interfaces for domain implementations
- `serialize/` package for struct-to-map conversion and JSON/YAML serialization
  - `ToMap`, `ToYAML`, `ToJSON` with naming convention options
  - `SnakeCase`, `CamelCase`, `PascalCase`, `OmitEmpty` options
- `lsp/` package for Language Server Protocol infrastructure
  - `NewServer`, `Diagnose`, `Complete`, `Hover`, `Definition` methods
  - `DiagnosticProvider`, `CompletionProvider`, `HoverProvider`, `DefinitionProvider` interfaces

## [1.1.0] - 2026-01-10

### Added
- Provider abstraction layer (`providers/`) for multi-backend AI support
- `providers.Provider` interface with `CreateMessage` and `StreamMessage` methods
- `providers/anthropic` package implementing the Provider interface
- `kiro/` package stub for future Kiro CLI integration
- `CreateDeveloperResponderWithProvider` function for provider-agnostic developer agents
- `Traits` field to `Persona` struct for persona characteristic tagging

### Changed
- `RunnerAgent` now accepts a configurable `Provider` instead of direct Anthropic client
- `RunnerConfig.Provider` field added for custom provider injection
- Tool definitions now use provider-agnostic `providers.Tool` type
- Streaming handler type aliased to `providers.StreamHandler`
- All 5 personas now include trait tags for programmatic access

## [1.0.0] - 2026-01-03

### Added
- CodeBot GitHub Actions workflow
- Comprehensive edge case test coverage for all modules
- Integration tests for agents module

### Changed
- License updated to Apache 2.0

## [0.1.1] - Previous

### Added
- GitHub Actions CI workflow
- Test and build automation

## [0.1.0] - Initial Release

### Added
- Core agent infrastructure (personas, scoring, results, orchestrator, agents)
- Anthropic SDK integration
- Developer/Runner agent coordination
