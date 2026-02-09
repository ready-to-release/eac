# linking

Link rewriting engine for the docprep pipeline that translates source-relative paths in markdown to staging-relative paths after files are copied to the staging directory.

## Key Types

- **`TranslationMap`** -- Manages per-file link translations from source to staging paths, with external link handling and PDF mode support
- **`LinkTranslation`** -- Holds old-to-new path mappings for a single staged file, including source and staging file paths
- **`CalculateRelativePath`** -- Computes a relative path from a staging markdown file to a target absolute path

## Patterns

- Two-phase translation: First builds a translation table from file mappings via `BuildTranslations`, then applies all translations to staged files via `ApplyAllTranslations`
- Code-block-safe rewriting: `ReplaceOutsideCodeBlocks` splits content around fenced code blocks to avoid corrupting code examples
- Link-context-only replacement: `ReplaceLinkPaths` only modifies paths inside markdown links, images, and HTML attributes
- External link stripping: In PDF mode (`StripExternal`), links to files outside staging are stripped (text preserved, href removed) to avoid dead links in the PDF output
- Anchor preservation: Path replacements preserve `#anchor` fragments attached to links
- Longest-first replacement: Translations are sorted by path length (longest first) to prevent partial replacements of overlapping paths
- Jinja2 template skipping: Links containing `{{ }}` template expressions are left untouched
- Docs-relative tracking: `TranslationMap` records docs-relative paths for each source file to enable external link resolution

## Internal Structure

| File | Responsibility |
| --- | --- |
| translator.go | `TranslationMap` with file mapping tracking, `BuildTranslations`, `ApplyAllTranslations`, and `CalculateRelativePath` |
| extractor.go | `ExtractRelativeLinks`, `StripCodeBlocks`, and link pattern definitions for markdown and HTML contexts |
| rewriter.go | `ReplaceOutsideCodeBlocks`, `ReplaceLinkPaths`, and `ReplaceMarkdownLinks` for safe path substitution |

## Dependencies

- `core/paths` -- `DocsSourcePath` for docs-relative path computation

## Role in System

The linking package implements phase 5 of the docprep preprocessing pipeline in `eac`. When documentation sources are copied from multiple directories into a flat staging structure, internal cross-references break because the relative path relationships between files change.

The `TranslationMap` receives file mapping data from the copy phase, computes the necessary path translations by comparing source and staging locations, then rewrites all markdown links, image references, and HTML src/href attributes to point to the correct staging-relative paths.

For links that reference files in the docs tree but outside the current book's staging, it rewrites them to absolute site URLs. In PDF mode, it additionally strips these external links entirely, preserving link text while removing the href to ensure clean PDF output.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified -- all three source files have matching test files, and the package is compact (476 lines total)

### Optimization Opportunities
- None identified -- code-block-safe rewriting and longest-first replacement are already well-optimized for correctness
