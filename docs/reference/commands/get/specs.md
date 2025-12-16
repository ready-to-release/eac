# get specs

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get specs <module> [version]`
**Purpose**: Get specifications data in structured format (YAML/JSON/TOML)
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get specs <module>
r2r eac get specs <module> <version>
r2r eac get specs <module> latest
r2r eac get specs <module> unreleased
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `module` | Yes | Module moniker (e.g., `ext-eac`, `eac-commands`) |
| `version` | No | Version number or special keyword (`latest`, `unreleased`) |

## Output Formats

| Flag | Format | Use Case |
|------|--------|-------------|
| *(none)* | YAML | Default, human-readable structured data |
| `--as-json` | JSON | Machine parsing, API integration |
| `--as-toml` | TOML | Configuration files |
| `--as-yaml` | YAML | Explicit YAML output |

## Special Keywords

| Keyword | Description | Output |
|---------|-------------|--------|
| `latest` | Most recent released version | Specs from that release |
| `unreleased` | Pending unreleased changes | Specs changed since last release |
| *(omit)* | Same as `unreleased` (default) | Specs changed since last release |

## Bundle Modules

For **container/bundle modules** with dependencies, specs are **aggregated from all dependent modules**.

**Example:** `ext-eac` depends on `eac-commands` and `r2r-cli`:

```bash
r2r eac get specs ext-eac --as-json
```

Returns specs from:
- `specs/eac-commands/` (dependency)
- `specs/r2r-cli/` (dependency)
- `specs/ext-eac/` (if any)

This provides a **complete view** of all specifications included in the release bundle.

**Regular modules** (without dependencies) only return specs from their own `specs/<module>/` directory.

## Data Structure

```yaml
module: "ext-eac"
version: "0.0.7"
added_count: 2
modified_count: 1
deleted_count: 0
total_scenarios: 15
spec_files:
  - file_path: "C:\\source\\simply-cli\\cli\\specs\\eac-commands\\show-specs\\specification.feature"
    relative_path: "specs/eac-commands/show-specs/specification.feature"
    module: "eac-commands"
    feature_name: "show-specs"
    title: "eac-commands_show-specs"
    status: "Added"
    scenario_count: 5
  - file_path: "C:\\source\\simply-cli\\cli\\specs\\eac-commands\\get-specs\\specification.feature"
    relative_path: "specs/eac-commands/get-specs/specification.feature"
    module: "eac-commands"
    feature_name: "get-specs"
    title: "eac-commands_get-specs"
    status: "Added"
    scenario_count: 4
  - file_path: "C:\\source\\simply-cli\\cli\\specs\\eac-commands\\build\\specification.feature"
    relative_path: "specs/eac-commands/build/specification.feature"
    module: "eac-commands"
    feature_name: "build"
    title: "eac-commands_build"
    status: "Modified"
    scenario_count: 6
```

## Examples

```bash
# Get specs as YAML (default)
r2r eac get specs ext-eac

# Get as JSON
r2r eac get specs ext-eac --as-json

# Get latest version specs
r2r eac get specs ext-eac latest --as-json

# Get unreleased specs
r2r eac get specs ext-eac unreleased --as-json

# Get specific version
r2r eac get specs ext-eac 0.0.7 --as-json

# Count total scenarios
r2r eac get specs ext-eac --as-json | jq '.total_scenarios'

# List all added specs
r2r eac get specs ext-eac latest --as-json | jq '.spec_files[] | select(.status == "Added") | .relative_path'

# Count specs by status
r2r eac get specs ext-eac --as-json | jq '{added: .added_count, modified: .modified_count, deleted: .deleted_count}'

# Get scenario count per spec
r2r eac get specs ext-eac --as-json | jq '.spec_files[] | {file: .relative_path, scenarios: .scenario_count}'
```

## File Location

Specifications are read from: `specs/<module>/`

**Example:** For module `eac-commands`, reads from `specs/eac-commands/`

## Error Handling

| Error | Exit Code | Solution |
|-------|-----------|----------|
| `module not found` | 1 | Verify module with `show modules` |
| `version not found` | 1 | List versions with `get changelog <module> --as-json \| jq '.versions[].number'` |
| `no released versions found` | 1 | Normal if no releases yet |
| `failed to open git repository` | 1 | Run from repository root |

## See Also

- [show specs](../show/specs.md) - Human-readable markdown output
- [get changelog](./changelog.md) - Changelog data for same version
- [How-To Guide](../../../how-to-guides/eac/commands/release-management/view-specifications.md)

{{ diataxis_footer() }}
