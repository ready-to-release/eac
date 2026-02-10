# update Commands

Commands that regenerate or refresh derived artifacts, caches, and documentation
from authoritative source files in the repository.

## Command Index

| Command | Package | Purpose |
| --- | --- | --- |
| `eac update ai-summary` | [ai-summary/](./ai-summary/) | Generate AI-powered analysis summaries for modules |
| `eac update cache-clear` | [cache-clear/](./cache-clear/) | Clear incremental cache state files to force full rebuilds |
| `eac update design` | [design/](./design/) | Update Structurizr DSL workspace files using AI analysis |
| `eac update docs` | [docs/](./docs/) | Regenerate documentation from source |
| `eac update docs-manifest` | [docs-manifest/](./docs-manifest/) | Update documentation assets manifest from docs/assets/ |
| `eac update evidence` | [evidence/](./evidence/) | Collect and update audit evidence files for compliance |
| `eac update go-mod-sums` | [go-mod-sums/](./go-mod-sums/) | Download dependencies and refresh go.sum across workspace modules |
| `eac update go-sums` | [go-sums/](./go-sums/) | Refresh go.sum files across workspace modules |
| `eac update go-tidy` | [go-tidy/](./go-tidy/) | Run go mod tidy across all workspace modules |
| `eac update pdf-screenshots` | [pdf-screenshots/](./pdf-screenshots/) | Extract PDF pages as PNG images for documentation cache |
| `eac update structurizr` | [structurizr/](./structurizr/) | Export Structurizr diagram views to SVG cache |

## Shared Infrastructure

| Package | Purpose |
| --- | --- |
| [internal/gowork/](./internal/gowork/) | Go workspace module discovery via go.work parsing |

## Architecture Notes

Update commands are idempotent regeneration operations. They read authoritative sources
(Go modules, DSL files, evidence configuration) and produce derived outputs (PDFs, SVGs,
manifests, checksums). Most commands operate at the workspace level, iterating across all
modules. The `cache-clear` command is the exception, removing state rather than producing it.
