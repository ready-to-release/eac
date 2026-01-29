# Testing Reference

Technical reference for EAC test suite configuration, execution, and CLI commands.

## In This Section

| Reference                           | Description                                            |
| ----------------------------------- | ------------------------------------------------------ |
| [Test Suites](./test-suites.md)     | Suite definitions, tag selection, and CD stage mapping |
| [Manual Testing](./manual-tests.md) | Manual test workflows, schemas, and CI/CD integration  |
| [Go Testing](./go/index.md)         | Go/Godog BDD implementation, step definitions          |

## Quick Reference

```bash
# Run test suites
r2r eac test <module> --suite unit
r2r eac test <module> --suite integration
r2r eac test <module> --suite acceptance
r2r eac test <module> --suite production-verification

# Debug test selection
r2r eac test <module> --suite unit --dry-run
r2r eac test <module> --suite unit --count
```

## Related Documentation

- [Test Command Reference](../commands/test/index.md) - Full test command options
- [Test Levels (Conceptual)](../../../explanation/specifications/taxonomy/test-levels.md) - L0-L4 environment concepts
