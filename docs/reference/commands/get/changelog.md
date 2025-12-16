# get changelog

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get changelog <module> [version]`
**Purpose**: Get changelog data in structured format (YAML/JSON/TOML)
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get changelog <module>
r2r eac get changelog <module> <version>
r2r eac get changelog <module> latest
r2r eac get changelog <module> unreleased
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `module` | Yes | Module moniker (e.g., `ext-eac`, `eac-commands`) |
| `version` | No | Version number, or special keyword (`latest`, `unreleased`) |

## Output Formats

| Flag | Format | Use Case |
|------|--------|----------|
| *(none)* | YAML | Default, human-readable structured data |
| `--as-json` | JSON | Machine parsing, API integration |
| `--as-toml` | TOML | Configuration files |
| `--as-yaml` | YAML | Explicit YAML output |

## Special Keywords

| Keyword | Description | Output |
|---------|-------------|--------|
| `latest` | Most recent released version | Single version object |
| `unreleased` | Pending unreleased changes | Single version object |
| *(omit)* | All versions | Full changelog object |

## Data Structure

### Full Changelog (no version specified)

```yaml
module: "ext-eac"
title: "Changelog"
description: "All notable changes..."
versiontype: 0  # 0=semver, 1=calver
unreleased:
  number: "Unreleased"
  date: "0001-01-01T00:00:00Z"
  added: []
  changed: []
  fixed: []
  # ...
versions:
  - number: "0.0.7"
    date: "2025-12-11T00:00:00Z"
    added: [...]
    changed: [...]
    fixed: [...]
    # ...
```

### Single Version (with version/latest/unreleased)

```yaml
number: "0.0.7"
date: "2025-12-11T00:00:00Z"
added:
  - description: "new feature"
    committype: "feat"
    scope: "module"
    commitsha: "abc123"
    breaking: false
changed: []
fixed:
  - description: "bug fix"
    committype: "fix"
    scope: "core"
    commitsha: "def456"
    breaking: false
```

## Examples

```bash
# Get full changelog as YAML
r2r eac get changelog ext-eac

# Get as JSON
r2r eac get changelog ext-eac --as-json

# Get latest version only
r2r eac get changelog ext-eac latest --as-json

# Get unreleased changes
r2r eac get changelog ext-eac unreleased --as-json

# Get specific version
r2r eac get changelog ext-eac 0.0.7 --as-json

# Extract version numbers
r2r eac get changelog ext-eac --as-json | jq -r '.versions[].number'

# Get latest version number
r2r eac get changelog ext-eac latest --as-json | jq -r '.number'

# Count unreleased changes
r2r eac get changelog ext-eac unreleased --as-json | jq '[.added, .changed, .fixed] | flatten | length'
```

## File Location

The command reads from: `release/<module>/CHANGELOG.md`

**Example:** For module `ext-eac`, reads from `release/ext-eac/CHANGELOG.md`

## Error Handling

| Error | Exit Code | Solution |
|-------|-----------|----------|
| `module not found` | 1 | Verify module with `show modules` |
| `failed to parse changelog` | 1 | Check CHANGELOG.md exists and is valid |
| `version not found` | 1 | List versions first |
| `no unreleased changes found` | 1 | Normal if no pending changes |
| `no released versions found` | 1 | Normal if no releases yet |

## See Also

- [show changelog](../show/changelog.md) - Human-readable output
- [get release-notes](./release-notes.md) - Release notes data
- [How-To Guide](../../../how-to-guides/eac/commands/release-management/view-changelog-release-notes.md)

{{ diataxis_footer() }}
