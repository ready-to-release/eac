# servers

Provides serve context configuration and adapter registrations for Docker-based development servers (MkDocs, nginx static sites, Structurizr Lite). Acts as the bridge between serve commands and the tool bridge serve infrastructure.

## Key Types

- **`ServeContext`** -- Configuration struct holding Docker, port, browser, watch, and module settings for serve operations
- **`ServeFunc`** -- Type alias for serve handler functions (delegated from tool bridge)
- **`ServeOptions`** -- Type alias for serve option configuration (delegated from tool bridge)
- **`ServeResult`** -- Type alias for serve result reporting (delegated from tool bridge)

## Key Functions

- **`DefaultServeContext()`** -- Create a `ServeContext` with sensible defaults (auto port, browser open, watch enabled)
- **`mkdocsLiveServeAdapter()`** -- Serve adapter for MkDocs live-reload documentation server via Docker
- **`staticSiteServeAdapter()`** -- Serve adapter for nginx-based static site server via Docker
- **`structurizrServeAdapter()`** -- Serve adapter for Structurizr Lite architecture viewer via Docker

## Patterns

- Global singleton: `GlobalServeContext` provides shared configuration across serve commands
- Adapter pattern: each server type implements a serve adapter function registered with the tool bridge
- Delegation to tool bridge: type aliases and convenience functions delegate to `tool.GlobalServeBridge()`
- Docker-based serving: all server types run as Docker containers with volume mounts

## Internal Structure

| File | Responsibility |
| --- | --- |
| context.go | `ServeContext` struct with Docker/port/browser/watch configuration and `GlobalServeContext` global |
| registry.go | Type aliases and convenience functions delegating to `tool.GlobalServeBridge()` |
| mkdocs_live.go | MkDocs live-reload Docker serve adapter |
| static_site.go | nginx static site Docker serve adapter |
| structurizr.go | Structurizr Lite Docker serve adapter |

## Dependencies

- `adapters/tool` -- tool bridge infrastructure for serve handler registration
- `core/logging` -- structured logging

## Role in System

The `servers` package provides the concrete Docker-based server implementations that power the `serve` command group. It defines the shared `ServeContext` configuration and registers adapter functions for each server type (MkDocs, static sites, Structurizr), enabling the serve infrastructure to launch and manage different documentation and visualization servers.

## Code Health

### Tech Debt
- `GlobalServeContext` is a mutable global variable, making concurrent access unsafe and testing harder

### Pain Points
- None identified.

### Optimization Opportunities
- Replace `GlobalServeContext` with dependency injection or context passing (moderate effort, improves testability)
