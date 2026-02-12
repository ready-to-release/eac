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

| File                 | Responsibility                                                                   |
| -------------------- | -------------------------------------------------------------------------------- |
| constants.go         | Package doc, repository path constants, git configuration                        |
| constants_app.go     | Application configuration and behavior constants                                 |
| constants_build.go   | Build, execution, and system configuration constants                             |
| constants_ci.go      | CI/CD platform detection and GitHub Actions metadata constants                   |
| constants_debug.go   | Debug flags and logging configuration constants                                  |
| constants_docker.go  | Docker and container runtime configuration constants                             |
| constants_testing.go | Test infrastructure, BDD framework constants (separate block for mock constants) |
| runtime.go           | `RuntimeEnv`, `DetectRuntime`, `IsCI`, `IsDevBox`                                |
| artifacts_mode.go    | `ArtifactsMode` with platform-aware build target filtering                       |
| contracts.go         | `EnvironmentContract`, `Environment`, YAML loading and validation                |
| memory.go            | `GetSystemMemoryBytes`, PDF concurrency, container memory allocation             |
| memory_darwin.go     | macOS memory detection via `syscall.Sysctl`                                      |
| memory_linux.go      | Linux memory detection via `/proc/meminfo`                                       |
| memory_windows.go    | Windows memory detection via `kernel32.GlobalMemoryStatusEx`                     |

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
- None identified

### Pain Points
- As a foundation package imported by nearly everything, adding a new constant here triggers recompilation of a large portion of the module graph

### Optimization Opportunities
- None identified
