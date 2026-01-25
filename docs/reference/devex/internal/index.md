# Internal: Monorepo Development

Reference documentation for developers contributing to the EAC monorepo.

## Overview

The EAC repository is a modular monorepo containing:

- **R2R CLI** - The core command-line framework
- **EAC Extension** - Containerized automation tools
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

All modules are defined in `.r2r/eac/repository.yml`. Each module has:

- Unique moniker (identifier)
- Type (from `module-types.yml`)
- File ownership patterns
- Dependencies on other modules

### Two-Tier Architecture

1. **R2R CLI (Tier 1)** - Framework running on the host
2. **Extensions (Tier 2)** - Tools running in containers

### Contract-Driven Configuration

Everything is configured via YAML contracts validated against JSON schemas:

- `repository.yml` - Module definitions
- `module-types.yml` - Type templates
- `environments.yml` - Test environments
- `test-suites.yml` - Test suite definitions

## Quick Commands

```bash
# View all modules
r2r eac show-modules

# Check dependencies
r2r eac show-dependencies

# Validate contracts
r2r eac validate

# Build a module
r2r eac build <module>

# Test a module
r2r eac test <module>
```

## Getting Started

1. Read [Repository Layout](./repository-layout.md) to understand the structure
2. Review [CLI vs Extensions](./cli-vs-extensions.md) for architecture context
3. Check [Workflow Conventions](./workflow-conventions.md) before contributing

## See Also

- [Creating Commands](../../eac/development/creating-commands.md) - Add new EAC commands
- [Contracts Reference](../../eac/contracts/) - Full contract documentation
- [Local Setup](../../../how-to-guides/local-setup/) - Development environment setup
