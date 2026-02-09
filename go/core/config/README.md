# config

Central configuration loading for the workspace. Reads YAML files from three
layers (contract defaults, user config, personal overrides) and merges them
into a unified `EACConfig` that the rest of the system consumes.

## Key Types

- `EACConfig` holds all loaded configuration sections and exposes load methods for each concern
- `LoadOptions` controls which config sections to load, schema validation, and lazy-load behavior
- `RepositoryConfig` represents the unified repository.yml including modules, paths, and conventions
- `Module` defines a deployable module with its components, versioning, and dependency declarations
- `ComponentEntry` describes a single component within a module (root, patterns, build config)
- `ComponentType` defines how to process files of a certain type (builders, linters, testers, scanners)
- `ComponentTypesConfig` maps component type names to their `ComponentType` definitions
- `ConfigCache` provides a process-level read-through cache with `sync.RWMutex` protection
- `ConfigLayer` enumerates the three merge layers: contract, user, and personal
- `LoadedFile` tracks which file a configuration value originated from

## Patterns

- Three-layer merge: contract defaults from embedded YAML, user config from `.eac/`, personal overrides from `.eac.personal/`
- Factory default singletons use `sync.Once` with clone-on-read to provide thread-safe default values
- Core configs (repository) fail fast on errors; optional configs (environments, books) log and continue
- `LoadAll` orchestrates loading all sections in sequence, separating core from optional
- `Global()` singleton provides a process-wide config instance with `SetGlobalForTesting()` for tests
- Schema validation delegates to `domain/schema.Validator` when `ValidateSchemas` is enabled
- `ApplyComponentDefaults` fills in default roots and patterns from `ComponentTypesConfig` definitions
- Config source tracking records which layer and file each value came from via `LoadedConfig`

## Internal Structure

| Path | Purpose |
|------|---------|
| `config.go` | Main entry point: `Load`, `LoadAll`, global cache |
| `repository.go` | Repository config with paths, conventions, remotes |
| `modules.go` | Module and component entry types, component discovery |
| `module_types.go` | Build artifacts, docker build config, type constants |
| `component_types.go` | Component type registry, file extension mapping |
| `defaults.go` | Contract default loading and merge functions |
| `factory_defaults.go` | Singleton factory defaults with `sync.Once` |
| `cache.go` | Process-level config cache with read-write locking |
| `validation.go` | Schema validation helpers and `MultiError` |
| `config_source.go` | Config layer enum and source tracking types |

## Dependencies

- `contracts/core` for embedded default YAML and contract filesystem
- `contracts/scanner` for security and risk config port interfaces
- `core/domain/schema` for JSON schema validation
- `core/domain` for shared domain types referenced by module config
- `core/paths` for workspace path resolution
- `core/workspace` for workspace root detection
- `core/resource` for pool allocation types used by component types

## Role in System

This package is the single source of truth for all workspace configuration.
Every command and subsystem reads its settings through `EACConfig` rather than
parsing YAML directly. The three-layer merge allows contract authors to ship
sensible defaults while letting teams and individuals override specific values.

## Code Health

### Tech Debt
- Multiple global mutable singletons: `global.go` (globalConfig), `cache.go` (globalConfigCache), `git_provider.go` (gitRemoteProvider) each with separate locking
- `ModuleComponents.Clone()` in `modules.go:426-479` is a 50+ line manual deep-copy; consider a code-generated or reflection-based approach
- `LoadRepository` in `config.go:254-343` is a 90-line, 10-step orchestration method that mixes loading, merging, and expansion

### Pain Points
- No TODO/FIXME markers exist, but the package spans ~56 files making navigation difficult; consider grouping related files into sub-packages (e.g., `config/risk`, `config/books`)
- `RepositoryConfig` in `repository.go` carries path-helper methods (30+) that could live in `core/paths` to reduce surface area

### Optimization Opportunities
- `GetModule`/`GetByMoniker` in `repository.go` perform linear scans over `[]Module`; a moniker-to-index map built once after load would be O(1) — low effort, high impact for large configs
- `GetByExtension` in `component_types.go:311-323` also does a linear scan; a pre-built extension map would help repos with many component types — low effort
