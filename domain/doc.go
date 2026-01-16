// Package domain provides the core types and interfaces for wetwire domain packages.
//
// This package defines the foundational abstractions that all wetwire domain
// implementations (aws, honeycomb, etc.) must satisfy. It includes:
//
// - Result and Error types for unified operation returns
// - Domain interface with compile-time enforcement
// - Optional capability interfaces (ImporterDomain, ListerDomain, GrapherDomain)
// - Operation interfaces (Builder, Linter, Initializer, Validator, etc.) with options structs
// - Context type for passing execution context to operations
//
// The design follows the principle of composition over inheritance, allowing
// domain packages to implement only the capabilities they support while
// maintaining a consistent API across all domains.
package domain
