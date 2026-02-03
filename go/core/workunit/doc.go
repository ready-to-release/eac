// Package workunit provides unified types for work unit handling across
// build, test, lint, and scan commands.
//
// # Core Types
//
// The package defines three primary types:
//
//   - [UnitID]: Uniquely identifies a unit of work
//   - [UnitSpec]: Complete specification for executing a work unit
//   - [Context]: Operation type (build, test, lint, scan)
//
// # Naming Methods
//
// UnitID provides multiple naming methods for different use cases:
//
//	id := workunit.UnitID{
//	    Context:   workunit.ContextBuild,
//	    Module:    "core",
//	    Component: "go",
//	    Tool:      "go",
//	}
//
// Longname - Unique identifier for matching and state tracking:
//
//	id.Longname() // "build:core:go:go"
//
// Used for: cache keys, state files, dependency matching, TUI tab matching.
// Format: "context:module:component:tool" or "module:spec:specname" for BDD tests.
//
// Shortname - Brief display name (deprecated):
//
//	id.Shortname() // "core:go"
//
// Deprecated: Use Path() for stable identifiers or DisplayName() for contextual display.
//
// Path - Component-level identification:
//
//	id.Path() // "core:go"
//
// Returns module:component. Not unique across contexts or tools - use Longname()
// when uniqueness is required.
//
// DisplayName - Context-aware display name:
//
//	id.DisplayName(false) // "go"       - just component
//	id.DisplayName(true)  // "core:go"  - module:component
//
// Use disambiguate=true when multiple modules have same component names.
// For BDD tests, returns spec name or "module:spec:specname".
//
// TabLabel - Truncated name for TUI tabs:
//
//	id.TabLabel(10) // "go"
//	id.TabLabel(5)  // "go"
//
// Returns component (or spec name for BDD tests), truncated with "..." if needed.
//
// # BDD Test Naming
//
// For BDD tests (godog, tscucumber), the Spec field enables special formatting:
//
//	id := workunit.UnitID{
//	    Module: "eac-cli",
//	    Spec:   "build-module",
//	}
//	id.Longname()       // "eac-cli:spec:build-module"
//	id.Shortname()      // "build-module"
//	id.DisplayName(true) // "eac-cli:spec:build-module"
//
// # When to Use Each Method
//
//   - Longname(): Cache keys, state tracking, dependency matching, unique IDs
//   - Path(): Grouping by component, when context/tool doesn't matter
//   - DisplayName(false): User-facing output when context is clear
//   - DisplayName(true): User-facing output when disambiguation needed
//   - TabLabel(n): TUI tabs with limited width
//
// # State Management
//
// The package provides [StateManager] for incremental execution tracking.
// State files are stored at paths derived from UnitID:
//
//	id.StateFile() // "out/build/core/go/state.json"
//
// See [StateManager] for cache detection and state persistence APIs.
package workunit
