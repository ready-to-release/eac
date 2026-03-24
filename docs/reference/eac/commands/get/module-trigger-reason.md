# get module-trigger-reason

<!-- book:cmd get module-trigger-reason -->

## Output

Plain text reason to stdout. Possible values:

| Output            | Meaning                                    |
| ----------------- | ------------------------------------------ |
| `files changed`   | Module files changed since last CI success |
| `{dep} changed`   | Dependency module triggered CI             |
| `no previous CI`  | No CI history for this module              |
| `CI query failed` | GitHub API query failed                    |
| `unknown`         | Module not found in status data            |

## See Also

- [get](get.md)
- [get changed-modules-ci](changed-modules-ci.md)
