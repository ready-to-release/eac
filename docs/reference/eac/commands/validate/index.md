# Validate Commands

The **validate** category contains commands for validating repository contracts, dependencies, and compliance.

**Key Features**:

- Comprehensive validation of repository contracts
- Go module dependency checking
- Gherkin specification validation
- OSCAL compliance validation
- Design and documentation verification

## Commands in this Category

| Command                                            | Purpose                                 |
| -------------------------------------------------- | --------------------------------------- |
| [validate](./validate.md)                          | Validate all repository contracts       |
| [validate books](./books.md)                       | Validate books.yml configuration        |
| [validate control-tags](./control-tags.md)         | Validate @control tags against OSCAL    |
| [validate design](./design.md)                     | Check workspace.dsl syntax              |
| [validate go-tidy](./go-tidy.md)                   | Validate Go dependencies are tidy       |
| [validate markdown](./markdown.md)                 | Validate markdown file syntax           |
| [validate module-files](./module-files.md)         | Validate module file ownership          |
| [validate release](./release.md)                   | Validate changelog format               |
| [validate release-version](./release-version.md)   | Validate release version format         |
| [validate risk-catalog](./risk-catalog.md)         | Validate OSCAL catalogs                 |
| [validate risk-profile](./risk-profile.md)         | Validate OSCAL profile documents        |
| [validate test-tags](./test-tags.md)               | Validate test tags against contract     |

## Common Use Cases

### Pre-commit Validation

```bash
eac validate
```

### Dependency Validation

```bash
eac validate dependencies
eac validate go-tidy
```

### Documentation Validation

```bash
eac validate specs
eac validate markdown
eac validate design
```

## See Also

- [show config](../show/config.md)
- [get modules](../get/modules.md)
- [scan Commands](../scan/index.md)
