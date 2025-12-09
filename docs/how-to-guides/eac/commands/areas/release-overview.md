<!-- EDITOR
# Editor: how-to-guides/commands/areas/release-overview.md

## Soul

CalVer-based release management system with AI-generated changelogs, automated validation, CI integration, and git tagging for traceable, compliant releases.

## Sections

1. What is Release Management?
2. When to Use Release Commands
3. Common Use Cases
4. Key Concepts
   - Calendar Versioning (CalVer)
   - Why CalVer?
   - Changelog Format
   - Release Tags
5. Workflow Overview
   - Standard Release Flow
   - Automated CI Release
   - Hotfix Release
6. Version Lifecycle
   - States
7. Integration Points
   - With CI/CD
   - With GitHub Releases
   - With Module Contracts
   - With Workspace
8. Special Cases
   - r2r-cli Releases
   - Multi-Module Releases
   - Version Extraction
9. Best Practices
10. Next Steps
11. Related Areas
-->

# Release Management

Release management in EAC provides CalVer-based versioning, automated changelog generation, and CI-integrated release workflows for consistent, traceable releases.

## What is Release Management?

EAC's release system enables you to:

- **Version modules** using Calendar Versioning (CalVer)
- **Generate changelogs** from git history with AI
- **Validate release readiness** including CI status
- **Create git tags** for released versions
- **Track pending releases** across modules

The system automates the release process while maintaining traceability and compliance requirements.

## When to Use Release Commands

Use release commands when you need:

| Scenario                      | Commands                         |
| ----------------------------- | -------------------------------- |
| Check if module needs release | `release pending`                |
| Generate/update changelog     | `release changelog`              |
| Validate changelog format     | `validate release`               |
| Check CI before release       | `release check-ci`               |
| Create release tag            | `release generate-module-calver`, `release this` |
| Find untagged versions        | `release tag-pending`            |

### Common Use Cases

- **Continuous delivery** - Automate releases on merge to main
- **Sprint releases** - Bundle changes into versioned releases
- **Hotfix releases** - Quick patches with proper versioning
- **Audit compliance** - Traceable release history

## Key Concepts

### Calendar Versioning (CalVer)

EAC uses CalVer format: `YYYY.MM.MICRO`

| Component | Description                | Example    |
| --------- | -------------------------- | ---------- |
| `YYYY`    | Four-digit year            | 2024       |
| `MM`      | Two-digit month            | 12         |
| `MICRO`   | Incremental release number | 1, 2, 3... |

Examples:

- `2024.12.1` - First release in December 2024
- `2024.12.2` - Second release in December 2024
- `2025.01.1` - First release in January 2025

### Why CalVer?

| Benefit        | Description                                  |
| -------------- | -------------------------------------------- |
| **Time-based** | Immediately know when a version was released |
| **Simple**     | No semantic versioning complexity            |
| **Consistent** | Same format across all modules               |
| **Monotonic**  | Versions always increase                     |

### Changelog Format

