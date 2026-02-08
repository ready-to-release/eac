# show approval-comments

<!-- book:cmd show approval-comments -->

## Bundle Modules

For **container/bundle modules** with dependencies, approvals are **aggregated from all dependent modules**.

Example: When querying `eac-ext` (which depends on `eac-commands` and `clie-cli`):

```bash
eac show approval-comments eac-ext
```

Shows approvals from PRs containing:

- `specs/eac-commands/` (dependency)
- `specs/clie-cli/` (dependency)
- `specs/eac-ext/` (if any)

**Regular modules** only show approvals from their own `specs/<module>/` directory.

## How It Works

1. Scans git commits for PR references (`#123` or `Merge pull request #123`)
2. Fetches PR data via GitHub CLI (`gh pr view`)
3. Filters to PRs containing `.feature` files in `specs/<module>/`
4. Extracts reviews (APPROVED by default, use `--include-all-reviews` for all states)

## Requirements

- **GitHub CLI (`gh`)**: Command degrades gracefully if unavailable
- **Authentication**: `gh auth login` required

## Troubleshooting

### No Approvals Shown

Possible causes:

- No PRs merged in the version range
- PRs don't contain `.feature` files
- No APPROVED reviews (try `--include-all-reviews`)
- GitHub CLI unavailable

### gh Command Failed

```bash
# Install and authenticate GitHub CLI
gh auth login
```

## See Also

- [get approval-comments](../get/approval-comments.md) - JSON/YAML output
- [show specs](./specs.md) - View specifications
- [show Commands](../categories/show.md)
