# Specifications Reference

Quick reference for specification formats, commands, and syntax.

---

## Architecture Overview

This project maintains clear separation between specifications and implementations:

| Concern        | Location   | Contents                      |
| -------------- | ---------- | ----------------------------- |
| Specifications | `specs/`   | Gherkin `.feature` files      |
| Implementations| `src/`     | Go test code, step definitions|

See [Gherkin Format](./gherkin-format.md) for specification syntax.
See [Godog Commands](./godog-commands.md) for running tests from `src/`.

---

## Naming Convention

All specification files MUST follow the kebab-case naming convention.

### Feature Name Format

```
[module-name_feature-name]
```

**Components**:
- **Module**: kebab-case identifier of owning module
- **Feature**: kebab-case description of feature
- **Separator**: Single underscore `_`

### Module Names

Use kebab-case for multi-word modules:

| Module | Correct | Incorrect |
|--------|---------|-----------|
| Source commands | `src-commands` | `src_commands`, `srcCommands` |
| VSCode extension | `vscode-extension` | `vscode_extension`, `vscodeExtension` |
| MCP server | `mcp-server` | `mcp_server`, `mcpServer` |
| CLI | `cli` | `CLI`, `Cli` |

### Feature Names

Use kebab-case for all features:

| Feature | Correct | Incorrect |
|---------|---------|-----------|
| Design command | `design-command` | `design_command`, `designCommand` |
| Commit workflow | `commit-workflow` | `commit_workflow`, `commitWorkflow` |
| Init project | `init-project` | `init_project`, `initProject` |
| GitHub integration | `github-integration` | `github_integration`, `githubIntegration` |

### Complete Examples

Correct format:
```
src-commands_design-command       ✅
vscode-extension_commit-workflow  ✅
cli_init-project                  ✅
mcp-server_github-integration     ✅
docs_architecture-guide           ✅
```

Incorrect format (avoid):
```
src_commands_design_command       ❌ (no kebab-case)
srcCommands_designCommand         ❌ (camelCase)
SrcCommands_DesignCommand         ❌ (PascalCase)
src-commands-design-command       ❌ (missing underscore separator)
SRC-COMMANDS_DESIGN-COMMAND       ❌ (uppercase)
```

### Rationale

Kebab-case provides:
- **Consistency**: Same format as file/directory names
- **Readability**: Words clearly separated by hyphens
- **CLI-friendly**: Works well in URLs, file paths, command names
- **Git-friendly**: No special escaping needed

### File Path Examples

```
specs/src-commands/design-command/specification.feature
specs/vscode-extension/commit-workflow/specification.feature
specs/cli/init-project/specification.feature
specs/mcp-server/github-integration/specification.feature
```

**Note**: File paths and feature names should match using the same kebab-case convention.

---

## Specification Format References

### [Gherkin Format](gherkin-format.md)

specification.feature file structure with Rule blocks for ATDD and Scenario blocks for BDD.

**Quick lookup:**

- Template structure with Rules and Scenarios
- Feature metadata and tags
- Feature naming convention (kebab-case)
- Rule blocks (ATDD - acceptance criteria)
- Scenario blocks (BDD - executable examples)
- Given/When/Then syntax
- Tag system (@ac1, @success, @error, @IV, @PV, @risk)
- Background sections
- Feature name traceability
- Complete examples

### [TDD Format](tdd-format.md)

Unit test structure and Go test conventions.

**Quick lookup:**

- Test file patterns and naming
- Arrange-Act-Assert pattern
- Table-driven tests
- Error handling and assertions
- Feature name traceability
- Running tests and coverage

---

## Command References

### [Godog Commands](godog-commands.md)

Running Godog tests with go test (executes both ATDD and BDD layers).

---

## Scenario Classification

### [Verification Tags](verification-tags.md)

Scenario classification for implementation reports (IV/OV/PV).

---

## Related Documents

- **Understanding-oriented?** See [Specifications Explanation](../../explanation/specifications/index.md)
- **Need to perform tasks?** See [Specifications How-to Guides](../../how-to-guides/specifications/index.md)
