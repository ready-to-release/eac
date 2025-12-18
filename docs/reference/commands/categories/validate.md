# validate Commands

## Overview

The **validate** category contains 15 commands for validating repository contracts, dependencies, and compliance.

## Commands

<!-- book:category-commands validate -->

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
