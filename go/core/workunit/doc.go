// Package workunit provides unified types for work unit handling across
// build, test, lint, and scan commands.
//
// # Core Types
//
// The package defines three primary types:
//
//   - [UnitID]: Uniquely identifies a unit of work
//   - [UnitSpec]: Complete specification for executing a work unit
//   - [core.ActionType]: Operation type (build, test, lint, scan)
//
// # Naming Methods
//
// UnitID provides multiple naming methods for different use cases:
//
//	id := workunit.UnitID{
//	    Action:    core.ActionBuild,
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
// Path - Component-level identification:
//
//	id.Path() // "core:go"
//
// Returns module:component. Not unique across contexts or tools - use Longname()
// when uniqueness is required.
//
// DisplayName - Context-aware compact display name:
//
//	buildID.DisplayName()    // "core: go" (build context, component=tool)
//	siteID.DisplayName()     // "docs: site: mkdocs" (build context, component!=tool)
//	testID.DisplayName()     // "impl-build: unit" (test context)
//	specID.DisplayName()     // "build-module: godog" (test context, BDD)
//	lintID.DisplayName()     // "lint:go:golangci-lint" (lint context)
//	scanID.DisplayName()     // "scan:go:sbom" (scan context)
//
// Format varies by context for optimal readability in TUI tabs and logs.
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
//	    Action:  core.ActionTest,
//	    Module:  "eac-cli",
//	    Tool:    "godog",
//	    Spec:    "build-module",
//	}
//	id.Longname()     // "eac-cli:spec:build-module"
//	id.DisplayName()  // "build-module: godog"
//
// # When to Use Each Method
//
//   - Longname(): Cache keys, state tracking, dependency matching, unique IDs
//   - Path(): Grouping by component, when context/tool doesn't matter
//   - DisplayName(): User-facing output (TUI tabs, logs)
//   - TabLabel(n): TUI tabs with limited width
//
// # State Management
//
// The package provides [StateManager] for incremental execution tracking.
// State files are stored under .cache/eac/incremental/ (separate from output
// artifacts in out/) so that deleting .cache/eac/ clears all incremental state:
//
//	id.StateFile()     // ".cache/eac/incremental/build/core/go-go/state.json"
//	id.StateCacheDir() // ".cache/eac/incremental/build/core/go-go"
//	id.OutDir()        // "out/build/core/go-go" (logs, results, locks)
//
// See [StateManager] for cache detection and state persistence APIs.
package workunit
