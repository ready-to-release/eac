# Contracts Overview

Summary of the EAC contract system for monorepo contributors.

## What Are Contracts?

Contracts are YAML configuration files that define how the EAC system behaves. They are:

- **Version-controlled** - Tracked in Git like code
- **Schema-validated** - Checked against JSON schemas
- **Self-documenting** - Human-readable configuration

## Core Contracts

| Contract         | File               | Purpose                                          |
| ---------------- | ------------------ | ------------------------------------------------ |
| **Repository**   | `repository.yml`   | Module definitions, dependencies, file ownership |
| **Module Types** | `module-types.yml` | Type templates with build/test behavior          |
| **Environments** | `environments.yml` | Test execution environments (L0-L4)              |
| **Test Suites**  | `test-suites.yml`  | Test suites with tag selectors                   |
| **Testing Tags** | `testing-tags.yml` | Valid test tag definitions                       |

**Location**: All contracts in `.r2r/eac/`

## Contract Relationships

```text
repository.yml ──references──> module-types.yml
      │                              │
      │ depends_on                   │ requires
      ▼                              ▼
  (other modules)          system-dependencies.yml

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
      root: go/eac/core
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
r2r eac validate
```

## Schema Location

JSON schemas live in `contracts/eac-core/0.1.0/`:

```text
contracts/
└── eac-core/
    └── 0.1.0/
        ├── repository.schema.json
        ├── module-types.schema.json
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

1. Edit `.r2r/eac/repository.yml`
2. Add module definition with moniker and type
3. Run `r2r eac validate` to verify

### Create a New Module Type

1. Edit `.r2r/eac/module-types.yml`
2. Define type with capabilities and defaults
3. Reference the type in modules

### Add Test Environment

1. Edit `.r2r/eac/environments.yml`
2. Define environment with level and type
3. Update test suites if needed

## Full Documentation

For complete contract reference with all fields and examples:

- [Contracts Reference](../../eac/contracts/) - Full documentation
- [Modules](../../eac/modules/) - Module system details
