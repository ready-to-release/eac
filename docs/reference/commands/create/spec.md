# create spec

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac create spec <module> <description>`
**Purpose**: Generate Gherkin specifications from natural language descriptions
**Category**: [create](../categories/create.md)

## Syntax

```bash
r2r eac create spec <module> <description>
```

## Examples

```bash
# Generate spec from description
r2r eac create spec src-auth "User can login with valid credentials"

# Generate multiple scenarios
r2r eac create spec src-api "API endpoints validate input and return proper errors"
```

## See Also

- [validate specs](../validate/specs.md)
- [test](../test/test.md)
- [create Commands](../categories/create.md)

{{ diataxis_footer() }}
