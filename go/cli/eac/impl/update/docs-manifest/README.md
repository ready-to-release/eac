# docs-manifest

Scans documentation assets, tracks their usage across markdown files, and maintains a two-file manifest system: a git-tracked descriptions file for human metadata and a git-ignored cache file for auto-generated usage and statistics.

## Key Types

- **`DescriptionsFile`** -- YAML file with human-maintained asset descriptions
- **`Description`** -- Active status and description for a single asset
- **`CacheManifest`** -- Auto-generated cache with usage, metadata, and stats
- **`CacheAsset`** -- Auto-generated metadata for a single asset (usage, hash, size)
- **`DiscoveredAsset`** -- Asset found during filesystem scanning
- **`UpdateResult`** -- Changes detected during manifest update
- **`ManifestStats`** -- Summary statistics by category (total, used, unused)

## Patterns

- Two-file manifest: descriptions.yml (human, git-tracked) + .manifest-cache.json (auto, git-ignored)
- Phased execution: load existing, scan assets, scan usage, merge, write
- Check mode: CI validation that manifest is up-to-date without writing
- Dry-run support: preview changes without writing files

## Internal Structure

| File | Responsibility |
| --- | --- |
| command.go | Command entry point, flag parsing, phased update execution |
| types.go | All manifest data types and result structures |
| scanner.go | Filesystem scanning for documentation assets |
| usage.go | Markdown file scanning for asset references |
| merge.go | Merge discovered assets with existing manifest data |
| descriptions.go | Load and save descriptions.yml |
| manifest.go | Load and save .manifest-cache.json |

## Dependencies

- `clibase/registry` -- command registration
- `clibase/flags` -- flag validation from registry metadata
- `core/logging` -- structured logging
- `core/paths` -- docs source and cache path resolution
- `core/repository` -- repository root discovery

## Role in System

The `docs-manifest` package provides asset inventory management for `eac-cli` documentation, enabling both human curation of asset descriptions and automated tracking of which assets are actually referenced. Its check mode integrates with CI to catch stale manifests before merge.

## Code Health

### Tech Debt
- `runUpdate` in command.go (~191 lines) handles loading, scanning, merging, writing, and reporting in one function
- No unit tests for command.go, descriptions.go, or manifest.go (only merge, scanner, and usage have tests)

### Pain Points
- `runUpdate` mixes orchestration with console output formatting, making it hard to test the update logic without capturing stdout
- descriptions.go and manifest.go perform file I/O directly without interfaces, limiting testability

### Optimization Opportunities
- Split `runUpdate` into discrete load/scan/merge/write/report functions and test each independently (high feasibility, function already follows phased design internally)
- Add unit tests for descriptions.go and manifest.go YAML/JSON serialization (high feasibility, small files with clear contracts)
