# util

Docker utility functions for Docker-in-Docker (DinD) path translation, volume formatting, and daemon availability checking.

## Key Functions

- `IsDinD` -- Detects Docker-in-Docker mode using multiple signals (env vars, filesystem)
- `GetHostRepoRoot` -- Returns the host repository root from env var
- `GetContainerRepoRoot` -- Returns the container's view of the repository root
- `TranslatePathForMount` -- Translates container paths to host paths for Docker volume mounts in DinD
- `FormatDockerVolume` -- Converts Windows paths (C:\path) to Docker format (/c/path)
- `IsDockerAvailable` -- Checks if Docker daemon is accessible

## Patterns

- Multi-signal DinD detection: Checks `CLIE_HOST_REPOROOT`, `CLIE_DOCKER_MODE`, Windows path heuristics, and `/.dockerenv`
- Cross-platform path translation: Handles Windows-to-Linux and Linux-to-Windows path conversions for volume mounts
- Pure functions: Path translation functions have no side effects

## Internal Structure

| File    | Responsibility                                                                        |
| ------- | ------------------------------------------------------------------------------------- |
| dind.go | DinD detection, path translation, volume formatting, and Docker availability checking |

## Dependencies

- None (leaf package; imports only stdlib)

## Role in System

The util sub-package handles the complexity of Docker volume mounting when the EAC CLI itself runs inside a Docker container (DinD). It translates paths between the container's filesystem and the host's filesystem so that nested Docker containers can correctly mount volumes from the original host.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- None identified
