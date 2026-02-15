# cleanup

Post-preprocessing cleanup that fixes broken internal links and removes unreferenced asset files from the staging directory before MkDocs rendering.

## Key Types

- **`AssetExtensions`** -- Map of file extensions considered assets (`.png`, `.jpg`, `.jpeg`, `.gif`, `.svg`, `.webp`, `.ico`, `.pdf`)
- **`LinkWithContextPattern`** -- Regex matching markdown links with preceding context for image-link skip detection
- **`CleanupUnreferencedAssets`** -- Function that scans markdown for asset references, deletes unreferenced files, and prunes empty directories
- **`FixBrokenInternalLinks`** -- Function that converts dead internal links to absolute site URLs for the configured `site_url`
- **`FixLinksInContent`** -- Pure function that processes a single file's content and returns modified content with fix count

## Patterns

- Two-pass asset cleanup: First collects all referenced asset paths from markdown, then deletes unreferenced files and prunes empty directories
- Conservative link fixing: Only rewrites links whose targets do not exist in staging, converting them to absolute site URLs
- Fallback matching: Asset reference detection uses both normalized absolute path and basename matching to avoid false positives from path differences
- Anchor preservation: Link fixing preserves `#anchor` fragments when rewriting the path portion to a site URL
- Iterative directory pruning: Repeatedly removes empty directories after asset deletion until no more empty directories remain, handling nested empties
- Image link exclusion: `LinkWithContextPattern` captures the preceding character to detect and skip image links (preceded by `!`)
- External link skipping: Both cleanup functions skip external URLs, mailto links, and anchor-only references
- Multi-resolution target checking: `FixLinksInContent` checks exact path, `.md` suffix, and `index.md` directory matches before declaring a link broken

## Internal Structure

| File | Responsibility |
| --- | --- |
| assets.go | `CleanupUnreferencedAssets` scans markdown for asset references, deletes unreferenced images, and prunes empty directories |
| linkfix.go | `FixBrokenInternalLinks` and `FixLinksInContent` convert dead internal links to absolute site URL references |

## Dependencies

- `docprep/staging` -- `FileIndex` for iterating markdown files in the staging directory

## Role in System

The cleanup package runs as the final two phases (13 and 14) of the docprep preprocessing pipeline in `eac`. After all content transformation, link rewriting, and diagram processing are complete, `FixBrokenInternalLinks` converts any remaining dead internal links to absolute site URLs using the book's configured `site_url`. It handles multiple target resolution strategies, checking for exact file matches, `.md` suffix matches, and `index.md` directory matches before determining a link is broken.

Then `CleanupUnreferencedAssets` removes images, PDFs, and other asset files that no markdown file references, along with any empty directories left behind.

These phases are intentionally last in the pipeline because earlier phases (diagram processing, content insertion) may add or remove asset references. Running cleanup after all other transformations ensures accurate reference counting and prevents stale content from reaching the final rendered output.

## Code Health

### Tech Debt

- None identified

### Pain Points

- None identified

### Optimization Opportunities

- None identified
