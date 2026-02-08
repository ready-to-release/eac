# output

Tracks Unit of Work (UoW) execution results, persists manifests to disk,
validates artifact integrity, and provides change detection at both UoW
and module granularity.

## Key Types

- **`UoWManifest`** -- Execution metadata, artifacts, and hashes for one UoW
- **`Artifact`** -- Single output file with SHA256 hash and size
- **`InMemoryTracker`** -- Records UoW start/complete with disk persistence
- **`DiskOutputReader`** -- Reads and aggregates manifests from `out/` tree
- **`ValidationResult`** -- Outcome of manifest and artifact validation
- **`UoWChangeResult`** -- Batch change detection at UoW granularity
- **`AggregatedChangeResult`** -- Module-level rollup of UoW change detection
- **`Moniker`** -- Colon-separated unit identifier parser
- **`ComponentView`** -- Aggregated UoW results for a single component
- **`ModuleView`** -- Aggregated component results for a single module

## Patterns

- Manifest-on-disk: UoW results persisted as `uow.manifest.json` files
- Hash-based caching: input hash comparison drives change detection
- Port adapters: `OutputReaderAdapter` and `UoWTrackerAdapter` bridge contracts
- Status aggregation: component and module status computed from UoW exit codes

## Internal Structure

| File | Responsibility |
| --- | --- |
| types.go | Core types: `UoWManifest`, `Artifact`, `Status`, view types |
| manifest.go | `Save`/`Load` for manifest JSON persistence |
| tracker.go | `InMemoryTracker` for recording UoW lifecycle |
| reader.go | `DiskOutputReader` for scanning and aggregating manifests |
| cache_detector.go | `DetectUoWChanges`, `IsModuleChanged` logic |
| validation.go | `HashFile`, `ValidateArtifacts`, `ComputeOutputHash` |
| aggregation.go | `AggregateUoWChanges` module-level rollup |
| buffer.go | `outputBuffer` for TUI stdout/stderr capture |
| passthrough.go | No-op buffer for console mode |
| moniker.go | `Moniker` parsing and display helpers |
| ports.go | Contract port adapters for reader, tracker, manifest |

## Dependencies

- `contracts/core` -- port interfaces (`OutputReaderPort`, `UoWTrackerPort`)
- `core/workunit` -- `UnitID` and `UoWAggregator` types

## Role in System

This package is the persistence and caching backbone of the `core` module.
Every build, test, lint, and scan operation writes a `UoWManifest` through
the tracker, and subsequent runs use the reader and change detector to skip
work whose inputs have not changed. The port adapters expose this
functionality through contract interfaces to the CLI layer.

## Code Health

### Tech Debt
- `ports.go` (319 lines): dominated by trivial getter methods to satisfy port interfaces -- consider code generation or embedding to reduce boilerplate
- `reader.go` (360 lines): `DiskOutputReader` handles reading, validation, integrity checks, and build-ID extraction in a single struct

### Pain Points
- Port adapter layer (`ports.go`) must be updated every time a field is added to `UoWManifest`, `Artifact`, or view types -- high coupling surface
- Validation logic in `reader.go` (`ValidateUoW`, `ValidateModule`, `VerifyModuleIntegrity`) mixes read concerns with verification concerns

### Optimization Opportunities
- Extract validation methods from `DiskOutputReader` into a dedicated `Validator` struct to separate read and verify responsibilities (low effort)
- No TODO/FIXME markers found -- codebase is clean of deferred work items