Changelogs follow [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
# Changelog

## [2024.12.2] - 2024-12-15

### Added
- New authentication module with JWT support
- API rate limiting middleware

### Changed
- Improved error messages for validation failures

### Fixed
- Race condition in cache invalidation

## [2024.12.1] - 2024-12-01

### Added
- Initial release with core functionality
```

### Release Tags

Git tags follow the pattern: `<moniker>/v<version>`

```text
eac-commands/v2024.12.1
eac-core/v2024.12.2
r2r-cli/v2024.12.3
```

## Workflow Overview

### Standard Release Flow

```bash
# 1. Check if module has pending changes
r2r eac release pending eac-commands

# 2. Generate/update changelog
r2r eac release changelog eac-commands

# 3. Review changelog
cat go/eac/commands/CHANGELOG.md

# 4. Validate changelog format
r2r eac validate release eac-commands

# 5. Check CI status
r2r eac release check-ci eac-commands

# 6. Create release
r2r eac release this eac-commands

# 7. Push tag
git push origin eac-commands/v2024.12.1
```

### Automated CI Release

```yaml
# On merge to main
- name: Check pending releases
  run: |
    PENDING=$(r2r eac release pending --all)
    if [ -n "$PENDING" ]; then
      echo "modules=$PENDING" >> $GITHUB_OUTPUT
    fi

- name: Release pending modules
  run: |
    for module in ${{ steps.check.outputs.modules }}; do
      r2r eac release changelog $module
      r2r eac release this $module
    done

- name: Push tags
  run: git push --tags
```

### Hotfix Release

```bash
# 1. Create workspace for hotfix
r2r eac work create hotfix/critical-fix

# 2. Make fix and commit
r2r eac work commit --all

# 3. Merge to main
r2r eac work merge
cd ../eac && git push

# 4. Release immediately
r2r eac release changelog eac-core
r2r eac release this eac-core
git push origin eac-core/v2024.12.3
```

## Version Lifecycle

```text
Development ──▶ Changelog ──▶ Validate ──▶ CI Check ──▶ Tag ──▶ Push
     │              │            │            │          │        │
     ▼              ▼            ▼            ▼          ▼        ▼
  Commits      AI-generated   Format OK   Green CI   Git tag   Released
```

### States

| State           | Description                 | Next Action         |
| --------------- | --------------------------- | ------------------- |
| **Development** | Active changes on main      | Continue or release |
| **Pending**     | Unreleased commits exist    | Generate changelog  |
| **Ready**       | Changelog updated, CI green | Create tag          |
| **Released**    | Tag created                 | Push to remote      |
| **Published**   | Tag pushed                  | Done                |

## Integration Points

### With CI/CD

Release validation in pipelines:

```yaml
- name: Validate changelogs
  run: r2r eac validate release

- name: Check CI status
  run: r2r eac release check-ci $MODULE

- name: Release on green
  if: success()
  run: r2r eac release this $MODULE
```

### With GitHub Releases

Create GitHub releases from tags:

```bash
# After tagging
gh release create eac-commands/v2024.12.1 \
  --title "eac-commands v2024.12.1" \
  --notes-file go/eac/commands/CHANGELOG.md
```

### With Module Contracts

Release commands read module configuration:

```yaml
# modules.yml
modules:
  - moniker: eac-commands
    type: go
    version_strategy: calver
    changelog: CHANGELOG.md
```

### With Workspace

Release after feature merge:

```bash
# In workspace
r2r eac work merge
cd ../eac

# Back in main
r2r eac release pending eac-commands  # Check if release needed
r2r eac release this eac-commands     # Release if pending
```

## Special Cases

### r2r-cli Releases

The CLI uses SemVer for compatibility:

```bash
r2r eac release r2r-cli
# Creates tag: r2r-cli/v1.2.3 (semver)
```

### Multi-Module Releases

Release multiple modules together:

```bash
# Check all pending
r2r eac release pending --all

# Release specific modules
r2r eac release this eac-commands eac-core

# Or release all pending
r2r eac release this --all
```

### Version Extraction

Get current version programmatically:

```bash
r2r eac release get-version eac-commands
# Output: 2024.12.1
```

## Best Practices

### Do's

- **Release frequently** - Small, focused releases are easier to debug
- **Update changelogs** - AI helps, but review before release
- **Check CI** - Never release with failing tests
- **Use meaningful entries** - Changelog is for users, not git log

### Don'ts

- **Don't skip validation** - Catches format issues early
- **Don't release without CI** - `release check-ci` exists for a reason
- **Don't forget to push** - Tags must be pushed to be useful

## Next Steps

- [Release Configuration](release-configuration.md) - Configure version patterns and changelog format
- [Release Commands](release-commands.md) - Full command reference

## Related Areas

- [Pipeline](pipeline-overview.md) - CI/CD integration for releases
- [Workspace](workspace-overview.md) - Development workflow before release
- [Validate Commands](validate-overview.md) - Contract validation before release
