# Contracts

{{ page_breadcrumb() }}

## Overview

The EAC contract system defines all configuration via **YAML contracts validated against JSON schemas**. Configuration, architecture, and behavior are version-controlled and validated before execution.

**Principle**: Configuration over Convention - Everything is explicit, not implicit.

**Components**:

1. **Contract Files** - YAML configuration in `.r2r/eac/*.yml`
2. **JSON Schemas** - Validation rules in `contracts/**/*.schema.json`
3. **Validation Engine** - Runtime validation in `eac-core`

---

## All Contracts

| Contract | File | Purpose |
|----------|------|---------|
| **Modules** | `modules.yml` | Module definitions, dependencies, file ownership |
| **Module Types** | `module-types.yml` | Type templates with build/test behavior |
| **Environments** | `environments.yml` | Test execution environments (L0-L4) |
| **Test Suites** | `test-suites.yml` | Test suites with tag selectors |
| **Testing Tags** | `testing-tags.yml` | Valid test tag definitions |
| **Books** | `books.yml` | Documentation book configuration |
| **Repository** | `repository.yml` | Repository metadata |
| **System Deps** | `system-dependencies.yml` | Required system tools (go, docker, etc.) |
| **Security Tools** | `security-tools.yml` | Security scanning tool configuration |
| **AI Config** | `ai-config.yml` | AI provider configuration |
| **EAC Config** | `eac-config.yml` | EAC system configuration |
| **Logging** | `logging.yml` | Logging configuration |

**Location**: All contracts in `.r2r/eac/`, validated against schemas in `contracts/`

---

## Modules Contract

**File**: `.r2r/eac/modules.yml`

**Purpose**: Central module registry defining module identities, dependencies, and file ownership

### Structure

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
    description: Core domain libraries for contract management
    type: go-library
    depends_on: [logging-go]
    files:
      root: go/eac/core
      source: ["**/*.go"]
      tests: ["**/*_test.go"]
      exclude: ["**/*_old.go"]
    versioning:
      scheme: semver
      prefix: v
    metadata:
      owner: platform-team
      criticality: high
```

### Key Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `moniker` | string | ✅ | Unique module identifier (kebab-case) |
| `type` | string | ✅ | Module type reference (from module-types.yml) |
| `name` | string | ❌ | Human-readable name |
| `description` | string | ❌ | Module purpose |
| `depends_on` | array | ❌ | Module dependencies (monikers) |
| `files` | object | ❌ | File ownership patterns (glob) |
| `versioning` | object | ❌ | Versioning configuration (semver/calver) |
| `metadata` | object | ❌ | Custom key-value pairs |

### File Ownership

```yaml
files:
  root: go/eac/core        # Base directory
  source: ["**/*.go"]      # Source patterns (glob)
  tests: ["**/*_test.go"]  # Test patterns
  exclude: ["**/vendor/**"] # Exclusions
```

**Validation Rule**: Each file must be claimed by exactly one module

**Command**: `r2r eac validate-module-files`

### Dependencies

**Example**:

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

**Build Order** (topological sort): logging-go → eac-core → eac-commands

**Validation**: `r2r eac validate-module-hierarchy` checks for circular dependencies

---

## Module Types Contract

**File**: `.r2r/eac/module-types.yml`

**Purpose**: Define reusable module type templates with build configuration, capabilities, and defaults

### Structure

**Minimal**:

```yaml
types:
  - name: my-type
    description: "My custom module type"
```

**Complete**:

```yaml
types:
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
      repo:
        specs: ["{specs_root}/{moniker}/**"]
```

### Key Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | ✅ | Unique type identifier (kebab-case) |
| `description` | string | ✅ | Human-readable description |
| `build_deps` | array | ❌ | System dependencies (go, docker, npm) |
| `capabilities` | array | ❌ | Type capabilities (executable, go_module, etc.) |
| `build` | object | ❌ | Build artifacts configuration |
| `defaults` | object | ❌ | Default values inherited by modules |
| `docker_build` | object | ❌ | Docker build config (r2r-extension only) |

### Capabilities

| Capability | Description |
|------------|-------------|
| `go_module` | Go module (workspace member) |
| `executable` | Produces executable binary |
| `cross_compile` | Supports cross-platform compilation |
| `container` | Docker container image |
| `documentation` | Documentation generation |

### Build Artifacts

```yaml
build:
  artifacts:
    - type: executable          # Artifact type
      pattern: "{moniker}-{os}-amd64{ext}"  # Path pattern
      platforms: [linux, windows, darwin]   # Target platforms
      verify: current_platform  # Verification mode
