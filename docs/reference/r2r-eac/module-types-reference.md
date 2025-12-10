# Module Types Reference

{{ page_breadcrumb() }}

EAC supports different module types, each optimized for specific languages and build requirements.

## Available Module Types

### `go` - Go Modules

**Language:** Go
**Capabilities:** `go_module`, `cross_compile`

**File Patterns:**

- Source: `**/*.go`, `**/*.go.txt`
- Tests: `**/*_test.go`
- Config: `go.mod`, `go.sum`

**Build Features:**

- Library compilation (verification only)
- Executable builds with version injection
- Cross-platform compilation (linux/darwin/windows × amd64/arm64)
- UPX compression support
- ldflags injection for version metadata

**Test Support:**

- `gotest` - Standard Go unit tests with JSON output
- `godog` - BDD/Gherkin tests with Cucumber integration

**Build Handler:** `GoHandler` - Supports libraries, executables, and cross-compilation

**Example:**

```yaml
- moniker: my-service
  type: go
  capabilities: [go_module, cross_compile]
  artifacts:
    - name: my-service
      type: executable
      platforms: [linux/amd64, darwin/arm64, windows/amd64]
```

---

### `typescript` - TypeScript/npm Modules

**Language:** TypeScript/JavaScript
**Capabilities:** `npm_package`, `typescript`

**File Patterns:**

- Source: `src/**/*.ts`, `features/**/*.ts`
- Tests: `**/*.test.ts`, `**/*.spec.ts`
- Config: `package.json`, `tsconfig.json`, `*.config.js`

**Build Features:**

- `npm install` - Dependency installation
- `npx tsc` - TypeScript compilation
- Output directory management (typically `dist/`)
- Node modules caching

**Test Support:**

- `mocha` - Unit tests via `npm test`
- `cucumber-js` - BDD/Gherkin tests with tag filtering

**Build Handler:** `NpmHandler` - Manages npm/TypeScript build pipeline

**Example:**

```yaml
- moniker: my-webapp
  type: typescript
  capabilities: [npm_package, typescript]
  artifacts:
    - name: dist
      type: directory
```

---

### `container` - Docker Containers

**Language:** Any (language-agnostic)
**Capabilities:** `buildx`

**File Patterns:**

- Source: `Dockerfile`, `**/*`
- Assets: All files in module directory (build context)

**Build Features:**

- Multi-platform image builds via Docker buildx
- Image tagging and versioning
- Registry push support
- Build context optimization
- Platform targeting (linux/amd64, linux/arm64, etc.)

**Test Support:**

- Depends on what's included in the container
- Can run tests inside container during build

**Build Handlers:**

- `BuildxHandler` - Multi-platform builds (preferred)
- `DockerHandler` - Single-platform builds (fallback)

**Example:**

```yaml
- moniker: my-api
  type: container
  capabilities: [buildx]
  artifacts:
    - name: my-api
      type: image
      platforms: [linux/amd64, linux/arm64]
      registries:
        - ghcr.io/myorg/my-api
```

**Use Cases:**

- Containerizing applications in any language (Python, Rust, Java, etc.)
- Creating deployment artifacts
- Multi-platform binary distribution

---

### `docs` - MkDocs Documentation

**Language:** Markdown/Python (MkDocs)
**Capabilities:** `mkdocs`

**File Patterns:**

- Source: `docs/**/*.md`
- Config: `mkdocs.yml`, `books.yml`

**Build Features:**

- MkDocs site generation
- PDF generation (dark and light themes)
- Books support (multiple documentation sets)
- Material theme integration
- Search indexing

**Build Handler:** `MkdocsHandler` - Manages MkDocs build and PDF generation

**Example:**

```yaml
- moniker: docs-site
  type: docs
  capabilities: [mkdocs]
  artifacts:
    - name: site
      type: directory
```

---

### `static` - Static Modules

**Language:** Any
**Capabilities:** None (or custom)

**File Patterns:**

- Defined by file ownership only (no defaults)

**Build Features:**

- No automatic build steps
- Can use custom scripts via module configuration
- Useful for configuration, contracts, or documentation

**Test Support:**

- None by default
- Can define custom test commands

**Build Handler:** `NoneHandler` - No-op handler, or `ScriptsHandler` with custom scripts

**Example:**

```yaml
- moniker: config
  type: static
  capabilities: []
  files:
    - .r2r/**/*
    - contracts/**/*
```

**Use Cases:**

