# Internal: Monorepo Development

Reference documentation for developers contributing to the EAC monorepo.

## Overview

The EAC repository is a modular monorepo containing:

- **CLIE CLI** - The core command-line framework
- **EAC CLI** - Automation tools (standalone + containerized)
- **Supporting Modules** - Shared libraries and infrastructure

## In This Section

| Topic                                             | Description                                 |
| ------------------------------------------------- | ------------------------------------------- |
| [CLI vs Extensions](./cli-vs-extensions.md)       | Two-tier architecture explained             |
| [Repository Layout](./repository-layout.md)       | Directory structure and module organization |
| [Contracts Overview](./contracts-overview.md)     | YAML contract system summary                |
| [Workflow Conventions](./workflow-conventions.md) | Git workflow for contributors               |

## Key Concepts

### Modular Monorepo

All modules are defined in `.eac/repository.yml`. Each module has:

- Unique moniker (identifier)
- Components (go, typescript, specs, etc.)
- File ownership patterns
- Dependencies on other modules

### Two-Tier Architecture

1. **CLIE CLI (Tier 1)** - Framework running on the host
2. **Extensions (Tier 2)** - Tools running in containers

### Contract-Driven Configuration

Everything is configured via YAML contracts validated against JSON schemas:

- `repository.yml` - Module definitions (in `.eac/`)
- `blueprints.yml` - Component kind definitions (system default)
- `environments.yml` - Test environments (system default)
- `test-suites.yml` - Test suite definitions (system default)

System defaults are in `contracts/core/0.1.0/schemas/defaults/`.

## Quick Commands

```bash
# View all modules
eac show modules

# Check dependencies
eac show dependencies

# Validate contracts
eac validate

# Build a module
eac build <module>

# Test a module
eac test <module>
```

## Getting Started

1. Read [Repository Layout](./repository-layout.md) to understand the structure
2. Review [CLI vs Extensions](./cli-vs-extensions.md) for architecture context
3. Check [Workflow Conventions](./workflow-conventions.md) before contributing

## See Also

- [Command Implementation](../../eac/architecture/command-implementation.md) - Add new EAC commands
- [Contracts Reference](../../eac/architecture/contracts.md) - Full contract documentation
- [Local Setup](../../../how-to-guides/local-setup/index.md) - Development environment setup
