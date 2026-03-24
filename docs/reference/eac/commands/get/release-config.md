# get release-config

<!-- book:cmd get release-config -->

## Output Variables

| Variable                | Description                                                 |
| ----------------------- | ----------------------------------------------------------- |
| `RELEASE_TYPE`          | `cli-binary`, `container`, `docs-site`, `bundle`, or `none` |
| `VERSION_TYPE`          | `semver`, `calver`, or `none`                               |
| `HAS_EVIDENCE`          | Whether the module has evidence-book components             |
| `AWAIT_MODULE_RELEASES` | `true` for bundle releases (awaits dependency releases)     |

## Release Type Resolution

| Condition                                   | Release Type |
| ------------------------------------------- | ------------ |
| `bundle` release type                       | `bundle`     |
| `published` + dockerfile with push          | `container`  |
| `published` + docs-site component           | `docs-site`  |
| `published` + go component with binary_name | `cli-binary` |
| `internal` or `none`                        | `none`       |

## See Also

- [get release-status](./release-status.md)
- [get release-notes](./release-notes.md)
- [get Commands](../get/index.md)
