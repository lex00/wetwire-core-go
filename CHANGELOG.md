# Changelog

## [Unreleased]

### Added
- `version/` package for exposing version info to dependent packages via `runtime/debug`
- `cmd/` package for shared CLI command framework with cobra
  - `NewRootCommand`, `NewBuildCommand`, `NewLintCommand`, `NewInitCommand`, `NewValidateCommand`
  - `Builder`, `Linter`, `Initializer`, `Validator` interfaces for domain implementations
- `serialize/` package for struct-to-map conversion and JSON/YAML serialization
  - `ToMap`, `ToYAML`, `ToJSON` with naming convention options
  - `SnakeCase`, `CamelCase`, `PascalCase`, `OmitEmpty` options

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
