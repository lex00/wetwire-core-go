# version

Runtime version information using Go's build info.

## Overview

The version package exposes version information for wetwire-core-go using Go's `runtime/debug` build info. This allows dependent packages to query the version at runtime.

## Usage

```go
import "github.com/lex00/wetwire-core-go/version"

// Get the module version
v := version.Version()
// Returns: "v1.2.3" (from go.mod) or "dev" (local builds)

// Get the module path
path := version.ModulePath()
// Returns: "github.com/lex00/wetwire-core-go"
```

## How It Works

- When built with `go build` or `go install`, Go embeds version info from the module
- `Version()` reads this info via `runtime/debug.ReadBuildInfo()`
- Returns "dev" for local development builds without version info

## Use Cases

- Display version in CLI `--version` output
- Include version in MCP server info
- Log version on startup for debugging
- API responses that include server version

## API

```go
// Version returns the module version or "dev" for local builds
func Version() string

// ModulePath returns the canonical module path
func ModulePath() string
```
