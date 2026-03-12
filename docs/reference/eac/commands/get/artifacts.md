# Get artifacts

<!-- book:cmd get artifacts -->

Returns resolved build artifacts for a module, including existence status, metadata overrides, and build mode breakdown.

## Usage

```bash
eac get artifacts <module> [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `module` | Module moniker (required) |

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--os` | string | Target OS (defaults to current) |
| `--arch` | string | Target architecture (defaults to current) |
| `--all-platforms` | bool | Show artifacts for all platforms |

## Output Structure

```yaml
module: clie
type: go
build_dir: out/build/clie
os: linux
arch: amd64
build_modes:
  default: [clie-linux-amd64]
  all: [clie-linux-amd64, clie-darwin-arm64, ...]
metadata: {}
artifacts:
  - resolved_name: clie-linux-amd64
    resolved_path: out/build/clie/clie-linux-amd64
    exists: true
    metadata_override: ""
summary:
  total: 5
  exists: 3
  missing: 2
  overrides: 0
```

## Examples

```bash
# Current platform artifacts
eac get artifacts clie

# Cross-platform artifacts
eac get artifacts clie --os linux --arch arm64

# All platform variants
eac get artifacts clie --all-platforms
```

## See Also

- [show artifacts](../show/artifacts.md) - Formatted table
- [build](../build/build.md)
