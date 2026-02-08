# domain

Core domain types shared across the module system. Defines the base contract
model, component structure, validation infrastructure, and error types that
all domain sub-packages build upon.

## Key Types

- `BaseContract` is the shared contract structure embedded by `ModuleContract` and other domain types
- `Contract` pairs a contract type (ai or domain) with its raw YAML node
- `ContractError` wraps errors with operation, path, and context for structured diagnostics
- `Loader` reads and unmarshals YAML files into typed structures
- `ValidatorRegistry` is a global registry mapping validator names to `BaseValidator` instances
- `ModuleComponents` is a map of component name to `ComponentEntry` with helper queries
- `ComponentEntry` describes one component including root, patterns, build config, and docker build
- `JSONSchemaValidator` validates documents against JSON Schema using gojsonschema
- `ValidationError` and `Validator` are facade types re-exported from `core/validation`

## Patterns

- `BaseContract` is designed for embedding: `ModuleContract` adds workspace context on top
- `ModuleComponents` provides query methods (`HasComponent`, `GetEnabled`, `GetAllRoots`) over its map
- Validation facade re-exports types from `core/validation` so consumers import only `domain`
- `ContractError` implements `error` and `Unwrap` for standard error chain inspection
- `ValidatorRegistry` uses a package-level map guarded for concurrent reads
- Shared enum constants in `shared_definitions.go` mirror JSON schema values for scanners, platforms, and severities
- `FormatValidationErrors` and `CountCriticalErrors` provide CLI-friendly error formatting

## Internal Structure

| Path | Purpose |
|------|---------|
| `types.go` | `BaseContract`, `ModuleVersioning`, top-level domain types |
| `contract.go` | `Contract` struct and `ContractType` enum |
| `errors.go` | `ContractError` with `IsNotFound` helper |
| `components.go` | `ModuleComponents`, `ComponentEntry`, `ComponentBuild`, `AmpConfig` |
| `registry.go` | `ValidatorRegistry` global validator store |
| `loader.go` | YAML file loading with glob pattern support |
| `validator.go` | `BaseValidator`, `ValidationContext` |
| `validation_facade.go` | Re-exports from `core/validation` |
| `json_validator.go` | JSON Schema validation via gojsonschema |
| `modules/` | Module contract type system and registry |
| `reports/` | Report generation for components, units, specs, changelogs |
| `schema/` | Embedded JSON schema compilation and validation |

## Dependencies

- `contracts/core` for contract filesystem and action type constants
- `core/validation` for validation error types re-exported as facades

## Role in System

This package defines the vocabulary that the entire module system speaks.
Configuration loading produces `BaseContract` values, the modules sub-package
wraps them with workspace context, and reports consume them for CLI output.
By centralizing component structure and validation types here, higher-level
packages avoid circular imports and share a single domain model.

## Code Health

### Tech Debt
- `ModuleComponents` and `ComponentEntry` in `components.go` are near-duplicates of the same types in `config/modules.go`; the two definitions drift independently
- `EACConfigRelPath` constant in `loader.go:16` is explicitly flagged as a duplication of `paths.EACConfigRelPath`
- `globalRegistry` in `registry.go:14` is package-level mutable state populated via `init()` calls; no dependency injection alternative exists

### Pain Points
- `BaseContract` in `types.go` has grown to 20+ methods (getters, artifact queries, book queries); splitting artifact logic into a helper would reduce surface area
- `shared_definitions.go` validation maps (`validScannerCategories`, etc.) are package-level vars returned by reference — callers could mutate them

### Optimization Opportunities
- `ValidScannerCategories()` and similar functions in `shared_definitions.go` return the map directly; returning a copy or using a frozen set would prevent accidental mutation — low effort
- None identified for performance; the package is primarily type definitions with no hot paths
