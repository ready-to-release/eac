# builders

Registry of build handler implementations that compile, render, or package module components. Each handler registers itself with the global build bridge at init time.

## Key Types

- **`Handler`** -- Type alias for `tool.BuildHandler` interface
- **`GoHandler`** -- Builds Go modules (libraries, CLIs, cross-platform)
- **`DockerHandler`** -- Builds Docker container images
- **`SiteHandler`** -- Unified docs-site handler (preprocessing + HTML render)
- **`PDFHandler`** -- Unified docs-pdf handler (preprocessing + PDF render)
- **`ScriptsHandler`** -- Copies script files to build output
- **`NoneHandler`** -- No-op handler for modules without a build step
- **`BookBuildState`** -- Content-hash tracking for incremental book builds

## Patterns

- Self-registration: Each handler calls `tool.GlobalBuildBridge().RegisterNativeHandler()` in `init()`
- Adapter pattern: MkDocs sub-package handlers are adapted to the `Handler` interface via lazy adapters
- Container delegation: Site and PDF handlers run preprocessing natively, then invoke container tools for rendering
- Content-hash caching: `BookBuildState` enables incremental builds by tracking staging directory hashes

## Internal Structure

| File/Sub-package | Responsibility |
| --- | --- |
| go.go | `GoHandler` for Go compilation with cross-platform and CGO support |
| docker.go | `DockerHandler` for Docker image builds |
| buildx.go | `BuildxHandler` for multi-platform Docker buildx |
| site.go | `SiteHandler` for unified doc-site preprocessing and render |
| pdf.go | `PDFHandler` for unified PDF preprocessing and render |
| scripts.go | `ScriptsHandler` for script file packaging |
| none.go | `NoneHandler` for modules with no build step |
| mkdocs.go | MkDocs shared utilities and Docker path helpers |
| mkdocs_adapter.go | Adapters bridging mkdocs sub-package to build handler interface |
| contenthash.go | `BookBuildState` for staging content hash tracking |
| helpers.go | `RunCommandWithLog`, `Logln`, and shared builder utilities |
| drawio.go | DrawIO diagram rendering via container |
| structurizr.go | Structurizr architecture diagram rendering to SVG |
| mkdocs/ | MkDocs-specific handlers (preprocess, site-render, pdf-render) |

## Dependencies

- `contracts/core` -- `ModuleContractPort` and action constants
- `core/tool` -- build bridge, handler registry, and tool execution
- `core/config` -- module and artifact configuration
- `core/adapters` -- module contract port adapters
- `impl/build/docprep` -- document preprocessing pipeline

## Role in System

The builders package provides all concrete `BuildHandler` implementations used by the build command in `eac`. When the build framework dispatches a component work unit, it looks up the appropriate handler from the global build bridge, validates the module, and invokes `Build()`. The mkdocs sub-package further decomposes documentation builds into preprocessing and container-based rendering steps.

## Code Health

### Tech Debt
- `buildx.go:297`: TODO to migrate to tool executor once stdin support is added to ExecutionContext
- `mkdocs-book.go:438`: TODO to support multiple books per module with `--book` flag
- `go.go` is 665 lines with many functions; `buildCrossCompiledFromArtifacts` (line 412) is ~97 lines with nested cross-platform loops
- `buildx.go`: `Build` method (~115 lines, line 55-170) inlines CI detection, registry auth, and arg assembly

### Pain Points
- No test files for `docker.go`, `drawio.go`, `structurizr.go`, `scripts.go`, `none.go`, or `mermaid-render.go`
- `buildx.go:84` detects CI via raw env-var checks (`os.Getenv("CI")`) instead of the shared `environment.Detect()` helper

### Optimization Opportunities
- Extract cross-compilation loop from `go.go` into a helper that builds a single target, reducing nesting -- moderate effort
- Unify CI-detection logic by using `environment.Detect()` consistently across all handlers -- low effort
