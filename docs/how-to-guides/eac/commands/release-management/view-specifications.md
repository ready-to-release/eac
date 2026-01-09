# View Release Specifications

## What You'll Accomplish

Learn how to view and analyze specification files (.feature files) that were added or modified for a specific release using the `show specs` and `get specs` commands.

## Prerequisites

- Repository with EAC configuration
- Module with specification files in `specs/<module>/` directory
- At least one commit with .feature file changes

## View Specifications for Unreleased Changes

Show specifications that have been added or modified since the last release:

```bash
# Human-readable markdown output
r2r eac show specs my-module

# or explicitly
r2r eac show specs my-module unreleased
```

**Output:**

```markdown
# Specifications: my-module (Unreleased)

**Summary:** 2 added, 1 modified, 0 deleted (15 total scenarios)

| File | Status | Scenarios | Feature |
|------|--------|-----------|---------|
| specs/my-module/show-specs/specification.feature | ✨ Added | 5 | show-specs |
| specs/my-module/build/specification.feature | 📝 Modified | 6 | build |
```

## View Specifications for Latest Release

Show specifications that were included in the most recent release:

```bash
r2r eac show specs my-module latest
```

## View Specifications for Specific Version

```bash
r2r eac show specs my-module 1.2.3
```

## Query from Different Branches

By default, commands query from the trunk branch (usually `main`). Use `--branch` to query from other branches:

```bash
# Query from main branch (default)
r2r eac show specs my-module

# Query from current branch (useful when working in feature branches)
r2r eac show specs my-module --branch HEAD

# Query from specific branch
r2r eac show specs my-module --branch develop
```

**When to use this:**

- Working in a feature branch and want to see specs relative to that branch
- Comparing specs across different branches
- CI/CD pipelines running on non-main branches

## Export Structured Data

### As JSON

```bash
r2r eac get specs my-module --as-json
```

### As YAML (default)

```bash
r2r eac get specs my-module

# or explicitly
r2r eac get specs my-module --as-yaml
```

### As TOML

```bash
r2r eac get specs my-module --as-toml
```

## Common Use Cases

### 1. Count Total Scenarios in Unreleased Specs

```bash
r2r eac get specs my-module unreleased --as-json | jq '.total_scenarios'
```

**Example Output:**

```text
15
```

### 2. List All Added Specifications

```bash
r2r eac get specs my-module latest --as-json | jq -r '.spec_files[] | select(.status == "Added") | .relative_path'
```

**Example Output:**

```text
specs/my-module/show-specs/specification.feature
specs/my-module/get-specs/specification.feature
```

### 3. Get Scenario Count Per Specification

```bash
r2r eac get specs my-module --as-json | jq '.spec_files[] | {file: .feature_name, scenarios: .scenario_count}'
```

**Example Output:**

```json
{
  "file": "show-specs",
  "scenarios": 5
}
{
  "file": "get-specs",
  "scenarios": 4
}
{
  "file": "build",
  "scenarios": 6
}
```

### 4. Count Specifications by Status

```bash
r2r eac get specs my-module --as-json | jq '{added: .added_count, modified: .modified_count, deleted: .deleted_count}'
```

**Example Output:**

```json
{
  "added": 2,
  "modified": 1,
  "deleted": 0
}
```

### 5. Find Specifications with Most Scenarios

```bash
r2r eac get specs my-module --as-json | jq '.spec_files | sort_by(.scenario_count) | reverse | .[0] | {file: .feature_name, scenarios: .scenario_count}'
```

**Example Output:**

```json
{
  "file": "build",
  "scenarios": 6
}
```

### 6. Check if Any Specs Were Deleted

```bash
r2r eac get specs my-module --as-json | jq '.deleted_count > 0'
```

**Example Output:**

```text
false
```

### 7. Generate Release Documentation

Create a summary for release notes:

```bash
echo "## Specification Changes"
echo ""
echo "- Added: $(r2r eac get specs my-module latest --as-json | jq '.added_count') specifications"
echo "- Modified: $(r2r eac get specs my-module latest --as-json | jq '.modified_count') specifications"
echo "- Total Scenarios: $(r2r eac get specs my-module latest --as-json | jq '.total_scenarios')"
echo ""
echo "### Added Specifications:"
r2r eac get specs my-module latest --as-json | jq -r '.spec_files[] | select(.status == "Added") | "- " + .feature_name + " (" + (.scenario_count | tostring) + " scenarios)"'
```

**Example Output:**

```markdown
## Specification Changes

- Added: 2 specifications
- Modified: 1 specifications
- Total Scenarios: 15

### Added Specifications:
- show-specs (5 scenarios)
- get-specs (4 scenarios)
```

## File Location Requirements

- Specifications must be in `specs/<module>/<feature>/` directory
- Files must have `.feature` extension
- Files must be tracked in git (committed)

## See Also

- [show specs Reference](../../../../reference/commands/show/specs.md) - Complete command reference
- [get specs Reference](../../../../reference/commands/get/specs.md) - JSON/YAML output reference
- [View Changelog and Release Notes](./view-changelog-release-notes.md) - View changes and notes
- [Generate Changelog](./generate-changelog.md) - Create changelog entries
