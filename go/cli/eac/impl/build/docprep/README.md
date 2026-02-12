# docprep

Document preprocessing pipeline that transforms source markdown and assets into a staging directory ready for MkDocs to render into HTML or PDF. All preprocessing runs as pure Go; containers only execute `mkdocs build`.

## Key Types

- **`Pipeline`** -- Orchestrates sequential execution of preprocessing phases
- **`Phase`** -- Interface for a single preprocessing step
- **`PhaseFunc`** -- Adapts a function to the `Phase` interface
- **`PreprocessContext`** -- Shared mutable state passed through all phases
- **`OutputMode`** -- Interface controlling format-specific behavior (site vs PDF)
- **`SiteMode`** -- `OutputMode` for HTML site builds
- **`PDFMode`** -- `OutputMode` for PDF builds with theme selection
- **`Logger`** -- Leveled log writer for phase output

## Patterns

- Pipeline pattern: `DefaultPipeline()` assembles 14 phases in dependency order, executed sequentially
- Context object: `PreprocessContext` carries immutable inputs and mutable shared state across phases
- Output mode strategy: `OutputMode` interface encapsulates site vs PDF behavioral differences
- Parallel sub-phases: Diagram processing runs mermaid, structurizr, and PlantUML concurrently via errgroup

## Internal Structure

| File/Sub-package | Responsibility |
| --- | --- |
| pipeline.go | `Pipeline` and `Phase` types with sequential execution |
| default_pipeline.go | `DefaultPipeline()` assembling all 14 phases from sub-packages |
| context.go | `PreprocessContext` with immutable inputs and mutable shared state |
| output_mode.go | `OutputMode` interface with `SiteMode` and `PDFMode` implementations |
| logger.go | `Logger` wrapping `io.Writer` with leveled formatting |
| doc.go | Package documentation |
| caching/ | Asset caching for diagram renders |
| cleanup/ | Broken link fixing and unreferenced asset removal |
| content/ | Content processing, command execution, image constraints |
| diagrams/ | Mermaid, PlantUML, and Structurizr diagram processing |
| linking/ | Source-to-staging link rewriting |
| navigation/ | Navigation structure and macro injection |
| staging/ | File copying, indexing, and asset reference scanning |

## Dependencies

- `core/config` -- book configuration (`config.Book`)
- `core/paths` -- markdown command fragment output paths

## Role in System

The docprep package is the preprocessing engine behind documentation builds in `eac`. The site and PDF build handlers in builders invoke `DefaultPipeline().Execute()` to transform source documentation into a staged directory before handing off to containerized MkDocs rendering. Each sub-package owns a distinct concern (staging, linking, diagrams, etc.) and exposes functions consumed by the pipeline phases.

## Code Health

### Tech Debt

- None identified

### Pain Points

- Phase ordering is implicit in `DefaultPipeline()` sequence; reordering requires understanding undocumented data dependencies between phases

### Optimization Opportunities

- None identified
