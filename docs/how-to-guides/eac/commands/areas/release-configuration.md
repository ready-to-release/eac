<!-- EDITOR
# Editor: how-to-guides/commands/areas/release-configuration.md

## Soul

Configuration reference for release management including CalVer/SemVer settings, Keep a Changelog format, AI changelog prompts, tag patterns, CI validation rules, and module-specific overrides.

## Sections

1. Configuration Files
2. Release Settings
   - Basic Configuration
3. CalVer Configuration
   - Format Options
   - Format Examples
   - Version Increment
4. SemVer Configuration
   - Module Override
   - SemVer Settings
   - SemVer Commands
5. Changelog Configuration
   - Keep a Changelog Format
   - Changelog Structure
   - Conventional Changelog
6. AI Configuration
   - Prompt Templates
   - Changelog Entry Prompt
7. Tag Configuration
   - Tag Pattern
   - Tag Signing
   - Tag Message
8. Validation Configuration
   - Release Validation
   - Pre-release Checks
9. CI Integration
   - Required Workflows
   - Branch Protection
10. Module-Specific Settings
    - Per-Module Override
    - Changelog Location
11. Environment Variables
12. Example Configurations
    - Minimal Configuration
    - Enterprise Configuration
    - Open Source Configuration
13. Troubleshooting
14. Related Documentation
-->

# Release Configuration

This guide covers configuration options for EAC's release management system, including CalVer settings, changelog format, and release validation rules.

## Configuration Files

| File                          | Purpose             |
| ----------------------------- | ------------------- |
| `.r2r/eac/release/config.yml` | Release settings    |
| `<module>/CHANGELOG.md`       | Module changelog    |
| `.r2r/eac/ai/release/`        | AI prompt templates |

## Release Settings

### Basic Configuration

`.r2r/eac/release/config.yml`:

```yaml
# Version strategy
versioning:
  # Default strategy: calver or semver
  strategy: calver

  # CalVer format
  calver:
    format: "YYYY.MM.MICRO"
    # YYYY - 4-digit year
    # YY - 2-digit year
    # MM - 2-digit month (01-12)
    # DD - 2-digit day (01-31)
    # MICRO - incremental number

  # SemVer settings (for specific modules)
  semver:
    # Modules using semver
    modules:
      - r2r-cli

# Changelog settings
changelog:
  # Filename
  filename: CHANGELOG.md

  # Format: keepachangelog, conventional, custom
  format: keepachangelog

  # Header template
  header: |
    # Changelog

    All notable changes to this project will be documented in this file.

    The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

# Tag settings
tags:
  # Tag format pattern
  pattern: "{moniker}/v{version}"

  # Sign tags with GPG
  sign: false

  # Push tags automatically
  auto_push: false

# CI integration
ci:
  # Check CI status before release
  require_green_ci: true

  # Required workflows
  required_workflows:
    - ci
    - test

  # Allow release from branches
  allowed_branches:
    - main
    - release/*
```

## CalVer Configuration

### Format Options

| Token   | Description           | Example               |
| ------- | --------------------- | --------------------- |
| `YYYY`  | 4-digit year          | 2024                  |
| `YY`    | 2-digit year          | 24                    |
| `MM`    | 2-digit month         | 12                    |
| `DD`    | 2-digit day           | 15                    |
| `MICRO` | Incremental number    | 1, 2, 3...            |
| `MINOR` | Monthly reset counter | 1 (resets each month) |

### Format Examples

```yaml
calver:
  # Default: 2024.12.1
  format: "YYYY.MM.MICRO"

  # Ubuntu style: 24.12
  format: "YY.MM"

  # Daily: 2024.12.15.1
  format: "YYYY.MM.DD.MICRO"

  # Short: 24.12.1
  format: "YY.MM.MICRO"
```

### Version Increment

```bash
# Current: 2024.12.1
# Same month release: 2024.12.2

# New month release: 2025.01.1
# (MICRO resets to 1)
```

## SemVer Configuration

