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
r2r eac show specs ext-eac

# or explicitly
r2r eac show specs ext-eac unreleased
```

**Output:**
```markdown
# Specifications: ext-eac (Unreleased)

**Summary:** 2 added, 1 modified, 0 deleted (15 total scenarios)

| File | Status | Scenarios | Feature |
|------|--------|-----------|---------|
| specs/eac-commands/show-specs/specification.feature | ✨ Added | 5 | show-specs |
| specs/eac-commands/build/specification.feature | 📝 Modified | 6 | build |
```

## View Specifications for Latest Release

Show specifications that were included in the most recent release:

```bash
r2r eac show specs ext-eac latest
```

## View Specifications for Specific Version

```bash
r2r eac show specs ext-eac 0.0.7
```

## Query from Different Branches

By default, commands query from the trunk branch (usually `main`). Use `--branch` to query from other branches:

```bash
# Query from main branch (default)
r2r eac show specs ext-eac

# Query from current branch (useful when working in feature branches)
r2r eac show specs ext-eac --branch HEAD

# Query from specific branch
r2r eac show specs ext-eac --branch develop
```

**When to use this:**
- Working in a feature branch and want to see specs relative to that branch
- Comparing specs across different branches
- CI/CD pipelines running on non-main branches

## Export Structured Data

### As JSON

```bash
r2r eac get specs ext-eac --as-json
```

### As YAML (default)

```bash
r2r eac get specs ext-eac

# or explicitly
r2r eac get specs ext-eac --as-yaml
```

### As TOML

```bash
r2r eac get specs ext-eac --as-toml
```

## Common Use Cases

### 1. Count Total Scenarios in Unreleased Specs

```bash
r2r eac get specs ext-eac unreleased --as-json | jq '.total_scenarios'
```

**Example Output:**
```
15
```

### 2. List All Added Specifications

```bash
r2r eac get specs ext-eac latest --as-json | jq -r '.spec_files[] | select(.status == "Added") | .relative_path'
```

**Example Output:**
```
specs/eac-commands/show-specs/specification.feature
specs/eac-commands/get-specs/specification.feature
```

### 3. Get Scenario Count Per Specification

```bash
r2r eac get specs ext-eac --as-json | jq '.spec_files[] | {file: .feature_name, scenarios: .scenario_count}'
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
r2r eac get specs ext-eac --as-json | jq '{added: .added_count, modified: .modified_count, deleted: .deleted_count}'
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
r2r eac get specs ext-eac --as-json | jq '.spec_files | sort_by(.scenario_count) | reverse | .[0] | {file: .feature_name, scenarios: .scenario_count}'
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
r2r eac get specs ext-eac --as-json | jq '.deleted_count > 0'
```

**Example Output:**
```
false
```

### 7. Generate Release Documentation

Create a summary for release notes:

```bash
echo "## Specification Changes"
echo ""
echo "- Added: $(r2r eac get specs ext-eac latest --as-json | jq '.added_count') specifications"
echo "- Modified: $(r2r eac get specs ext-eac latest --as-json | jq '.modified_count') specifications"
echo "- Total Scenarios: $(r2r eac get specs ext-eac latest --as-json | jq '.total_scenarios')"
echo ""
echo "### Added Specifications:"
r2r eac get specs ext-eac latest --as-json | jq -r '.spec_files[] | select(.status == "Added") | "- " + .feature_name + " (" + (.scenario_count | tostring) + " scenarios)"'
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

## Troubleshooting

### "module not found" Error

**Problem:** The module moniker is invalid or doesn't exist.

**Solution:** List available modules:
```bash
r2r eac show modules
```

### "version not found" Error

**Problem:** The specified version doesn't exist in the changelog.

**Solution:** List available versions:
```bash
r2r eac get changelog ext-eac --as-json | jq -r '.versions[].number'
```

### "no released versions found" Error

**Problem:** Using `latest` keyword but module has no releases yet.

**Solution:** This is normal for new modules. Use `unreleased` or omit version parameter:
```bash
r2r eac show specs ext-eac unreleased
```

### No Specifications Shown

**Problem:** Git history doesn't show any .feature file changes for the version.

**Possible Causes:**
- No spec files were actually modified in git commits
- Spec files exist but weren't committed
- Looking at wrong version

**Solution:** Verify spec files were committed:
```bash
# Check if specs directory exists
ls specs/<module>/

# Check git history for .feature files
git log --oneline --name-status | grep ".feature"
```

### Bundle Modules Show Specs from Multiple Directories

**Question:** Why do I see specs from `eac-commands` when querying `ext-eac`?

**Answer:** Container/bundle modules like `ext-eac` automatically **aggregate specs from all their dependencies**. This is intentional and provides a complete view of all specifications included in the release bundle.

**Example:**
- `ext-eac` depends on `eac-commands` and `r2r-cli`
- Running `r2r eac show specs ext-eac` shows specs from all three modules
- This ensures you see the full scope of the release

**To see only a specific module's specs:**
```bash
# Query the dependency directly
r2r eac show specs eac-commands
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
