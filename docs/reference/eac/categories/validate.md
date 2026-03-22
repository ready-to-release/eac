# validate Commands

## Overview

The **validate** category contains commands for validating repository contracts, dependencies, and compliance.

## Commands

<!-- book:category-commands validate -->

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

## Key Features

- Comprehensive validation of repository contracts
- Go module dependency checking
- Gherkin specification validation
- OSCAL compliance validation
- Design and documentation verification

## See Also

- [show config](../commands/show/config.md)
- [get modules](../commands/get/modules.md)
- [scan Commands](./scan.md)