```

**Artifact Types**: `executable`, `library`, `marker`, `image`, `site`

**Pattern Variables**: `{moniker}`, `{os}`, `{arch}`, `{ext}`, `{root}`

### Module Type Families

**Go Family**:

- `go-cli` - CLI application with cross-platform builds
- `go-library` - Library package (no executable)
- `go-commands` - Library with CLI invoke wrapper
- `go-mcp` - MCP server module
- `go-tests` - Test-only module

**Docker Family**:

- `r2r-extension` - R2R extension packaged as container

**Documentation Family**:

- `mkdocs-site` - MkDocs HTML documentation
- `mkdocs-pdf` - MkDocs PDF generation

**Infrastructure Family**:

- `configuration` - Configuration files
- `scripts-package` - Shell scripts
- `templates` - Template files

### Inheritance Example

```yaml
# Type defines defaults
types:
  - name: go-library
    defaults:
      files:
        source: ["**/*.go"]

# Module inherits automatically
modules:
  - moniker: my-lib
    type: go-library
    # Automatically has files.source = ["**/*.go"]

# Module can override
modules:
  - moniker: my-lib2
    type: go-library
    files:
      source: ["cmd/**/*.go"]  # Override
```

---

## Environments Contract

**File**: `.r2r/eac/environments.yml`

**Purpose**: Define test execution environments organized in testing pyramid hierarchy (L0-L4)

### Structure

**Minimal**:

```yaml
environments:
  - moniker: l00-01
    name: "L0 Environment 01"
    level: "L0"
    type: "unit"
```

**Complete**:

```yaml
environments:
  - moniker: plte01
    name: "PLTE Environment 01 - Kubernetes"
    description: "Kubernetes-based production-like test environment"
    level: "L3"
    type: "plte"
    system_deps: [kubectl, helm]
```

### Key Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `moniker` | string | ✅ | Unique environment identifier (kebab-case) |
| `name` | string | ✅ | Human-readable environment name |
| `level` | string | ✅ | Test level: L0, L1, L2, L3, L4 |
| `type` | string | ✅ | Environment type (unit, docker, plte, production) |
| `description` | string | ❌ | Detailed environment description |
| `system_deps` | array | ❌ | Required system dependencies |

### Testing Pyramid Levels

| Level | Speed | Execution Time | Type | System Deps | Use Cases |
|-------|-------|----------------|------|-------------|-----------|
| **L0** | Very Fast | < 100ms | unit | None | Pure logic, no I/O |
| **L1** | Fast | 100ms - 1s | unit | go | Component tests, mocked deps |
| **L2** | Moderate | 1s - 30s | docker | docker | Service integration, API contracts |
| **L3** | Slow | 30s - 5min | plte | kubectl, helm | End-to-end tests, PLTE |
| **L4** | Variable | Variable | production | kubectl, helm | Smoke tests, production |

**Recommended Distribution**: 54% L0, 30% L1, 10% L2, 5% L3, 1% L4

### Environment Types

| Type | Description | Typical Levels |
|------|-------------|----------------|
| `unit` | In-process or fast unit tests | L0, L1 |
| `docker` | Single Docker container | L2 |
| `docker-compose` | Multi-container orchestration | L2 |
| `plte` | Production-Like Test Environment (Kubernetes) | L3 |
| `production` | Live production | L4 |

### Examples

**L0 (Very Fast)**:

```yaml
- moniker: l00-01
  name: "L0 Environment 01 - In-process Unit Tests"
  level: "L0"
  type: "unit"
  system_deps: []
```

**L2 (Docker)**:

```yaml
- moniker: local01
  name: "Local Environment 01 - Docker Container"
  level: "L2"
  type: "docker"
  system_deps: [docker]
```

**L3 (PLTE)**:

```yaml
- moniker: plte01
  name: "PLTE Environment 01 - Kubernetes"
  level: "L3"
  type: "plte"
  system_deps: [kubectl, helm]
```

---

## Test Suites Contract

**File**: `.r2r/eac/test-suites.yml`

**Purpose**: Define test suites with tag-based selection criteria

### Structure

```yaml
suites:
  - moniker: commit
    name: Commit Tests
    description: "Fast tests for commit validation"
    selectors:
      - any_of_tags: ["@L0", "@L1"]
        exclude_tags: ["@L2", "@L3", "@L4"]

  - moniker: acceptance
    name: Acceptance Tests
    description: "Comprehensive acceptance tests"
    selectors:
      - require_tags: ["@L0", "@L1", "@L2"]
        exclude_tags: ["@wip"]
```

### Selector Types

| Selector | Description | Example |
|----------|-------------|---------|
| `require_tags` | Must have ALL tags | `["@L0", "@smoke"]` |
| `any_of_tags` | Must have ANY tag | `["@L0", "@L1"]` |
| `exclude_tags` | Must NOT have tags | `["@wip", "@skip"]` |

---

## Contract Relationships

```mermaid
graph TB
    Modules[modules.yml] -->|references| Types[module-types.yml]
    Modules -->|depends_on| Modules
    Types -->|requires| SysDeps[system-dependencies.yml]
    Suites[test-suites.yml] -->|selects| Tags[testing-tags.yml]
    Suites -->|runs in| Envs[environments.yml]
    Envs -->|requires| SysDeps

    style Modules fill:#e1f5ff
    style Types fill:#ffe1e1
    style Suites fill:#e1ffe1
