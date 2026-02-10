# docker

Docker container runtime adapter that wraps the Docker SDK to provide
container lifecycle management, image operations, port allocation, and
serving capabilities.

## Key Types

- **`ContainerAdapter`** -- Implements `ContainerPort` using Docker SDK
- **`DockerClient`** -- Interface abstracting Docker SDK operations
- **`RealDockerClient`** -- Production wrapper around official Docker client
- **`SimpleMockDockerClient`** -- BDD-friendly mock with disk-persisted state
- **`ServeConfig`** -- Configuration for long-running container services
- **`ServeResult`** -- Result of starting a serve container
- **`RunConfig`** -- Configuration for one-shot container execution
- **`ImageOperation`** -- Describes an image to ensure (build or pull)

## Patterns

- Port interface: `ContainerAdapter` implements `ContainerPort` from contracts
- Factory function: `NewDockerClient()` selects real or mock client via env var
- Global singleton: `GlobalContainer()` with lazy init and double-checked locking
- Port reservation: Atomic find-and-reserve prevents race conditions on parallel start
- Retry with backoff: `Execute` retries on transient Windows/Docker errors

## Internal Structure

| File | Responsibility |
| --- | --- |
| client.go | `DockerClient` interface and client factory |
| container.go | `ContainerAdapter` implementing `ContainerPort` |
| serve.go | Long-running container management (start, stop, list) |
| run.go | One-shot container execution (`RunContainer`) |
| scan.go | OWASP ZAP security scanning via Docker |
| image.go | Parallel image ensure with concurrency control |
| ports_network.go | Port availability checking and range configuration |
| port_reservation.go | Atomic port reservation with TTL cleanup |
| browser.go | Platform-aware browser launch for served content |
| global.go | Global `ContainerPort` singleton management |
| simple_mock_client.go | Mock Docker client with JSON state persistence |
| doc.go | Package documentation |
| mocks/ | Testify-based mock client for unit tests |
| util/ | DinD path translation and volume formatting |

## Dependencies

- `contracts/container-runtime/0.1.0` -- `ContainerPort` interface
- `core/config` -- port reservation TTL configuration
- `core/environments` -- environment variable constants

## Role in System

The `docker-eac` module is the primary implementation of the
`ContainerPort` contract, enabling the core orchestrator to run builds,
tests, scans, and documentation servers in Docker containers. It handles
DinD path translation for CI environments and provides both one-shot
execution and long-running serve capabilities used by build, scan, and
serve commands.

## Code Health

### Tech Debt
- `DockerClient` interface in client.go has 13 methods; consider splitting into smaller role interfaces (e.g., ImageClient, ContainerClient)
- `executeOnce` in container.go (~175 lines) handles mounts, env, config, creation, attach, wait, and cleanup in one function
- Four package-level mutable variables: `globalContainer` (global.go), `reservedPorts` (port_reservation.go), `createDockerClient` (serve.go), `scanContainerCounter` (scan.go)

### Pain Points
- scan.go and browser.go have no unit tests; scan.go relies on env-var-gated mock path only
- `ensureImage` in serve.go writes directly to `os.Stdout` via `fmt.Printf`, bypassing any TUI or log-writer pattern used elsewhere

### Optimization Opportunities
- Extract `executeOnce` into smaller helpers (buildMounts, buildContainerConfig, waitForExit) for readability; low risk, purely structural
- Port reservation cleanup runs on every `ReservePort` call; a background goroutine with periodic sweep would reduce per-call overhead for high-concurrency scenarios
