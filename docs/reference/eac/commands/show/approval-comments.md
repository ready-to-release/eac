# show approval-comments

<!-- book:cmd show approval-comments -->

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
