# Module Reference

This section provides detailed architecture documentation for each module in the EAC ecosystem. Each module's architecture is defined using the [C4 model](https://c4model.com/) via Structurizr DSL.

## Architecture Diagram Levels

The C4 model provides four levels of abstraction:

1. **System Context** - Shows how the module fits into the overall system landscape
2. **Container** - Shows the high-level technical building blocks
3. **Component** - Shows how containers are made up of components
4. **Code** - Shows how components are implemented (not typically documented)

## Core Modules

| Module                                                                 | Description                  |
| ---------------------------------------------------------------------- | ---------------------------- |
| [eac-commands](../../eac/modules/eac-commands.md)         | CLI command implementations  |
| [eac-core](../../eac/modules/eac-core.md)                 | Core libraries and contracts |
| [eac-mcp-commands](../../eac/modules/eac-mcp-commands.md) | MCP server integration       |
| [r2r-cli](../../r2r/architecture/module.md)                            | Ready-to-Release CLI         |

## Infrastructure Modules

| Module                                               | Description                         |
| ---------------------------------------------------- | ----------------------------------- |
| [ext-eac](../../eac/modules/ext-eac.md) | Docker extension image              |
| [repository](repository.md)                          | Repository contracts and validation |

## Supporting Modules

| Module                                                                   | Description                          |
| ------------------------------------------------------------------------ | ------------------------------------ |
| [Supporting Modules](../../eac/modules/index.md) | docs, templates, r2r-installer, etc. |

## Viewing Diagrams Interactively

For interactive exploration of architecture diagrams:

```bash
# Start Structurizr Lite for a specific module
r2r eac serve-design --module eac-commands
```

## Design File Locations

All architecture definitions are stored in Structurizr DSL format:

```text
specs/[module-name]/.design/workspace.dsl
```

## Updating Architecture Diagrams

When workspace.dsl files change, regenerate the cached SVGs:

```bash
# Update all modules
r2r eac update structurizr

# Update specific module
r2r eac update structurizr --module eac-commands
```
