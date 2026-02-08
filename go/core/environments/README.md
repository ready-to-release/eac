# environments

Centralized environment variable constants, runtime detection (DevBox vs CI),
environment contract management, artifact mode control, and memory-based
resource calculation.

## Key Types

- **`RuntimeEnv`** -- CLI execution context enum (DevBox, CI)
- **`ArtifactsMode`** -- Artifact generation scope enum (All, Reduced)
- **`EnvironmentContract`** -- Parsed environments.yml with metadata and environment list
- **`Environment`** -- Test execution environment with moniker, level, type, and system deps
- **`Metadata`** -- Contract version and scope information

## Patterns

- Single source of truth: all application environment variable names defined as constants
- Runtime detection: `IsCI` and `IsDevBox` check CI/GITHUB_ACTIONS env vars
- Artifact mode: CI defaults to all targets, DevBox defaults to current platform only
- Memory-based concurrency: PDF export concurrency calculated from system RAM with turbo option
- Platform-aware memory: `getSystemMemoryBytes` has Darwin, Linux, and Windows implementations
- Contract validation: `ValidateContract` enforces required fields and valid level values

## Internal Structure

| File | Responsibility |
| --- | --- |
| constants.go | All `Env*` constants for application environment variables |
| runtime.go | `RuntimeEnv`, `DetectRuntime`, `IsCI`, `IsDevBox` |
| artifacts_mode.go | `ArtifactsMode` with platform-aware build target filtering |
| contracts.go | `EnvironmentContract`, `Environment`, YAML loading and validation |
| memory.go | `GetSystemMemoryBytes`, PDF concurrency, container memory allocation |
| memory_darwin.go | macOS memory detection via `syscall.Sysctl` |
| memory_linux.go | Linux memory detection via `/proc/meminfo` |
| memory_windows.go | Windows memory detection via `kernel32.GlobalMemoryStatusEx` |

## Dependencies

- `core/paths` -- `EACConfigPath` for locating environments.yml

## Role in System

This package is the foundation layer that nearly every other package imports.
The environment variable constants eliminate hardcoded strings across the
codebase. The runtime detection drives behavioral differences between local
development and CI execution. The memory detection feeds into the resource
pool system for capacity-aware parallel execution.

## Code Health

### Tech Debt
- `constants.go`: single `const` block with 100+ constants spanning test mocks, CI metadata, debug flags, and app config -- grouping into separate files by domain (e.g., `constants_ci.go`, `constants_mock.go`) would improve discoverability
- `memory_windows.go:29`: uses `unsafe.Sizeof` with `kernel32.GlobalMemoryStatusEx` -- platform-specific code that cannot be tested on other operating systems

### Pain Points
- As a foundation package imported by nearly everything, adding a new constant here triggers recompilation of a large portion of the module graph
- Mock-related constants (`EnvCLIEMock*`, 15+ entries) are mixed with production constants, blurring the boundary between test infrastructure and runtime config

### Optimization Opportunities
- Split constants into domain-scoped files to reduce diff noise and merge conflicts when multiple contributors add constants simultaneously (low effort)
- No TODO/FIXME markers found -- codebase is clean of deferred work items
