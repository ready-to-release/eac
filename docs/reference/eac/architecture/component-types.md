# Component Types Reference

EAC modules contain **components**, and each component has a **type** that determines its build behavior,
file patterns, and tooling. Component types are defined in `component-types.yml` and provide the building blocks for modules.

## Key Concepts

- **Module** - A logical unit with a moniker (e.g., `eac-core`, `eac-commands`)
- **Component** - Something a module contains (e.g., Go code, TypeScript code, specs)
- **Component Type** - Defines behavior for a component (e.g., `go`, `typescript`, `book`)

A single module can contain multiple components of different types:

```yaml
modules:
  - moniker: eac-commands
    name: EAC Command Implementations
    components:
      go: go/cli/eac       # Go component
      specs: null               # Gherkin specs (uses default path)
      design: null              # Structurizr design (uses default path)
```

---

## Buildable Component Types

### `go` - Go Code

**Builder:** `go`
**Requirements:** Go toolchain

**Default File Patterns:**

- Source: `**/*.go`, `**/*.go.txt`
- Tests: `**/*_test.go`
- Config: `go.mod`, `go.sum`

**Build Features:**

- Library compilation (verification only)
- Executable builds with version injection via ldflags
- Cross-platform compilation (linux/darwin/windows × amd64/arm64)
- UPX compression support

**Test Support:**

- `gotest` - Standard Go unit tests with JSON output
- `godog` - BDD/Gherkin tests with Cucumber integration

**Example:**

```yaml
- moniker: my-service
  name: My Service
  components:
    go: go/my/service
```

---

### `typescript` - TypeScript/JavaScript

**Builder:** `npm`
**Requirements:** npm

**Default File Patterns:**

- Source: `src/**/*.ts`, `features/**/*.ts`
- Tests: `**/*.test.ts`
- Config: `package.json`, `tsconfig.json`, `*.config.js`

**Build Features:**

- `npm install` - Dependency installation
- `npx tsc` - TypeScript compilation
- Output directory management (typically `dist/`)
- Node modules caching

**Test Support:**

- `mocha` - Unit tests via `npm test`
- `cucumber-js` - BDD/Gherkin tests with tag filtering

**Example:**

```yaml
- moniker: vscode-ext
  name: VSCode Extension
  components:
    typescript: typescript/vscode-ext
```

---

### `dockerfile` - Container Images

**Builder:** `buildx`
**Requirements:** Docker

**Default File Patterns:**

- Source: `Dockerfile`

**Build Features:**

- Multi-platform image builds via Docker buildx
- Image tagging and versioning
- Registry push support
- Platform targeting (linux/amd64, linux/arm64)

**Example:**

```yaml
- moniker: ext-eac
  name: EAC Extension Container
  components:
    dockerfile: containers/ext-eac
```

---

### `book` - MkDocs Documentation

**Builder:** `mkdocs`
**Requirements:** MkDocs

Books are defined in `books.yml` and built as documentation artifacts.

**Build Features:**

- MkDocs site generation
- PDF generation (dark and light themes)
- Material theme integration
- Search indexing

**Example:**

```yaml
- moniker: docs
  name: Documentation
  components:
    book: null  # Uses books.yml configuration
```

---

### `pwsh` - PowerShell Scripts

**Builder:** `scripts`
**Requirements:** PowerShell 7.4+

**Default File Patterns:**

- Source: `**/*.ps1`, `**/*.psm1`
- Tests: `**/*.Tests.ps1`
- Config: `**/*.psd1`

---

### `bash` - Shell Scripts

**Builder:** `scripts`
**Requirements:** Bash

**Default File Patterns:**

- Source: `**/*.sh`, `**/*.bash`

---

## Non-Buildable Component Types

These component types define file ownership but don't have build steps.

### `gherkin` - BDD Specifications

**File Patterns:** `**/*.feature`

Gherkin specifications for BDD testing. Typically mapped via the `specs` component shorthand.

```yaml
components:
  specs: null  # Uses default: specs/{moniker}
```

---

### `structurizr` - Architecture Diagrams

**File Patterns:** `workspace.dsl`, `**/*.dsl`

C4 architecture diagrams in Structurizr DSL. Typically mapped via the `design` component shorthand.

```yaml
components:
  design: null  # Uses default: specs/{moniker}/.design
```

