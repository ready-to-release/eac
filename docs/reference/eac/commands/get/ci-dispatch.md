# get ci-dispatch

<!-- book:cmd get ci-dispatch -->

## How It Works

1. Directly changed modules are always added to the dispatch list
2. Invalidated modules are checked against GitHub for successful CI at the HEAD SHA
3. Modules with valid CI are skipped; others are dispatched
4. The dispatch list is topologically sorted (modules with no CI deps first)

## Output Fields (structured)

| Field             | Description                              |
| ----------------- | ---------------------------------------- |
| `dispatch`        | Modules to dispatch, in dependency order |
| `skipped`         | Modules with valid CI at HEAD            |
| `reasons`         | Per-module reasoning for dispatch/skip   |
| `ci_dependencies` | Dependency graph within the dispatch set |
| `head_sha`        | The SHA used for checking                |
| `total_modules`   | Total modules considered                 |

## Shell Format Output

```text
DISPATCH="mod1 mod2 mod3"
SKIPPED="mod4"
DISPATCH_COUNT=3
SKIPPED_COUNT=1
CI_DEPS_JSON='{"mod3":["mod1"]}'
```

## See Also

- [get changed-modules-ci](./changed-modules-ci.md)
- [pipeline ci](../pipeline/ci.md)
- [get Commands](../get/index.md)
