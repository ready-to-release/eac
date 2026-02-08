# Naming Conventions

> **Feature and file naming standards**

Consistent naming conventions enable traceability and organization across the codebase.

---

## Feature Name Format

Feature names use kebab-case: `[module-name_feature-name]`

**Structure**:

- Module name in kebab-case (e.g., `eac-commands`, `vscode-extension`)
- Underscore separator
- Feature name in kebab-case (e.g., `design-command`, `init-project`)

---

## Examples

```gherkin
Feature: cli_init-project
Feature: eac-commands_design-command
Feature: vscode-extension_status-bar
Feature: eac-mcp-server_tools-registration
```

---

## Why This Format?

### Traceability

Feature name appears in:

- Specification file path: `specs/cli/init-project/specification.feature`
- Step definition comments: `// Feature: cli_init-project`
- Unit test comments: `// Feature: cli_init-project`
- Test reports and logs

### Benefits

**Find all tests** for a feature:

```bash
grep -r "Feature: cli_init-project" src/
```

**Module context** visible in the name itself:

- `cli_init-project` immediately tells you this is a CLI module feature

**Unique identifiers** across the codebase:

- No feature name collisions between modules
- Clear ownership and organization

**Consistent** with file system paths:

- Feature name matches directory structure
- Easy to navigate between specs and implementation

---

## File Path Convention

### Specification Files

**Pattern**:

```text
specs/<module>/<feature>/specification.feature
```

**Examples**:

```text
specs/cli/init-project/specification.feature
specs/eac-commands/design-command/specification.feature
specs/vscode-extension/status-bar/specification.feature
```

### Relationship Between Name and Path

| Feature Name                  | File Path                                                 |
| ----------------------------- | --------------------------------------------------------- |
| `cli_init-project`            | `specs/cli/init-project/specification.feature`            |
| `eac-commands_design-command` | `specs/eac-commands/design-command/specification.feature` |
| `vscode-extension_status-bar` | `specs/vscode-extension/status-bar/specification.feature` |

**Pattern**: `module_feature-name` → `specs/module/feature-name/specification.feature`

---

## Naming Rules

### Module Name

**Format**: Lowercase, kebab-case

**Examples**:

- `cli`
- `eac-commands`
- `vscode-extension`
- `eac-mcp-server`

### Feature Name

**Format**: Lowercase, kebab-case

**Guidelines**:

- Use noun-verb or action-oriented names
- Be specific but concise
- Focus on user-facing capability

**Good examples**:

- `init-project`
- `design-command`
- `status-bar`
- `tools-registration`

**Bad examples**:

- `feature1` (meaningless)
- `InitProject` (wrong case)
- `init_project` (underscore instead of hyphen)
- `the-init-project-feature` (too verbose)

---

## Complete Naming Example

```gherkin
@cli @critical
Feature: cli_init-project
  # Feature name: module_feature-name
  # File path: specs/cli/init-project/specification.feature

  As a developer
  I want to initialize a CLI project
  So that I can quickly start development

  Rule: Creates project directory structure

    @ov
    Scenario: Initialize in empty directory
      Given I am in an empty folder
      When I run "r2r init"
      Then a file named "r2r.yaml" should be created
```

---

## Tag Naming Conventions

### Module Tags

**Format**: Lowercase, matches module name

**Examples**:

```gherkin
@cli
@eac-commands
@vscode-extension
@eac-mcp-server
```

**Usage**: Applied at Feature level to identify which module owns the feature.

### Priority Tags

**Format**: Lowercase, one of: `@critical`, `@high`, `@medium`, `@low`

**Examples**:

```gherkin
@critical  # Must work for basic functionality
@high      # Important but not blocking
@medium    # Nice to have
@low       # Future enhancement
```

### Acceptance Criteria Tags

**Format**: `@ac<number>` where number is 1-indexed

**Examples**:

```gherkin
@ac1  # Links to first acceptance criterion (Rule)
@ac2  # Links to second acceptance criterion (Rule)
@ac3  # Links to third acceptance criterion (Rule)
```

**Usage**: Applied at Scenario level to link scenarios back to Rule blocks.

---

## Best Practices

### DO

- ✅ Use lowercase for all names
- ✅ Use kebab-case (hyphens) within names
- ✅ Use underscore to separate module from feature
- ✅ Match feature name to file path structure
- ✅ Use descriptive, specific names
- ✅ Keep feature names concise (2-4 words)

### DON'T

- ❌ Use uppercase or mixed case
- ❌ Use underscores within module or feature names (only between them)
- ❌ Use spaces in feature names
- ❌ Create generic names like "feature1" or "test-feature"
- ❌ Make names too long (more than 5 words)
- ❌ Duplicate feature names across modules

---

## Naming Checklist

Use this checklist when naming a new feature:

- [ ] Is the module name in kebab-case?
- [ ] Is the feature name in kebab-case?
- [ ] Is there an underscore separating module and feature?
- [ ] Is the name all lowercase?
- [ ] Does the file path match the naming pattern?
- [ ] Is the feature name specific enough to understand purpose?
- [ ] Is the feature name concise (2-4 words)?
- [ ] Is this feature name unique across the codebase?

---

## Related Documentation

- [File Structure](./file-structure.md) - Where to place specification files
- [Size Guidelines](./size-guidelines.md) - How to name when splitting features
- [Organizing Rules](./organizing-rules.md) - Naming Rules within features
