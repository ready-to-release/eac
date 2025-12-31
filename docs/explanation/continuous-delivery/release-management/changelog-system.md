# Changelog-Based Release System

## Overview

This repository uses a **changelog-based release system**. Releases are triggered by updating changelogs and merging to main - no manual tagging required.

**Key principle:** The changelog is the source of truth for releases. When a new version appears in a changelog and merges to main, the release workflow automatically creates a git tag and triggers the appropriate release pipeline.

---

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

# 4. After PR merges, release-auto workflow creates the tag
# 5. Tag triggers the module's release workflow to build and publish
```

---

## How It Works

```text
Developer
    │
    ▼
┌──────────────────┐
│ release pending  │  Check for unreleased changes
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ release this     │  Update changelog with new version
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Commit & PR      │  Create PR for review
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Merge to main    │  After approval
└────────┬─────────┘
         │
─────────┼─────────────────────────────────────────
         │  CI/CD (automated)
         ▼
┌──────────────────┐
│ release-auto.yml │  Detects new version, creates git tag
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ release-*.yml    │  Builds and publishes artifacts
└──────────────────┘
```

---

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
- `change_summary`: breakdown by change type

### `release this`

Finalize the changelog and prepare a module for release.

```bash
# Update changelog
r2r release this r2r-cli

# Preview without writing
r2r release this r2r-cli --dry-run

# Output as JSON
r2r release this r2r-cli --json
```

This command:

1. Analyzes commits since the last release tag
2. Generates changelog entries from conventional commits
3. Calculates the next version (respecting constraints)
4. Updates the changelog with the new version

### `release tag-pending`

Check for changelog versions without corresponding git tags. Used by CI.

```bash
r2r release tag-pending r2r-cli
r2r release tag-pending --all
```

### `release validate`

Validate changelog format and structure.

```bash
r2r release validate r2r-cli
r2r release validate --all
```

Checks: file exists, valid header format, valid semver/calver, versions in order, no duplicates.

---

## Changelog Format

Changelogs follow [Keep a Changelog](https://keepachangelog.com/) format.

```markdown
# Changelog

All notable changes to **module-name** will be documented in this file.

## [Unreleased]

### Added
- Manual entry: something not in a commit

## [0.1.0] - 2024-01-15

### Added
- feat: new feature description

### Fixed
- fix: bug fix description

## [0.0.1] - 2024-01-01

### Added
- Initial release

[Unreleased]: https://github.com/org/repo/compare/module/0.1.0...HEAD
[0.1.0]: https://github.com/org/repo/compare/module/0.0.1...module/0.1.0
```

### Entry Generation

- Most entries are **auto-generated from conventional commits**
- The `[Unreleased]` section is for **manual additions** (optional)
- When you run `release this`, manual entries merge with commit-generated entries
- After release, `[Unreleased]` becomes empty

### Change Types

| Type       | Description                       | Conventional Commit  |
| ---------- | --------------------------------- | -------------------- |
| Added      | New features                      | `feat:`              |
| Changed    | Changes to existing functionality | `refactor:`, `perf:` |
| Deprecated | Soon-to-be removed features       | `deprecate:`         |
| Removed    | Removed features                  | `remove:`            |
| Fixed      | Bug fixes                         | `fix:`               |
| Security   | Security fixes                    | `security:`          |

---

## Versioning

### Semver (r2r-cli, ext-eac)

Semantic versioning: `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes (`feat!:`, `fix!:`, or `BREAKING CHANGE:`)
- **MINOR**: New features (`feat:`)
- **PATCH**: Bug fixes (`fix:`)

### Constraints

The `.r2r/eac/repository.yml` file can constrain version bumps:

```yaml
versioning:
  constraint: patch-only    # Only allow patch bumps
  constraint: calver-only   # Only allow date-based versions
```

---

## Directory Structure

```text
release/
├── r2r-cli/
│   └── CHANGELOG.md    # CLI changelog (semver)
├── ext-eac/
    └── CHANGELOG.md    # Extension changelog (semver)
```

**Releasable modules** place changelogs in `release/<module>/CHANGELOG.md`. The `release-auto` workflow watches this directory.

**Non-releasable modules** can have changelogs elsewhere for documentation only.

---

## GitHub Workflows

| Workflow              | Trigger                                            | Action                                      |
| --------------------- | -------------------------------------------------- | ------------------------------------------- |
| `release-auto.yml`    | Push to main with `release/*/CHANGELOG.md` changes | Creates git tag                             |
| `release-r2r-cli.yml` | `r2r-cli/*` tag                                    | Builds CLI binaries, creates GitHub release |
| `release-ext-eac.yml` | `ext-eac/*` tag                                    | Retags container images                     |

### Tag Format

Tags follow the pattern `<module>/<version>`:

- `r2r-cli/0.1.0`
- `ext-eac/1.0.0`

---

## Troubleshooting

### "No conventional commits found"

Ensure commits follow the format:

```text
type(scope): description

feat: add new feature
fix: resolve bug
refactor: improve code structure
```

### "No commits affecting module found"

Commits are filtered by module file patterns. Ensure changes touch files owned by the module.

### Changelog validation errors

Run `r2r release validate <module>` to check for format issues:

- Invalid version format
- Duplicate versions
- Versions out of order

---

## Next Steps

- [Release Notes](./release-notes.md) - How to write effective release notes
- [Release Evidence](./release-evidence.md) - What evidence is required for approval
- [Commit Messages](../workflow/commit-messages.md) - Conventional commit format
