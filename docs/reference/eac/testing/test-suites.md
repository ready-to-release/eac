# Test Suites

> **Unit, integration, acceptance, and production verification suites**

Test suites select tests by tags for execution at specific CD Model stages.

**Note**: All test suites automatically exclude tests tagged with `@ignore`.

---

## unit

**Selects**: `@L0`, `@L1`
**Excludes**: `@L2`, `@L3`, `@L4`, `@ignore`
**Time**: 2-5 minutes
**Purpose**: Fast module-level validation
**Environment**: DevBox or Build Agent
**Run**: `r2r eac test <module> --suite unit`

### What It Tests

- Very fast unit tests with no I/O (L0)
- Fast unit tests with minimal I/O (L1)

### Example

```bash
# Run unit suite
r2r eac test unit

# Runs all scenarios with:
# - @L0 or @L1
# - Excludes @L2, @L3, @L4, @ignore
```

---

## integration

**Selects**: `@L2`
**Excludes**: `@L0`, `@L1`, `@L3`, `@L4`, `@ignore`
**Time**: 5-15 minutes
**Purpose**: Emulated system tests with Docker
**Environment**: Build Agent with Docker
**Run**: `r2r eac test <module> --suite integration`

### What It Tests

- Emulated system tests (L2)
- Docker-based integration tests
- Tests requiring containers or network simulation

### Example

```bash
# Run integration suite
r2r eac test integration

# Runs all scenarios with:
# - @L2
# - Excludes @L0, @L1, @L3, @L4, @ignore
```

---

## acceptance

**Selects**: `@L3`
**Excludes**: `@L0`, `@L1`, `@L2`, `@L4`, `@ignore`
**Time**: 1-2 hours
**Purpose**: Production-like system tests in PLTE
**Environment**: PLTE (Production-Like Test Environment)
**Run**: `r2r eac test <module> --suite acceptance`

### What It Tests

- Installation verification (@iv) - deployment succeeded
- Operational verification (@ov) - features work in production-like environment
- Performance verification (@pv) - meets SLA/SLI/SLO's

### Example

```bash
# Run acceptance suite in PLTE
r2r eac test acceptance

# Runs all scenarios with:
# - @L3
# - Excludes @L0, @L1, @L2, @L4, @ignore
```

---

## production-verification

**Selects**: `@L4` AND `@piv`
**Excludes**: `@ignore`
**Time**: Continuous
**Purpose**: Production smoke tests
**Environment**: Production
**Run**: `r2r eac test <module> --suite production-verification`

### What It Tests

- Production installation verification (@piv)
- Production performance verification (@ppv)
- Continuous monitoring
- Post-deployment validation

### Example

```bash
# Run production verification suite
r2r eac test production-verification

# Runs all scenarios with:
# - @L4 AND @piv
# - Excludes @ignore
```

---

## Test Suite Selection Logic

### unit Suite

```gherkin
@L0 @ov
Scenario: Very fast unit test
  # SELECTED (L0)

@L1 @ov
Scenario: Fast unit test
  # SELECTED (L1)

@L2 @ov
Scenario: Docker-based test
  # NOT SELECTED (L2 excluded)

@ignore @L1 @ov
Scenario: Ignored test
  # NOT SELECTED (@ignore)
```

### integration Suite

```gherkin
@L2 @ov
Scenario: Docker integration test
  # SELECTED (L2)

@L1 @ov
Scenario: Unit test
  # NOT SELECTED (L1 excluded)

@L3 @iv
Scenario: PLTE deployment
  # NOT SELECTED (L3 excluded)

@ignore @L2 @ov
Scenario: Ignored test
  # NOT SELECTED (@ignore)
```

### acceptance Suite

```gherkin
@L3 @ov
Scenario: Production-like functional test
  # SELECTED (@L3)

@L3 @iv
Scenario: Deployment check
  # SELECTED (@L3)

@L2 @ov
Scenario: Emulated test
  # NOT SELECTED (L2 excluded)

@ignore @L3 @ov
Scenario: Ignored test
  # NOT SELECTED (@ignore)
```

### production-verification Suite

```gherkin
@L4 @piv
Scenario: Production smoke test
  # SELECTED (@L4 + @piv)

@L4 @ppv
Scenario: Production monitoring
  # SELECTED (@L4 + @ppv)

@L3 @iv
Scenario: PLTE test
  # NOT SELECTED (L3, not production)

@ignore @L4 @piv
Scenario: Ignored test
  # NOT SELECTED (@ignore)
```

---

## Running Default Suites

```bash
# Run all suites (single init, single summary)
r2r eac test

# Run all suites for a specific module
r2r eac test my-module
```

This runs tests from all three suites while routing output to the module's test folder:

- `out/test/<module>/` - All test results for the module (unit, integration, acceptance)

Benefits:

- Single initialization phase
- Single summary showing all results
- Faster than running three separate commands
- Useful for local development comprehensive testing

