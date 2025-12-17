# Size Guidelines
> **Rule and scenario count limits**

Guidelines for maintaining readable and maintainable feature files.

---

## Rule Count Per Feature

| Rule Count | Assessment | Action |
|-----------|-----------|--------|
| **2-4** | ✅ Ideal | Good feature scope |
| **5-6** | ✅ Acceptable | Monitor complexity |
| **7-10** | ⚠️ Large | Consider splitting |
| **>10** | ❌ Too large | Must split feature |

### Why Limit Rules?

**Cognitive load**: More than 6 acceptance criteria becomes hard to reason about

**Feature scope**: Large Rule count suggests feature does too much

**Maintainability**: Smaller features are easier to review and update

**Testing**: Fewer Rules = faster, more focused test execution

---

## Scenario Count Per Feature

| Scenario Count | Assessment | Action |
|---------------|-----------|--------|
| **10-15** | ✅ Ideal | Optimal readability |
| **15-20** | ✅ Acceptable | Still manageable |
| **20-30** | ⚠️ Large | Should split |
| **>30** | ❌ Too large | Must split |

### Why Limit Scenarios?

**Readability**: Files over 20 scenarios become hard to navigate

**Execution time**: Large files take longer to run

**Merge conflicts**: Multiple developers editing same large file

**Focus**: Many scenarios suggest feature does too much

---

## Scenarios Per Rule

**Target**: 2-4 scenarios per Rule

Each Rule should have:

1. **Happy path** - Primary use case
2. **Error cases** - How system handles failures
3. **Edge cases** - Boundary conditions, special states

---

## How to Split Large Features

When your feature exceeds recommended limits, use one of these strategies:

### Strategy 1: Split by Rule

If you have 8+ Rules, create separate features:

**Before**: `audit-logging.feature` (40 scenarios, 8 Rules)

**After**:
```text
├── audit-logging_user-actions.feature (15 scenarios, 3 Rules)
├── audit-logging_admin-actions.feature (15 scenarios, 3 Rules)
└── audit-logging_system-events.feature (10 scenarios, 2 Rules)
```

**When to use**: Feature has many distinct acceptance criteria that can be logically grouped.

---

### Strategy 2: Split by Workflow

If scenarios cluster around different user workflows:

**Before**: `user-management.feature` (35 scenarios, 6 Rules)

**After**:
```text
├── user-management_registration.feature (12 scenarios, 2 Rules)
├── user-management_authentication.feature (15 scenarios, 2 Rules)
└── user-management_profile.feature (8 scenarios, 2 Rules)
```

**When to use**: Feature covers multiple user workflows or use cases.

---

### Strategy 3: Split by Scenario Type

If you have many error scenarios:

**Before**: `config-validation.feature` (30 scenarios, 4 Rules)

**After**:
```text
├── config-validation.feature (15 scenarios, 4 Rules - happy paths)
└── config-validation_errors.feature (15 scenarios, 4 Rules - error cases)
```

**When to use**: Feature has extensive error handling or edge case coverage.

---

## Splitting Example

### Before Splitting (Too Large)

```gherkin
Feature: project_validation  # 12 Rules, 40 scenarios

  Rule: File structure must be valid              # 5 scenarios
  Rule: Configuration must be valid               # 4 scenarios
  Rule: Dependencies must be valid                # 3 scenarios
  Rule: Module boundaries must be respected       # 4 scenarios
  Rule: Naming conventions must be followed       # 3 scenarios
  Rule: Documentation must exist                  # 3 scenarios
  Rule: Tests must exist                          # 4 scenarios
  Rule: Coverage must meet threshold              # 3 scenarios
  Rule: Performance benchmarks must pass          # 3 scenarios
  Rule: Security checks must pass                 # 4 scenarios
  Rule: Linting must pass                         # 2 scenarios
  Rule: Type checking must pass                   # 2 scenarios
```

**Problems**:
- 12 Rules (too many acceptance criteria)
- 40 Scenarios (file too long)
- Mixed concerns (structure, quality, security)
- Hard to maintain and understand

### After Splitting (Focused Features)

```gherkin
Feature: project_structure-validation  # 4 Rules, 15 scenarios

  Rule: File structure must be valid              # 5 scenarios
  Rule: Configuration must be valid               # 4 scenarios
  Rule: Module boundaries must be respected       # 4 scenarios
  Rule: Naming conventions must be followed       # 3 scenarios
```

