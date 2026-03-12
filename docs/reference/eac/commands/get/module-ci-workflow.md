# Get module-ci-workflow

<!-- book:cmd get module-ci-workflow -->

Returns the CI workflow filename for a module. Derives the workflow from convention (`ci-{moniker}.yaml`) or from the module's explicit workflow configuration. Exits with code 1 if the workflow file does not exist.

## Usage

```bash
eac get module-ci-workflow <moniker> [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `moniker` | Module moniker (required) |

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--basename` | bool | Output only the basename (default) |
| `--full-path` | bool | Output the full relative path |

## Output

Plain text workflow filename to stdout.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Workflow found |
| 1 | Module not found or no CI workflow file exists |

## Examples

```bash
eac get module-ci-workflow clie    # ci-clie.yaml
eac get module-ci-workflow core    # ci-core.yaml
```

## See Also

- [get](get.md)
- [get ci-workflows](ci-workflows.md)