```

---

## Validation System

### Validation Levels

| Level | Validates | Command |
|-------|-----------|---------|
| **Schema** | YAML structure against JSON schema | `validate-contracts` |
| **Cross-reference** | Dependencies and references exist | `validate-dependencies` |
| **File ownership** | Files claimed by exactly one module | `validate-module-files` |
| **Hierarchy** | Dependency graph is acyclic | `validate-module-hierarchy` |
| **Specs** | Gherkin syntax, tag validity | `validate-specs` |
| **Design** | Structurizr DSL syntax | `validate-design` |

### Commands

```bash
# Validate all contracts
r2r eac validate

# Schema validation
r2r eac validate-contracts

# Cross-references
r2r eac validate-dependencies

# File ownership
r2r eac validate-module-files

# Dependency graph
r2r eac validate-module-hierarchy

# Gherkin specs
r2r eac validate-specs

# Structurizr DSL
r2r eac validate-design
```

---

## Configuration Precedence

**Hierarchy** (highest to lowest priority):

1. **Personal config** (`.personal.yml` files, not in Git)
2. **Shared config** (standard `.yml` files in `.r2r/eac/`)
3. **Type defaults** (from `module-types.yml`)
4. **System defaults** (hardcoded in eac-core)

**Example**: `.r2r/eac/modules.personal.yml` overrides `.r2r/eac/modules.yml`

---

## Contract Lifecycle

**1. Definition**: Create/edit YAML files in `.r2r/eac/`

**2. Validation**: Validate against JSON schema

```bash
r2r eac validate-contracts
```

**3. Loading**: Contracts loaded at command runtime

```text
Command Start → Load Contracts → Validate → Build Registry → Execute
```

**4. Execution**: Contracts drive behavior (builds, tests, commands)

---

## IDE Integration

YAML Language Server support via schema reference:

```yaml
# yaml-language-server: $schema=../../contracts/eac-core/0.1.0/modules.schema.json
modules:
  - moniker: |  # IDE provides auto-completion
```

**Features**:

- Auto-completion for fields
- Inline validation
- Hover documentation
- Enum value suggestions

**Supported IDEs**: VS Code, IntelliJ, Neovim (with LSP)

---

## Schema Versioning

**Schema organization**:

```text
contracts/
└── eac-core/
    └── 0.1.0/
        ├── modules.schema.json
        ├── module-types.schema.json
        ├── environments.schema.json
        └── ...
```

**Future**: Schema versioning with migration tools for contract upgrades

---

## Display & Query Commands

### Display Commands (Human-Readable)

```bash
r2r eac show-modules            # Module table
r2r eac show-dependencies       # Dependency graph
r2r eac show-environments       # Environment table
r2r eac show-moduletypes        # Module type table
r2r eac show-config             # All configuration
```

### Get Commands (JSON Output)

```bash
r2r eac get-modules             # Modules JSON
r2r eac get-dependencies        # Dependencies JSON
r2r eac get-environments        # Environments JSON
r2r eac get-config              # Config JSON
```

---

## Common Patterns

### Module Definition Pattern

```yaml
# Go library
- moniker: my-lib
  type: go-library
  depends_on: [other-lib]
  files:
    root: go/my/lib
    source: ["**/*.go"]
    tests: ["**/*_test.go"]

# Go CLI with versioning
- moniker: my-cli
  type: go-cli
  files:
    root: go/my/cli
  versioning:
    scheme: semver
    prefix: v
```

### Module Type Pattern

```yaml
# Custom type template
- name: my-type
  description: "My custom type"
  build_deps: [go]
  capabilities: [go_module]
  build:
    artifacts:
      - type: marker
        pattern: ".built"
  defaults:
    files:
      source: ["**/*.go"]
```

### Environment Pattern

```yaml
# Fast unit tests
- moniker: l00-01
  level: "L0"
  type: "unit"

# Docker integration tests
- moniker: local01
  level: "L2"
  type: "docker"
  system_deps: [docker]
```

### Test Suite Pattern

```yaml
# Fast commit tests
- moniker: commit
  selectors:
    - any_of_tags: ["@L0", "@L1"]
      exclude_tags: ["@L2", "@L3", "@L4", "@wip"]

# Production verification
- moniker: production-verification
  selectors:
    - require_tags: ["@L4", "@piv"]
```

---

## Benefits

| Benefit | Description |
|---------|-------------|
| **Version Control** | Track changes, code review, rollback capabilities |
| **Self-Documenting** | Human-readable YAML with schema validation |
| **Early Validation** | Catch errors pre-commit, not at runtime |
| **Consistency** | Single source of truth, type templates ensure uniformity |
| **Tool Integration** | IDE support, AI-friendly structured data |
| **Explicit Configuration** | No implicit conventions, everything defined |

---

## Related Documentation

- [Overview](./index.md) - System overview and key concepts
- [Architecture](./architecture.md) - System architecture and components
- [Modules](./modules.md) - Module system and dependency management

{{ diataxis_footer() }}
