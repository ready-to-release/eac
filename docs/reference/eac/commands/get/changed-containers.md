# get changed-containers

<!-- book:cmd get changed-containers -->

## How It Works

Determines which container components in a multi-container module need rebuilding:

- **Registry Check**: Queries the container registry (GHCR) for the last-build SHA tag
- **Per-Component Detection**: Checks if any of each component's files changed since the last build
- **Shared File Awareness**: Detects changes to module-level files not owned by any specific component
- **Fail-Open Design**: Any error (registry query, git diff) triggers a rebuild for safety

## Output Formats

- `--format github-output`: KEY=value lines for `$GITHUB_OUTPUT`
- `--format shell`: Shell variable assignments
- Default: YAML/JSON/TOML via standard get command

## See Also

- [get changed-modules](./changed-modules.md) - Module-level change detection
- [get changed-modules-ci](./changed-modules-ci.md) - For CI pipelines
- [build](../build/build.md)
