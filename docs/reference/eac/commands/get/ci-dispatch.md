# Get ci-dispatch

<!-- book:cmd get ci-dispatch -->

Filters modules for CI dispatch by checking whether invalidated modules already have valid CI at the current HEAD. Directly changed modules are always dispatched. The dispatch list is topologically sorted by CI dependencies.

## Usage

```bash
eac get ci-dispatch [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--directly-changed` | string | Space-separated list of directly changed modules (always dispatched) |
| `--invalidated` | string | Space-separated list of invalidated modules (checked for valid CI) |
| `--head-sha` | string | HEAD SHA to check against (defaults to auto-detection) |
| `--mock` | string | Mock CI status as JSON for testing (e.g. `'{"mod": true}'`) |
| `--format` | string | `shell` for eval-friendly output; otherwise standard get formats |
| `--as-yaml` | | Output as YAML (default) |
| `--as-json` | | Output as JSON |

## How It Works

1. Directly changed modules are always added to the dispatch list
2. Invalidated modules are checked against GitHub for successful CI at the HEAD SHA
3. Modules with valid CI are skipped; others are dispatched
4. The dispatch list is topologically sorted (modules with no CI deps first)

## Output Fields (structured)

| Field | Description |
|-------|-------------|
| `dispatch` | Modules to dispatch, in dependency order |
| `skipped` | Modules with valid CI at HEAD |
| `reasons` | Per-module reasoning for dispatch/skip |
| `ci_dependencies` | Dependency graph within the dispatch set |
| `head_sha` | The SHA used for checking |
| `total_modules` | Total modules considered |

## Shell Format Output

```
DISPATCH="mod1 mod2 mod3"
SKIPPED="mod4"
DISPATCH_COUNT=3
SKIPPED_COUNT=1
CI_DEPS_JSON='{"mod3":["mod1"]}'
```

## Examples

```bash
# Normal CI usage
eac get ci-dispatch --directly-changed "core" --invalidated "eac-cli docs"

# Shell format for workflow scripting
eval $(eac get ci-dispatch --format shell --invalidated "clie eac-ext")
for module in $DISPATCH; do
  gh workflow run "ci-${module}.yaml" --ref main
done

# Local testing with mock data
eac get ci-dispatch --directly-changed "core" --invalidated "eac-cli docs" \
  --mock '{"eac-cli": true, "docs": false}'
```

## See Also

- [get changed-modules-ci](./changed-modules-ci.md)
- [pipeline ci](../pipeline/ci.md)
- [get Commands](../../categories/get.md)
