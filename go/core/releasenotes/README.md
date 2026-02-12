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

| File         | Responsibility                                                 |
| ------------ | -------------------------------------------------------------- |
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

- None identified

### Pain Points

- validator_test.go is 397 lines, exceeds 300-line threshold

### Optimization Opportunities

- None identified
