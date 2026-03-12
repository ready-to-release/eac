# Validate dependabot

<!-- book:cmd validate dependabot -->

Validates that `.github/dependabot.yml` includes entries for all dependency sources discovered in the repository.

## Usage

```bash
eac validate dependabot
```

## What It Checks

Scans the repository for dependency sources and compares them against declared dependabot entries:

- **Go modules** -- Entries from `go.work` need `gomod` ecosystem entries.
- **npm packages** -- `package.json` files need `npm` ecosystem entries.
- **Python packages** -- `requirements.txt` files need `pip` ecosystem entries.
- **Docker images** -- `Dockerfile` files need `docker` ecosystem entries.
- **GitHub Actions** -- Workflows need a `github-actions` ecosystem entry.

Reports both missing entries (sources without coverage) and extra entries (no matching source found).

## Examples

```bash
eac validate dependabot
```

## Common Errors

- **Missing dependabot entries** -- A dependency source has no corresponding entry. Fix with `eac update dependabot`.
- **Extra dependabot entries** -- An entry has no matching source. Fix with `eac update dependabot --prune`.

## See Also

- [validate](./validate.md)
