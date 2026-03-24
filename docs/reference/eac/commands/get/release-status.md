# get release-status

<!-- book:cmd get release-status -->

## Output Structure

**Default (JSON/YAML):**

```yaml
modules:
  clie:
    tag: clie/1.0.0
    version: "1.0.0"
    released: true
  docs:
    tag: ""
    version: ""
    released: false
released: [clie]
missing: [docs]
```

**`--format shell`:**

```bash
RELEASED="clie eac-ext"
MISSING="docs"
ALL_RELEASED="false"
```

## See Also

- [get](get.md)
- [release](../release/index.md)
