# Repository Layout and Module Structure

## Overview

This repository is organized as a **modular monorepo** with clearly defined module boundaries using the EAC (Everything as Code) system. All modules are defined in `.eac/repository.yml`, which serves as the central contract for module ownership, dependencies, and build configuration.

**Authoritative documentation**:

- [Repository Layout](../../eac/architecture/repository-layout.md) - Full directory tree and file organization
- [EAC Overview](../../eac/index.md) - System overview and key capabilities
- [Modules](../../eac/modules/index.md) - Module system and dependency management
- [Contracts](../../eac/architecture/contracts.md) - Module contracts and configuration

## Adding a New Module

When adding a module to this repository:

1. Add an entry to `.eac/repository.yml` with `moniker`, `template`, `depends_on`, and `components`
2. Create the module directory at the path specified in `components.go.root` (or equivalent)
3. Run `eac validate` to verify the configuration

See [Modules Reference](../../eac/modules/index.md) for the complete module catalog and [Contracts Reference](../../eac/architecture/contracts.md) for all configuration fields.