```gherkin
Feature: project_quality-validation  # 4 Rules, 13 scenarios

  Rule: Tests must exist with minimum coverage    # 4 scenarios
  Rule: Documentation must exist for public APIs  # 3 scenarios
  Rule: Performance benchmarks must pass          # 3 scenarios
  Rule: Linting must pass                         # 3 scenarios
```

```gherkin
Feature: project_security-validation  # 4 Rules, 12 scenarios

  Rule: Dependencies must have no vulnerabilities # 3 scenarios
  Rule: Security checks must pass (SAST)          # 4 scenarios
  Rule: Secrets must not be committed             # 3 scenarios
  Rule: Authentication must be enforced           # 2 scenarios
```

**Benefits**:
- Each feature has 4 Rules (ideal range)
- Each feature has 12-15 scenarios (optimal readability)
- Clear separation of concerns
- Easier to maintain and review
- Can be worked on in parallel

---

## Refactoring Steps

When splitting a large feature, follow this procedure:

1. **Identify split points**: Group related Rules/Scenarios together
2. **Create new feature files**: One per group, following naming conventions
3. **Move content**: Cut and paste Rules with their Scenarios to new files
4. **Update feature names**: Follow `module_feature-name` convention
5. **Update tags**: Ensure all scenarios have required verification tags
6. **Update step definitions**: Check for shared steps that need refactoring
7. **Run tests**: Verify nothing broke after the split
8. **Commit together**: Commit all files in a single commit for traceability

### Example Refactoring Workflow

```bash
# 1. Create new feature files
touch specs/project/structure-validation/specification.feature
touch specs/project/quality-validation/specification.feature
touch specs/project/security-validation/specification.feature

# 2. Move Rules and Scenarios to new files
# (edit files manually)

# 3. Run tests to verify
godog ./specs/project/...

# 4. Commit the refactoring
git add specs/project/
git commit -m "refactor: split project_validation into focused features"
```

---

## Size Warning Signs

Watch for these warning signs that a feature needs splitting:

### Rule-Level Warnings

- ⚠️ More than 6 Rules in a feature
- ⚠️ Rules covering multiple distinct concerns
- ⚠️ Rules that aren't closely related
- ⚠️ More than 5 scenarios in a single Rule

### Scenario-Level Warnings

- ⚠️ More than 20 scenarios in a feature
- ⚠️ Scenarios clustered into distinct groups
- ⚠️ File length exceeding 500 lines
- ⚠️ Scenarios requiring different test data setups

### Team Warnings

- ⚠️ Multiple merge conflicts on same feature file
- ⚠️ Feature takes >10 minutes to run all scenarios
- ⚠️ Team members avoid working on the feature (too complex)
- ⚠️ New scenarios feel like they don't "belong"

---

## Best Practices

### DO

- ✅ Monitor Rule and scenario counts regularly
- ✅ Split features before they exceed limits
- ✅ Group related Rules and scenarios together
- ✅ Aim for 2-4 Rules and 10-15 scenarios per feature
- ✅ Review file size during code reviews
- ✅ Use splitting strategies when features grow

### DON'T

- ❌ Let features grow beyond 10 Rules
- ❌ Create features with 30+ scenarios
- ❌ Wait for merge conflicts before splitting
- ❌ Mix unrelated concerns in one feature
- ❌ Add "just one more scenario" repeatedly
- ❌ Ignore warning signs of feature complexity

---

## Size Checklist

Use this checklist to evaluate feature size:

**Rule Count**:
- [ ] Does the feature have 2-6 Rules?
- [ ] Are all Rules closely related?
- [ ] Can each Rule be understood independently?

**Scenario Count**:
- [ ] Does the feature have 10-20 scenarios?
- [ ] Does each Rule have 2-4 scenarios?
- [ ] Is the file under 500 lines?

**Maintainability**:
- [ ] Can a new team member understand the feature quickly?
- [ ] Can scenarios run in under 5 minutes?
- [ ] Do scenarios naturally group together?

If you answer "no" to multiple questions, consider splitting the feature.

---

## Related Documentation

- [Organizing Rules](./organizing-rules.md) - How to create focused Rules
- [Organizing Scenarios](./organizing-scenarios.md) - How to write concise scenarios
- [Naming Conventions](./naming-conventions.md) - How to name split features
