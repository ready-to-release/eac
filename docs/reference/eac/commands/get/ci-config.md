# get ci-config

<!-- book:cmd get ci-config -->

## Output Variables

| Variable                | Description                                                                  |
| ----------------------- | ---------------------------------------------------------------------------- |
| `IS_CONTAINER`          | Has dockerfile component with docker_build                                   |
| `IS_MULTI_CONTAINER`    | Module has multiple pushable container components                            |
| `CONTAINER_COMPONENTS`  | JSON array of container component names                                      |
| `CONTAINER_PUSH`        | Whether to push container images                                             |
| `HAS_TESTS`             | Has components with testers defined                                          |
| `TEST_ON_WINDOWS`       | Artifact matrix includes Windows targets                                     |
| `TEST_ON_MACOS`         | Artifact matrix includes macOS targets                                       |
| `SCANS`                 | Comma-separated scanners from component kinds                                |
| `SCAN_FAIL_MODE`        | Scan failure mode (default: `warn`)                                          |
| `BUILD_EVIDENCE`        | Has evidence-book components                                                 |
| `CROSS_COMPILE_WINDOWS` | Artifact matrix is cross-platform                                            |
| `BUILD_ARGS`            | CI build arguments (always `--all`)                                          |
| `DOWNLOAD_MODULES`      | Space-separated CI dependency modules to download artifacts from             |
| `CONTAINER_TEST_SCRIPT` | Path to container test script (convention: `containers/<module>/ci-test.sh`) |
| `TEST_SUITES`           | PR test suites (default: `unit,integration`)                                 |
| `TEST_SUITES_FULL`      | Full CI test suites (default: `unit,integration,acceptance`)                 |

## See Also

- [get ci-dispatch](./ci-dispatch.md)
- [get ci-workflows](./ci-workflows.md)
- [get Commands](../get/index.md)
