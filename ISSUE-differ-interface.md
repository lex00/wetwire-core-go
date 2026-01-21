# Issue: Add Differ Interface to Domain and Share K8s Implementation

## Summary

Add a `Differ` interface to wetwire-core-go's domain package that domains can optionally implement. The K8s manifest differ implementation should live in wetwire-k8s-go and be reusable by all K8s-based domains (GCP, K8s, future ACK/ASO).

## Problem

Currently, differ implementations are duplicated and inconsistent:

| Domain | Has Differ | Location | Format |
|--------|-----------|----------|--------|
| AWS | Yes | `internal/differ/` | CloudFormation |
| GCP | Yes | `internal/differ/` | K8s manifests |
| K8s | No | - | K8s manifests |
| Azure | No | - | ARM templates |
| Observability | No | - | Prometheus/Grafana |

Issues:
1. **No interface** - Differs are ad-hoc, not part of the domain contract
2. **Duplication** - GCP's K8s manifest differ duplicates logic that K8s domain should provide
3. **Inconsistent CLI** - Some domains have `diff` command, others don't

## Solution

### Phase 1: Add Differ Interface to wetwire-core-go

Add to `domain/domain.go`:

```go
// DiffOpts configures the differ.
type DiffOpts struct {
    // IgnoreOrder ignores array element order in comparisons
    IgnoreOrder bool
    // Format specifies output format (text, json)
    Format string
}

// DiffEntry represents a single difference.
type DiffEntry struct {
    Resource string   // Resource identifier
    Type     string   // Resource type
    Action   string   // "added", "removed", "modified"
    Changes  []string // Field-level changes for modified resources
}

// DiffResult contains the comparison result.
type DiffResult struct {
    Entries []DiffEntry
    Summary struct {
        Added    int
        Removed  int
        Modified int
        Total    int
    }
}

// Differ compares two outputs and returns semantic differences.
type Differ interface {
    Diff(ctx *Context, file1, file2 string, opts DiffOpts) (*DiffResult, error)
}

// DifferDomain is an optional interface for domains that support diff.
type DifferDomain interface {
    Domain
    Differ() Differ
}
```

Update `domain/run.go` to auto-register `diff` command if domain implements `DifferDomain`.

### Phase 2: Create Shared K8s Differ in wetwire-k8s-go

Create `differ/differ.go` in wetwire-k8s-go:

```go
package differ

import (
    "github.com/lex00/wetwire-core-go/domain"
)

// K8sDiffer implements domain.Differ for Kubernetes manifests.
type K8sDiffer struct{}

var _ domain.Differ = (*K8sDiffer)(nil)

func New() *K8sDiffer {
    return &K8sDiffer{}
}

func (d *K8sDiffer) Diff(ctx *domain.Context, file1, file2 string, opts domain.DiffOpts) (*domain.DiffResult, error) {
    // Implementation moved from wetwire-gcp-go/internal/differ
    // Identifies resources by apiVersion/kind/namespace/name
    // ...
}
```

### Phase 3: Update Domain Packages

**wetwire-gcp-go:**
```go
import "github.com/lex00/wetwire-k8s-go/differ"

func (d *GCPDomain) Differ() domain.Differ {
    return differ.New()  // Reuse K8s differ
}
```

**wetwire-k8s-go:**
```go
func (d *K8sDomain) Differ() domain.Differ {
    return differ.New()
}
```

**wetwire-aws-go:**
```go
func (d *AWSDomain) Differ() domain.Differ {
    return &awsDiffer{}  // Keep CloudFormation-specific implementation
}
```

**wetwire-azure-go:** (future)
```go
func (d *AzureDomain) Differ() domain.Differ {
    return &armDiffer{}  // ARM template-specific implementation
}
```

## Implementation Tasks

### wetwire-core-go
- [ ] Add `DiffOpts`, `DiffEntry`, `DiffResult` types to `domain/`
- [ ] Add `Differ` interface to `domain/`
- [ ] Add `DifferDomain` optional interface to `domain/`
- [ ] Update `domain.Run()` to auto-register `diff` command for `DifferDomain`
- [ ] Add tests for diff command registration

### wetwire-k8s-go
- [ ] Create `differ/` package at top level (exported, not internal)
- [ ] Move K8s manifest comparison logic from wetwire-gcp-go
- [ ] Implement `domain.Differ` interface
- [ ] Add comprehensive tests
- [ ] Update K8sDomain to implement `DifferDomain`

### wetwire-gcp-go
- [ ] Remove `internal/differ/` package
- [ ] Import `github.com/lex00/wetwire-k8s-go/differ`
- [ ] Update GCPDomain to implement `DifferDomain`
- [ ] Update diff.go to use domain's Differ

### wetwire-aws-go
- [ ] Update AWSDomain to implement `DifferDomain`
- [ ] Refactor `internal/differ/` to implement `domain.Differ` interface
- [ ] Update diff.go to use domain's Differ

## Benefits

1. **Consistent API** - All domains with diff support use the same interface
2. **Auto CLI registration** - `diff` command added automatically for DifferDomain
3. **Code reuse** - K8s-based domains share the manifest differ
4. **Extensible** - New domains just implement the interface
5. **Testable** - Standard interface enables consistent testing

## Verification

For each domain implementing DifferDomain:
```bash
wetwire-{domain} diff file1 file2           # Text output
wetwire-{domain} diff file1 file2 -f json   # JSON output
wetwire-{domain} diff file1 file2 --ignore-order
```

## Labels

`enhancement`, `architecture`, `domain-interface`
