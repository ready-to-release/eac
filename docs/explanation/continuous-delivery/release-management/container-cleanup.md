# Container Registry Cleanup

This guide explains how to safely clean up old container images while protecting released packages.

## Overview

CI/CD pipelines generate many container images over time:

- Every commit creates a `sha-*` tagged image
- Every PR creates a `pr-*` tagged image
- Every release creates a version-tagged image

Without cleanup, registries accumulate thousands of images, increasing storage costs and making it harder to find relevant images.

## Understanding Tags

Container registries use **two distinct tagging systems** that are often confused:

### Image Tags (OCI/Docker)

These are the tags visible on the container image itself:

```text
ghcr.io/org/package:sha-abc1234    # CI build
ghcr.io/org/package:v1.0.0         # Release tag
ghcr.io/org/package:latest         # Mutable pointer
```

Image tags are:

- Assigned when pushing to the registry
- Visible in `docker images` and registry UI
- What you specify in `docker pull`

### GitHub Release Tags

These are tags created via the GitHub Releases API:

```text
ext-eac/1.0.0      # Module release
r2r-cli/2.3.4      # CLI release
r2r-eac-bundle/2025.01.15  # Bundle release
```

Release tags are:

- Created when publishing a GitHub Release
- Visible in the Releases section of a repository
- Correlate with git tags

### The Connection

When you create a release, the CI pipeline typically:

1. Creates a GitHub Release with tag `module/version`
2. Pushes a container image with the **same tag** as an image tag

This creates a **correlation** between the GitHub Release and the container image. The cleanup system uses this correlation to protect released images.

## Safety Model

The cleanup system uses multiple layers of protection:

```text
┌─────────────────────────────────────────────────────────────┐
│                    PROTECTION LAYERS                        │
├─────────────────────────────────────────────────────────────┤
│  1. Image Tag Patterns (preserve)                           │
│     └─ "v*", "latest", "[0-9]*.[0-9]*.[0-9]*"               │
│                                                             │
│  2. GitHub Release API Correlation (ALWAYS ON)              │
│     └─ ext-eac/1.0.0 → protected                            │
│                                                             │
│  3. Release Bundle References (ALWAYS ON)                   │
│     └─ Versions in bundle release notes → protected         │
│                                                             │
│  4. Digest Matching                                         │
│     └─ Same digest as protected version → protected         │
│                                                             │
│  5. Minimum Age                                             │
│     └─ Created < min_age_days ago → protected               │
│                                                             │
│  6. Prune Patterns (must match to be candidate)             │
│     └─ "sha-*", "dev-*", "pr-*", "ci"                       │
└─────────────────────────────────────────────────────────────┘
```

### Non-Configurable Safety

Some protections cannot be disabled:

| Protection                 | Why Non-Configurable                         |
| -------------------------- | -------------------------------------------- |
| GitHub Release correlation | Released packages must never be auto-deleted |
| Release bundle references  | Bundle contents must remain available        |

To delete a released package, use manual methods (GitHub UI or API).

## Release Bundles

Release bundles aggregate multiple module releases:

```yaml
# In repository.yml
- moniker: r2r-eac-bundle
  release_bundle:
    headline:
      r2r: r2r-cli
      eac: ext-eac
    categories:
      - name: Core Tools
        modules: [r2r-cli, ext-eac]
```

When a bundle is released, its release notes list the included module versions:

```markdown
## Core Tools
- r2r-cli: v2.3.4
- ext-eac: v1.0.0
```

The cleanup system parses these notes and protects all referenced versions, even if they're older than the keep limit.

## Configuration

### Image Tags Section

Controls which container image tags to preserve or prune:

```yaml
image_tags:
  preserve:           # NEVER delete these
    - "v*"            # Release tags like v1.0.0
    - "latest"        # Latest pointer
    - "[0-9]*.[0-9]*.[0-9]*"  # Bare semver
  prune:              # Candidates for cleanup
    - "sha-*"         # CI builds
    - "dev-*"         # Dev builds
    - "pr-*"          # PR builds
    - "ci"            # CI tag
```

**Logic:**

- If ANY tag matches a preserve pattern → protected
- If NO tag matches a prune pattern → protected
- Only versions matching prune patterns are cleanup candidates

### GitHub Releases Section

Configures GitHub Release API correlation:

```yaml
github_releases:
  # Released packages are ALWAYS protected (non-configurable)
  tag_format: "{module}/{version}"  # How releases are tagged
```

The `tag_format` helps the system understand which image tags correspond to GitHub Releases.

### General Settings

```yaml
cleanup:
  enabled: true       # Master switch
  keep: 10            # Keep newest N prunable versions
  min_age_days: 7     # Don't prune versions younger than this
```

## Cleanup Strategy

The default "keep-latest-n" strategy:

1. Identify all prunable versions (match prune patterns, not protected)
2. Sort by creation date (newest first)
3. Keep the newest N versions
4. Delete the rest

```text
Versions (sorted newest first):
  sha-abc (2 days ago)  → KEEP (within keep limit)
  sha-def (5 days ago)  → KEEP (within keep limit)
  sha-ghi (8 days ago)  → KEEP (within keep limit)
  ...
  sha-xyz (30 days ago) → DELETE (outside keep limit)
```

## Manual Cleanup

For released packages that must be deleted (e.g., security issues):

### Using GitHub UI

1. Go to repository → Packages
2. Find the package
3. Click on the version
4. Delete (requires confirmation)

### Using gh CLI

```bash
# List versions
gh api /orgs/{org}/packages/container/{package}/versions

# Delete specific version
gh api -X DELETE /orgs/{org}/packages/container/{package}/versions/{version_id}
```

### Using Docker Registry API

```bash
# Get token
TOKEN=$(gh auth token)

# Delete manifest
curl -X DELETE \
  -H "Authorization: Bearer $TOKEN" \
  "https://ghcr.io/v2/{org}/{package}/manifests/{digest}"
```

## Monitoring

Track cleanup effectiveness:

| Metric          | Target            | Action if Exceeded         |
| --------------- | ----------------- | -------------------------- |
| Total versions  | < 100 per package | Increase cleanup frequency |
| Storage usage   | < quota           | Reduce keep count          |
| Protected ratio | < 50%             | Review preserve patterns   |

## See Also

- [release prune-packages](../../../reference/eac/commands/release/prune-packages.md) - Command reference
- [Release Evidence](./release-evidence.md) - What evidence is kept for releases
- [Changelog System](./changelog-system.md) - How releases are created
