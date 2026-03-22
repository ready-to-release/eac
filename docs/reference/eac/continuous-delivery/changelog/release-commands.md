# Release Commands

CLI commands for the changelog-based release system.

## Quick Reference

```bash
# Check for pending changes
eac release pending <module>
eac release pending --all

# Finalize changelog for release
eac release this <module>

# Check for untagged versions
eac release tag-pending <module>

# Validate changelog format
eac release validate <module>
```

---

## `release pending`

Check if a module has unreleased changes since the last release tag.

```bash
# Check single module
eac release pending clie

# Check all modules
eac release pending --all

# Quiet mode (exit code only: 0=has changes, 1=no changes)
eac release pending clie --quiet
```

**Output includes**:

- `has_changes`: whether there are releasable changes
- `current_version`: the current released version
- `next_version`: the calculated next version
- `change_summary`: breakdown by change type

---

## `release this`

Finalize the changelog and prepare a module for release.

```bash
# Update changelog
eac release this clie

# Preview without writing
eac release this clie --dry-run

# Output as JSON
eac release this clie --json
```

**What it does**:

1. Analyzes commits since the last release tag
2. Generates changelog entries from conventional commits
3. Calculates the next version (respecting constraints)
4. Updates the changelog with the new version

---

## `release tag-pending`

Check for changelog versions without corresponding git tags. Used by CI.

```bash
eac release tag-pending clie
eac release tag-pending --all
```

---

## `release validate`

Validate changelog format and structure.

```bash
eac release validate clie
eac release validate --all
```

**Checks**:

- File exists
- Valid header format
- Valid semver/calver
- Versions in order
- No duplicates

---

## Directory Structure

```text
release/
├── clie/
│   └── CHANGELOG.md    # CLI changelog (semver)
├── eac-ext/
    └── CHANGELOG.md    # Extension changelog (semver)
```

**Releasable modules** place changelogs in `release/<module>/CHANGELOG.md`. The `release-auto` workflow watches this directory.

---

## Release Workflows and Tagging

For GitHub workflow triggers and tag format details, see:

- [Release Workflows Reference](../workflows/release-workflows.md) — GitHub Actions workflow files and triggers
- [Versioning Reference](./versioning.md) — Tag format (`<module>/<version>`) and version constraints

---

## Versioning Constraints

The `.eac/repository.yml` file can constrain version bumps:

```yaml
versioning:
  constraint: patch-only    # Only allow patch bumps
  constraint: calver-only   # Only allow date-based versions
```

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

Run `eac release validate <module>` to check for format issues:

- Invalid version format
- Duplicate versions
- Versions out of order

---

## Related Documentation

- [Format Specification](./format-specification.md) - Keep a Changelog format details
- [Versioning](./versioning.md) - Semantic versioning rules
- [Release Command Reference](../../commands/release/index.md) - Full release command options
