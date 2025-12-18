# Organizing Rules
>
> **Rule blocks and measurable acceptance criteria**

Rules define acceptance criteria - the conditions that must be satisfied for a feature to be accepted.

---

## What Makes a Good Rule?

Rules define **acceptance criteria** - the conditions that must be satisfied for the feature to be accepted by stakeholders.

**Good Rules are**:

- **Measurable**: Objective criteria, not subjective opinions
- **Business-focused**: Expressed in domain language, not technical terms
- **Independent**: Each Rule can be validated separately
- **Granular**: One clear acceptance criterion per Rule

---

## Examples of Good vs Bad Rules

### Bad Rules (subjective, vague)

```gherkin
Rule: The system is user-friendly

Rule: Performance is acceptable

Rule: Error handling works correctly
```

**Problems**:

- Subjective ("user-friendly", "acceptable")
- Not measurable
- Unclear success criteria

### Good Rules (measurable, specific)

```gherkin
Rule: All modules must be shown with their type and path

Rule: Command completes in under 2 seconds for projects with 100 files

Rule: Missing configuration file displays error with path and suggestion
```

**Why these are good**:

- Objective and testable
- Specific metrics when relevant
- Clear pass/fail criteria

---

## How Many Rules Per Feature?

| Rule Count | Assessment | Action |
|-----------|-----------|--------|
| **2-4** | ✅ Ideal | Good feature scope |
| **5-6** | ✅ Acceptable | Monitor complexity |
| **7-10** | ⚠️ Large | Consider splitting |
| **>10** | ❌ Too large | Must split feature |

---

## Why Limit Rules?

### Cognitive Load

More than 6 acceptance criteria becomes hard to reason about. Stakeholders and developers lose track of what the feature actually does.

### Feature Scope

Large Rule count suggests the feature does too much. Features should have a focused purpose.

### Maintainability

Smaller features are easier to:

- Review and understand
- Update when requirements change
- Test thoroughly
- Debug when issues arise

### Testing

Fewer Rules = faster, more focused test execution.

---

## How to Split Large Features

Create focused features by domain concept or user workflow.

### Before (12 Rules - too many)

```text
Feature: project_validation
  Rule: File structure must be valid
  Rule: Configuration must be valid
  Rule: Dependencies must be valid
  Rule: Module boundaries must be respected
  Rule: Naming conventions must be followed
  Rule: Documentation must exist
  Rule: Tests must exist
  Rule: Coverage must meet threshold
  Rule: Performance benchmarks must pass
  Rule: Security checks must pass
  Rule: Linting must pass
  Rule: Type checking must pass
```

**Problems**:

- Too many concerns in one feature
- Hard to understand at a glance
- Long test execution time
- Difficult to maintain

### After (split into 3 features with 4 Rules each)

```text
Feature: project_structure-validation
  Rule: File structure must be valid
  Rule: Configuration must be valid
  Rule: Module boundaries must be respected
  Rule: Naming conventions must be followed

Feature: project_quality-validation
  Rule: Tests must exist with minimum coverage
  Rule: Documentation must exist for public APIs
  Rule: Performance benchmarks must pass
  Rule: Linting must pass

Feature: project_security-validation
  Rule: Dependencies must have no known vulnerabilities
  Rule: Security checks must pass (SAST)
  Rule: Secrets must not be committed
  Rule: Authentication must be enforced
```

**Benefits**:

- Each feature has a clear focus
- Easier to understand and review
- Faster test execution (can run in parallel)
- Better organization and maintainability

---

## Rule Patterns

### Pattern 1: Input Validation

```gherkin
Rule: Email address must be in valid format
Rule: Password must meet complexity requirements
Rule: Username must be unique
```

### Pattern 2: Business Logic

```gherkin
Rule: Order total must include tax for customer's location
Rule: Discount cannot exceed 50% of original price
Rule: Free shipping applies for orders over $50
```

### Pattern 3: System Behavior

```gherkin
Rule: Session expires after 30 minutes of inactivity
Rule: Failed login attempts are logged
Rule: API rate limit is 100 requests per minute
```

### Pattern 4: Error Handling

```gherkin
Rule: Missing configuration file displays helpful error message
Rule: Network timeout retries up to 3 times
Rule: Invalid input returns 400 status with error details
```

---

## Scenarios Per Rule

Each Rule typically needs **2-4 scenarios**:

1. **Happy path** - Primary use case
2. **Error cases** - How system handles failures
3. **Edge cases** - Boundary conditions

**Example**:

```gherkin
Rule: Configuration file must contain valid module definitions

  @ov
  Scenario: Valid configuration with all required fields
    Given I have a configuration with module name and path
    When I validate the configuration
    Then validation should succeed

  @ov
  Scenario: Configuration missing required module name
    Given I have a configuration without a module name
    When I validate the configuration
    Then I should see error "module.name is required"

  @ov
  Scenario: Configuration with invalid module path
    Given I have a configuration with non-existent path
    When I validate the configuration
    Then I should see error "module.path does not exist: src/invalid"
```

---

## Best Practices

### DO

- ✅ Write measurable, objective Rules
- ✅ Use domain language (not technical jargon)
- ✅ Keep 2-6 Rules per feature
- ✅ Make each Rule independently testable
- ✅ Map Rules to Example Mapping blue cards

### DON'T

- ❌ Write subjective Rules ("user-friendly", "performant")
- ❌ Create features with >10 Rules
- ❌ Make Rules dependent on each other
- ❌ Mix multiple concerns in one Rule
- ❌ Use technical implementation details in Rule statements

---

## Rule Writing Checklist

Use this checklist when writing Rules:

- [ ] Is this Rule measurable? (Can I objectively determine pass/fail?)
- [ ] Is this Rule expressed in business language?
- [ ] Is this Rule independent of other Rules?
- [ ] Does this Rule represent one clear acceptance criterion?
- [ ] Can this Rule be tested with 2-4 scenarios?
- [ ] Does this Rule map to a blue card from Example Mapping?

---

## Related Documentation

- [Organizing Scenarios](./organizing-scenarios.md) - How to write scenarios for Rules
- [Size Guidelines](./size-guidelines.md) - Rule and scenario count limits
- [Example Mapping](../discovery/example-mapping.md) - Discovering Rules through workshops
