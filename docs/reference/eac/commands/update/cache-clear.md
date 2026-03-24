# update cache-clear

<!-- book:cmd update cache-clear -->

## Cache Types

| Type          | Description                                                         |
| ------------- | ------------------------------------------------------------------- |
| `state`       | Incremental state (build/lint/test state.json, capacity semaphores) |
| `asset`       | Rendered assets (mermaid, drawio, structurizr caches)               |
| `work`        | Ephemeral work directories (npm work dirs)                          |
| `registry`    | Docker image cache (runs `docker image prune`)                      |
| `layer`       | Docker builder cache (runs `docker builder prune`)                  |
| `all`         | Everything                                                          |
| `local:state` | Fine-grained: local state only                                      |

## See Also

- [build](../build/build.md)
- [test](../test/test.md)
- [lint](../lint.md)
- [update Commands](../update/index.md)
