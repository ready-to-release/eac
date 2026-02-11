# internal

Provides shared infrastructure for GET and SHOW commands, including artifact resolution with metadata overrides, effective module configuration with path variable substitution, and build artifact validation with dependency traversal.

## Key Types

- **`ResolvedArtifact`** -- An artifact with metadata overrides applied, existence checked, and paths resolved
- **`ArtifactResolutionSummary`** -- Summary statistics for artifact resolution (total, exists, missing, overrides)
- **`ModuleValidationResult`** -- Validation result for a single module's artifacts
- **`ValidationResults`** -- Aggregated validation results across a module and all its transitive dependencies
- **`EffectiveModule`** -- Module configuration with resolved paths and package roots
- **`PathVariables`** -- Repository-wide path variables for template substitution
- **`PlatformInfo`** -- OS/architecture platform descriptor
- **`ArtifactInfo`** -- Complete artifact descriptor with type, ID, path, size, SHA256, and container registry info

## Key Functions

- **`ResolveArtifactsForModule()`** -- Resolve all artifacts for a module with metadata overrides applied
- **`ResolveArtifactsForModuleWithConfig()`** -- Resolve artifacts with optional book configuration for PDF expansion
- **`ValidateArtifactsTargetOnly()`** -- Validate artifacts for a single module without dependency checking
- **`ValidateArtifactsWithDependencies()`** -- Validate artifacts for a module and all transitive dependencies
- **`DetermineRequestedArtifacts()`** -- Determine which artifact IDs should be built based on module and mode
- **`GetEffectiveModuleConfig()`** -- Get module configuration with resolved paths and package roots
- **`GetPathVariables()`** -- Extract path variables from repository configuration
- **`FormatArtifactStatus()`** -- Format human-readable status string for an artifact
- **`FormatArtifactSize()`** -- Format file size in human-readable units (KB, MB, GB)
- **`expandBookArtifacts()`** -- Expand wildcard PDF patterns to specific book PDF artifacts
- **`addDependenciesRecursive()`** -- Recursively collect all transitive dependencies of a module
- **`deriveArtifactID()`** -- Derive an artifact ID from type, platform, and compression settings
- **`extractPlatformsFromModuleView()`** -- Extract platform info from UoW manifest artifact paths
- **`inferPlatformFromPath()`** -- Infer OS/architecture from artifact file path patterns

## Patterns

- Transitive dependency resolution: recursively collects all dependencies for validation
- Metadata override merging: module-level metadata overrides type-level artifact definitions
- Platform inference: extracts OS/architecture from file path patterns (e.g., `linux-amd64`)
- UoW manifest integration: reads Unit-of-Work manifests to determine which artifacts were actually produced
- Cross-platform validation: supports validating artifacts built on different platforms than the current one

## Internal Structure

| File | Responsibility |
| --- | --- |
| artifact_helpers.go | Shared artifact utilities: formatting, book expansion, platform inference, and ID derivation |
| artifact_resolution.go | Artifact resolution types (`ResolvedArtifact`, `ArtifactResolutionSummary`) and resolution functions |
| artifact_validation.go | Artifact validation types (`ModuleValidationResult`, `ValidationResults`) and dependency traversal |
| effective_config.go | Effective module configuration with path variable substitution |

## Dependencies

- `contracts/core/0.1.0` -- action constants for UoW manifest reading
- `core/config` -- module configuration, artifact types, and build artifact resolution
- `core/domain/modules` -- module registry for dependency graph traversal
- `core/output` -- UoW manifest reading for artifact validation

## Role in System

The `impl/internal` package provides the artifact resolution and validation infrastructure shared by GET and SHOW commands. It bridges the configuration layer (module definitions, artifact patterns) with the filesystem layer (checking file existence, computing hashes) and the UoW manifest layer (tracking what was actually built).

## Code Health

### Tech Debt
- None identified

### Pain Points
- Platform inference relies on string pattern matching against file paths, which is fragile for non-standard naming

### Optimization Opportunities
- No urgent opportunities identified
