# Contracts Overview

Summary of the EAC contract system for monorepo contributors.

## What Are Contracts?

Contracts are YAML configuration files that define how the EAC system behaves. They are:

- **Version-controlled** - Tracked in Git like code
- **Schema-validated** - Checked against JSON schemas
- **Self-documenting** - Human-readable configuration

## Core Contracts

| Contract            | File                  | Location                  | Purpose                                          |
| ------------------- | --------------------- | ------------------------- | ------------------------------------------------ |
| **Repository**      | `repository.yml`      | `.eac/`               | Module definitions, dependencies, file ownership |
| **Component Types** | `component-types.yml` | `contracts/.../defaults/` | Component type definitions with build behavior   |
| **Tool Config**     | `tool-config.yml`     | `contracts/.../defaults/` | Tool definitions and resources                   |
| **Environments**    | `environments.yml`    | `contracts/.../defaults/` | Test execution environments (L0-L4)              |
| **Test Suites**     | `test-suites.yml`     | `contracts/.../defaults/` | Test suites with tag selectors                   |
| **Testing Tags**    | `testing-tags.yml`    | `contracts/.../defaults/` | Valid test tag definitions                       |

**Location**: User configs in `.eac/`, system defaults in `contracts/eac-core/0.1.0/defaults/`

## Contract Relationships

```text
repository.yml ──defines──> components
      │                          │
      │ depends_on               │ use types from
      ▼                          ▼
  (other modules)        component-types.yml
                                 │
                                 │ use tools from
                                 ▼
                           tool-config.yml

test-suites.yml ──selects──> testing-tags.yml
      │
      │ runs in
      ▼
environments.yml
```

## Module Contract Example

**Minimal** (only 2 required fields):

```yaml
modules:
  - moniker: my-module
    type: go-library
```

**Complete**:

```yaml
modules:
  - moniker: eac-core
    name: EAC Core Libraries
    type: go-library
    depends_on: [logging-go]
    files:
      root: go/core
      source: ["**/*.go"]
      tests: ["**/*_test.go"]
```

## Validation

Contracts are validated at multiple levels:

| Level          | What It Checks         | Command                     |
| -------------- | ---------------------- | --------------------------- |
| **Schema**     | YAML structure         | `validate-contracts`        |
| **References** | Dependencies exist     | `validate-dependencies`     |
| **Files**      | Ownership is exclusive | `validate-module-files`     |
| **Hierarchy**  | No circular deps       | `validate-module-hierarchy` |

Run all validations:

```bash
eac validate
```

## Schema Location

JSON schemas live in `contracts/eac-core/0.1.0/`:

```text
contracts/
└── eac-core/
    └── 0.1.0/
        ├── repository.schema.json
        ├── component-types.schema.json
        ├── environments.schema.json
        └── ...
```

## IDE Support

Add schema reference for auto-completion:

```yaml
# yaml-language-server: $schema=../../contracts/eac-core/0.1.0/repository.schema.json
modules:
  - moniker: |  # IDE provides auto-completion here
```

## Common Tasks

### Add a New Module

1. Edit `.eac/repository.yml`
2. Add module definition with moniker and type
3. Run `eac validate` to verify

### Create a New Component Type

1. Copy `contracts/eac-core/0.1.0/defaults/component-types.yml` to `.eac/component-types.yml`
2. Define new type with builder and file patterns
3. Reference the type in module components

### Add Test Environment

1. Edit `.eac/environments.yml`
2. Define environment with level and type
3. Update test suites if needed

## Full Documentation

For complete contract reference with all fields and examples:

- [Contracts Reference](../../eac/architecture/contracts.md) - Full documentation
- [Modules](../../eac/modules/index.md) - Module system details
