# extensions

Extension image installation orchestration wrapping the Docker container host.

## Key Types

- **`Installer`** -- Wraps `ContainerHost` to provide extension-aware image install operations

## Key Functions

- **`NewInstaller`** -- Creates Installer by constructing a new ContainerHost
- **`EnsureExtensionImage`** -- Ensures an extension image is locally available; returns true if image was updated
- **`InstallExtension`** -- Installs a single extension by name
- **`InstallAllExtensions`** -- Installs all configured extensions with success/failure tracking
- **`GetContainerHost`** -- Exposes the underlying ContainerHost for advanced operations

## Patterns

- Thin wrapper: Installer delegates all Docker operations to ContainerHost, adding extension-level semantics
- Update detection: Compares image ID and repo digest before/after pull to determine if an update occurred
- Resource cleanup: Close() propagates to ContainerHost for Docker client cleanup

## Internal Structure

| File         | Responsibility                                                    |
| ------------ | ----------------------------------------------------------------- |
| installer.go | Installer type, image ensure/install operations, update detection |

## Dependencies

- `internal/conf` -- Extension configuration from Global singleton
- `internal/docker` -- ContainerHost for Docker operations
- `internal/logging` -- Debug and error logging

## Role in System

The extensions package provides a higher-level abstraction over raw Docker operations for the install and run commands. It encapsulates the logic of finding an extension in config, checking if its image needs updating, and reporting whether a pull actually changed the local image. The cmd package uses Installer for both `clie install` and `clie run` flows.

## Code Health

### Tech Debt

- None identified.

### Pain Points

- No test files exist for this package. `installer.go` (137 lines) lacks corresponding unit tests.

### Optimization Opportunities

- None identified.
