# EAC Configuration

How to configure EAC for your project.

## Configuration Files

EAC uses YAML configuration files with a layered system:

**User configs** (`.r2r/eac/`):

| File             | Purpose              | Required |
| ---------------- | -------------------- | -------- |
| `repository.yml` | Module definitions   | Yes      |
| `books.yml`      | Book configuration   | No       |
| `ai-provider.yml`| AI provider settings | No       |

**System defaults** (`contracts/eac-core/0.1.0/defaults/`):

| File                  | Purpose                    |
| --------------------- | -------------------------- |
| `component-types.yml` | Component type definitions |
| `tool-config.yml`     | Tool configurations        |
| `environments.yml`    | Test environments          |
| `test-suites.yml`     | Test suite definitions     |

## Module Configuration

### Minimal Module

```yaml
# .r2r/eac/repository.yml
modules:
  - moniker: my-service
    type: go-cli
```

### Complete Module

```yaml
modules:
  - moniker: my-service
    name: My Service
    description: Main application service
    type: go-cli
    depends_on: [my-library]
    files:
      root: cmd/service
      source: ["**/*.go"]
      tests: ["**/*_test.go"]
    versioning:
      scheme: semver
      prefix: v
```

### Key Fields

| Field        | Required | Description                            |
| ------------ | -------- | -------------------------------------- |
| `moniker`    | Yes      | Unique identifier (kebab-case)         |
| `type`       | Yes      | Module type (go-cli, go-library, etc.) |
| `name`       | No       | Human-readable name                    |
| `depends_on` | No       | Module dependencies                    |
| `files`      | No       | File ownership patterns                |

## Available Module Types

| Type            | Purpose                |
| --------------- | ---------------------- |
| `go-cli`        | Go CLI application     |
| `go-library`    | Go library             |
| `go-commands`   | Go commands library    |
| `mkdocs-site`   | MkDocs documentation   |
| `r2r-extension` | R2R extension (Docker) |
| `configuration` | Configuration files    |

## Dependencies

Define build order with `depends_on`:

```yaml
modules:
  - moniker: core-lib
    type: go-library

  - moniker: utils-lib
    type: go-library
    depends_on: [core-lib]

  - moniker: my-app
    type: go-cli
    depends_on: [core-lib, utils-lib]
```

Build order is computed automatically via topological sort.

## File Ownership

Define which files belong to a module:

```yaml
files:
  root: go/myapp           # Base directory
  source: ["**/*.go"]      # Source patterns
  tests: ["**/*_test.go"]  # Test patterns
  exclude: ["**/vendor/**"] # Exclusions
```

**Rule**: Each file must belong to exactly one module.

Validate with:

```bash
r2r eac validate-module-files
```

## AI Provider Configuration

For AI-powered features:

### Using Environment Variables (Recommended)

```yaml
# .r2r/eac/ai-provider.yml
provider: claude-api
api_key_env: ANTHROPIC_API_KEY
model: claude-sonnet-4-5
```

Then set the environment variable:

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

### Using Personal Config (Local Only)

```yaml
# .r2r/eac/ai-provider.personal.yml (gitignored)
provider: claude-api
api_key: sk-ant-...
model: claude-sonnet-4-5
```

## Environment Configuration

Define test execution environments:

```yaml
# .r2r/eac/environments.yml
environments:
  - moniker: unit
    name: Unit Tests
    level: L0
    type: unit

  - moniker: integration
    name: Integration Tests
    level: L2
    type: docker
    system_deps: [docker]
```

## Test Suites

Group tests by purpose:

```yaml
# .r2r/eac/test-suites.yml
suites:
  - moniker: commit
    name: Commit Tests
    selectors:
      - any_of_tags: ["@L0", "@L1"]
        exclude_tags: ["@wip"]

  - moniker: full
    name: Full Test Suite
    selectors:
      - any_of_tags: ["@L0", "@L1", "@L2"]
```

## Configuration Precedence

Settings are merged in order (later overrides earlier):

1. Hardcoded defaults (in eac-core code)
2. System defaults (`contracts/eac-core/0.1.0/defaults/`)
3. User config (`.r2r/eac/*.yml`)
4. Personal config (`.r2r/eac/*.personal.yml`)

## Validation

Always validate after changes:

```bash
# Validate all contracts
r2r eac validate

# Specific validations
r2r eac validate-contracts      # Schema validation
r2r eac validate-dependencies   # Dependency references
r2r eac validate-module-files   # File ownership
```

## Common Configurations

### Go Monorepo

```yaml
modules:
  - moniker: pkg-core
    type: go-library
    files:
      root: pkg/core

  - moniker: pkg-api
    type: go-library
    depends_on: [pkg-core]
    files:
      root: pkg/api

  - moniker: cmd-server
    type: go-cli
    depends_on: [pkg-core, pkg-api]
    files:
      root: cmd/server
```

### Documentation Site

```yaml
modules:
  - moniker: docs
    type: mkdocs-site
    files:
      root: docs
      source: ["**/*.md"]
```

### Mixed Repository

```yaml
modules:
  - moniker: backend
    type: go-cli
    files:
      root: backend

  - moniker: docs
    type: mkdocs-site
    files:
      root: docs

  - moniker: config
    type: configuration
    files:
      root: .r2r
```

## See Also

- [Project Structure](./project-structure.md) - Directory organization
- [Contracts Reference](../../eac/contracts/index.md) - Full contract documentation
- [Module Types](../../eac/architecture/component-types.md) - All available types
