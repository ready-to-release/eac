# Get ci-config

<!-- book:cmd get ci-config -->

Derives CI configuration for a module from the core config system. All values are computed from `repository.yml` and `blueprints.yml`, eliminating duplicate configuration in workflow files.

## Usage

```bash
eac get ci-config --module <moniker> [--format shell|github-output]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--module` | string | Module moniker (required) |
| `--format` | string | Output format: `shell`, `github-output`, or standard get formats (yaml/json/toml) |

## Output Variables

| Variable | Description |
|----------|-------------|
| `IS_CONTAINER` | Has dockerfile component with docker_build |
| `IS_MULTI_CONTAINER` | Module has multiple pushable container components |
| `CONTAINER_COMPONENTS` | JSON array of container component names |
| `CONTAINER_PUSH` | Whether to push container images |
| `HAS_TESTS` | Has components with testers defined |
| `TEST_ON_WINDOWS` | Artifact matrix includes Windows targets |
| `TEST_ON_MACOS` | Artifact matrix includes macOS targets |
| `SCANS` | Comma-separated scanners from component kinds |
| `SCAN_FAIL_MODE` | Scan failure mode (default: `warn`) |
| `BUILD_EVIDENCE` | Has evidence-book components |
| `CROSS_COMPILE_WINDOWS` | Artifact matrix is cross-platform |
| `BUILD_ARGS` | CI build arguments (always `--all`) |
| `DOWNLOAD_MODULES` | Space-separated CI dependency modules to download artifacts from |
| `CONTAINER_TEST_SCRIPT` | Path to container test script (convention: `containers/<module>/ci-test.sh`) |
| `TEST_SUITES` | PR test suites (default: `unit,integration`) |
| `TEST_SUITES_FULL` | Full CI test suites (default: `unit,integration,acceptance`) |

## Examples

```bash
# Get CI config as YAML
eac get ci-config --module eac-cli

# Shell format for eval in CI scripts
eval $(eac get ci-config --module core --format shell)
echo "Has tests: $HAS_TESTS"

# GitHub Actions output format
eac get ci-config --module docs --format github-output >> $GITHUB_OUTPUT
```

## See Also

- [get ci-dispatch](./ci-dispatch.md)
- [get ci-workflows](./ci-workflows.md)
- [get Commands](../categories/get.md)
