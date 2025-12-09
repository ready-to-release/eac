# Contributing to EAC

{{ page_breadcrumb() }}

Technical reference for developers contributing to the EAC project.

## In This Section

| Document | Description |
|----------|-------------|
| [Creating Commands](./creating-commands.md) | Build new CLI commands for EAC |

## Overview

This section contains technical documentation for developers who want to extend or contribute to the EAC system itself. If you're looking to **use** EAC commands, see the [How-To Guides](../../../how-to-guides/eac/).

## Contribution Areas

### Command Development

Learn how to create new CLI commands:

- **Command Structure** - File organization and patterns
- **Help System Integration** - Structured comment headers
- **Flag Definitions** - Type-safe configuration
- **Registry Integration** - Auto-discovery mechanism

See: [Creating Commands](./creating-commands.md)

### Module System

Understanding the module architecture:

- **Module Contracts** - YAML schema and validation
- **Build Pipeline** - Lifecycle phases
- **Artifact Generation** - Output specifications
- **Dependency Resolution** - Graph traversal

### MCP Server Development

Extending the MCP server:

- **Tool Definitions** - Exposing commands via MCP
- **Protocol Handlers** - Request/response processing
- **Error Handling** - Consistent error reporting
- **Testing MCP Tools** - Validation strategies

### Core Libraries

Contributing to shared packages:

- **Repository Package** - File system operations
- **Contracts Package** - Schema definitions
- **AI Package** - LLM integrations
- **Pipeline Package** - Execution orchestration

## Development Guidelines

### Code Standards

- **Go Version**: ≥ 1.21
- **Formatting**: `gofmt`, `go vet`
- **Testing**: Table-driven tests required
- **Documentation**: Godoc comments for exported APIs

### Three Rules of Vibe Coding

All contributions must follow:

1. **Easy to Understand** - Clear, idiomatic Go
2. **Easy to Change** - Stable boundaries, minimal coupling
3. **Hard to Break** - Comprehensive tests, input validation

### Contribution Workflow

1. Fork and branch
2. Write specifications (BDD/Gherkin)
3. Write tests first (TDD)
4. Implement functionality
5. Validate all contracts
6. Submit pull request

## Related Documentation

- [Architecture Documentation](../../explanations/eac/) (Coming)
- [Module Reference](../modules/) (Coming)
- [Command Reference](../commands/)

{{ diataxis_footer() }}
