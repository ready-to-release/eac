# Test Suites

{{ page_breadcrumb() }}

> **Pre-commit, acceptance, and production verification suites**

Test suites select tests by tags for execution at specific CD Model stages.

**Note**: All test suites automatically exclude tests tagged with `@ignore`.

---

## pre-commit

**Selects**: `@L0`, `@L1`, `@L2`
**Excludes**: `@ignore`
**Time**: 5-10 minutes
**Purpose**: Fast pre-commit validation
**Environment**: DevBox or Build Agent
**Run**: `r2r eac test pre-commit`

### What It Tests

- Fast unit tests (L0)
- Unit tests with minimal I/O (L1)
- Emulated system tests (L2)

### Example

```bash
# Run pre-commit suite
r2r eac test pre-commit

# Runs all scenarios with:
# - @L0, @L1, or @L2
# - Excludes @ignore
```

---

## acceptance

**Selects**: `@iv`, `@ov`, `@pv`
**Excludes**: `@ignore`
**Infers**: `@L3` from `@iv` and `@pv`
**Time**: 1-2 hours
**Purpose**: PLTE deployment validation
**Environment**: PLTE (Production-Like Test Environment)
**Run**: `r2r eac test acceptance`

### What It Tests

- Installation verification (@iv) - deployment succeeded
- Operational verification (@ov) - features work
- Performance verification (@pv) - meets SLA

### Example

```bash
# Run acceptance suite in PLTE
r2r eac test acceptance

# Runs all scenarios with:
# - @iv, @ov, or @pv
# - Infers @L3 for @iv and @pv
# - Excludes @ignore
```

---

## production-verification

**Selects**: `@L4` AND `@piv`
**Excludes**: `@ignore`
**Time**: Continuous
**Purpose**: Production smoke tests
**Environment**: Production
**Run**: `r2r eac test production-verification`

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
# - @L4 AND (@piv OR @ppv)
# - Excludes @ignore
```

---

## Test Suite Selection Logic

### pre-commit Suite

```gherkin
@L0 @ov
Scenario: Fast unit test
  # ✅ SELECTED (L0)

@L1 @ov
Scenario: Unit test
  # ✅ SELECTED (L1)

@L2 @ov
Scenario: Emulated test
  # ✅ SELECTED (L2)

@L3 @iv
Scenario: PLTE deployment
  # ❌ NOT SELECTED (L3)

@ignore @L1 @ov
Scenario: Ignored test
  # ❌ NOT SELECTED (@ignore)
```

### acceptance Suite

```gherkin
@ov
Scenario: Functional test
  # ✅ SELECTED (@ov)

@iv
Scenario: Deployment check
  # ✅ SELECTED (@iv, infers @L3)

@pv
Scenario: Performance test
  # ✅ SELECTED (@pv, infers @L3)

@L2 @ov
Scenario: Emulated test
  # ❌ NOT SELECTED (explicit @L2, not @L3)

@ignore @ov
Scenario: Ignored test
  # ❌ NOT SELECTED (@ignore)
```

### production-verification Suite

```gherkin
@L4 @piv
Scenario: Production smoke test
  # ✅ SELECTED (@L4 + @piv)

@L4 @ppv
Scenario: Production monitoring
  # ✅ SELECTED (@L4 + @ppv)

@piv
Scenario: Missing L4
  # ✅ SELECTED (@piv infers @L4)

@L3 @iv
Scenario: PLTE test
  # ❌ NOT SELECTED (L3, not production)

@ignore @L4 @piv
Scenario: Ignored test
  # ❌ NOT SELECTED (@ignore)
```

---

## CD Model Stage Mapping

| CD Stage | Test Suite | Tags Selected | Environment |
|----------|-----------|---------------|-------------|
| **Build** | pre-commit | `@L0`, `@L1`, `@L2` | DevBox/Agent |
| **Acceptance** | acceptance | `@iv`, `@ov`, `@pv` | PLTE |
| **Production** | production-verification | `@L4` + `@piv` | Production |

---

## Best Practices

### Test Suite Organization

✅ **DO**:

- Run pre-commit before every commit
- Run acceptance in PLTE after deployment
- Run production-verification continuously in production
- Ensure fast feedback (keep pre-commit < 10 minutes)
- Balance test coverage across suites

❌ **DON'T**:

- Skip pre-commit tests (catch issues early)
- Run production tests in PLTE (environment mismatch)
- Run PLTE tests in production (excessive load)
- Include slow tests in pre-commit (breaks feedback loop)

### Tag Selection

✅ **DO**:

- Use appropriate verification tags (@ov, @iv, @pv, @piv, @ppv)
- Set correct test levels (@L0-L4)
- Let test suites select automatically by tags
- Review test distribution across suites

❌ **DON'T**:

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

### pre-commit (Target: < 10 minutes)

- **L0**: < 1 minute (microseconds per test)
- **L1**: 2-3 minutes (milliseconds per test)
- **L2**: 5-7 minutes (seconds per test)
- **Total**: 8-10 minutes maximum

**If exceeding 10 minutes**:
- Move slow L2 tests to acceptance suite (change to @L3)
- Optimize test doubles and mocks
- Run tests in parallel
- Review test necessity

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
r2r eac test pre-commit --dry-run

# Show test count by suite
r2r eac test pre-commit --count
r2r eac test acceptance --count
r2r eac test production-verification --count
```

### Common Issues

**Issue**: Test not running in expected suite

**Solution**: Check effective tags (see [Tag Inheritance](./tag-inheritance.md))

**Issue**: Test running in multiple suites

**Solution**: Review tag combinations - may be intentional or need refinement

**Issue**: Test not running in any suite

**Solution**: Verify test has required tags (test level + verification tag)

---

## Related Documentation

- [Test Levels](./test-levels.md) - L0-L4 execution environments
- [Verification Tags](./verification-tags.md) - @ov, @iv, @pv, @piv, @ppv
- [Tag Inheritance](./tag-inheritance.md) - How tags accumulate
- [Execution Control](./execution-control-tags.md) - @ignore and @Manual

{{ diataxis_footer() }}
