# Feature Naming Reference

Naming conventions for Gherkin feature files.

## Feature Name Format

**Pattern**: `[module-name_feature-name]`

**Structure**:

- Module name in kebab-case (e.g., `eac-commands`, `vscode-extension`)
- Underscore separator
- Feature name in kebab-case (e.g., `design-command`, `init-project`)

## Examples

```gherkin
Feature: cli_init-project
Feature: eac-commands_design-command
Feature: vscode-extension_status-bar
Feature: mcp-server_tools-registration
```

## Why This Format?

**Traceability**: Feature name appears in:

- Specification file path: `specs/cli/init-project/specification.feature`
- Step definition comments: `// Feature: cli_init-project`
- Unit test comments: `// Feature: cli_init-project`
- Test reports and logs

**Benefits**:

- **Find all tests** for a feature: `grep -r "Feature: cli_init-project" src/`
- **Module context** visible in the name itself
- **Unique identifiers** across the codebase
- **Consistent** with file system paths

## File Path Convention

```text
specs/<module>/<feature>/specification.feature
```

Example:

```text
specs/cli/init-project/specification.feature
```

## Related

- [Gherkin Limits](gherkin-limits.md)
- [Gherkin Concepts](../../explanation/specifications/gherkin-concepts.md)