---

### `markdown` - Documentation Files

**File Patterns:** `**/*.md`, `**/*.markdown`

---

### `yaml` - Configuration Files

**File Patterns:** `**/*.yml`, `**/*.yaml`

---

### `json` - Data Files

**File Patterns:** `**/*.json`

---

### `workflows` - GitHub Actions

**Default Root:** `.github/workflows`
**File Patterns:** `**/*.yml`, `**/*.yaml`
**Requirements:** actionlint

---

### `docs-assets` - Documentation Images

**File Patterns:** `**/*.png`, `**/*.jpg`, `**/*.svg`

---

### `testdata` - Test Fixtures

**File Patterns:** `**/*.txt`, `**/*.go.txt`

---

## Language Support Matrix

<!-- markdownlint-disable MD060 -->

| Language       | Component Type | Build      | Test                  | Cross-Compile | Notes                                     |
| -------------- | -------------- | ---------- | --------------------- | ------------- | ----------------------------------------- |
| **Go**         | `go`           | ✅ Full    | ✅ gotest, godog      | ✅ Yes        | Native support with cross-platform builds |
| **TypeScript** | `typescript`   | ✅ Full    | ✅ mocha, cucumber-js | ❌ No         | npm and tsc integration                   |
| **JavaScript** | `typescript`   | ✅ npm     | ✅ mocha, cucumber-js | ❌ No         | Use typescript type, skip tsc if no TS    |
| **Python**     | `dockerfile`   | ⚠️ Custom   | ⚠️ Custom              | ❌ No         | Use Dockerfile with Python base image     |
| **Rust**       | `dockerfile`   | ⚠️ Custom   | ⚠️ Custom              | ❌ No         | Use Dockerfile with Rust toolchain        |
| **Java**       | `dockerfile`   | ⚠️ Custom   | ⚠️ Custom              | ❌ No         | Use Dockerfile with Maven/Gradle          |
| **Markdown**   | `book`         | ✅ MkDocs  | ❌ No                 | ❌ No         | Documentation generation only             |
| **PowerShell** | `pwsh`         | ✅ Scripts | ✅ Pester             | ❌ No         | Cross-platform PowerShell 7.4+            |
| **Bash**       | `bash`         | ✅ Scripts | ❌ No                 | ❌ No         | POSIX shell scripts                       |

<!-- markdownlint-enable MD060 -->

**Legend:**

- ✅ **Full** - Native handler with comprehensive support
- ⚠️ **Custom** - Requires Dockerfile or custom scripts
- ❌ **No** - Not supported

---

## Resource Configuration

Components consume resources during build, test, lint, and scan operations. EAC provides two mechanisms to control resource allocation:

1. **Tool Resources** - Base resource limits defined per tool in `tool-config.yml`
2. **Component Amp** - Per-component multiplier defined in `repository.yml`

### Tool Resources

Each tool can define resource requirements in `tool-config.yml`:

```yaml
tools:
  mkdocs-system:
    type: container
    resources:
      cpus: 4        # CPU cores (also used as scheduling weight)
      memory: "4g"   # Memory limit (e.g., "512m", "2g", "8g")
      shm_size: "1g" # Shared memory size (optional, for browsers/heavy tools)
```

**Resource Fields:**

| Field | Description | Default | Example |
|-------|-------------|---------|---------|
| `cpus` | CPU cores allocated to container; also used as scheduling weight | 1 | `4` |
| `memory` | Memory limit for container | unlimited | `"4g"`, `"512m"` |
| `shm_size` | Shared memory (`/dev/shm`) size | Docker default | `"1g"` |

**Typical Resource Values by Tool Type:**

| Tool Category | CPUs | Memory | Notes |
|--------------|------|--------|-------|
| Linters | 1-2 | 2-4g | Fast, low memory |
| Go builds | 2-4 | 4g | Parallelizes well |
| TypeScript builds | 2 | 4g | Node.js memory usage |
| MkDocs/Books | 4 | 4g | PDF generation is heavy |
| Docker buildx | 4 | 4g+ | Layer caching helps |
| Security scanners | 2 | 2-4g | Database lookups |

### Component Amp (Resource Amplifier)

The `amp` field lets you adjust resource allocation per-component without modifying tool definitions. This is useful when specific components need more or fewer resources than the default.

