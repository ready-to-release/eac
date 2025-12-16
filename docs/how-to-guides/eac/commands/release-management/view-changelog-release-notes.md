# View Changelog and Release Notes

{{ page_breadcrumb() }}

Learn how to view changelog entries and release notes for modules in both human-readable and structured formats.

## What You'll Accomplish

- View changelog entries for a module
- Display release notes for a module
- Export changelog data in structured formats (JSON/YAML/TOML)
- Query specific versions

## Prerequisites

- Repository with modules that have `CHANGELOG.md` and/or `RELEASE-NOTES.md` files
- Module files located in their root directory

## View Changelog

### Display All Versions

Show all changelog versions in human-readable markdown format:

```bash
r2r eac show changelog <module>
```

**Example:**

```bash
r2r eac show changelog eac-commands
```

**Output:**

```markdown
# Changelog: eac-commands

## [0.0.2] - 2025-12-01

### Added

| Description                                  | Type | Scope        | Commit |
| -------------------------------------------- | ---- | ------------ | ------ |
| add risks assessment and control creation... | feat | multi-module | -      |
| add dual-output logging with debug support   | feat | multi-module | -      |
...

### Changed
...

### Fixed
...
```

### Display Specific Version

Show changelog for a specific version:

```bash
r2r eac show changelog <module> <version>
```

**Example:**

```bash
r2r eac show changelog eac-commands 0.0.2
```

### Display Latest Release

Show the most recent released version:

```bash
r2r eac show changelog <module> latest
```

**Example:**

```bash
r2r eac show changelog ext-eac latest
```

### Display Unreleased Changes

Show pending changes not yet released:

```bash
r2r eac show changelog <module> unreleased
```

## View Release Notes

### Display Latest Release

Show release notes for the most recent version:

```bash
r2r eac show release-notes <module>
```

**Example:**

```bash
r2r eac show release-notes ext-eac
```

**Output:**

```markdown
## [0.0.7] - 2025-12-11

### Conclusion on Fitness for Intended Use

This release enhances the EAC extension...

### Impact on Business Process

The changes improve workflow automation...
```

### Display Latest Release (Explicit)

Show the most recent released version explicitly:

```bash
r2r eac show release-notes <module> latest
```

**Example:**

```bash
r2r eac show release-notes ext-eac latest
```

### Display Specific Version

Show release notes for a specific version:

```bash
r2r eac show release-notes <module> <version>
```

**Example:**

```bash
r2r eac show release-notes ext-eac 0.0.6
```

## Export Structured Data

### Export Changelog as JSON

Get changelog data in JSON format for scripting and automation:

```bash
r2r eac get changelog <module> --as-json
```

**Example:**

```bash
r2r eac get changelog eac-commands --as-json | jq '.versions[0].number'
```

### Export Changelog as YAML

Get changelog data in YAML format (default):

```bash
r2r eac get changelog <module>
```

or explicitly:

```bash
r2r eac get changelog <module> --as-yaml
```

### Export Specific Version

Get structured data for a specific version:

```bash
r2r eac get changelog <module> <version> --as-json
```

**Example:**

```bash
r2r eac get changelog eac-commands 0.0.2 --as-json
```

### Export Release Notes

Get release notes in structured format:

```bash
r2r eac get release-notes <module> --as-json
r2r eac get release-notes <module> --as-yaml
r2r eac get release-notes <module> --as-toml
```

## Common Use Cases

### Review Changes Before Release

Check what's changed since the last release:

```bash
r2r eac show changelog <module> unreleased
```

### Compare Versions

View two different versions side by side:

```bash
# PowerShell
r2r eac show changelog eac-commands 0.0.1 > old.md
r2r eac show changelog eac-commands 0.0.2 > new.md
code --diff old.md new.md

# Bash
r2r eac show changelog eac-commands 0.0.1 > old.md
r2r eac show changelog eac-commands 0.0.2 > new.md
diff old.md new.md
```

### Extract Version Number

Get the latest version programmatically:

```bash
# PowerShell
$version = (r2r eac get changelog eac-commands --as-json | ConvertFrom-Json).versions[0].number

# Bash
version=$(r2r eac get changelog eac-commands --as-json | jq -r '.versions[0].number')
```

### Generate Release Report

Combine changelog and release notes for a report:

```bash
r2r eac show release-notes ext-eac 0.0.7 > report.md
r2r eac show changelog ext-eac 0.0.7 >> report.md
```

## Troubleshooting

### "Module not found" Error

**Problem:** Command reports module doesn't exist.

**Solution:** Verify module moniker:

```bash
r2r eac show modules | grep <module>
```

### "Failed to parse changelog" Error

**Problem:** Changelog file doesn't exist or has incorrect format.

**Solution:** Check file exists and follows [Keep a Changelog](https://keepachangelog.com/) format:

```bash
ls <module-root>/CHANGELOG.md
```

### "Version not found" Error

**Problem:** Requested version doesn't exist in changelog.

**Solution:** List all available versions:

```bash
r2r eac get changelog <module> --as-json | jq '.versions[].number'
```

## File Location Requirements

Commands expect files in these locations:

- **Changelog:** `<module-root>/CHANGELOG.md`
- **Release Notes:** `<module-root>/RELEASE-NOTES.md`

**Example:** For module `eac-commands` with root `go/eac/commands`:

- Changelog: `go/eac/commands/CHANGELOG.md`
- Release Notes: `go/eac/commands/RELEASE-NOTES.md`

## See Also

- [Generate Changelog](./generate-changelog.md) - Create changelog from commits
- [Prepare Module Release](./prepare-module-release.md) - Complete release checklist
- [show Commands Reference](../../../../reference/commands/show/index.md)
- [get Commands Reference](../../../../reference/commands/get/index.md)

{{ diataxis_footer() }}
