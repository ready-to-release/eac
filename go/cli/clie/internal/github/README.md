# github

GitHub CLI authentication and GitHub Container Registry (GHCR) operations.

## Key Types

- **`CLIAuth`** -- Authentication credentials obtained from the GitHub CLI (`gh`)
- **`RegistryClient`** -- HTTP client for GHCR operations with dual API strategy
- **`Tag`** -- Container image tag with name and update timestamp
- **`ExtensionInfo`** -- Discovered extension metadata (name, description, image path)

## Key Functions

- **`GetCLIAuth`** -- Retrieves token via `gh auth token` and username via `gh auth status`
- **`NewRegistryClient`** -- Creates client supporting both authenticated and unauthenticated access
- **`ListTags`** -- Lists tags using GitHub API (private) with OCI Registry fallback (public)
- **`GetLatestStableTag`** -- Finds latest stable tag prioritizing SHA for extensions, then run-N, then semver
- **`ListExtensions`** -- Discovers available extensions by querying GHCR for `ext-*` packages

## Patterns

- Dual API strategy: GitHub API for private packages (requires token), OCI Registry API as fallback for public packages
- Anonymous token: When OCI Registry returns 401, automatically obtains anonymous bearer token and retries
- Tag prioritization: SHA tags preferred for extension pinning, then run-N tags, then semver tags
- Multi-source auth: Environment variables checked first, GitHub CLI as fallback

## Internal Structure

| File        | Responsibility                                                   |
| ----------- | ---------------------------------------------------------------- |
| auth.go     | CLIAuth type, GetCLIAuth, IsCLIAvailable, IsCLIAuthenticated    |
| registry.go | RegistryClient, ListTags (dual API), GetLatestStableTag, ListExtensions |

## Dependencies

- `internal/logging` -- Debug logging for API requests and fallback paths

## Role in System

The github package handles all interactions with GitHub's authentication and container registry systems. It provides credentials for Docker image pulls (used by the docker package's `CreateGitHubAuthConfig`) and registry browsing for the list and install commands. The dual API strategy ensures the CLI works both for private organizational packages and public community packages.

## Code Health

### Tech Debt

_None identified._

### Pain Points

_None identified._

### Optimization Opportunities

_None identified._
