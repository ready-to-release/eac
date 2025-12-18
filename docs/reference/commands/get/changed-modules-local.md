# get changed-modules-local

<!-- book:cmd get changed-modules-local -->

## Use Cases

### Local Development

Only rebuild modules that changed since last build:

```bash
r2r eac get changed-modules-local | jq -r '.modules[]' | xargs -L1 r2r eac build
```

### Build Status Reporting

```bash
RESULT=$(r2r eac get changed-modules-local)
echo "Modules needing rebuild: $(echo "$RESULT" | jq -r '.modules | length')"
```

## Comparison

| Command | Purpose | Use When |
|---------|---------|----------|
| `get changed-modules-local` | Modules needing rebuild based on build state | Local development |
| `get changed-modules` | Modules affected by git changes | Working with uncommitted changes |
| `get changed-modules-ci` | Modules needing rebuild in CI | CI/CD pipelines |

## See Also

- [get changed-modules](./changed-modules.md) - Git-based change detection
- [get changed-modules-ci](./changed-modules-ci.md) - CI pipeline change detection
- [build](../build/build.md) - Build modules
- [get Commands](../categories/get.md)
