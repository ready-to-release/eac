# Show specs

<!-- book:cmd show specs -->

## Special Keywords

| Keyword      | Description                                             |
| ------------ | ------------------------------------------------------- |
| `latest`     | Show specs included in the most recent released version |
| `unreleased` | Show specs added/modified since last release (default)  |
| _(omit)_     | Same as `unreleased` - specs changed since last release |

## Bundle Modules

For **container/bundle modules** with dependencies, specs are **aggregated from all dependent modules**.

**Example:** When querying `ext-eac` (which depends on `eac-commands` and `r2r-cli`):

```bash
r2r eac show specs ext-eac
```

**Shows specs from:**

- `specs/eac-commands/` (dependency)
- `specs/r2r-cli/` (dependency)
- `specs/ext-eac/` (if any)

This provides a **complete view** of all specifications included in the release bundle.

**Regular modules** (without dependencies) only show specs from their own `specs/<module>/` directory.

## Output Format

Markdown-formatted table with:

- Version header (`# Specifications: module (version)`)
- Summary line with counts (added, modified, deleted, total scenarios)
- Table columns: File, Status, Scenarios, Feature
- Status icons: ✨ Added, 📝 Modified, 🗑️ Deleted

## Example Output

```markdown
# Specifications: ext-eac (0.0.7)

**Summary:** 2 added, 1 modified, 0 deleted (15 total scenarios)

| File                                                | Status      | Scenarios | Feature    |
| --------------------------------------------------- | ----------- | --------- | ---------- |
| specs/eac-commands/show-specs/specification.feature | ✨ Added    | 5         | show-specs |
| specs/eac-commands/get-specs/specification.feature  | ✨ Added    | 4         | get-specs  |
| specs/eac-commands/build/specification.feature      | 📝 Modified | 6         | build      |
```

## File Location

Specifications are read from: `specs/<module>/`

**Example:** For module `eac-commands`, reads from `specs/eac-commands/`

## Error Handling

| Error                           | Cause                               | Solution                                                                  |
| ------------------------------- | ----------------------------------- | ------------------------------------------------------------------------- |
| `module not found`              | Invalid module moniker              | Check `show modules` for valid names                                      |
| `version not found`             | Version doesn't exist in changelog  | Verify with `get changelog <module> --as-json \| jq '.versions[].number'` |
| `no released versions found`    | No releases yet when using `latest` | Normal for new modules                                                    |
| `failed to open git repository` | Not in a git repository             | Run from repository root                                                  |

## See Also

- [get specs](../get/specs.md) - Structured JSON/YAML output
- [show changelog](./changelog.md) - View changelog for same version
- [How-To Guide](../../../how-to-guides/eac/commands/release-management/view-specifications.md)
