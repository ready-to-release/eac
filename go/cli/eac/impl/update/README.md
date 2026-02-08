# update Commands

Commands that regenerate or refresh derived artifacts, caches, and documentation
from authoritative source files in the repository.

## Command Index

| Command | Package | Purpose |
| --- | --- | --- |
| `eac update cache-clear` | [cache-clear/](./cache-clear/) | Clear incremental cache state files to force full rebuilds |
| `eac update docs` | [docs/](./docs/) | Regenerate documentation from source |
| `eac update docs-manifest` | [docs-manifest/](./docs-manifest/) | Update documentation assets manifest from docs/assets/ |
| `eac update lint` | [lint/](./lint/) | Update lint-related generated files |

## Small Commands

Commands documented inline rather than in their own README.

### evidence

Generates evidence PDFs from a module's `evidence_books` configuration. Evidence books
are markdown-based documentation packages that aggregate test results, scan reports,
and compliance artifacts into reviewable PDF documents.

### go-tidy

Runs `go mod tidy` across all workspace modules, ensuring all Go modules have
clean `go.mod` and `go.sum` files with no unused or missing dependencies.

### go-mod-sums

Downloads all declared dependencies and refreshes `go.sum` checksums across workspace
modules without modifying `go.mod` files. Useful for syncing checksums after dependency updates.

### go-sums

Runs `go mod tidy` across all workspace modules to ensure tidy `go.mod` and `go.sum` files.
Functionally similar to `go-tidy` but operates as a separate entry point.

### pdf-screenshots

Extracts PDF pages as PNG images for the documentation cache. Scans `out/build` folders
for generated PDF books and stores page images in `.cache/eac/pdf-screenshots/`.

### structurizr

Exports all Structurizr views from `workspace.dsl` files to SVG format and stores them
in the `docs/assets/cache/structurizr/` directory for documentation embedding.

### ai-summary

Generates AI-powered analysis summaries for modules. The analysis covers architecture
(DSL), specifications (Gherkin), test results, and code structure to produce
comprehensive status reports.

### design

Updates Structurizr DSL workspace files for modules by re-analyzing source code with AI.
Follows the same AI-driven pattern as `create design` but operates as an update to
existing design artifacts.

## Shared Infrastructure

| Package | Purpose |
| --- | --- |
| [internal/](./internal/) | Shared helpers used across update subcommands |

## Architecture Notes

Update commands are idempotent regeneration operations. They read authoritative sources
(Go modules, DSL files, evidence configuration) and produce derived outputs (PDFs, SVGs,
manifests, checksums). Most commands operate at the workspace level, iterating across all
modules. The `cache-clear` command is the exception, removing state rather than producing it.
