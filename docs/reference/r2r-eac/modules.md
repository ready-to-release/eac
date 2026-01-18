# Modules

## Overview

The EAC module system provides **independently buildable, testable units** with explicit contracts, dependency management, and file ownership. Modules are defined in YAML contracts, validated against schemas, and built in topological order based on their dependencies.

**Key concepts**:

- **Modules** - Independently buildable units with explicit identity
- **Module Types** - Reusable templates with build/test behavior
- **Dependencies** - Explicit declaration enables topological build ordering
- **File Ownership** - Each file claimed by exactly one module

---

## Module Architectures

Each module includes C4 architecture diagrams documenting its design. View them interactively:

```bash
r2r eac serve-design
# Opens http://localhost:8080
```

**Design files in repository:**

- All modules: `specs/[module]/.design/workspace.dsl`
- View in GitHub: [specs/\*/​.design/](https://github.com/ready-to-release/eac/tree/main/specs)

See [Viewing Architecture](./viewing-architecture.md) for detailed instructions.

---

## Module Registry

**File**: `.r2r/eac/repository.yml`

Modules are registered with:

- **Moniker** - Unique identifier (e.g., `eac-core`)
- **Type** - Module type reference (e.g., `go-library`)
- **Dependencies** - Module dependencies (e.g., `depends_on: [logging-go]`)
- **File Ownership** - Glob patterns defining owned files

**Example**:

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

---

## Module Types

**File**: `.r2r/eac/module-types.yml`

Module types define:

- **Build Dependencies** - System dependencies (go, docker, npm)
- **Capabilities** - Type capabilities (executable, go_module, etc.)
- **Build Artifacts** - Expected artifacts and verification
- **Defaults** - Default values inherited by modules

### Go Family

| Type            | Purpose                                    | Artifacts                                     |
| --------------- | ------------------------------------------ | --------------------------------------------- |
| **go-cli**      | CLI application with cross-platform builds | Platform executables (linux, windows, darwin) |
| **go-library**  | Library package (no executable)            | Marker file                                   |
| **go-commands** | Library with CLI invoke wrapper            | Single executable                             |
| **go-mcp**      | MCP server module                          | MCP server executable                         |
| **go-tests**    | Test-only module (BDD infrastructure)      | None (tests only)                             |

### Example (go-cli)

```yaml
- name: go-cli
  description: "Go CLI application with cross-platform builds"
  build_deps: [go]
  capabilities: [go_module, executable, cross_compile]
  build:
    artifacts:
      - type: executable
        pattern: "{moniker}-{os}-amd64{ext}"
        platforms: [linux, windows, darwin]
        verify: current_platform
  defaults:
    files:
      source: ["**/*.go"]
      tests: ["**/*_test.go"]
```

### Other Families

**Docker**: `r2r-extension` - Container image with multi-platform builds

**Documentation**: `mkdocs-site`, `mkdocs-pdf` - MkDocs HTML/PDF generation

**Infrastructure**: `configuration`, `scripts-package`, `templates` - Non-buildable modules

---

## Dependency Management

### Dependency Graph

**Declaration**:

```yaml
modules:
  - moniker: logging-go
    type: go-library

  - moniker: eac-core
    type: go-library
    depends_on: [logging-go]

  - moniker: eac-commands
    type: go-commands
    depends_on: [eac-core]
```

**Build Order** (topological sort):

1. `logging-go`
2. `eac-core`
3. `eac-commands`

### Commands

```bash
# Show dependency graph
r2r eac show-dependencies

# Get execution order for specific modules
r2r eac get-execution-order eac-commands

# Validate dependencies
r2r eac validate-dependencies

# Check for circular dependencies
r2r eac validate-module-hierarchy
```

### Dependency Rules

- Dependencies must exist in `repository.yml`
- No circular dependencies allowed
- Topological sort determines build order
- Changed modules trigger rebuild of dependents

---

## File Ownership

### Glob Patterns

```yaml
files:
  root: go/eac/core        # Base directory
  source: ["**/*.go"]      # All .go files
  tests: ["**/*_test.go"]  # All test files
  exclude: ["**/vendor/**"] # Exclude vendor
```

**Pattern Variables**:

- `**` - Recursive match
- `*` - Single-level match
- `{specs_root}`, `{test_impl_root}` - Template variables

### Validation

**Rule**: Each file claimed by exactly one module

**Command**:

```bash
r2r eac validate-module-files
```

**Error Example**:

```text
❌ File 'go/util/helper.go' claimed by multiple modules:
  - eac-core (pattern: go/**/*.go)
  - util-lib (pattern: go/util/**/*.go)
```

**Query Commands**:

```bash
# Show all files with ownership
r2r eac show-files

# Show changed files with ownership
r2r eac show-files-changed

# Show staged files with ownership
r2r eac show-files-staged
```

---

## Module Lifecycle

### 1. Discovery

**Load contracts**:

```bash
r2r eac get-modules
```

**Output**: All modules from `repository.yml` with resolved dependencies

### 2. Build

**Build single module**:

```bash
r2r eac build <module>
```

**Build with dependencies**:

```bash
r2r eac build <module> --deps
```

**Build Flow**:

1. Load contracts from `.r2r/eac/`
2. Resolve dependencies
3. Topological sort
4. Build in dependency order
5. Verify artifacts
6. Update cache

### 3. Test

**Test single module**:

```bash
r2r eac test <module>
```

**Test suite** (multiple modules):

```bash
r2r eac test-suite unit
```

**Test Flow**:

1. Load test suites from `.r2r/eac/test-suites.yml`
2. Select tests by tags (e.g., `@L0`, `@L1`)
3. Run tests in parallel
4. Collect results in `out/test/`

### 4. Validation

**Schema validation**:

```bash
r2r eac validate-contracts
```

**Dependency validation**:

```bash
r2r eac validate-dependencies
```

**File ownership validation**:

```bash
r2r eac validate-module-files
```

### 5. Release

**Check pending releases**:

```bash
r2r eac release-pending <module>
```

**Generate changelog**:

```bash
r2r eac release-changelog <module>
```

**Create release**:

```bash
r2r eac release-this <module>
```

---

## Working with Modules

### Creating a New Module

**1. Add to repository.yml**:

```yaml
modules:
  - moniker: my-new-module
    name: My New Module
    type: go-library
    depends_on: [eac-core]
    files:
      root: go/my/module
      source: ["**/*.go"]
      tests: ["**/*_test.go"]
```

**2. Validate**:

```bash
r2r eac validate-contracts
r2r eac validate-module-files
```

**3. Build**:

```bash
r2r eac build my-new-module
```

### Modifying Dependencies

**1. Update repository.yml**:

```yaml
depends_on: [logging-go, config-go]  # Add config-go
```

**2. Validate**:

```bash
r2r eac validate-dependencies
r2r eac validate-module-hierarchy  # Check for cycles
```

**3. Rebuild**:

```bash
r2r eac build my-module --deps
```

### Finding Module Information

```bash
# List all modules
r2r eac show-modules

# Show dependency graph
r2r eac show-dependencies

# Show files owned by modules
r2r eac show-files

# Get module details (JSON)
r2r eac get-modules

# Get build dependencies
r2r eac get-build-deps my-module
```

---

## Build System

### Artifacts

Each module type defines expected artifacts:

**Executable (go-cli)**:

```yaml
artifacts:
  - type: executable
    pattern: "{moniker}-{os}-amd64{ext}"
    platforms: [linux, windows, darwin]
    verify: current_platform
```

**Marker (go-library)**:

```yaml
artifacts:
  - type: marker
    pattern: ".built"
    verify: existence
```

**Container Image (r2r-extension)**:

```yaml
artifacts:
  - type: image
    pattern: "{moniker}:latest"
```

### Artifact Location

**Build Artifacts**: `out/build/{module}/`

**Verification**:

```bash
r2r eac show-artifacts <module>
r2r eac validate-artifacts <module>
```

### Build Cache

**Location**: `.r2r/cache/build/`

**Invalidation**:

- File changes (detected via git)
- Dependency changes
- Force rebuild flag

**Usage**:

```bash
# Use cache (default)
r2r eac build my-module

# Force rebuild
r2r eac build my-module --force
```

### Incremental Builds

**Changed Modules Detection**:

```bash
# Detect changed modules since last successful CI
r2r eac get-changed-modules-ci

# Get affected modules (dependents)
r2r eac get-changed-modules --with-dependents
```

**CI Workflow**:

```bash
# 1. Detect changes
CHANGED=$(r2r eac get-changed-modules-ci)

# 2. Build affected modules
r2r eac build $CHANGED

# 3. Test affected modules
r2r eac test $CHANGED
```

---

## Module Designs (Architecture)

### Structurizr C4 Diagrams

**Location**: `specs/{module}/.design/workspace.dsl`

**Format**: C4 model DSL (System Context → Containers → Components)

### Viewing Designs

**Method 1: Browser (Recommended)**:

```bash
# Start Structurizr Lite server
r2r eac serve-design

# Access: http://localhost:8080

# Stop server
r2r eac serve-design --stop
```

**Method 2: VS Code Extension**:

- Install: `systemticks.c4-dsl-extension`
- Open: `specs/{module}/.design/workspace.dsl`

### Design Commands

```bash
# Validate design
r2r eac validate-design <module>

# Generate design (AI-powered)
r2r eac create-design <module>

# Update existing design (AI-powered)
r2r eac update-design <module>
```

### C4 Model Levels

| Level                 | Focus                                    | Audience   | Diagram Type  |
| --------------------- | ---------------------------------------- | ---------- | ------------- |
| **1. System Context** | System boundaries, external dependencies | Everyone   | systemContext |
| **2. Container**      | Major subsystems, applications           | Technical  | container     |
| **3. Component**      | Internal modules, packages               | Developers | component     |
| **4. Code**           | Classes, functions                       | Developers | (code only)   |

---

## Module Organization

### Core System Modules

| Module               | Type        | Purpose                                                                  |
| -------------------- | ----------- | ------------------------------------------------------------------------ |
| **eac-core**         | go-library  | Core libraries (contracts, repository, git)                              |
| **eac-commands**     | go-commands | Command implementations with integrated AI providers (Anthropic, OpenAI) |
| **eac-specs**        | go-library  | BDD test infrastructure (Godog)                                          |
| **eac-mcp-commands** | go-mcp      | MCP server (LLM tool integration)                                        |

### CLI and Extensions

| Module      | Type          | Purpose                              |
| ----------- | ------------- | ------------------------------------ |
| **r2r-cli** | go-cli        | CLI framework (Docker orchestration) |
| **ext-eac** | r2r-extension | EAC Docker extension image           |

### Libraries

| Module            | Type       | Purpose                       |
| ----------------- | ---------- | ----------------------------- |
| **logging-go**    | go-library | Structured logging (Go)       |
| **config-go**     | go-library | Configuration management (Go) |
| **validation-go** | go-library | Validation utilities (Go)     |

### Documentation

| Module        | Type        | Purpose                   |
| ------------- | ----------- | ------------------------- |
| **docs-site** | mkdocs-site | MkDocs documentation site |

---

## Module Commands Reference

### Discovery

```bash
r2r eac show-modules               # Module table
r2r eac show-dependencies          # Dependency graph
r2r eac show-files                 # File ownership
r2r eac show-component-types        # Component type table
r2r eac get-modules                # Modules JSON
r2r eac get-dependencies           # Dependencies JSON
```

### Build

```bash
r2r eac build <module>             # Build single module
r2r eac build <module> --deps      # Build with dependencies
r2r eac get-artifacts <module>     # List artifacts
r2r eac show-artifacts <module>    # Show artifact status
r2r eac show-build-summary <module> # Build summary
```

### Test

```bash
r2r eac test <module>              # Test single module
r2r eac test-suite <suite>         # Run test suite
r2r eac test-debug                 # Debug test failures
r2r eac show-test-summary <module> # Test summary
```

### Validation

```bash
r2r eac validate                   # Validate all
r2r eac validate-contracts         # Schema validation
r2r eac validate-dependencies      # Dependency validation
r2r eac validate-module-files      # File ownership validation
r2r eac validate-module-hierarchy  # Circular dependency check
r2r eac validate-artifacts <module> # Artifact validation
```

### Design

```bash
r2r eac validate-design <module>   # Validate Structurizr DSL
r2r eac create-design <module>     # Generate design (AI)
r2r eac update-design <module>     # Update design (AI)
r2r eac serve-design               # Serve designs in browser
```

### Release

```bash
r2r eac release-changelog <module> # Generate changelog
r2r eac release-pending <module>   # Check pending releases
r2r eac release-this <module>      # Create release
```

---

## Module Best Practices

### Module Granularity

- **Small, focused modules** - Single responsibility
- **Clear dependencies** - Explicit, minimal coupling
- **Shared libraries** - Extract common code into libraries
- **Avoid circular dependencies** - Refactor to break cycles

### Naming Conventions

- **Moniker**: kebab-case (e.g., `eac-core`)
- **Type prefix**: Language or tech (e.g., `go-library`)
- **Descriptive names**: Clear purpose (e.g., `logging-go`)

### File Organization

- **Module root**: `go/{namespace}/{module}/`
- **Source**: `**/*.go`
- **Tests**: `**/*_test.go`
- **Specs**: `specs/{module}/`
- **Design**: `specs/{module}/.design/`

### Dependency Management

- **Minimize dependencies** - Only depend on what you need
- **Layer architecture** - Libraries → Commands → CLI
- **Shared contracts** - Use eac-core for shared types
- **Avoid deep chains** - Flatten dependency trees

---

## Related Documentation

- [Overview](./index.md) - System overview and key concepts
- [Architecture](./architecture.md) - System architecture and components
- [Contracts](./contracts.md) - Contract system and YAML schemas
