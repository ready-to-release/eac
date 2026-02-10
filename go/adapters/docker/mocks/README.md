# mocks

Testify-based mock implementation of the Docker client interface for unit testing.

## Key Types

- **`MockDockerClient`** -- Testify mock implementing `docker.DockerClient` with all 13 methods

## Patterns

- Testify mocking: Uses `mock.Mock` embedding for method stubbing and call assertions
- Interface compliance: Implements every method of `DockerClient` including image, container, and lifecycle operations
- Nil-safe returns: Methods guard against nil type assertions for optional return values

## Internal Structure

| File      | Responsibility                                                                      |
| --------- | ----------------------------------------------------------------------------------- |
| client.go | `MockDockerClient` with testify mock implementations for all `DockerClient` methods |

## Dependencies

- None (leaf package within the docker adapter; imports only Docker SDK types and testify)

## Role in System

The mocks sub-package provides a testify-based mock for `DockerClient` that is used in unit tests across the docker adapter. It enables testing of container lifecycle, image operations, and serve functionality without requiring a running Docker daemon.

## Code Health

### Tech Debt

- None

### Pain Points

- None identified

### Optimization Opportunities

- None identified
