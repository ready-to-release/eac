# get approval-comments

<!-- book:cmd get approval-comments -->

## Bundle Modules

For **container/bundle modules** with dependencies, approvals are **aggregated from all dependent modules**.

Example: `ext-eac` depends on `eac-commands` and `r2r-cli`:

```bash
eac get approval-comments ext-eac --as-json
```

Returns approvals for PRs containing spec files from:

- `specs/eac-commands/` (dependency)
- `specs/r2r-cli/` (dependency)
- `specs/ext-eac/` (if any)

**Regular modules** only return approvals from their own `specs/<module>/` directory.

## Requirements

- **GitHub CLI (`gh`)**: Command degrades gracefully if unavailable
- **Authentication**: `gh auth login` required

## See Also

- [show approval-comments](../show/approval-comments.md) - Human-readable output
- [get specs](./specs.md) - Specification data
- [get Commands](../categories/get.md)
