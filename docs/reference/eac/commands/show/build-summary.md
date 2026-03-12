# Show build-summary

<!-- book:cmd show build-summary -->

Generate a build summary for a module, formatted as Markdown suitable for `$GITHUB_STEP_SUMMARY`.

Reads UoW manifests from `out/build/<module>/` to derive build status automatically. On success, shows artifact metrics. On failure, shows diagnostic information including build log tails.

## Usage

```
eac show build-summary <module> [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `module` | Module name (required) |

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--run-id` | string | GitHub Actions run ID for linking to workflow |

## Output Sections

**On success:**
- **Header**: module name with build emoji
- **Status**: component types that were built
- **Build Output**: artifact verification table (artifact pattern, type, size/file count)
- **Artifacts**: artifact bundle name for download
- **Build Configuration** (collapsible): component types, container runtime, output directory

**On failure:**
- **Header**: module name with failure indicator
- **Status**: failure message
- **Diagnostics**: last 30 lines of each component build log
- **Timing**: build timing data if available
- **Build Configuration** (collapsible): same as success

## Examples

```bash
# Generate build summary for a module
eac show build-summary core

# With run ID for CI linking
eac show build-summary eac --run-id=12345678

# Redirect to GitHub Actions step summary
eac show build-summary core >> "$GITHUB_STEP_SUMMARY"
```

## See Also

- [build](../build/build.md) - Build modules
- [show build-times](./build-times.md) - Performance analysis
- [get build-times](../get/build-times.md) - JSON output
