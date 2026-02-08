# releasenotes

Parses, validates, and generates RELEASE-NOTES.md files for versioned
module releases.

## Key Types

- **`ReleaseNotes`** -- Parsed release notes with version entries
- **`ReleaseNotesVersion`** -- Single version entry with date and sections
- **`ReleaseNotesSection`** -- Section header and content within a version

## Patterns

- Regex-based parsing: extracts `## [version] - date` and `### section` headers
- Template generation: creates new release notes from configurable templates
- Version validation: checks that a version exists and has non-empty content

## Internal Structure

| File | Responsibility |
| --- | --- |
| validator.go | `Parse`, `ParseContent`, `ValidateVersion`, `GenerateTemplate` |

## Dependencies

- `core/config` -- repository config for template path resolution
- `core/environments` -- container root env var for template lookup

## Role in System

The `releasenotes` package supports the release pipeline in `core`. The
`release-changelog` command uses it to validate that release notes exist
for a version before publishing, and `GenerateTemplate` scaffolds new
release note files from repository templates.

## Code Health

### Tech Debt
- `validator.go:52-54`: regexes are recompiled on every `ParseContent()` call; should be package-level `var` for efficiency

### Pain Points
- `ParseContent()` uses string concatenation (`currentSection.Content += line + "\n"`) for building content; `strings.Builder` would be more efficient for large files

### Optimization Opportunities
- Hoist `versionHeaderRegex` and `sectionHeaderRegex` to package-level compiled vars (low effort, avoids repeated compilation)
- Package is well-tested (397-line test file, ~1.9x test-to-code ratio); no major structural issues
