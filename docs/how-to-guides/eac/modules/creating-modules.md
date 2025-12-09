# Creating Modules

{{ page_breadcrumb() }}

**Problem**: You want to add a new component to your repository with proper file ownership, dependencies, and build configuration.

**Solution**: Define a module contract in `.r2r/eac/modules.yml`.

## What is a Module?

A module is a logical unit of code with:

- **Moniker**: Unique identifier (e.g., `eac-core`, `my-service`)
- **Type**: Classification that provides defaults (e.g., `go-library`, `go-cli`)
- **Files**: Ownership boundaries (which files belong to this module)
- **Dependencies**: Other modules this depends on

## Quick Start

Add a module to `.r2r/eac/modules.yml`:

```yaml
modules:
  - moniker: my-service
    name: My Service
    type: go-library
    description: Core business logic service
    files:
      root: src/my-service
```

Verify with:

```bash
r2r eac show modules           # List all modules
r2r eac get files my-service   # Show files owned by module
```

## Module Contract Structure

### Minimal Module

```yaml
modules:
  - moniker: my-module        # Required: unique identifier
    name: My Module           # Required: human-readable name
    type: go-library          # Required: module type (from module-types.yml)
    files:
      root: src/my-module     # Required: root directory
```

### Full Module

```yaml
modules:
  - moniker: my-service
    name: My Service
    type: go-cli
    description: API service for user management

    depends_on:               # Other modules this depends on
      - eac-core
      - eac-ai

    metadata:                 # Custom key-value data
      exe: my-service         # Binary name for go-cli types
      team: platform

    files:
      root: src/my-service

      # Override type defaults
      source:
        - "**/*.go"
        - "!**/*_test.go"
      config:
        - "config/*.yaml"
      tests:
        - "**/*_test.go"
        - "tests/**/*.go"
      assets:
        - "assets/**/*"
      exclude:
        - "vendor/**"

      # CI/CD workflows
      workflows:
        ci: .github/workflows/ci-my-service.yaml
        release: .github/workflows/release-my-service.yaml

      # Repository-level files (outside module root)
      repo:
        specs:
          - "specs/my-service/**/*.feature"
        design: "specs/my-service/.design/workspace.dsl"
        test_impl: "go/eac/specs/impl/my-service"
```

## Common Module Types

| Type | Description | Build Deps |
|------|-------------|------------|
| `go-library` | Go package (no binary) | go |
| `go-cli` | Go CLI with cross-platform builds | go |
| `go-commands` | Go library with CLI wrapper | go |
| `go-mcp` | Go MCP server | go |
| `r2r-extension` | r2r extension container | docker |
| `mkdocs-site` | Documentation site | docker |
| `scripts-package` | Shell scripts | none |
| `configuration` | Config files only | none |

See available types:

```bash
r2r eac show moduletypes
```

## Dependencies

Declare dependencies between modules:

```yaml
modules:
  - moniker: my-api
    depends_on:
      - eac-core      # Foundation library
      - my-models     # Data models
```

View dependency graph:

```bash
r2r eac show dependencies
r2r eac get dependencies --format=json
```

## File Ownership

Each module owns files matching its patterns:

```yaml
files:
  root: src/my-service    # Base directory

  # Patterns relative to root
  source:
    - "**/*.go"           # All Go files
    - "!**/*_test.go"     # Except tests
  tests:
    - "**/*_test.go"      # Test files
  config:
    - "*.yaml"            # Config files
    - "Dockerfile"
```

Check file ownership:

```bash
r2r eac show files                    # All files with owners
r2r eac show files-changed            # Changed files with owners
r2r eac get changed-modules           # Modules affected by changes
```

## Workflows

Link CI/CD workflows to modules:

```yaml
files:
  workflows:
    ci: .github/workflows/ci-my-service.yaml
    release: .github/workflows/release-my-service.yaml
```

## Repository-Level Files

Own files outside the module root:

```yaml
files:
  repo:
    specs:
      - "specs/my-service/**/*.feature"    # Gherkin specs
    design: "specs/my-service/.design/workspace.dsl"  # Architecture
    test_impl: "go/eac/specs/impl/my-service"  # Step definitions
    other:
      - "docs/my-service/**/*.md"          # Documentation
```

## Build and Test

After defining a module:

```bash
# Build
r2r eac build my-service

# Test
r2r eac test module my-service

# Validate contracts
r2r eac validate contracts
```

## Best Practices

1. **Unique monikers**: Use descriptive, lowercase names with hyphens
2. **Correct type**: Choose the type that matches your module's purpose
3. **Explicit dependencies**: Declare all module dependencies
4. **Clear boundaries**: Each file should belong to exactly one module
5. **Minimal scope**: Keep modules focused on a single responsibility

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Overlapping files | Use `exclude` patterns or adjust ownership |
| Missing type | Add type to `module-types.yml` or use existing |
| Build fails | Check `build_deps` match available system dependencies |
| Wrong files listed | Verify glob patterns and root directory |

## See Also

- [Creating Module Types](creating-module-types.md) - Define new module type templates

{{ diataxis_footer() }}
