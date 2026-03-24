# get artifacts

<!-- book:cmd get artifacts -->

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

## See Also

- [show artifacts](../show/artifacts.md) - Formatted table
- [build](../build/build.md)
