# Update Dependabot

<!-- book:cmd update dependabot -->

Scans the repository for all dependency sources and updates `.github/dependabot.yml` to match.

Preserves existing entries and their customizations (schedule, labels, etc.). New entries are added for discovered but uncovered dependency sources. This is the fix counterpart to `validate dependabot` (which only checks).

## Usage

```bash
eac update dependabot [flags]
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--dry-run` | | Show what would change without modifying the file |
| `--prune` | | Remove entries that have no matching dependency source |
| `--verbose` | `-v` | Show detailed discovery output |

## Examples

```bash
eac update dependabot                # Add missing entries
eac update dependabot --prune        # Also remove stale entries
eac update dependabot --dry-run      # Preview changes
```

## See Also

- [validate dependabot](../validate/dependabot.md)
- [update Commands](../../categories/update.md)
