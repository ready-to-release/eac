# get changed-modules-ci

<!-- book:cmd get changed-modules-ci -->

## Detection Logic

For each module:

1. Query the module's last successful CI run via GitHub Actions API
2. If CI passed at current HEAD SHA -- skip (no rebuild needed)
3. If CI passed at a different SHA -- check if the module's files changed since then
4. If no CI history -- module needs CI

Transitive dependents of changed modules are also invalidated iteratively.

## Output Structure

```yaml
modules: [core, clie]           # All modules needing CI
directly_changed: [core]        # Files changed in these modules
invalidated: [clie]             # Transitive dependents
base_sha: per-module
head_sha: abc1234...
is_bootstrap: false
changed_file_count: 5
files_by_module:
  core: [go/core/config/config.go]
module_status:
  core:
    has_valid_ci: false
    last_success_sha: def5678...
    reason: files_changed_since_def5678
    files_changed: 3
skipped: [docs]
```

With `--format shell`:

```bash
MODULES="core clie"
DIRECTLY_CHANGED="core"
INVALIDATED="clie"
BASE_SHA="per-module"
IS_BOOTSTRAP="false"
CHANGED_FILE_COUNT="5"
```

## See Also

- [get changed-modules](./changed-modules.md) - Local changes
