# navigation

Navigation generation and macro management for the docprep pipeline, ensuring every staging directory has valid `.nav.yml` files and handling Jinja2 macro injection or stripping based on output mode.

## Key Types

- **`NavFile`** -- Parsed `.nav.yml` structure with title and nav entries, serialized via YAML tags
- **`DiataxisSections`** -- Map of Diataxis documentation sections that receive breadcrumb and footer macros

## Patterns

- Auto-generation: Directories without `.nav.yml` get one generated from their markdown files, sorted alphabetically with titles extracted from headings
- Validation and cleanup: Existing `.nav.yml` entries referencing missing files are removed; unreferenced markdown files are appended via `ValidateAndCleanNav`
- Mode-dependent macros: Site mode injects breadcrumb and footer macros; PDF mode strips all Jinja2 macro calls and nav titles
- Root index guarantee: `EnsureRootIndex` generates a root `index.md` with auto-generated table of contents if none exists from source
- Title extraction: Uses frontmatter `title:` field first, falls back to first `#` heading, then filename-based title via `FilenameToTitle`
- Order-aware sorting: Files from command sources receive explicit sort order from book configuration; other files default to order 500
- Content-aware directory inclusion: Subdirectories are only added to navigation if they contain markdown content, checked recursively via `HasMarkdownContent`

## Internal Structure

| File | Responsibility |
| --- | --- |
| generator.go | `EnsureNavigationStructure`, `GenerateNavForDir`, title extraction, and file ordering helpers |
| index.go | `EnsureRootIndex` creates root `index.md` with auto-generated table of contents |
| macros.go | `InjectMacros`, `StripMacros`, and `StripNavTitles` for Jinja2 handling by output mode |
| validator.go | `ValidateNavEntry` and `ValidateAndCleanNav` for `.nav.yml` integrity checking and repair |

## Dependencies

- `core/config` -- book configuration for site URL, copy sources, and file ordering
- `core/paths` -- `.nav.yml` path conventions and navigation config file resolution
- `docprep/staging` -- `FileIndex` for markdown file iteration across the staging directory

## Role in System

The navigation package implements phases 7 and 10 of the docprep preprocessing pipeline in `eac-cli`. Phase 7 ensures every directory in staging has a valid `.nav.yml` and that a root `index.md` exists, providing MkDocs with the navigation structure it needs. It validates existing nav files against the actual directory contents and generates new ones where missing.

Phase 10 handles output-mode-specific macro processing. For site builds, it injects breadcrumb and Diataxis footer macros into markdown files in the standard documentation sections. For PDF builds, it strips all Jinja2 macro calls and nav titles since the PDF rendering container does not support the macros plugin.

Hidden directories (dot-prefixed) and `assets` directories are excluded from both navigation generation and directory walking.

## Code Health

### Tech Debt
- `generator.go`: `GenerateNavForDir` is ~109 lines (73-181) with scanning, sorting, title extraction, and YAML writing in one function
- Single test file `navigation_test.go` covers all four source files; per-file test files would improve isolation

### Pain Points
- Five compiled regex patterns spread across `macros.go` and `index.go` as package-level vars; adding a new macro pattern requires knowing which file to update

### Optimization Opportunities
- Split `GenerateNavForDir` into scan, sort, and emit steps to improve testability of each step independently -- low effort
