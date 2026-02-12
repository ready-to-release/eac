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
| go.go | `GoHandler` registration, interface methods, and tool execution |
| go_build.go | Go build orchestration: library, test, and single-binary builds |
| go_cross.go | Cross-compilation for multiple GOOS/GOARCH targets |
| go_artifacts.go | Artifact listing from per-module definitions |
| go_version.go | Version injection, ldflags, and changelog parsing |
| go_checksum.go | SHA256 checksum generation for release artifacts |
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
| builders_unit_test.go | Unit tests for `Logln`, `substituteVars`, `isBuildMetadataFile` |
| mkdocs/ | MkDocs-specific handlers (preprocess, site-render, pdf-render) |

## Dependencies

- `contracts/core` -- `ModuleContractPort` and action constants
- `core/tool` -- build bridge, handler registry, and tool execution
- `core/config` -- module and artifact configuration
- `core/adapters` -- module contract port adapters
- `core/environments` -- CI detection via `environments.IsCI()`
- `impl/build/docprep` -- document preprocessing pipeline

## Role in System

The builders package provides all concrete `BuildHandler` implementations used by the build command in `eac`. When the build framework dispatches a component work unit, it looks up the appropriate handler from the global build bridge, validates the module, and invokes `Build()`. The mkdocs sub-package further decomposes documentation builds into preprocessing and container-based rendering steps.

## Code Health

### Tech Debt

- `mkdocs.go` (743 lines), `pdf-render-tool.go` (742 lines), `mkdocs-book.go` (496 lines), `helpers.go` (394 lines), `buildx.go` (353 lines), `structurizr.go` (334 lines), `pdf.go` (320 lines), `docker.go` (311 lines) exceed 300 lines

### Pain Points

- No test coverage for `drawio.go`, `go_artifacts.go`, `go_build.go`, `go_checksum.go`, `go_cross.go`, `go_version.go`, `mermaid-render.go`, `mkdocs-book.go`, `none.go`, `python.go`, `scripts.go`, `structurizr.go`, `types.go`

### Optimization Opportunities

- None identified
