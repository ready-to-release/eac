# Go Testing Reference

Technical reference for Go/Godog BDD implementation in EAC.

## In This Section

| Reference                                   | Description                               |
| ------------------------------------------- | ----------------------------------------- |
| [Overview](./overview.md)                   | Go testing architecture and patterns      |
| [File Organization](./file-organization.md) | Directory structure and file naming       |
| [Step Definitions](./step-definitions.md)   | Writing Godog step definitions            |
| [Test Levels](./test-levels.md)             | L0-L4 test environment implementation     |
| [Best Practices](./best-practices.md)       | Guidelines for writing maintainable tests |

## Quick Start

```bash
# Run Go tests for a module
eac test eac-commands

# Run specific suite
eac test eac-commands --suite unit

# Run with verbose output
eac test eac-commands --verbose
```

## Related Documentation

- [Test Suites](../test-suites.md) - Suite definitions and configuration
- [Specifications](../../specifications/index.md) - Gherkin specification reference
