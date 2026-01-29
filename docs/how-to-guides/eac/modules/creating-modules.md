# Creating Modules

**Problem**: You want to add a new component to your repository with proper file ownership, dependencies, and build configuration.

**Solution**: Define a module contract in `.r2r/eac/repository.yml`.

## What is a Module?

A module is a logical unit of code with:

- **Moniker**: Unique identifier (e.g., `eac-core`, `my-service`)
- **Type**: Component type that provides defaults (e.g., `go-cli`, `go-library`, `mkdocs-site`)
- **Files**: Ownership boundaries (which files belong to this module)
- **Dependencies**: Other modules this depends on

## Choosing a Component Type

Select the appropriate type based on your module's purpose:

| Purpose               | Component Type  | Build Support                           | Test Support          |
| --------------------- | --------------- | --------------------------------------- | --------------------- |
| Go CLI application    | `go-cli`        | ✅ Full (cross-compile, version inject) | ✅ gotest, godog      |
| Go library            | `go-library`    | ✅ Go module                            | ✅ gotest, godog      |
| Go commands library   | `go-commands`   | ✅ Go module with CLI wrapper           | ✅ gotest, godog      |
| Documentation site    | `mkdocs-site`   | ✅ MkDocs HTML                          | ❌ No tests           |
| R2R extension         | `r2r-extension` | ✅ Docker buildx                        | Depends on container  |
| Configuration files   | `configuration` | ❌ No build                             | ❌ No tests           |

See [Component Types Reference](../../../reference/eac/architecture/component-types.md) for all available types and configuration options.

## Quick Start

Add a module to `.r2r/eac/repository.yml`:

```yaml
modules:
  - moniker: my-service
    name: My Service
    type: go-cli
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
    type: go-library          # Required: component type
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

    build:                    # Build configuration (defines executables)
      artifacts:
        - id: linux-amd64
          type: executable
          pattern: "my-service-linux-amd64"

    metadata:                 # Custom key-value data
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

## Component Types

Component types define build behavior and defaults for modules:

| Type            | Description                    | Capabilities                    |
| --------------- | ------------------------------ | ------------------------------- |
| `go-cli`        | Go CLI with cross-compilation  | go_module, executable           |
| `go-library`    | Go library package             | go_module                       |
| `go-commands`   | Go library with CLI wrapper    | go_module                       |
| `mkdocs-site`   | MkDocs documentation           | documentation                   |
| `r2r-extension` | Docker container extension     | container, buildx               |
| `configuration` | Configuration files            | none                            |

See available types:

```bash
r2r eac show component-types
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

| Problem            | Solution                                               |
| ------------------ | ------------------------------------------------------ |
| Overlapping files  | Use `exclude` patterns or adjust ownership             |
| Missing type       | Check component type in `component-types.yml`          |
| Build fails        | Check `build_deps` are available (docker, go)          |
| Wrong files listed | Verify glob patterns and root directory                |
