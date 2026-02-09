# View Changelog and Release Notes

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
eac show changelog <module>
```

**Example:**

```bash
eac show changelog my-module
```

**Output:**

```markdown
# Changelog: my-module

## [1.2.2] - 2025-12-01

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
eac show changelog <module> <version>
```

**Example:**

```bash
eac show changelog my-module 1.2.2
```

### Display Latest Release

Show the most recent released version:

```bash
eac show changelog <module> latest
```

**Example:**

```bash
eac show changelog my-module latest
```

### Display Unreleased Changes

Show pending changes not yet released:

```bash
eac show changelog <module> unreleased
```

## View Release Notes

### Display Latest Release

Show release notes for the most recent version:

```bash
eac show release-notes <module>
```

**Example:**

```bash
eac show release-notes my-module
```

**Output:**

```markdown
## [1.2.3] - 2025-12-11

### Conclusion on Fitness for Intended Use

This release enhances the module...

### Impact on Business Process

The changes improve workflow automation...
```

### Display Latest Release (Explicit)

Show the most recent released version explicitly:

```bash
eac show release-notes <module> latest
```

**Example:**

```bash
eac show release-notes my-module latest
```

### Display Specific Version

Show release notes for a specific version:

```bash
eac show release-notes <module> <version>
```

**Example:**

```bash
eac show release-notes my-module 1.2.2
```

## Export Structured Data

### Export Changelog as JSON

Get changelog data in JSON format for scripting and automation:

```bash
eac get changelog <module> --as-json
```

**Example:**

```bash
eac get changelog my-module --as-json | jq '.versions[0].number'
```

### Export Changelog as YAML

Get changelog data in YAML format (default):

```bash
eac get changelog <module>
```

or explicitly:

```bash
eac get changelog <module> --as-yaml
```

### Export Specific Version

Get structured data for a specific version:

```bash
eac get changelog <module> <version> --as-json
```

**Example:**

```bash
eac get changelog my-module 1.2.2 --as-json
```

### Export Release Notes

Get release notes in structured format:

```bash
eac get release-notes <module> --as-json
eac get release-notes <module> --as-yaml
eac get release-notes <module> --as-toml
```

## Common Use Cases

### Review Changes Before Release

Check what's changed since the last release:

```bash
eac show changelog <module> unreleased
```

### Compare Versions

View two different versions side by side:

```bash
# PowerShell
eac show changelog my-module 1.2.1 > old.md
eac show changelog my-module 1.2.2 > new.md
code --diff old.md new.md

# Bash
eac show changelog my-module 1.2.1 > old.md
eac show changelog my-module 1.2.2 > new.md
diff old.md new.md
```

### Extract Version Number

Get the latest version programmatically:

```bash
# PowerShell
$version = (eac get changelog my-module --as-json | ConvertFrom-Json).versions[0].number

# Bash
version=$(eac get changelog my-module --as-json | jq -r '.versions[0].number')
```

### Generate Release Report

Combine changelog and release notes for a report:

```bash
eac show release-notes my-module 1.2.3 > report.md
eac show changelog my-module 1.2.3 >> report.md
```

## File Location Requirements

Commands automatically discover files through module contracts in `.eac/repository.yml`.

**Standard locations** (most modules):

- **Changelog:** `release/<module>/CHANGELOG.md`
- **Release Notes:** `release/<module>/RELEASE-NOTES.md`

**Example:** For module `clie`:

- Changelog: `release/clie/CHANGELOG.md`
- Release Notes: `release/clie/RELEASE-NOTES.md`

**Why centralized?** The `release/` folder keeps all release artifacts in one discoverable location. See [Understanding the Release Folder](./understanding-release-folder.md) for details on folder structure and how module contracts link to changelog files.

**Path resolution**: Commands use the `versioning.changelog` property from module contracts, falling back to `release/<module>/CHANGELOG.md` if not specified.

## File Types

**CHANGELOG.md** contains technical changes for developers:

- Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
- Categories: Added, Changed, Fixed, etc.
- Generated from conventional commits
- Used by: Developers, users, CI versioning

**RELEASE-NOTES.md** contains business assessment:

- Format: Custom fitness assessment
- Sections: Summary, fitness conclusion, business impact
- Manually authored by release manager
- Used by: Approval process, compliance, stakeholders

See [Understanding the Release Folder](./understanding-release-folder.md) for complete details on file types and structure.

## See Also

- [Understanding the Release Folder](./understanding-release-folder.md) - Folder structure and file types
- [Generate Changelog](./generate-changelog.md) - Create changelog from commits
- [Prepare Module Release](./prepare-module-release.md) - Complete release checklist
- [show Commands Reference](../../../../reference/eac/commands/show/index.md)
- [get Commands Reference](../../../../reference/eac/commands/get/index.md)
