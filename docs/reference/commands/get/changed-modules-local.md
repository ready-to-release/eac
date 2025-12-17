# Get changed-modules-local

<!-- book:cmd get changed-modules-local -->

## Use Cases

### Local Development

Use this command during local development to avoid unnecessary rebuilds:

```bash
# Only rebuild what changed since last build
r2r eac get changed-modules-local | jq -r '.modules[]' | xargs -L1 r2r eac build
```

### Pre-Commit Hook

Integrate into pre-commit hooks to validate only affected modules:

```bash
# Validate only modules that changed locally
CHANGED=$(r2r eac get changed-modules-local | jq -r '.modules[]')
for module in $CHANGED; do
  r2r eac validate artifacts $module
done
```

### Build Status Reporting

Report on build freshness and required work:

```bash
# Generate build status report
RESULT=$(r2r eac get changed-modules-local)
echo "Modules needing rebuild: $(echo "$RESULT" | jq -r '.modules | length')"
echo "Up-to-date modules: $(echo "$RESULT" | jq -r '.up_to_date | length')"
echo "$RESULT" | jq -r '.change_reasons | to_entries[] | "\(.key): \(.value)"'
```

## Comparison with Related Commands

| Command                     | Purpose                                      | Use When                                       |
| --------------------------- | -------------------------------------------- | ---------------------------------------------- |
| `get changed-modules-local` | Modules needing rebuild based on build state | Local development, incremental builds          |
| `get changed-modules`       | Modules affected by git changes              | Working with uncommitted changes               |
| `get changed-modules-ci`    | Modules needing rebuild in CI                | CI/CD pipelines, comparing against base commit |

## See Also

- [get changed-modules](./changed-modules.md) - Git-based change detection
- [get changed-modules-ci](./changed-modules-ci.md) - CI pipeline change detection
- [build](../build/build.md) - Build modules
- [validate artifacts](../validate/artifacts.md) - Validate build outputs
