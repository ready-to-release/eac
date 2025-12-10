# scan sbom

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac scan-sbom [target]`
**Purpose**: Generate Software Bill of Materials (SBOM)
**Category**: [scan](../categories/scan.md)

## Syntax

```bash
r2r eac scan-sbom [target]
```

## Examples

```bash
# Generate SBOM
r2r eac scan-sbom .

# Generate for container
r2r eac scan-sbom docker.io/myapp:latest

# Output format
r2r eac scan-sbom . --format cyclonedx
```

## See Also

- [scan vuln](./vuln.md)
- [scan Commands](../categories/scan.md)

{{ diataxis_footer() }}