### Module Override

```yaml
versioning:
  strategy: calver

  semver:
    modules:
      - r2r-cli  # Uses semver instead
```

### SemVer Settings

```yaml
semver:
  modules:
    - r2r-cli

  settings:
    r2r-cli:
      # Start version
      initial: "1.0.0"

      # Pre-release identifiers
      prerelease:
        - alpha
        - beta
        - rc

      # Build metadata
      include_build_metadata: false
```

### SemVer Commands

```bash
# Specific to r2r-cli
r2r eac release r2r-cli

# Creates: r2r-cli/v1.2.3
```

## Changelog Configuration

### Keep a Changelog Format

```yaml
changelog:
  format: keepachangelog

  # Change types
  types:
    - Added       # New features
    - Changed     # Changes in existing functionality
    - Deprecated  # Soon-to-be removed features
    - Removed     # Removed features
    - Fixed       # Bug fixes
    - Security    # Security fixes

  # Link format for versions
  links:
    compare: "https://github.com/{owner}/{repo}/compare/{prev}...{current}"
    release: "https://github.com/{owner}/{repo}/releases/tag/{tag}"
```

### Changelog Structure

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added
- New feature description

### Fixed
- Bug fix description

## [2024.12.1] - 2024-12-15

### Added
- Initial release feature

### Changed
- Modified behavior

[Unreleased]: https://github.com/owner/repo/compare/v2024.12.1...HEAD
[2024.12.1]: https://github.com/owner/repo/releases/tag/v2024.12.1
```

### Conventional Changelog

```yaml
changelog:
  format: conventional

  # Commit types to include
  types:
    feat: Added
    fix: Fixed
    perf: Changed
    refactor: Changed
    docs: Documentation
    style: Style
    test: Testing
    chore: Maintenance

  # Scope mapping
  scopes:
    api: "API"
    cli: "CLI"
    core: "Core"
```

## AI Configuration

### Prompt Templates

Location: `.r2r/eac/ai/release/`

```text
.r2r/eac/ai/release/
├── changelog-entry.md      # Changelog entry generation
├── release-summary.md      # Release summary
└── breaking-changes.md     # Breaking change detection
```

### Changelog Entry Prompt

```markdown
# Changelog Entry Generation

## Context
Generate changelog entries from git commits.

## Commits Since Last Release
{{range .Commits}}
- {{.Hash}}: {{.Message}}
  Files: {{.Files}}
{{end}}

## Previous Changelog
{{.PreviousChangelog}}

## Guidelines

### Change Types
- **Added**: New features, capabilities
- **Changed**: Modifications to existing features
- **Deprecated**: Features marked for removal
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Security improvements

### Writing Style
1. Start with verb (Add, Fix, Change, Remove)
2. Be specific but concise
3. Include ticket/issue references
4. Note breaking changes prominently

### Breaking Changes
If a commit includes BREAKING CHANGE, add to special section.

## Output
Changelog entries in Keep a Changelog format.
```

## Tag Configuration

### Tag Pattern

```yaml
tags:
  # Default: eac-commands/v2024.12.1
  pattern: "{moniker}/v{version}"

  # Alternative: v2024.12.1-eac-commands
  # pattern: "v{version}-{moniker}"

  # Simple: v2024.12.1
  # pattern: "v{version}"
```

### Tag Signing

```yaml
tags:
  sign: true

  # GPG key ID (optional, uses default if not specified)
  gpg_key: "ABCD1234"
```

### Tag Message

```yaml
tags:
  # Include message with tag
  message: true

  # Message template
  message_template: |
    Release {moniker} {version}

    Changes in this release:
    {changelog_excerpt}
```

## Validation Configuration

### Release Validation

```yaml
validation:
  # Require changelog entry
  require_changelog: true

  # Validate changelog format
  validate_format: true

  # Check for version in changelog
  check_version_exists: true

  # Require date in changelog
  require_date: true

  # Date format
  date_format: "2006-01-02"  # Go date format
