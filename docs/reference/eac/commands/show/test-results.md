# show test-results

<!-- book:cmd show test-results -->

## Report Sections

The output includes:

- **Module Overview** - Pass/fail counts per module with duration
- **Specification Coverage** - Godog feature coverage
- **Test Results by Module** - Complete test listing with all tags
- **Summary** - Aggregations by type, suite, and control

## Tags Display

Tags are displayed without the @ prefix for readability:

- `L0`, `L1`, `L2`, `L4` - Test levels
- `ov`, `iv`, `pv`, `piv` - Verification types
- `control:ai-2` - Control references
- `deps:go` - System dependencies
- `depm:eac-core` - Module dependencies

## Prerequisites

Test manifests must exist from previous test runs:

```bash
# Run tests first
r2r eac test <module>

# Then show results
r2r eac show test-results
```

## See Also

- [get test-results](../get/test-results.md) - Structured data
- [show tests](./tests.md) - Test definitions
- [test](../test/test.md) - Run tests
