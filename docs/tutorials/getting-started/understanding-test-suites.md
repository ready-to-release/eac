# Understanding Test Suites

{{ page_breadcrumb() }}

**Status:** Placeholder - Content coming soon

**Estimated time:** 20 minutes
**Prerequisites:** [Your First Module](./first-module.md)

## Planned Content

This tutorial will teach you about the test pyramid, test levels, and how to choose the right test suite for different scenarios.

### What You'll Learn

- Understand the test pyramid (L0 unit → L4 smoke)
- Learn verification types (@ov, @iv, @pv)
- Use the three built-in test suites: commit, acceptance, production-verification
- Know when to use each test level and suite
- Run specific test suites with `r2r test --suite <name>`
- Configure custom test suites in contracts

### Tutorial Structure

1. **The Test Pyramid**
   - L0: Unit tests (fast, isolated, deterministic)
   - L1: Integration tests (component interactions)
   - L2: Component tests (full component behavior)
   - L3: System tests (end-to-end scenarios)
   - L4: Smoke tests (production verification)
   - Why the pyramid shape matters

2. **Verification Types**
   - @ov: Operational Verification (functionality)
   - @iv: Interface Verification (contracts/APIs)
   - @pv: Performance Verification (scalability)
   - @piv: Production Interface Verification
   - When to use each type

3. **Built-in Test Suites**
   - **commit**: L0-L2 tests (fast feedback, pre-commit)
   - **acceptance**: IV/OV/PV tests (PLTE acceptance)
   - **production-verification**: L4+PIV (smoke tests)
   - Trade-offs: speed vs. confidence

4. **Running Test Suites**
   - Default: `r2r test` (runs commit suite)
   - Specific suite: `r2r test --suite acceptance`
   - Specific module: `r2r test eac-commands --suite commit`
   - List suites: `r2r test list-suites`

5. **Understanding Test Results**
   - Test summary: `r2r show test-summary <module>`
   - Test timings: `r2r show test-timings`
   - Debug failures: `r2r test debug`

6. **Custom Test Suites**
   - Define custom suites in contracts
   - Filter by tags and test levels
   - Create scenario-specific suites

### Example Scenarios

The tutorial will walk through practical examples:

- **Pre-commit hook**: Use `commit` suite for fast validation
- **Pull request CI**: Use `acceptance` suite for thorough validation
- **Production deployment**: Use `production-verification` for smoke tests
- **Performance testing**: Use `@pv` tagged tests

### Key Concepts Covered

- Test pyramid and why it matters
- Test levels and their trade-offs
- Verification types for different concerns
- Built-in test suites and when to use them
- Custom test suite configuration

### Best Practices

- Run L0-L2 tests locally (fast feedback)
- Run L3 tests in CI/CD (integration confidence)
- Run L4 tests post-deployment (smoke tests)
- Tag tests appropriately for discoverability
- Keep test suites focused and fast

### Next Steps

You now understand testing fundamentals. Continue to [Core Workflows](../core-workflows/) to learn daily development practices with TDD and specifications.

{{ diataxis_footer() }}
