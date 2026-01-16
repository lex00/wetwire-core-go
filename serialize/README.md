# serialize

Struct-to-map conversion with configurable naming conventions.

## Overview

The serialize package converts Go structs to maps and serializes them to JSON/YAML with control over field naming conventions. Useful for generating configuration files or API payloads with consistent naming.

## Usage

### Basic Conversion

```go
import "github.com/lex00/wetwire-core-go/serialize"

type Config struct {
    MaxRetries  int
    EnableCache bool
    APIKey      string
}

config := Config{MaxRetries: 3, EnableCache: true, APIKey: "secret"}

// Convert to map with snake_case keys
m := serialize.ToMap(config, serialize.SnakeCase)
// Result: {"max_retries": 3, "enable_cache": true, "api_key": "secret"}

// Convert to map with camelCase keys
m := serialize.ToMap(config, serialize.CamelCase)
// Result: {"maxRetries": 3, "enableCache": true, "apiKey": "secret"}
```

### Serialization

```go
// To YAML with snake_case
yaml, _ := serialize.ToYAML(config, serialize.SnakeCase)
// max_retries: 3
// enable_cache: true
// api_key: secret

// To JSON with camelCase
json, _ := serialize.ToJSON(config, serialize.CamelCase)
// {"maxRetries":3,"enableCache":true,"apiKey":"secret"}
```

### Options

| Option | Description |
|--------|-------------|
| `SnakeCase` | Convert field names to snake_case (MaxRetries → max_retries) |
| `CamelCase` | Convert field names to camelCase (MaxRetries → maxRetries) |
| `PascalCase` | Keep field names as PascalCase (default) |
| `OmitEmpty` | Omit fields with zero values |

Options can be combined:

```go
m := serialize.ToMap(config, serialize.SnakeCase, serialize.OmitEmpty)
```

### Nested Structs

Nested structs are converted recursively:

```go
type Server struct {
    Host   string
    Config Config
}

server := Server{Host: "localhost", Config: Config{MaxRetries: 3}}
m := serialize.ToMap(server, serialize.SnakeCase)
// Result: {"host": "localhost", "config": {"max_retries": 3, ...}}
```

## API

```go
// ToMap converts a struct to a map
func ToMap(v any, opts ...Option) map[string]any

// ToYAML serializes a struct to YAML bytes
func ToYAML(v any, opts ...Option) ([]byte, error)

// ToJSON serializes a struct to JSON bytes
func ToJSON(v any, opts ...Option) ([]byte, error)
```
