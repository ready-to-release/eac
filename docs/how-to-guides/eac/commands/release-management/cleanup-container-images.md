# Clean Up Container Images

Remove old container image versions from GHCR while protecting released packages.

## Prerequisites

- GHCR configured in `repository.yml` with cleanup policy
- `gh` CLI authenticated with package delete permissions

## Quick Start

```bash
# Dry-run (see what would be deleted)
r2r release prune-packages ext-eac

# Actually delete
r2r release prune-packages ext-eac --force
```

---

## Step 1: Configure Cleanup Policy

Add registry configuration to `.r2r/eac/repository.yml`:

```yaml
registries:
  ghcr.io:
    org: your-org
    cleanup:
      enabled: true
      keep: 10                    # Keep newest 10 prunable versions
      min_age_days: 7             # Don't delete versions < 7 days old
      image_tags:
        preserve:                 # Never delete these
          - "v*"
          - "latest"
          - "[0-9]*.[0-9]*.[0-9]*"
        prune:                    # Candidates for cleanup
          - "sha-*"
          - "dev-*"
          - "pr-*"
      github_releases:
        tag_format: "{module}/{version}"
```

---

## Step 2: Preview Changes (Dry-Run)

Always run a dry-run first:

```bash
r2r release prune-packages ext-eac
```

**Expected output:**

```text
Processing package: ext-eac
  Total versions: 45
  Protected: 12
  Prunable: 33 (keeping 10)
  Versions to delete: 23
  Would delete: sha256:abc12345 [sha-abc1234]
  Would delete: sha256:def67890 [sha-def6789]
  ...
DRY RUN complete. Would delete 23 versions.
Use --force to actually delete.
```

---

## Step 3: Review Protected Versions

See why versions are protected:

```bash
r2r release prune-packages ext-eac --verbose
```

**Protection reasons:**

| Reason                               | Meaning                        |
| ------------------------------------ | ------------------------------ |
| `tag matches preserve pattern`       | Tag like `v1.0.0` matches `v*` |
| `associated with GitHub release`     | Tag matches a GitHub Release   |
| `referenced by release bundle`       | Included in a bundle release   |
| `digest matches released version`    | Same content as a release      |
| `created less than min_age_days ago` | Too recent to delete           |
| `no tags match prune patterns`       | Not a CI build image           |

---

## Step 4: Execute Cleanup

When satisfied with dry-run results:

```bash
r2r release prune-packages ext-eac --force
```

---

## Common Scenarios

### Clean All Packages

```bash
# Dry-run all packages
r2r release prune-packages --all

# Delete from all packages
r2r release prune-packages --all --force
```

### Override Keep Count

Keep fewer versions for this run:

```bash
r2r release prune-packages ext-eac --keep 5 --force
```

### Get JSON Output

For scripting or analysis:

```bash
r2r release prune-packages ext-eac --json
```

---

## Understanding Image Tags vs Release Tags

| Type             | Source              | Examples                          |
| ---------------- | ------------------- | --------------------------------- |
| **Image Tags**   | Container registry  | `sha-abc1234`, `v1.0.0`, `latest` |
| **Release Tags** | GitHub Releases API | `ext-eac/1.0.0`, `r2r-cli/2.0.0`  |

The cleanup command correlates these to protect released images.

---

## Safety Features

**Released packages are ALWAYS protected.** This cannot be disabled.

To delete a released package manually:

```bash
# List versions
gh api /orgs/{org}/packages/container/{package}/versions

# Delete specific version (use with caution!)
gh api -X DELETE /orgs/{org}/packages/container/{package}/versions/{id}
```

---

## Troubleshooting

### "No registry configuration found"

Add `registries` section to `repository.yml`. See Step 1.

### "Cleanup is disabled"

Set `cleanup.enabled: true` in the registry config.

### Nothing Being Deleted

Check your prune patterns. Only images with tags matching prune patterns (like `sha-*`) are candidates.

---

## See Also

- [Container Registry Cleanup](../../../../explanation/continuous-delivery/release-management/container-cleanup.md) - Conceptual guide
- [release prune-packages](../../../../reference/eac/commands/release/prune-packages.md) - Command reference
