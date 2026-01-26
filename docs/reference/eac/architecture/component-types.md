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
      go: go/eac/commands       # Go component
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

## Component Configuration

### Simple Form (Path Only)

When a component just needs a root path:

```yaml
components:
  go: go/eac/core
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
    root: go/eac/core
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

Implement a handler in `go/eac/commands/impl/build/builders/`:

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

Implement a runner in `go/eac/commands/impl/test/runners/` if the component type has specific test tooling.

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
- [Contracts](../contracts/index.md) - Contract system and YAML schemas
- [Dependencies](./dependencies.md) - Dependency resolution details
