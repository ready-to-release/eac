# pdf-screenshots

Implements the `update pdf-screenshots` command that extracts PDF pages as PNG images for documentation caching. Scans build output for PDF books and uses Docker-based `pdftoppm` to extract pages with SHA256 hash-based cache invalidation.

## Key Types

- **`PDFInfo`** -- Metadata for a discovered PDF file (path, module, book name, hash, cache directory)

## Key Functions

- **`UpdatePDFScreenshots()`** -- Entry point for the `update pdf-screenshots` command
- **`scanForPDFs()`** -- Walk `out/build/` to discover PDF files, optionally filtered by module
- **`hashFile()`** -- Compute SHA256 hash of a file (first 12 characters) for cache invalidation
- **`cacheValidWithHash()`** -- Check if cache directory has correct hash marker and PNG files
- **`countPages()`** -- Count PNG files in a cache directory
- **`ensurePDFToolsImage()`** -- Build the `pdf-cli-oci` Docker image if not already present
- **`extractPages()`** -- Use Docker container with `pdftoppm` to extract PDF pages as PNG images
- **`renameToZeroPadded()`** -- Rename extracted page files to zero-padded format for proper sorting

## Patterns

- `init()` registration: registers command function with the global registry
- Hash-based caching: SHA256 hash markers detect when PDFs have changed and pages need re-extraction
- Docker-based tool execution: uses `pdf-cli-oci` Docker image for PDF processing with volume mounts
- Dry-run support: `--dry-run` flag shows what would be done without making changes
- Module filtering: `--module` flag limits processing to a single module's PDFs

## Internal Structure

| File | Responsibility |
| --- | --- |
| update.go | PDF discovery, cache validation, Docker-based page extraction, and zero-padded renaming (490 lines) |

## Dependencies

- `adapters/docker` -- Docker client and container execution
- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration
- `core/logging` -- structured logging
- `core/paths` -- build output and cache directory path resolution
- `core/repository` -- repository root discovery

## Role in System

The `pdf-screenshots` sub-package maintains a cache of PDF page images for documentation workflows. When PDF books are generated during the build process, this command extracts individual pages as PNG images that can be embedded in web-based documentation or used for visual regression testing.

## Code Health

### Tech Debt
- `update.go` (490 lines) handles discovery, caching, Docker management, extraction, and renaming in a single file

### Pain Points
- Requires Docker to be running and the `pdf-cli-oci` image to be buildable

### Optimization Opportunities
- Extract PDF scanning and cache validation into separate files (moderate effort, improves readability)
