# Building and Testing

{{ page_breadcrumb() }}

Learn how to build modules and run tests to verify your changes.

## In This Section

| Guide | What You'll Accomplish |
|-------|------------------------|
| [Build Single Module](./build-single-module.md) | Compile a module and generate artifacts |
| [Build Changed Modules](./build-changed-modules.md) | Build only affected modules for efficiency |
| [Run Tests for Module](./run-tests-for-module.md) | Execute tests and view results |
| [Debug Test Failures](./debug-test-failures.md) | Identify and fix failing tests |
| [Run Test Suites](./run-test-suites.md) | Execute specific test suites |

## Build and Test Workflow

### Local Development

1. Make code changes
2. [Build module](./build-single-module.md) to compile
3. [Run tests](./run-tests-for-module.md) to verify
4. [Debug failures](./debug-test-failures.md) if needed
5. Commit changes

### CI Integration

- [Build only changed modules](./build-changed-modules.md) for efficiency
- [Run test suites](./run-test-suites.md) in parallel
- Generate test summaries and coverage reports

{{ diataxis_footer() }}