**Configuration in `repository.yml`:**

```yaml
modules:
  - moniker: my-module
    components:
      # Heavy PDF book needs more build resources
      howto:
        type: book
        amp:
          build: 2.0    # Double the base tool resources

      # Standard Go component with heavy tests
      go:
        root: go/my-module
        amp:
          build: 1.0    # Normal build resources
          test: 2.0     # Double resources for parallel tests
          lint: 0.5     # Half resources (linting is light)

      # Trivial markdown docs
      readme:
        type: markdown
        amp:
          lint: 0.5     # Minimal resources needed
```

**Amp Operations:**

| Operation | When Applied |
|-----------|--------------|
| `build` | During `eac build` |
| `test` | During `eac test` |
| `lint` | During `eac lint` |
| `scan` | During `eac scan` |

**Amp Value Effects:**

| Amp Value | Effect | Use Case |
|-----------|--------|----------|
| `1.0` | No change (default) | Most components |
| `2.0` | Double resources | Heavy builds, large test suites |
| `0.5` | Half resources | Lightweight linting, small components |
| `0.1` | Minimum (10%) | Trivial operations |
| `10.0` | Maximum (10x) | Extreme resource needs |

**How Amp Works:**

1. **Scheduling Weight**: `base_cpus × amp` determines parallel slot allocation
2. **Container Resources**: CPU and memory limits are multiplied by amp

```
Example: Tool has cpus: 4, memory: 4g
         Component has amp.build: 2.0

Result:  Scheduling weight = 8
         Container gets 8 CPUs, 8GB memory
```

**When to Use Amp:**

- ✅ One component is slower than others of same type
- ✅ Test suite needs more parallelism
- ✅ Linting a small component doesn't need full resources
- ❌ All components of a type need more resources (update tool-config.yml instead)

---

## Component Configuration

### Simple Form (Path Only)

When a component just needs a root path:

```yaml
components:
  go: go/core
  typescript: typescript/vscode-ext
```

### Default Path (Null)

Use `null` to apply the component type's default root path (with `{moniker}` substitution):

```yaml
components:
  specs: null      # Resolves to specs/{moniker}
  design: null     # Resolves to specs/{moniker}/.design
```

### Extended Form (Custom Patterns)

Override default file patterns when needed:

```yaml
components:
  go:
    root: go/core
    patterns:
      source: ["**/*.go"]
      tests: ["**/*_test.go"]
      config: ["go.mod"]
```

### Multiple Components of Same Type

Use the `type` field to specify the component type when the key differs:

```yaml
components:
  main-app:
    type: go
    root: go/app/main
  admin-app:
    type: go
    root: go/app/admin
```

---

## Adding Custom Component Types

To add support for a new language or build system:

### 1. Define Component Type

Add to `component-types.yml`:

```yaml
component-types:
  python:
    extensions: [".py"]
    builder: python
    scanners: [sbom, vuln, secrets, sast]
    requirements: [python]
    files:
      source: ["**/*.py"]
      tests: ["tests/**/*.py", "**/*_test.py"]
      config: ["pyproject.toml", "setup.py", "requirements.txt"]
```

### 2. Create Build Handler

Implement a handler in `go/cli/eac/impl/build/builders/`:

```go
type PythonHandler struct{}

func (h *PythonHandler) Name() string {
    return "python"
}

func (h *PythonHandler) Build(ctx context.Context, module *contracts.Module) error {
    // Build logic here
}
```

### 3. Register Handler

Add to `init()` in your handler file:

```go
func init() {
    registry.RegisterBuildHandler(&PythonHandler{})
}
```

### 4. Add Test Runner (Optional)

Implement a runner in `go/cli/eac/impl/test/runners/` if the component type has specific test tooling.

---

## Commands

```bash
# List all component types
r2r eac show-component-types

# Show modules with their components
r2r eac show-modules
```

---

## Related Documentation

- [Modules](./index.md) - Module system and dependency management
- [Contracts](./contracts.md) - Contract system and YAML schemas
- [Dependencies](./dependencies.md) - Dependency resolution details
- `contracts/eac-core/0.1.0/defaults/tool-config.yml` - Tool definitions with resource settings
- `contracts/eac-core/0.1.0/repository.schema.json` - Schema for amp configuration
