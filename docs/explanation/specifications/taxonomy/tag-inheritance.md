# Tag Inheritance

> **How tags accumulate and override**

Tags accumulate from Feature → Rule → Scenario levels with specific override rules.

---

## Accumulation Rules

Tags flow down from Feature → Rule → Scenario:

```gherkin
@L2 @deps:docker
Feature: Container Tests

  @ov
  Rule: Container operations

    Scenario: Start container
      # Effective tags: @L2, @deps:docker, @ov
```

---

## Override Rules

### Test Level Tags (`@L0`-`@L4`)

- Scenario level overrides feature level
- Allows mixing test levels within a feature

**Example**:

```gherkin
@L2
Feature: Mixed-Level Tests
  # Feature says L2 (emulated system)

  @ov
  Scenario: Fast emulated test
    # Uses L2 from feature
    # Effective tags: @L2, @ov

  @L3 @iv
  Scenario: Deployment test in PLTE
    # Overrides to L3 (PLTE environment)
    # Effective tags: @L3, @iv
```

### Dependencies (`@deps:*`)

- Accumulate (additive)
- Scenario inherits all feature dependencies
- Can add more dependencies at scenario level

**Example**:

```gherkin
@deps:docker
Feature: Container Tests

  @deps:git
  Rule: Container builds

    @ov
    Scenario: Build from repository
      # Effective dependencies: @deps:docker, @deps:git
```

### Verification Tags (`@ov`, `@iv`, etc.)

- Accumulate (additive)
- Scenario can add additional verification types
- Multiple verification tags can coexist

**Example**:

```gherkin
@ov
Feature: Deployment Validation

  Rule: Service deployment

    @iv
    Scenario: Deploy and validate
      # Effective tags: @ov, @iv
      # Tests both operational and installation aspects
```

---

## Effective Tag Calculation

Tags are calculated in this order:

1. **Feature level tags** - Base tags for all scenarios
2. **Rule level tags** - Add or override feature tags
3. **Scenario level tags** - Add or override rule tags

---

## Complete Example

```gherkin
@L2 @deps:docker @ov
Feature: auth_api-service
  API service authentication tests

  @control:ac-2
  Rule: Account management

    @ov
    Scenario: Create account
      # Effective tags:
      # - @L2 (from feature)
      # - @deps:docker (from feature)
      # - @ov (from feature + scenario)
      # - @control:ac-2 (from rule)

    @L3 @iv @deps:kubectl
    Scenario: Deploy to PLTE
      # Effective tags:
      # - @L3 (scenario overrides feature L2)
      # - @deps:docker (from feature)
      # - @deps:kubectl (scenario adds)
      # - @iv (from scenario)
      # - @control:ac-2 (from rule)
```

---

## Tag Override Priority

From highest to lowest priority:

1. **Scenario tags** - Most specific, highest priority
2. **Rule tags** - Override feature, but not scenario
3. **Feature tags** - Base level, lowest priority

---

## When Tags Override vs Accumulate

| Tag Type                                   | Behavior   | Reason                                  |
| ------------------------------------------ | ---------- | --------------------------------------- |
| Test Level (`@L0`-`@L4`)                   | Override   | Only one execution environment per test |
| Verification (`@ov`, `@iv`, etc.)          | Accumulate | Tests can verify multiple aspects       |
| Dependencies (`@deps:`, `@depm:`, `@env:`) | Accumulate | Tests can need multiple tools           |
| Control (`@control:`)                      | Accumulate | Tests can satisfy multiple controls     |
| Execution Control (`@ignore`, `@Manual`)   | Accumulate | Multiple execution modifiers allowed    |

---

## Best Practices

### Feature-Level Tags

Use for tags that apply to **all scenarios**:

```gherkin
@L2 @deps:docker @ov
Feature: Container Operations
  # All scenarios need Docker
  # All scenarios are L2 operational tests
```

### Rule-Level Tags

Use for tags that apply to **scenarios in that rule**:

```gherkin
Feature: Security Tests

  @control:ac-2
  Rule: Access control
    # All scenarios in this rule relate to AC-2

  @control:au-3
  Rule: Audit logging
    # All scenarios in this rule relate to AU-3
```

### Scenario-Level Tags

Use for tags **specific to that scenario**:

```gherkin
@L2
Feature: Mixed Tests

  @ov
  Rule: Test operations

    @ov
    Scenario: Normal test
      # Uses L2 from feature

    @L3 @iv @deps:kubectl
    Scenario: Deployment test
      # Overrides to L3
      # Adds kubectl dependency
      # Changes to installation verification
```

---

## Common Patterns

### Pattern 1: Common Level, Different Verifications

```gherkin
@L2 @deps:docker
Feature: API Tests

  @ov
  Scenario: Functional test
    # @L2, @deps:docker, @ov

  @pv
  Scenario: Performance test
    # @L2, @deps:docker, @pv
```

### Pattern 2: Different Levels, Same Verification

```gherkin
@ov
Feature: Auth Service

  @L2
  Scenario: Unit test with mocks
    # @L2, @ov

  @L3 @env:plte
  Scenario: Integration test in PLTE
    # @L3, @ov, @env:plte
```

### Pattern 3: Progressive Dependencies

```gherkin
@deps:go
Feature: Build Pipeline

  @deps:git
  Rule: Source builds
    # Needs: go, git

    @deps:docker
    Scenario: Container build
      # Needs: go, git, docker
```

---

## Debugging Tag Inheritance

### Viewing Effective Tags

Use the validation command to see effective tags:

```bash
# Show effective tags for all scenarios
r2r eac validate specs --show-effective-tags

# Example output:
# Feature: auth_api-service
#   Scenario: Create account
#     Effective Tags: @L2, @deps:docker, @ov, @control:ac-2
```

### Common Issues

| Issue                                  | Solution                                                                         |
| -------------------------------------- | -------------------------------------------------------------------------------- |
| Scenario not running in expected suite | Check if feature-level test level tag is inherited when you expected an override |
| Missing dependencies in test execution | Verify feature-level dependencies are declared, or add them at scenario level    |

---

## Reference Documentation

For validation commands, see:

**[Specifications Reference](../../../reference/eac/specifications/index.md)** - Commands including:

- `r2r eac validate specs` - Validate Gherkin syntax
- `r2r eac validate specs --show-effective-tags` - Show calculated tags

---

## Related Documentation

- [Test Levels](./test-levels.md) - Understanding test level tags
- [Verification Tags](./verification-tags.md) - Types of validation tags
- [Dependency Tags](./dependency-tags.md) - System and module dependencies
- [Test Suites](./test-suites.md) - How tags are used in test selection
