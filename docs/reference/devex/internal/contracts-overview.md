# Contracts Overview

Summary of the EAC contract system for monorepo contributors.

## What Are Contracts?

Contracts are YAML configuration files that define how the EAC system behaves. They are:

- **Version-controlled** - Tracked in Git like code
- **Schema-validated** - Checked against JSON schemas
- **Self-documenting** - Human-readable configuration

## Core Contracts

EAC uses six core contract files: `repository.yml`, `blueprints.yml`, `tool-config.yml`, `environments.yml`, `test-suites.yml`, and `testing-tags.yml`. User-specific configs live in `.eac/`; system defaults are in `contracts/eac-core/0.1.0/defaults/`.

For the complete contract field reference and schema details, see [Contracts Reference](../../eac/architecture/contracts.md).

## Contract Relationships

```text
repository.yml ──defines──> components
      │                          │
      │ depends_on               │ use types from
      ▼                          ▼
  (other modules)        blueprints.yml
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
        ├── blueprints.schema.json
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

### Create a New Component Kind

1. Copy `contracts/core/0.1.0/schemas/defaults/blueprints.yml` to `.eac/blueprints.yml`
2. Define new kind under `component-kinds` section with builder and file patterns
3. Reference the kind in module components

### Add Test Environment

1. Edit `.eac/environments.yml`
2. Define environment with level and type
3. Update test suites if needed

## Full Documentation

For complete contract reference with all fields and examples:

- [Contracts Reference](../../eac/architecture/contracts.md) - Full documentation
- [Modules](../../eac/modules/index.md) - Module system details
