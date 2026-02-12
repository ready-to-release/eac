# changelog

Parsing, generation, and version management for Keep a Changelog format files,
with conventional commit integration and support for both semantic versioning
and calendar versioning.

## Key Types

- **`Changelog`** -- Complete parsed changelog with unreleased section, version history, and repo URL
- **`Version`** -- Single version entry with categorized entries (Added, Changed, Fixed, etc.)
- **`Entry`** -- Individual changelog item with optional conventional commit metadata
- **`Commit`** -- Parsed conventional commit with type, scope, and breaking change detection
- **`ChangeType`** -- Category enum (Added, Changed, Deprecated, Removed, Fixed, Security)
- **`VersionType`** -- Versioning scheme enum (Semver, Calver)
- **`BumpType`** -- Version bump level enum (None, Patch, Minor, Major)

## Patterns

- Keep a Changelog format: `## [version] - date` headers with `### Category` sections
- Conventional commit parsing: `type(scope)!: description` format with body BREAKING CHANGE detection
- Dual versioning: semver bump calculation and calver date-based generation with collision suffixes
- Roundtrip fidelity: `Parse` and `String` produce consistent markdown output
- Commit-to-entry mapping: `CommitTypeToChangeType` converts commit types to changelog categories
- Unreleased promotion: `PromoteUnreleased` moves pending entries to a new version
- Module filtering: `FilterCommitsByModule` scopes commits to files matching module glob patterns

## Internal Structure

| File | Responsibility |
| --- | --- |
| types.go | `Changelog`, `Version`, `Entry`, `ChangeType`, `VersionType` |
| parser.go | `Parse`, `ParseString`, `ParseReader` with regex-based line parsing |
| writer.go | `String`, `Write` with link definition generation |
| conventional.go | `Commit`, `ParseCommitMessage`, `CommitsToVersion`, `FilterCommitsByModule` |
| semver.go | `BumpType`, `ParseSemver`, `BumpSemver`, `NextCalver`, `VersionCalcOptions`, `CalculateNextVersion`, `CalculateNextVersionConstrained` |
| testhelpers_test.go | Shared test fixtures (`makeEntry`, `makeCommit`, `makeVersion`, `testDate`, `minimalChangelog`) |

## Dependencies

No internal repository imports. This is a leaf package.

## Role in System

This package provides the changelog infrastructure used by the release pipeline.
The `release changelog` command parses existing CHANGELOG.md files, collects
conventional commits since the last release, generates new version entries, and
writes updated changelogs. The semver and calver support ensures each module
can use its preferred versioning scheme.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- Clean leaf package with good separation of concerns; no structural changes needed
- Excellent test coverage with conventional_test.go, parser_test.go, semver_test.go, and testhelpers_test.go
- All files are reasonably sized (largest is parser.go at ~250 lines, semver.go at ~280 lines)
