# mkdocs

MkDocs documentation build handlers implementing a multi-stage pipeline for preprocessing, HTML site rendering, and PDF generation.

## Key Types

- **`PreprocessHandler`** -- First stage of the build pipeline; copies static files, applies link translations, processes diagrams, and writes staged markdown ready for MkDocs
- **`SiteRenderHandler`** -- Second stage; consumes base-site output and runs MkDocs in Docker to generate an HTML site
- **`PDFRenderHandler`** -- Alternative render stage; consumes base-site output and generates PDF via Docker with retry logic for transient Playwright timeouts
- **`Handler`** -- Interface that all mkdocs handlers implement (Name, Build, ListArtifacts, Requirements, ValidateModule, IsContainer, IsHostInstalled)
- **`ComponentManifest`** -- Tracks the state of a built component for manifest-based caching; stores input/output hashes, timestamps, artifacts, and dependencies
- **`ManifestStore`** -- Manages component manifests in the `.cache/eac/components/` directory
- **`BuildOptions`** -- Configuration flags for controlling the build process (force, PDF mode, tidy, reproducible, weight, artifacts mode)
- **`BuildResult`** -- Outcome of a component build including exit code, cache status, manifest, and artifacts
- **`ConfigOptions`** -- Options for generating an mkdocs.yml config (site name, theme, docs dir, PDF concurrency, page limit)
- **`TemplateType`** -- Enum distinguishing site vs PDF mkdocs templates
- **`CacheStats`** -- Tracks cache hit/miss statistics for reporting

## Key Functions

- `Handlers` -- Returns all mkdocs handlers (preprocess, site-render, pdf-render) for registration by the builders package
- `HandlerNames` -- Returns the names of all mkdocs handlers
- `GenerateMkDocsConfig` -- Generates a final mkdocs.yml by loading a template and injecting options via string replacement, preserving Python-specific YAML tags
- `WriteMkDocsConfig` -- Writes a generated mkdocs.yml to the specified path
- `LoadMkDocsTemplate` -- Loads the mkdocs.yml template for a given output type
- `GetThemeFromOutput` -- Extracts theme name from output format (e.g., "pdf-dark" -> "dark")

## Patterns

- **Multi-stage pipeline**: Build is split into preprocessing (pure Go, no Docker), site rendering (Docker), and PDF rendering (Docker), enabling caching at each stage
- **Manifest-based caching**: Each stage writes a `ComponentManifest` with input/output hashes; downstream stages use the upstream output hash as their input hash for cache invalidation
- **Template string replacement**: mkdocs.yml generation uses regex-based string replacement instead of YAML parsing to preserve Python-specific YAML tags like `!!python/name:`
- **Handler registration via factory**: The `Handlers()` function returns a map of handlers, avoiding circular imports with the builders package
- **Base-site dependency resolution**: `resolveBaseSiteComponent` resolves which base-site a render handler depends on via `depends_on`, manifest lookup, naming convention, or default fallback

## Internal Structure

| File | Responsibility |
| --- | --- |
| types.go | Core types: ComponentManifest, BuildResult, ManifestStore, ComponentType constants, CacheStats |
| register.go | Handler interface definition, handler factory (Handlers), and handler name listing |
| preprocess.go | PreprocessHandler implementation for base-site preprocessing pipeline |
| site.go | SiteRenderHandler implementation for HTML site generation via Docker, plus resolveBaseSiteComponent logic |
| pdf.go | PDFRenderHandler implementation for PDF generation via Docker with retry logic |
| mkdocsconfig.go | MkDocs configuration template loading, generation, and writing with regex-based value replacement |

## Dependencies

- `contracts/core/0.1.0` -- ModuleContractPort interface for module abstraction
- `go/cli/eac/impl/build/docprep` -- Document preprocessing pipeline (DefaultPipeline, PreprocessContext)
- `go/core/adapters` -- UnwrapModule to get concrete module from contract port
- `go/core/config` -- Loading repository and book configuration
- `go/core/environments` -- PDF export concurrency and artifacts mode settings
- `go/core/paths` -- Canonical file path resolution (staging cache, template paths, workspace DSL)
- `go/core/tool` -- Tool bridge for executing Docker-based builds (GlobalHandlerToolBridge)

## Role in System

This package is the core documentation build engine for the eac system. It implements a three-stage pipeline (preprocess, site-render, pdf-render) that transforms source markdown into HTML sites and PDF documents using MkDocs via Docker containers. The manifest-based caching system ensures only changed content is rebuilt, and the handler registration pattern allows the builders package to orchestrate these handlers alongside other build types.

## Code Health

### Tech Debt

- `preprocess.go` (429 lines), `pdf.go` (336 lines) exceed 300 lines

### Pain Points

- No test coverage for `mkdocsconfig.go`, `pdf.go`, `register.go`, `site.go`

### Optimization Opportunities

- None identified
