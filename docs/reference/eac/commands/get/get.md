# get

<!-- book:cmd get -->

## Subcommands

All get commands return JSON output for automation and scripting. Process with `jq` or pipe to other tools.

See the [get Commands Category](./index.md) for complete list of subcommands.

## Common Patterns

```bash
# Cache expensive queries
eac get files > files.json
jq '.files[] | select(.module == "src-auth")' files.json

# Build changed modules
CHANGED=$(eac get changed-modules | jq -r '.changed_modules[]')
eac build $CHANGED
```

## See Also

- [show Commands](../show/index.md) - Human-readable output
