# validate Commands

Validate repository contracts, dependencies, and compliance for quality gates.

## Commands in this Category

| Command                                            | Purpose                                 |
| -------------------------------------------------- | --------------------------------------- |
| [validate](./validate.md)                          | Validate all repository contracts       |
| [validate artifacts](./artifacts.md)               | Validate build artifacts exist          |
| [validate books](./books.md)                       | Validate books.yml configuration        |
| [validate contracts](./contracts.md)               | Validate contracts against JSON schemas |
| [validate control-tags](./control-tags.md)         | Validate @control tags against OSCAL    |
| [validate dependencies](./dependencies.md)         | Validate module dependencies            |
| [validate design](./design.md)                     | Check workspace.dsl syntax              |
| [validate go-tidy](./go-tidy.md)                   | Validate Go dependencies are tidy       |
| [validate markdown](./markdown.md)                 | Validate markdown file syntax           |
| [validate module-files](./module-files.md)         | Validate module file ownership          |
| [validate module-hierarchy](./module-hierarchy.md) | Validate dependency graph structure     |
| [validate release](./release.md)                   | Validate changelog format               |
| [validate release-version](./release-version.md)   | Validate release version format         |
| [validate risk-catalog](./risk-catalog.md)         | Validate OSCAL catalogs                 |
| [validate risk-profile](./risk-profile.md)         | Validate OSCAL profile documents        |
| [validate specs](./specs.md)                       | Validate Gherkin specifications         |
| [validate test-tags](./test-tags.md)               | Validate test tags against contract     |

## Quick Examples

```bash
# Validate all contracts
eac validate

# Validate dependencies
eac validate dependencies
```

## See Also

- [Category Overview](../categories/validate.md)
- [scan Commands](../scan/index.md)
