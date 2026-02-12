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

- None identified.

### Pain Points

- No test files exist for this package. `spinner.go` lacks corresponding unit tests for the Bubble Tea model.

### Optimization Opportunities

- None identified.
