# get release-notes

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get release-notes <module> [version]`
**Purpose**: Get release notes data in structured format (YAML/JSON/TOML)
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get release-notes <module>
r2r eac get release-notes <module> <version>
r2r eac get release-notes <module> latest
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `module` | Yes | Module moniker (e.g., `ext-eac`, `eac-commands`) |
| `version` | No | Version number or `latest` keyword |

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
| *(omit)* | All versions | Full release notes object |

## Data Structure

### Full Release Notes (no version specified)

```yaml
versions:
  - number: "0.0.7"
    date: "2025-12-11T00:00:00Z"
    sections:
      - header: "Conclusion on Fitness for Intended Use"
        content: "This release is fit for..."
      - header: "Impact on Business Process"
        content: "The changes improve..."
  - number: "0.0.6"
    date: "2025-12-10T00:00:00Z"
    sections: [...]
```

### Single Version (with version/latest)

```yaml
number: "0.0.7"
date: "2025-12-11T00:00:00Z"
sections:
  - header: "Conclusion on Fitness for Intended Use"
    content: "This release is fit for production..."
  - header: "Impact on Business Process"
    content: "The changes improve workflow..."
  - header: "Risk Assessment"
    content: "Low risk deployment..."
```

## Examples

```bash
# Get all release notes as YAML
r2r eac get release-notes ext-eac

# Get as JSON
r2r eac get release-notes ext-eac --as-json

# Get latest version only
r2r eac get release-notes ext-eac latest --as-json

# Get specific version
r2r eac get release-notes ext-eac 0.0.7 --as-json

# Extract version numbers
r2r eac get release-notes ext-eac --as-json | jq -r '.versions[].number'

# Extract specific section
r2r eac get release-notes ext-eac latest --as-json | jq -r '.sections[] | select(.header == "Conclusion on Fitness for Intended Use") | .content'

# Count sections
r2r eac get release-notes ext-eac latest --as-json | jq '.sections | length'
```

## File Location

The command reads from: `release/<module>/RELEASE-NOTES.md`

**Example:** For module `ext-eac`, reads from `release/ext-eac/RELEASE-NOTES.md`

## Error Handling

| Error | Exit Code | Solution |
|-------|-----------|----------|
| `module not found` | 1 | Verify module with `show modules` |
| `failed to parse release notes` | 1 | Check RELEASE-NOTES.md exists and is valid |
| `version not found` | 1 | List versions first |
| `no released versions found` | 1 | Normal if no releases yet |

## See Also

- [show release-notes](../show/release-notes.md) - Human-readable output
- [get changelog](./changelog.md) - Changelog data
- [How-To Guide](../../../how-to-guides/eac/commands/release-management/view-changelog-release-notes.md)

{{ diataxis_footer() }}
