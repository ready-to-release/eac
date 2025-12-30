# Release System

This repository uses a changelog-based release system. Releases are triggered by updating changelogs and merging to main.

## Changelog Location

Any module can have a changelog. The changelog location is defined in the module contract via the `files.changelog` field (defaults to `CHANGELOG.md` in the module's root).

**Releasable modules** place their changelogs in `release/<module>/CHANGELOG.md`. The `release-auto` workflow watches this directory and creates git tags when new versions are merged.

**Non-releasable modules** can have changelogs elsewhere (e.g., in their module root). These are for documentation only and don't trigger releases.

## Quick Start

```bash
# 1. Check if there are changes to release
r2r release pending r2r-cli

# 2. Update the changelog with a new version
r2r release this r2r-cli

# 3. Commit and create a PR
git add release/r2r-cli/CHANGELOG.md
git commit -m "release(r2r-cli): 0.1.0"
git push && gh pr create

# 4. After PR is merged, release-auto workflow creates the tag
# 5. Tag triggers the module's release workflow to build and publish
```

## How It Works

```text
┌──────────────────────────────────────────────────────────────────────────┐
│                           RELEASE FLOW                                   │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Developer                                                               │
│     │                                                                    │
│     ▼                                                                    │
│  ┌──────────────────┐                                                    │
│  │ release pending  │  Check for unreleased changes                      │
│  └────────┬─────────┘                                                    │
│           │                                                              │
│           ▼                                                              │
│  ┌──────────────────┐                                                    │
│  │ release this     │  Update changelog with new version                 │
│  └────────┬─────────┘                                                    │
│           │                                                              │
│           ▼                                                              │
│  ┌──────────────────┐                                                    │
│  │ Commit & PR      │  Create PR for review                              │
│  └────────┬─────────┘                                                    │
│           │                                                              │
│           ▼                                                              │
│  ┌──────────────────┐                                                    │
│  │ Merge to main    │  After approval                                    │
│  └────────┬─────────┘                                                    │
│           │                                                              │
│  ─────────┼───────────────────────────────────────────────────────────   │
│           │  CI/CD (automated)                                           │
│           ▼                                                              │
│  ┌──────────────────┐                                                    │
│  │ release-auto.yml │  Detects new version, creates git tag              │
│  └────────┬─────────┘                                                    │
│           │                                                              │
│           ▼                                                              │
│  ┌──────────────────┐                                                    │
│  │ release-*.yml    │  Builds and publishes artifacts                    │
│  └──────────────────┘                                                    │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

## Commands

### `release pending`

Check if a module has unreleased changes since the last release tag.

```bash
# Check single module
r2r release pending r2r-cli

# Check all modules
r2r release pending --all

# Quiet mode (exit code only: 0=has changes, 1=no changes)
r2r release pending r2r-cli --quiet
```

Output includes:
- `has_changes`: whether there are releasable changes
- `current_version`: the current released version
- `next_version`: the calculated next version
- `change_summary`: breakdown by change type (added, fixed, changed, etc.)

### `release this`

Finalize the changelog and prepare a module for release.

```bash
# Update changelog
r2r release this r2r-cli

# Preview without writing
r2r release this r2r-cli --dry-run

# Output as JSON
r2r release this r2r-cli --json

# Override release date
r2r release this r2r-cli --date 2024-01-15
```

This command:
1. Analyzes commits since the last release tag
2. Generates changelog entries from conventional commits
3. Calculates the next version (respecting constraints)
4. Updates the changelog with the new version
5. Outputs next steps for committing and creating a PR

### `release tag-pending`

Check for changelog versions that don't have corresponding git tags. Used by CI.

```bash
# Check single module
r2r release tag-pending r2r-cli

# Check all modules
r2r release tag-pending --all
```

### `release validate`

Validate changelog format and structure.

```bash
# Validate single module
r2r release validate r2r-cli

# Validate all modules
r2r release validate --all

# Output as JSON
r2r release validate r2r-cli --json
```

Checks performed:
- File exists at expected path
- Valid Keep a Changelog header format
- Version entries have valid format
- Version numbers are valid semver or calver
- Versions are in descending order
- No duplicate version numbers

## Changelog Format

Changelogs follow the [Keep a Changelog](https://keepachangelog.com/) format.

**How entries are generated:**
- Most entries are **auto-generated from conventional commits**
- The `## [Unreleased]` section is for **manual additions** (optional)
- When you run `release this`, manual entries in `[Unreleased]` are **merged** with commit-generated entries
- After release, `[Unreleased]` becomes empty again

**When to use `[Unreleased]`:**
- Adding entries not tied to a specific commit
- Rewording auto-generated entries before release
- Adding context or details beyond commit messages

```markdown
# Changelog

All notable changes to **module-name** will be documented in this file.

## [Unreleased]

### Added
- Manual entry: something not in a commit

## [0.1.0] - 2024-01-15

### Added
- feat: new feature description

### Changed
- refactor: changed behavior description

### Fixed
- fix: bug fix description

## [0.0.1] - 2024-01-01

### Added
- Initial release

[Unreleased]: https://github.com/org/repo/compare/module/0.1.0...HEAD
[0.1.0]: https://github.com/org/repo/compare/module/0.0.1...module/0.1.0
[0.0.1]: https://github.com/org/repo/releases/tag/module/0.0.1
```

### Change Types

| Type | Description | Conventional Commit |
|------|-------------|---------------------|
| Added | New features | `feat:` |
| Changed | Changes to existing functionality | `refactor:`, `perf:` |
| Deprecated | Soon-to-be removed features | `deprecate:` |
| Removed | Removed features | `remove:` |
| Fixed | Bug fixes | `fix:` |
| Security | Security fixes | `security:` |

## Versioning

### Semver (r2r-cli, ext-eac)

Semantic versioning: `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes (`feat!:`, `fix!:`, or `BREAKING CHANGE:`)
- **MINOR**: New features (`feat:`)
- **PATCH**: Bug fixes (`fix:`)

### Calver (docs)

Calendar versioning: `YYYY.MM.DD` or `YYYY.MM.DD.N` for multiple releases per day.

## Versioning Constraints

The `.r2r/definitions.yml` file can constrain version bumps:

```yaml
# Only allow patch version bumps (no minor or major)
versioning:
  constraint: patch-only

# Only allow calver (date-based versions)
versioning:
  constraint: calver-only
```

## Directory Structure

```
release/
├── README.md           # This file
├── r2r-cli/
│   └── CHANGELOG.md    # CLI changelog (semver) - RELEASABLE
├── ext-eac/
│   └── CHANGELOG.md    # Extension changelog (semver) - RELEASABLE
└── docs/
    └── CHANGELOG.md    # Documentation changelog (calver) - RELEASABLE

# Non-releasable modules can have changelogs in their own directories:
go/eac/core/
└── CHANGELOG.md        # Core library changelog - NOT releasable (documentation only)
```

### Making a Module Releasable

To make a module releasable via the automated workflow:

1. Set the changelog path in the module contract:
   ```yaml
   files:
     changelog: ../../release/<module>/CHANGELOG.md
   ```

2. Create the changelog file at `release/<module>/CHANGELOG.md`

3. The `release-auto` workflow will now watch for changes and create tags

## GitHub Workflows

### `release-auto.yml`

Triggers on push to main when `release/*/CHANGELOG.md` changes. Detects new versions and creates git tags.

**This workflow only watches `release/*/CHANGELOG.md`**. Changelogs in other locations are not monitored by this workflow.

### `release-r2r-cli.yml`

Triggers on `r2r-cli/*` tag push. Builds CLI binaries for multiple platforms and creates GitHub release.

### `release-ext-eac.yml`

Triggers on `ext-eac/*` tag push. Retags container images with release version.

### `release-docs.yml`

Triggers on `docs/*` tag push. Deploys documentation.

## Tag Format

Tags follow the pattern `<module>/<version>`:

- `r2r-cli/0.1.0`
- `ext-eac/1.0.0`
- `docs/2024.01.15`

## Manual Release (Emergency)

If you need to release without the automated flow:

```bash
# 1. Update changelog manually or with release this
r2r release this r2r-cli

# 2. Commit the changelog
git add release/r2r-cli/CHANGELOG.md
git commit -m "release(r2r-cli): 0.1.0"

# 3. Create and push the tag manually
git tag -a "r2r-cli/0.1.0" -m "Release r2r-cli v0.1.0"
git push origin main
git push origin "r2r-cli/0.1.0"
```

## Troubleshooting

### "No conventional commits found"

The release system requires conventional commit messages. Ensure commits follow the format:

```
type(scope): description

feat: add new feature
fix: resolve bug
refactor: improve code structure
```

### "No commits affecting module found"

Commits are filtered by module file patterns. Ensure your changes touch files owned by the module.

### Changelog validation errors

Run `r2r release validate <module>` to check for format issues. Common problems:
- Invalid version format
- Duplicate versions
- Versions out of order

### Tag already exists

If a tag already exists for a version, the release workflow will fail. Either:
- Delete the existing tag and release
- Increment the version number
