# validate Commands

{{ page_breadcrumb() }}

## Overview

The **validate** category contains 15 commands for validating repository contracts, dependencies, and compliance.

## Commands

| Command | Purpose |
|---------|---------|
| [validate](../validate/validate.md) | Validate all repository contracts and dependencies |
| [validate contracts](../validate/contracts.md) | Validate contracts against JSON schemas |
| [validate dependencies](../validate/dependencies.md) | Validate module dependencies against contracts |
| [validate specs](../validate/specs.md) | Validate Gherkin specifications |
| [validate artifacts](../validate/artifacts.md) | Validate build artifacts exist |
| [validate module-files](../validate/module-files.md) | Validate module file ownership |
| [validate module-hierarchy](../validate/module-hierarchy.md) | Validate dependency graph structure |
| [validate go-tidy](../validate/go-tidy.md) | Validate Go dependencies are tidy |
| [validate design](../validate/design.md) | Check workspace.dsl syntax |
| [validate books](../validate/books.md) | Validate books.yml configuration |
| [validate markdown](../validate/markdown.md) | Validate markdown file syntax |
| [validate test-tags](../validate/test-tags.md) | Validate test tags against contract |
| [validate control-tags](../validate/control-tags.md) | Validate @control tags against OSCAL |
| [validate risk-profile](../validate/risk-profile.md) | Validate OSCAL profile documents |
| [validate risk-catalog](../validate/risk-catalog.md) | Validate OSCAL catalogs |

## Common Use Cases

### Pre-commit Validation

```bash
r2r eac validate
```

### Dependency Validation

```bash
r2r eac validate dependencies
r2r eac validate go-tidy
```

### Documentation Validation

```bash
r2r eac validate specs
r2r eac validate markdown
r2r eac validate design
```

## Key Features

- Comprehensive validation of repository contracts
- Go module dependency checking
- Gherkin specification validation
- OSCAL compliance validation
- Design and documentation verification

## See Also

- [show config](../show/config.md)
- [get modules](../get/modules.md)
- [scan Commands](./scan.md)

{{ diataxis_footer() }}
