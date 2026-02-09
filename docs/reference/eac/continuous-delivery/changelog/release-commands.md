# Release Commands

CLI commands for the changelog-based release system.

## Quick Reference

```bash
# Check for pending changes
clie release pending <module>
clie release pending --all

# Finalize changelog for release
clie release this <module>

# Check for untagged versions
clie release tag-pending <module>

# Validate changelog format
clie release validate <module>
```

---

## `release pending`

Check if a module has unreleased changes since the last release tag.

```bash
# Check single module
clie release pending clie

# Check all modules
clie release pending --all

# Quiet mode (exit code only: 0=has changes, 1=no changes)
clie release pending clie --quiet
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
clie release this clie

# Preview without writing
clie release this clie --dry-run

# Output as JSON
clie release this clie --json
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
clie release tag-pending clie
clie release tag-pending --all
```

---

## `release validate`

Validate changelog format and structure.

```bash
clie release validate clie
clie release validate --all
```

**Checks**:

- File exists
- Valid header format
- Valid semver/calver
- Versions in order
- No duplicates

---

## Release Workflow

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

## GitHub Workflows

| Workflow              | Trigger                                            | Action                                      |
| --------------------- | -------------------------------------------------- | ------------------------------------------- |
| `release-auto.yml`    | Push to main with `release/*/CHANGELOG.md` changes | Creates git tag                             |
| `release-clie.yml` | `clie/*` tag                                    | Builds CLI binaries, creates GitHub release |
| `release-eac-ext.yml` | `eac-ext/*` tag                                    | Retags container images                     |

### Tag Format

Tags follow the pattern `<module>/<version>`:

- `clie/0.1.0`
- `eac-ext/1.0.0`

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

Run `clie release validate <module>` to check for format issues:

- Invalid version format
- Duplicate versions
- Versions out of order

---

## Related Documentation

- [Format Specification](./format-specification.md) - Keep a Changelog format details
- [Versioning](./versioning.md) - Semantic versioning rules
- [Release Command Reference](../../commands/release/index.md) - Full release command options
