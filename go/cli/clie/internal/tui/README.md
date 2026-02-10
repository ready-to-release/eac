# tui

Bubble Tea TUI model for Docker image pull progress display.

## Key Types

- **`model`** -- Bubble Tea model with spinner, image name, done/error state
- **`dockerPullMsg`** -- Message type carrying pull completion result

## Key Functions

- **`NewPullModel`** -- Creates a Bubble Tea model for pulling an extension image
- **`Auth`** -- Retrieves base64-encoded GitHub auth string via docker.CreateGitHubAuthConfig
- **`pullImage`** -- Performs the Docker image pull operation

## Patterns

- Bubble Tea architecture: Init/Update/View lifecycle for interactive terminal UI
- Async pull: Image pull runs in a background tea.Cmd, spinner animates while waiting
- Quit handling: Supports q, esc, ctrl+c for user cancellation

## Internal Structure

| File       | Responsibility                                                 |
| ---------- | -------------------------------------------------------------- |
| spinner.go | Bubble Tea model, Init/Update/View, pullImage, Auth helper     |

## Dependencies

- `internal/conf` -- Extension type for NewPullModel parameter
- `internal/docker` -- CreateGitHubAuthConfig for registry authentication
- `internal/logging` -- Fatal logging on auth failure

## Role in System

The tui package provides visual feedback during Docker image pulls. It is an alternative pull display mechanism alongside the docker package's `DisplayDockerProgress`. The spinner model shows an animated dot spinner with the extension name while the image downloads.

## Code Health

### Tech Debt

- `spinner.go:58-72` -- The `pullImage` function creates its own Docker client (`client.NewClientWithOpts`) instead of accepting a `DockerClient` interface or using the shared `ContainerHost`. This bypasses the mock-friendly interface used everywhere else in the codebase.

### Pain Points

_None identified._

### Optimization Opportunities

- Refactoring `pullImage` to accept a `DockerClient` interface would allow testing without a live Docker daemon and would be consistent with the rest of the docker package's architecture.