- Configuration-only modules
- Contract definitions
- Static assets
- Documentation without build steps

---

## Language Support Matrix

| Language | Module Type | Build | Test | Cross-Compile | Notes |
|----------|-------------|-------|------|---------------|-------|
| **Go** | `go` | ✅ Full | ✅ gotest, godog | ✅ Yes | Native support with cross-platform builds |
| **TypeScript** | `typescript` | ✅ Full | ✅ mocha, cucumber-js | ❌ No | npm and tsc integration |
| **JavaScript** | `typescript` | ✅ npm | ✅ mocha, cucumber-js | ❌ No | Use typescript type, skip tsc if no TS |
| **Python** | `container` | ⚠️ Custom | ⚠️ Custom | ❌ No | Use Dockerfile with Python base image |
| **Rust** | `container` | ⚠️ Custom | ⚠️ Custom | ❌ No | Use Dockerfile with Rust toolchain |
| **Java** | `container` | ⚠️ Custom | ⚠️ Custom | ❌ No | Use Dockerfile with Maven/Gradle |
| **Markdown** | `docs` | ✅ MkDocs | ❌ No | ❌ No | Documentation generation only |
| **Any** | `static` | ❌ No | ❌ No | ❌ No | File ownership only |

**Legend:**

- ✅ **Full** - Native handler with comprehensive support
- ⚠️ **Custom** - Requires Dockerfile or custom scripts
- ❌ **No** - Not supported

---

## Capability-Based Dispatch

EAC uses a **capability-based dispatch system** to match module types to build/test handlers:

1. **Module declares capabilities** in `.r2r/eac/modules.yml`:

   ```yaml
   capabilities: [go_module, cross_compile]
   ```

2. **Handlers register capabilities** in their `init()` function:

   ```go
   registry.RegisterHandler(&GoHandler{
       Capabilities: []string{"go_module", "cross_compile"},
   })
   ```

3. **Commands dispatch to handlers** that match module capabilities:

   ```go
   handler := GetHandlerForModule(module)
   ```

This architecture allows new languages to be added by:

- Creating a new handler (e.g., `PythonHandler`, `RustHandler`)
- Registering capabilities
- Defining module type defaults

---

## Adding Custom Module Types

To add support for a new language:

### 1. Define Module Type

Create or update `.r2r/eac/module-types.yml`:

```yaml
types:
  - name: python
    capabilities: [python_package]
    defaults:
      files:
        source:
          - "**/*.py"
        tests:
          - "tests/**/*.py"
        config:
          - "pyproject.toml"
          - "setup.py"
```

### 2. Create Build Handler

Implement a handler in `go/eac/commands/impl/build/builders/`:

```go
type PythonHandler struct {
    Capabilities []string
}

func (h *PythonHandler) GetCapabilities() []string {
    return []string{"python_package"}
}

func (h *PythonHandler) Build(ctx context.Context, module *contracts.Module) error {
    // Build logic here
}
```

### 3. Register Handler

Add to `init()` in your handler file:

```go
func init() {
    registry.RegisterHandler(&PythonHandler{})
}
```

### 4. Add Test Runner (Optional)

Implement a runner in `go/eac/commands/impl/test/runners/` if needed.

---

## Best Practices

### Choosing Module Types

| Scenario | Recommended Type | Rationale |
|----------|------------------|-----------|
| Go library/service | `go` | Native build/test support, cross-compilation |
| TypeScript library/app | `typescript` | npm and tsc integration |
| Python service | `container` | No native handler yet, use Docker |
| Multi-language app | `container` | Bundle all dependencies in image |
| Static config | `static` | No build needed |
| Documentation | `docs` | MkDocs integration |

### File Ownership

Define explicit file patterns for clarity:

```yaml
files:
  source:
    - "cmd/**/*.go"
    - "internal/**/*.go"
  tests:
    - "**/*_test.go"
  config:
    - "go.mod"
    - "go.sum"
```

### Artifacts

Specify expected build outputs:

```yaml
artifacts:
  - name: my-binary
    type: executable
    platforms: [linux/amd64]
  - name: coverage.out
    type: file
```

---

## Related Documentation

- [Modules](./modules.md) - Module system and dependency management
- [Contracts](./contracts.md) - Module contract specification
- [Creating Modules](../../how-to-guides/eac/modules/creating-modules.md) - Step-by-step guide
- [Creating Module Types](../../how-to-guides/eac/modules/creating-module-types.md) - Custom type definitions

{{ diataxis_footer() }}
