# Get module-trigger-reason

<!-- book:cmd get module-trigger-reason -->

Extracts the CI trigger reason for a specific module from MODULE_STATUS JSON. Used in CI summaries to explain why each module was included in a CI run.

## Usage

```bash
eac get module-trigger-reason <module> [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `module` | Module name (required) |

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--json` | string | MODULE_STATUS JSON string (defaults to `MODULE_STATUS` env var) |

## Output

Plain text reason to stdout. Possible values:

| Output | Meaning |
|--------|---------|
| `files changed` | Module files changed since last CI success |
| `{dep} changed` | Dependency module triggered CI |
| `no previous CI` | No CI history for this module |
| `CI query failed` | GitHub API query failed |
| `unknown` | Module not found in status data |

## Examples

```bash
# From environment variable
export MODULE_STATUS='{"docs":{"reason":"files_changed_since_abc1234"}}'
eac get module-trigger-reason docs

# From explicit JSON
eac get module-trigger-reason docs --json "$MODULE_STATUS"
```

## See Also

- [get](get.md)
- [get changed-modules-ci](changed-modules-ci.md)
