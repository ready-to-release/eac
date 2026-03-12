# Validate artifacts

<!-- book:cmd validate artifacts -->

Validates that build artifacts exist for a module and all its transitive dependencies. Ensures the build-to-test flow has all necessary artifacts before running tests.

## Usage

```bash
eac validate artifacts <module> [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `module` | Module moniker to validate (required) |

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--skip-depm` | bool | Skip validation of transitive module dependencies (for release workflows) |
| `--os` | string | Target OS for platform-specific artifacts (default: current OS) |
| `--arch` | string | Target architecture (default: current arch) |

## What It Checks

- Target module artifacts in `out/build/` (executables, files, directories).
- All transitive dependency artifacts (recursive, unless `--skip-depm`).
- Platform-specific artifacts for the target OS/architecture.
- Marker files for modules with no traditional build outputs.

## Examples

```bash
# Validate module and all dependencies
eac validate artifacts eac-cli

# Cross-platform check
eac validate artifacts clie --os linux --arch amd64

# Release context (skip dependency check)
eac validate artifacts docs --skip-depm
```

## Common Errors

- **module not found** -- The moniker does not exist. Check `eac show modules`.
- **Missing artifacts detected** -- Required build outputs are missing. Run `eac build <module>`.

## See Also

- [show artifacts](../show/artifacts.md)
- [validate](./validate.md)
