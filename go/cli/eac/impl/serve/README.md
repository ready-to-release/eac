# serve

Starts Docker-based servers for module build output, supporting static site serving, MkDocs live-reload development mode, Structurizr Lite design viewing, and Gource git history visualization. Handles container lifecycle with auto-rebuild, staleness detection, and browser opening.

## Key Types

- **`ModuleServeConfig`** -- Resolved serve configuration with module moniker, content path, and site flag
- **`DockerClient`** -- Wrapper for Docker container operations (start, stop, status, logs, staleness)
- **`ServeContext`** -- Execution-time configuration for servers with Docker images, ports, and watch settings
- **`ServeFunc`** -- Type alias for server function signature from the tool system
- **`servableItem`** -- Component that can be served, with name and site/PDF distinction

## Patterns

- Dual serve modes: live-reload development (MkDocs dev server) vs. full build then static serve
- Auto-staleness detection: checks container image and build output freshness, auto-restarts if stale
- Tool-system integration: server Docker images and configuration resolved from tool-config.yml
- Progressive readiness polling: waits for server HTTP 200 with exponential backoff before opening browser
- Module config resolution: identifies servable components by type (site-render, pdf-render, docs-site, book)

## Internal Structure

| File | Responsibility |
| --- | --- |
| serve.go | Command entry point, flag parsing, module config resolution, dev/build mode dispatch |
| docker.go | `DockerClient` wrapper for container lifecycle, staleness checks, and log streaming |

## Dependencies

- `contracts/core` -- action type constants and services options
- `adapters/docker` -- Docker container serve, stop, status, and browser operations
- `adapters/eac` -- EAC CLI port for triggering module builds
- `cli/eac/impl/design/helper` -- Structurizr Lite start/stop for serve design sub-command
- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration
- `clibase/services` -- service initialization (workspace, config, tools, logging)
- `core/config` -- EAC configuration and module lookup
- `core/environments` -- environment variable constants
- `core/logging` -- structured logging
- `core/output` -- UoW manifest reader for build staleness detection
- `core/paths` -- build output, design, and workspace path resolution
- `core/repository` -- repository root discovery
- `core/tool` -- tool definition lookup, serve bridge, and Docker image resolution

## Role in System

The `serve` package provides local server capabilities for `eac`, enabling developers to preview documentation sites, PDF outputs, architecture diagrams, and git history visualizations through Docker containers. Its dual-mode approach (live-reload for rapid iteration, static for production preview) and auto-staleness management make it the primary interface for local content viewing during development.

## Code Health

### Tech Debt
- No test for serve.go (804 lines)
- No test for docker.go (220 lines)
- No test for design/serve.go (191 lines)
- No test for gource/serve.go (540 lines)

### Pain Points
- serve.go (804 lines) exceeds 300-line guideline
- gource/serve.go (540 lines) exceeds 300-line guideline
- Only 3 test files for 9 non-test Go files across all subdirectories

### Optimization Opportunities
- None identified.
