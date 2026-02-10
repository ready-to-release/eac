# design

Provides the `serve design` command for launching Structurizr Lite as a Docker container to interactively view and edit architecture diagrams from `workspace.dsl` files.

## Key Types

None (command-only package).

## Key Functions

- **`ServeDesign()`** -- Entry point for the `serve design` command; launches Structurizr Lite via Docker with module selection
- **`formatModuleList()`** -- Format available modules for display in usage output
- **`printServeUsage()`** -- Display command usage including available modules

## Patterns

- `init()` registration: registers `ServeDesign` command function with the global registry
- Docker-based tool execution: runs Structurizr Lite container with workspace.dsl volume mount
- Interactive module selection: supports `--module` flag or defaults to auto-detected modules

## Internal Structure

| File | Responsibility |
| --- | --- |
| serve.go | Structurizr Lite Docker launch with module selection and workspace mounting |

## Dependencies

- `clibase/registry` -- command registration
- `core/logging` -- structured logging
- `design/helper` -- Structurizr workspace discovery and Docker management

## Role in System

The `design` sub-package provides interactive architecture diagram viewing within the `serve` command group. It enables developers to browse and edit Structurizr C4 model diagrams locally via a Docker-hosted Structurizr Lite instance.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
