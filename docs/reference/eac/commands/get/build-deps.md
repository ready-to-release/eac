# get build-deps

<!-- book:cmd get build-deps -->

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

## See Also

- [build](../build/build.md)
