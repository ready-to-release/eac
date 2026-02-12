# docs/helper

Provides a Docker client wrapper for managing MkDocs live-reload containers used by the documentation serve workflow. Handles container lifecycle (start, stop, status check, log streaming), MkDocs configuration generation, and cross-platform browser opening.

## Key Types

- **`Client`** -- High-level facade for MkDocs container operations backed by a `docker.DockerClient`
- **`ContainerInfo`** -- Running container metadata with name, URL, and host port

## Patterns

- Facade over Docker adapter: `Client` delegates all container operations to functions in `adapters/docker`
- MkDocs config generation: generates `mkdocs.yml` from site template with relative `docs_dir` paths for Docker volume mounts
- Cross-platform browser detection: uses `cmd /c start` on Windows, `open` on macOS, `xdg-open` on Linux
- DinD-aware browser fallback: `OpenBrowserWithFallback` skips browser opening when running inside Docker-in-Docker
- Container image defaults: hardcoded `mkdocs-render-oci` container name, image, and Dockerfile path

## Internal Structure

| File | Responsibility |
| --- | --- |
| types.go | `ContainerInfo` type definition |
| client.go | `Client` facade with start, stop, status, log streaming, and browser opening |
| container.go | Docker container lifecycle: start with config generation, stop, status check, log streaming |
| browser.go | Cross-platform browser detection and URL opening |

## Dependencies

- `adapters/docker` -- Docker container serve, stop, status, log streaming, and browser fallback operations
- `cli/eac/impl/build/builders/mkdocs` -- MkDocs config writing from template options
- `core/logging` -- structured logging
- `core/paths` -- serve output path and MkDocs config path resolution
- `core/repository` -- repository root discovery

## Role in System

The `docs/helper` package encapsulates MkDocs Docker container management for the `update docs` and documentation serve commands in `eac`. It generates MkDocs configuration, launches live-reload containers with proper volume mounts, and provides browser integration, serving as the shared infrastructure between documentation update and interactive documentation preview workflows.

## Code Health

### Tech Debt
- No unit tests for container.go (226 lines), client.go (87 lines), or browser.go (59 lines); only BDD-level tests exist
- Hardcoded container name and image constants in container.go

### Pain Points
- browser.go (59 lines) uses platform-specific command detection without testable abstraction

### Optimization Opportunities
- None identified
