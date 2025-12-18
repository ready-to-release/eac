# Dependency Tags
>
> **System, module, and environment dependencies**

Declare required tooling and dependencies for test execution.

---

## System Dependencies (`@deps:`)

System dependency tags declare required tooling for test execution.

### Available System Dependencies

**`@deps:docker`** - Docker engine required
**`@deps:git`** - Git CLI required
**`@deps:go`** - Go toolchain required
**`@deps:az-cli`** - Azure CLI required

### Dependency Checking

**Local Development**:

- Warning + skip tests with missing dependencies
- Allows development without all tools installed

**CI Environment**:

- Fail fast on missing dependencies
- Ensures CI has required tooling

**Override**: `--dep-check=warn|fail` flag

### Example

```gherkin
@L2 @deps:docker @deps:git @ov
Feature: Container Build Pipeline
  Tests requiring Docker and Git for artifact builds

  @ov
  Scenario: Build container from Git repository
    Given I have a Git repository with Dockerfile
    When I run the container build
    Then the container image should be created
    And the image should pass security scan
```

---

## Module Dependencies (`@depm:`)

Module dependency tags declare required Go modules for test execution.

### Format

**Single Module**: `@depm:<module-moniker>`
**Multiple Modules**: `@depm:<module1>` `@depm:<module2>`

### Example

```gherkin
@L2 @depm:cli-core @depm:parser @ov
Feature: CLI Command Processing
  Tests requiring CLI core and parser modules

  @ov
  Scenario: Parse complex command
    Given I have the CLI core module
    And I have the parser module
    When I parse "deploy --env prod --region us-east-1"
    Then the command should be parsed correctly
```

---

## Environment Dependencies (`@env:`)

Environment dependency tags declare required infrastructure or environment configurations.

### Common Environment Dependencies

**`@env:plte`** - Production-Like Test Environment required
**`@env:prod`** - Production environment required
**`@env:dev`** - Development environment required
**`@env:cloud`** - Cloud infrastructure required

### Example

```gherkin
@L3 @env:plte @iv
Feature: Deployment Verification
  Tests requiring PLTE infrastructure

  @iv
  Scenario: Deploy to PLTE
    Given the PLTE environment is available
    When I deploy version 1.2.3
    Then the deployment should succeed
    And health checks should pass
```

---

## Dependency Accumulation

Dependencies accumulate from Feature → Rule → Scenario levels.

```gherkin
@L2 @deps:docker
Feature: Container Tests

  @deps:git
  Rule: Container builds from Git

    @ov
    Scenario: Build container from repository
      # Effective dependencies: @deps:docker, @deps:git
```

---

## Best Practices

### System Dependencies

✅ **DO**:

- Declare all required system tools
- Use standard dependency names
- Document tool version requirements in comments
- Test locally with and without dependencies

❌ **DON'T**:

- Assume tools are installed
- Use custom dependency names
- Forget to declare transitive dependencies
- Skip dependency checks in CI

### Module Dependencies

✅ **DO**:

- Declare required Go modules explicitly
- Use module monikers consistently
- Keep dependency list minimal
- Review dependencies regularly

❌ **DON'T**:

- Over-declare dependencies (creates false requirements)
- Use full module paths (use monikers)
- Create circular dependencies

### Environment Dependencies

✅ **DO**:

- Clearly specify environment requirements
- Document environment prerequisites
- Fail gracefully when environment unavailable
- Use environment-specific test suites

❌ **DON'T**:

- Run environment-specific tests in wrong environment
- Hardcode environment URLs or credentials
- Skip environment validation

---

## Dependency Checking

### Local Development Mode

```bash
# Warn and skip tests with missing dependencies
go test ./... --dep-check=warn

# Example output:
# ⚠️ Skipping test: Docker not available
# ✅ 10 tests passed
# ⚠️ 3 tests skipped
```

### CI Mode

```bash
# Fail fast on missing dependencies
go test ./... --dep-check=fail

# Example output:
# ❌ Error: Required dependency 'docker' not found
# Pipeline failed
```

---

## Multiple Dependency Types

Combine different dependency types:

```gherkin
@L2 @deps:docker @deps:git @depm:build-tools @ov
Feature: Complete Build Pipeline
  Tests requiring Docker, Git, and build tools module

  @ov
  Scenario: Full build and test
    Given I have all dependencies
    When I run the complete build pipeline
    Then all stages should succeed
```

---

## Dependency Discovery

Find all tests requiring specific dependencies:

```bash
# Find all Docker-dependent tests
grep -r "@deps:docker" specs/

# Find all tests requiring specific module
grep -r "@depm:cli-core" specs/
```

---

## Related Documentation

- [Test Levels](./test-levels.md) - Where dependencies are used
- [Tag Inheritance](./tag-inheritance.md) - How dependencies accumulate
- [Test Suites](./test-suites.md) - Dependency handling in test suites