```

### Pre-release Checks

```yaml
validation:
  pre_release:
    # Check CI status
    - ci_status

    # Validate contracts
    - contracts

    # Run tests
    - tests

    # Check for uncommitted changes
    - clean_working_tree

    # Verify on allowed branch
    - branch_check
```

## CI Integration

### Required Workflows

```yaml
ci:
  require_green_ci: true

  required_workflows:
    - ci          # Main CI workflow
    - test        # Test workflow
    - security    # Security scan

  # Workflow check timeout
  timeout: 30m

  # Retry on pending
  retry_pending: true
  retry_interval: 30s
```

### Branch Protection

```yaml
ci:
  allowed_branches:
    - main
    - release/*
    - hotfix/*

  # Require branch to be up-to-date
  require_up_to_date: true
```

## Module-Specific Settings

### Per-Module Override

```yaml
# modules.yml
modules:
  - moniker: r2r-cli
    type: go
    release:
      strategy: semver
      changelog: CHANGELOG.md
      require_ci: true

  - moniker: eac-commands
    type: go
    release:
      strategy: calver
      changelog: CHANGELOG.md
      auto_changelog: true
```

### Changelog Location

```yaml
# Default: <module-root>/CHANGELOG.md
modules:
  - moniker: eac-commands
    release:
      changelog: docs/CHANGELOG.md  # Custom location
```

## Environment Variables

| Variable            | Description         | Default  |
| ------------------- | ------------------- | -------- |
| `RELEASE_STRATEGY`  | Version strategy    | `calver` |
| `RELEASE_SIGN_TAGS` | Sign tags with GPG  | `false`  |
| `RELEASE_AUTO_PUSH` | Auto-push tags      | `false`  |
| `GPG_KEY_ID`        | GPG key for signing | -        |

## Example Configurations

### Minimal Configuration

```yaml
versioning:
  strategy: calver
  calver:
    format: "YYYY.MM.MICRO"

changelog:
  format: keepachangelog

tags:
  pattern: "{moniker}/v{version}"
```

### Enterprise Configuration

```yaml
versioning:
  strategy: calver
  calver:
    format: "YYYY.MM.MICRO"
  semver:
    modules:
      - r2r-cli

changelog:
  format: keepachangelog
  filename: CHANGELOG.md
  header: |
    # Changelog

    All notable changes documented here.

tags:
  pattern: "{moniker}/v{version}"
  sign: true
  auto_push: false

ci:
  require_green_ci: true
  required_workflows:
    - ci
    - test
    - security
    - compliance
  allowed_branches:
    - main
    - release/*

validation:
  require_changelog: true
  validate_format: true
  pre_release:
    - ci_status
    - contracts
    - tests
    - clean_working_tree
    - branch_check
```

### Open Source Configuration

```yaml
versioning:
  strategy: calver
  calver:
    format: "YYYY.MM.MICRO"

changelog:
  format: keepachangelog
  links:
    compare: "https://github.com/org/repo/compare/{prev}...{current}"
    release: "https://github.com/org/repo/releases/tag/{tag}"

tags:
  pattern: "{moniker}/v{version}"
  sign: false
  message: true
  message_template: |
    Release {version}

    See CHANGELOG.md for details.

ci:
  require_green_ci: true
  required_workflows:
    - ci
  allowed_branches:
    - main
```

## Troubleshooting

| Issue                      | Cause                         | Solution                |
| -------------------------- | ----------------------------- | ----------------------- |
| Version not incrementing   | Same month, MICRO not updated | Check existing tags     |
| Changelog validation fails | Format error                  | Run `validate release`  |
| CI check fails             | Workflows not complete        | Wait or check status    |
| Tag exists                 | Version already released      | Increment version       |
| GPG signing fails          | Key not available             | Check GPG configuration |

## Related Documentation

- [Release Overview](release-overview.md) - Concepts and workflows
- [Release Commands](release-commands.md) - Command reference
- [Pipeline Configuration](pipeline-configuration.md) - CI/CD settings
