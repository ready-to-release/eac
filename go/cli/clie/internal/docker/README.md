# docker

Docker client abstraction, container lifecycle management, image pulling, and environment configuration.

## Key Types

- **`ContainerHost`** -- Central type managing Docker client, context, root directory, and lazy connectivity
- **`DockerClient`** -- Interface abstracting Docker SDK operations for testability
- **`RealDockerClient`** -- Production implementation wrapping the official Docker SDK client
- **`MockDockerClient`** -- Testify mock implementation of DockerClient
- **`ContainerMode`** -- Enum distinguishing run mode from interactive mode
- **`ExtensionConfig`** -- Resolved extension configuration passed to container operations
- **`ContainerCleanupOptions`** -- Options for container cleanup (extensions-only, age filter, dry-run)
- **`CleanupResult`** -- Result of cleanup operation with counts and errors
- **`AnsiFilter`** -- Writer wrapper that strips problematic terminal escape sequences

## Key Functions

- **`NewContainerHost`** -- Creates ContainerHost with lazy Docker connectivity (~800ms savings)
- **`EnsureImageExists`** -- Multi-strategy image pull: AutoDetect, Always, IfNotPresent, Never
- **`CreateContainerConfig`** -- Builds Docker container.Config from extension settings
- **`CreateHostConfig`** -- Builds Docker container.HostConfig with mounts, ports, resources
- **`CreateGitHubAuthConfig`** -- Multi-source auth: env vars, then GitHub CLI fallback
- **`IsRunningInContainer`** -- Detects Docker-in-Docker via .dockerenv, cgroup, env vars
- **`DisplayDockerProgress`** -- Parses Docker JSON pull stream for user-facing progress

## Patterns

- Lazy connectivity: `EnsureConnected()` uses `sync.Once` to defer Docker Ping until first real operation
- Client interface: `DockerClient` interface enables unit testing without Docker daemon
- Docker-in-Docker: Host path propagation via `CLIE_HOST_REPOROOT` and `CLIE_CONTAINER_REPOROOT`
- Metadata caching: Extension metadata cached with digest-based invalidation to avoid repeated container exec
- Parallel image pull: Semaphore-based concurrency control for pulling multiple extension images

## Internal Structure

| File                  | Responsibility                                                      |
| --------------------- | ------------------------------------------------------------------- |
| hosting.go            | ContainerHost constructor, EnsureConnected, ValidateExtensions, FindExtension |
| hosting-lifecycle.go  | CreateContainerConfig, CreateHostConfig, CreateContainer, StartContainer, AttachToContainer, WaitForContainer, StopContainer, GetContainerSnapshot, WarnAboutNewContainers |
| hosting-image.go      | InspectImage, GetImageDigest, CreateGitHubAuthConfig, EnsureImageExists, cacheImageDigest |
| hosting-env.go        | BuildEnvironmentVars, CI detection, color settings, terminal dimension passthrough |
| hosting-metadata.go   | ExecuteMetadataCommand, GetExtensionMetadata, parseExtensionMetadata, MergeMetadataEnv |
| hosting-cleanup.go    | ContainerCleanupOptions, CleanupResult, CleanupContainers and convenience wrappers |
| cleanup.go            | IsRunningInContainer, CleanupChildContainers, CleanupOrphanedContainers, helper string functions |
| client_interface.go   | DockerClient interface definition                                   |
| client_real.go        | RealDockerClient wrapping official Docker SDK                       |
| client_mock.go        | MockDockerClient with testify mock                                  |
| image_parallel.go     | ParallelEnsureImages with semaphore concurrency                     |
| progress.go           | DisplayDockerProgress parsing Docker JSON pull stream               |
| ansi_filter.go        | AnsiFilter stripping problematic terminal escape sequences          |

## Dependencies

- `internal/cache` -- Registry and metadata caching
- `internal/conf` -- Global config and extension definitions
- `internal/github` -- GitHub CLI auth fallback for registry operations
- `internal/logging` -- Debug, info, warning, and error logging
- `internal/envconsts` -- Environment variable constant names

## Role in System

The docker package is the execution engine of the clie CLI. It handles everything from authenticating with GHCR, pulling images, creating and configuring containers, attaching I/O streams, to cleaning up after execution. It supports both standard run mode (non-interactive command execution) and interactive mode (shell-attached sessions). Docker-in-Docker is supported through host path propagation for volume mounts.

## Code Health

### Tech Debt

- `hosting-image.go:210-221` -- Duplicates Ping logic that already exists in `EnsureConnected()`. The `EnsureImageExists` function calls `ch.client.Ping()` directly instead of using `ch.EnsureConnected()`, resulting in redundant Docker service error handling.
- `cleanup.go:167-178` -- Reimplements `strings.Contains` and `strings.LastIndex` as custom `contains`, `containsSubstring`, and `lastIndex` functions. These should use the standard library.

### Pain Points

- `hosting-image.go:99-261` -- `EnsureImageExists` is 162 lines with deeply nested conditionals for the AutoDetect pull policy strategy.
- The package has 13 non-test source files, making it the largest package in the clie module.

### Optimization Opportunities

- The custom string helper functions in `cleanup.go` (`contains`, `containsSubstring`, `lastIndex`, `splitLines`) should be replaced with `strings.Contains`, `strings.LastIndex`, and `strings.Split` from the standard library.
