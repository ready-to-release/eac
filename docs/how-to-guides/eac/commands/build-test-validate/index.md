# Build, Test & Validate

{{ page_breadcrumb() }}

Learn how to build modules, run tests, and validate quality before committing.

## In This Section

### Building Modules

| Guide | What You'll Accomplish |
|-------|------------------------|
| [Build Single Module](./build-single-module.md) | Compile a module and generate artifacts |
| [Build Changed Modules](./build-changed-modules.md) | Build only affected modules for efficiency |

### Testing

| Guide | What You'll Accomplish |
|-------|------------------------|
| [Run Tests for Module](./run-tests-for-module.md) | Execute tests and view results |
| [Debug Test Failures](./debug-test-failures.md) | Identify and fix failing tests |
| [Run Test Suites](./run-test-suites.md) | Execute specific test suites |

### Validation and Quality

| Guide | What You'll Accomplish |
|-------|------------------------|
| [Validate Before Commit](./validate-before-commit.md) | Run comprehensive quality checks |
| [Check Dependencies](./check-dependencies.md) | Verify dependency contracts |
| [Validate Specifications](./validate-specifications.md) | Check Gherkin specs for quality |
| [Scan for Security Issues](./scan-for-security-issues.md) | Detect vulnerabilities and secrets |
| [Manage Risk Compliance](./manage-risk-compliance.md) | Track OSCAL compliance with automated evidence |

## Complete Development Workflow

### Local Development Cycle

1. **Make Changes** - Edit code, add features, fix bugs
2. **Build** - [Compile module](./build-single-module.md) to verify syntax
3. **Test** - [Run tests](./run-tests-for-module.md) to verify behavior
4. **Debug** - [Fix failures](./debug-test-failures.md) if needed
5. **Validate** - [Run quality checks](./validate-before-commit.md)
6. **Commit** - Push changes with confidence

### Pre-Commit Quality Gates

Run these validations before every commit:

- **Contract validation** - Ensure module contracts are valid
- **Dependency checking** - Verify dependencies are correct
- **Specification validation** - Check Gherkin quality
- **Security scanning** - Detect vulnerabilities and secrets
- **Test execution** - Verify all tests pass

See: [Validate Before Commit](./validate-before-commit.md)

### CI/CD Integration

Optimize your pipeline with:

- **Smart builds** - [Build only changed modules](./build-changed-modules.md)
- **Parallel testing** - [Run test suites](./run-test-suites.md) concurrently
- **Quality metrics** - Generate coverage and timing reports
- **Security scans** - Automated [vulnerability detection](./scan-for-security-issues.md)

## Quick Commands

```bash
# Build a module
go run ./go/eac/commands build <module-name>

# Run tests
go run ./go/eac/commands test <module-name>

# Validate everything
go run ./go/eac/commands validate

# Security scan
go run ./go/eac/commands scan
```

{{ diataxis_footer() }}
