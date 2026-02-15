# diagrams

Diagram processing for the docprep pipeline, scanning markdown for mermaid, PlantUML, Structurizr, and DrawIO diagrams, resolving pre-rendered SVGs, and replacing code blocks with image references.

## Key Types

- **`MermaidBlock`** -- A mermaid code block with content hash, file position, and optional size preset
- **`CacheStatus`** -- Cache state for a mermaid block (hit/miss with resolved SVG path)
- **`PlantUMLBlock`** -- A PlantUML code block with content hash and file position
- **`PlantUMLCacheStatus`** -- Cache state for a PlantUML block (hit/miss with resolved SVG path)
- **`StructurizrMarker`** -- A `<!-- structurizr:module:viewKey -->` marker in markdown with module and view key
- **`DrawioImage`** -- A discovered `.drawio.png` file with content hash and source path

## Patterns

- Scan-then-replace: Diagrams are first scanned and indexed, then code blocks are replaced with `<img>` tags in a second pass
- Builder index lookup: Pre-rendered SVGs are located via JSON index manifests written by diagram builder handlers
- Content-hash filenames: SVG filenames include a content hash suffix for cache busting and deduplication
- Parallel execution: Mermaid, PlantUML, and Structurizr processing run concurrently via errgroup in the pipeline
- Size presets: Mermaid blocks support `%%{size:small}%%` directives mapped to CSS width percentages for responsive rendering

## Internal Structure

| File | Responsibility |
| --- | --- |
| mermaid.go | Scan, size presets, cache lookup, and block-to-image replacement for mermaid diagrams |
| plantuml.go | Scan and block-to-image replacement for PlantUML diagrams |
| structurizr.go | Marker extraction and SVG embedding for Structurizr architecture views |
| drawio.go | Discovery, content hashing, and width constraints for `.drawio.png` files |

## Dependencies

- `docprep/staging` -- `FileIndex` for iterating markdown files in the staging directory
- `docprep/linking` -- `CalculateRelativePath` for computing relative image reference paths
- `core/paths` -- builder output and cache directory paths for SVG index manifest lookup

## Role in System

The diagrams package implements the diagram processing phase (phase 11) of the docprep pipeline in `eac`. It runs three diagram types in parallel using errgroup. For each type, it scans staged markdown for diagram code blocks or HTML comment markers, looks up pre-rendered SVGs from builder output directories via JSON index manifests, and replaces the source blocks with `<img>` tags pointing to the SVG files.

This architecture decouples expensive container-based rendering (handled by builder handlers in the builders package) from the pure-Go preprocessing pipeline. The builder handlers render diagrams to SVGs and write index manifests during the build phase; the diagrams package then consumes those manifests to locate and embed the rendered output.

## Code Health

### Tech Debt

- `diagram.go` (345 lines) exceeds 300 lines

### Pain Points

- No test coverage for `diagram.go`, `drawio.go`

### Optimization Opportunities

- None identified
