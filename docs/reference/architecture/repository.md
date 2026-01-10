# repository

The `repository` module defines the contract system and validation rules for the EAC repository structure. It ensures consistency across all modules and enforces architectural constraints.

## System Context

Shows how the repository module governs the overall repository structure.

<!-- structurizr:repository:SystemContext -->

## Container Architecture

High-level view of the repository contract system.

<!-- structurizr:repository:Containers -->

## Component Architecture

### Contracts Components

Contract definitions and schema validation.

<!-- structurizr:repository:ContractsComponents -->

### Validation Components

Repository-wide validation and consistency checks.

<!-- structurizr:repository:ValidationComponents -->

## Design File

- **Location**: `specs/repository/.design/workspace.dsl`
- **Interactive**: `r2r eac serve-design --module repository`

## Contract Types

| Contract | Purpose |
|----------|---------|
| `repository.yml` | Global repository configuration |
| `modules.yml` | Module registry and metadata |
| `books.yml` | Documentation book definitions |
| `environments.yml` | Environment configurations |

## Validation Rules

The repository module enforces:

- Module dependency hierarchy (no circular dependencies)
- File ownership (every file belongs to exactly one module)
- Contract schema compliance
- Changelog format consistency
- Test tag validity
