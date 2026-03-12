# Get build-deps

<!-- book:cmd get build-deps -->

Returns aggregated build dependencies for a module, collected from the module and all its transitive dependencies. Dependencies include system tools like `docker` and optional tools like `upx` (when artifact compression is configured).

## Usage

```bash
eac get build-deps <module> [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `module` | Module moniker (required) |

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--format` | string | Output format: `shell`, `space`, `yaml`, `json` |

## Output Structure

Default output (YAML):

```yaml
module: clie
type: go
build_deps:
  - docker
  - upx
```

With `--format shell`:

```bash
MODULE_PACKAGES="go"
BUILD_DEPS="docker,upx"
```

## Examples

```bash
# YAML output
eac get build-deps clie

# Shell-friendly output for CI scripts
eac get build-deps clie --format shell
```

## See Also

- [build](../build/build.md)
