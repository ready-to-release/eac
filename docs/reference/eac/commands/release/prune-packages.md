# release prune-packages

<!-- book:cmd release prune-packages -->

## Synopsis

```bash
eac release prune-packages <package> [flags]
eac release prune-packages --all [flags]
```

## Description

The `prune-packages` command cleans up old container image versions from GHCR. It implements multiple safety layers to prevent accidental deletion of released or important packages.

**Default mode is dry-run** - use `--force` to actually delete versions.

### Safety Checks (in order)

1. **Preserve patterns** - Image tags matching patterns like `v*`, `latest` are never deleted
2. **GitHub Release tags** - Versions with tags matching GitHub Releases are always protected
3. **Release bundle references** - Versions referenced in release bundle notes are protected
4. **Digest matching** - Versions sharing a digest with a protected version are protected
5. **Minimum age** - Versions created less than `min_age_days` ago are protected
6. **Prune patterns** - Only versions matching prune patterns (e.g., `sha-*`) are candidates

### Non-configurable Safety

**Released packages are ALWAYS protected.** This is a non-configurable safety feature. If a container image tag matches any GitHub Release tag, or is referenced by any release bundle, it will never be auto-deleted.

To manually delete a released package, use the GitHub UI or `gh api` directly.

## Configuration

Configure in `.eac/repository.yml`:

```yaml
registries:
  ghcr.io:
    org: your-org
    cleanup:
      enabled: true
      keep: 10                    # Number of prunable versions to keep
      min_age_days: 7             # Minimum age before pruning
      image_tags:
        preserve:                 # Image tags (OCI/Docker) to never delete
          - "v*"
          - "latest"
          - "[0-9]*.[0-9]*.[0-9]*"
        prune:                    # Image tags eligible for cleanup
          - "sha-*"
          - "dev-*"
          - "pr-*"
          - "ci"
      github_releases:
        # Released packages are ALWAYS protected (non-configurable)
        tag_format: "{module}/{version}"  # e.g., eac-ext/1.0.0
```

## Understanding Tags

This command works with two different types of tags:

| Type             | Source                          | Examples                          | Purpose                            |
| ---------------- | ------------------------------- | --------------------------------- | ---------------------------------- |
| **Image Tags**   | Container registry (OCI/Docker) | `sha-abc1234`, `v1.0.0`, `latest` | Identify specific container images |
| **Release Tags** | GitHub Releases API             | `eac-ext/1.0.0`, `clie/2.0.0`     | Mark official releases             |

The `image_tags` configuration controls which container image tags to preserve or prune.
The `github_releases` configuration correlates container images with GitHub Releases.

See [Container Registry Cleanup](../../../../explanation/continuous-delivery/release-management/container-cleanup.md) for detailed concepts.

## See Also

- [release cleanup](cleanup.md) - Clean up orphaned tags after failed releases
- [release prune](prune.md) - Prune old pre-releases
- [Container Registry Cleanup](../../../../explanation/continuous-delivery/release-management/container-cleanup.md) - Conceptual guide