!!! note "CI Pipelines"
    CI pipelines typically run suites separately (`--suite unit`, `--suite integration`, etc.) for better failure isolation and parallel job distribution.

---

## CD Model Stage Mapping

| CD Stage                 | Test Suite              | Tags Selected   | Environment    |
| ------------------------ | ----------------------- | --------------- | -------------- |
| **Pre-commit/MR/Commit** | unit                    | `@L0`, `@L1`    | DevBox/Agent   |
| **Integration**          | integration             | `@L2`           | Agent + Docker |
| **Acceptance**           | acceptance              | `@L3`           | PLTE           |
| **Production**           | production-verification | `@L4` + `@piv`  | Production     |

---

## Best Practices

### Test Suite Organization

**DO**:

- Run unit tests before every commit (fast feedback)
- Run integration tests before merging (Docker validation)
- Run acceptance tests in PLTE after deployment
- Run production-verification continuously in production
- Keep unit suite < 5 minutes
- Keep integration suite < 15 minutes

**DON'T**:

- Skip unit tests (catch issues early)
- Skip integration tests (Docker issues caught here)
- Run production tests in PLTE (environment mismatch)
- Run PLTE tests in production (excessive load)
- Include slow tests in commit suite (breaks feedback loop)

### Tag Selection

**DO**:

- Use appropriate test levels (@L0-L4)
- Use verification tags (@ov, @iv, @pv, @piv, @ppv)
- Let test suites select automatically by tags
- Review test distribution across suites

**DON'T**:

- Manually filter test suites (use tags)
- Mix test levels inappropriately
- Forget verification tags (required)
- Over-tag tests (complicates selection)

---

## Custom Test Suites

You can create custom test suites with specific tag combinations:

```bash
# Run all L2 tests with Docker dependency
godog run --tags="@L2 && @deps:docker"

# Run all control tests for AC family
godog run --tags="@control:ac-"

# Run operational tests excluding manual
godog run --tags="@ov && !@Manual"
```

---

## Test Suite Execution Time Guidelines

### commit (Target: < 5 minutes)

- **L0**: < 30 seconds (microseconds per test)
- **L1**: 2-4 minutes (milliseconds per test)
- **Total**: 5 minutes maximum

**If exceeding 5 minutes**:

- Move slow L1 tests to integration suite (change to @L2)
- Optimize test doubles and mocks
- Run tests in parallel
- Review test necessity

### integration (Target: < 15 minutes)

- **L2**: 5-15 minutes (seconds per test)
- **Total**: 15 minutes maximum

**If exceeding 15 minutes**:

- Parallelize Docker container tests
- Optimize container startup
- Review test coverage (remove redundant tests)
- Consider pre-built test containers

### acceptance (Target: 1-2 hours)

- **Installation** (@iv): 10-20 minutes
- **Operational** (@ov): 40-60 minutes
- **Performance** (@pv): 20-40 minutes
- **Total**: 70-120 minutes

**If exceeding 2 hours**:

- Parallelize test execution
- Optimize test data setup
- Review test coverage (remove redundant tests)
- Consider splitting into multiple acceptance environments

### production-verification (Continuous)

- **Smoke tests** (@piv): 2-5 minutes per run
- **Monitoring** (@ppv): Every 5-15 minutes
- **Frequency**: Continuous (24/7)

**If tests are too slow**:

- Simplify smoke tests (only critical paths)
- Reduce monitoring frequency for non-critical checks
- Use read-only operations only

---

## Debugging Test Suite Selection

### Check which tests are selected

```bash
# Dry run - show which tests would run
r2r eac test commit --dry-run
r2r eac test integration --dry-run
r2r eac test acceptance --dry-run
r2r eac test production-verification --dry-run

# Show test count by suite
r2r eac test commit --count
r2r eac test integration --count
r2r eac test acceptance --count
r2r eac test production-verification --count
```

### Common Issues

**Issue**: Test not running in expected suite

**Solution**: Check effective tags (see [Tag Inheritance](../../../explanation/specifications/taxonomy/tag-inheritance.md))

**Issue**: Test running in multiple suites

**Solution**: Review tag combinations - may be intentional or need refinement

**Issue**: Test not running in any suite

**Solution**: Verify test has required tags (test level + verification tag)

---

## Related Documentation

- [Test Levels (Conceptual)](../../../explanation/specifications/taxonomy/test-levels.md) - L0-L4 execution environments
- [Verification Tags (Conceptual)](../../../explanation/specifications/taxonomy/verification-tags.md) - @ov, @iv, @pv, @piv, @ppv
- [Tag Inheritance (Conceptual)](../../../explanation/specifications/taxonomy/tag-inheritance.md) - How tags accumulate
- [Execution Control (Conceptual)](../../../explanation/specifications/taxonomy/execution-control-tags.md) - @ignore and @Manual
- [Test Command Reference](../commands/test/index.md) - Full test command options
