# Quality & Maintenance
> **Maintaining healthy, living specifications**

Learn how to keep specifications synchronized with implementation through continuous review and iteration.

---

## Overview

This section covers:

- **Review and Iterate** - Continuous refinement practices for living specifications
- **Quality Checklist** - Health indicators and red flags for specification quality
- **Maintenance Practices** - Regular review ceremonies and processes

Specifications aren't written once - they evolve with understanding, implementation, and production feedback.

---

## In This Section

| Topic | Description |
|-------|-------------|
| [Review and Iterate](./review-and-iterate.md) | Continuous specification refinement practices |
| [Spec Quality Checklist](./spec-quality-checklist.md) | Health indicators and quality assessment |

---

## Key Principles

### 1. Specifications are Living Documents

```text
Discovery → Specification → Implementation → Review → Iterate
                ↑_______________________________________↓
                          Feedback Loop
```

Specifications improve through:

- **Discovery**: Example Mapping workshops reveal requirements
- **Implementation**: Edge cases emerge during development
- **Production**: Bugs expose missing scenarios
- **Feedback**: Team reviews identify improvements

### 2. Review Cadence

**Weekly (Active Development)**:

- Review new/changed scenarios
- Discuss ambiguous language
- Identify missing coverage
- Refactor verbose scenarios

**Monthly (Maintenance)**:

- Check scenario relevance
- Verify language matches code
- Consolidate similar scenarios
- Update evolved terminology

**Quarterly (Comprehensive)**:

- Review all specs in module
- Split files >20 scenarios
- Update Ubiquitous Language
- Remove outdated scenarios

### 3. Specification Health Indicators

**Red Flags 🚩:**

- Scenarios unchanged >6 months while code evolved
- All scenarios always passing (testing nothing)
- Testing implementation details ("database updated")
- \>30 scenarios in single file
- Technical jargon ("POST to /api/users")

**Green Indicators ✅:**

- Specs committed with code changes
- Scenarios catch regression bugs
- Business can read and validate
- Clear traceability spec → step → code
- Consistent Ubiquitous Language

---

## Review Triggers

### When to Review Specifications

1. **After Implementation** - Did spec match reality?
2. **When Tests Fail** - Is spec wrong or code wrong?
3. **Requirements Change** - New regulations, processes, features
4. **Regular Cadence** - Weekly/Monthly/Quarterly reviews
5. **Before Extensions** - Refresh understanding before building on top
6. **Bugs/Defects** - Is specification incomplete? Missing corner cases?

---

## Quick Start

### Running Your First Review

**Step 1: Gather specs to review**:

```bash
# Find recently changed specs
git log --oneline --since="1 week ago" -- specs/

# Find large files (>500 lines)
find specs/ -name "*.feature" -exec wc -l {} \; | awk '$1 > 500'

# Find stale specs (>6 months unchanged)
find specs/ -name "*.feature" -mtime +180
```

**Step 2: Check for red flags**:

For each spec file:

- [ ] Scenarios have meaningful assertions
- [ ] Language matches domain glossary
- [ ] No implementation details exposed
- [ ] Scenarios test behavior, not code
- [ ] File size is manageable (<30 scenarios)

**Step 3: Validate with stakeholders**:

- Can Product Owner read and understand?
- Do scenarios match expected behavior?
- Are edge cases covered?
- Is language consistent with business terms?

**Step 4: Document actions**:

For each issue found:

1. Create task or ticket
2. Assign owner
3. Set deadline
4. Track to completion

---

## Common Quality Issues

### Specification Debt

**Symptoms:**

- Scenarios always pass but don't verify behavior
- Ambiguous steps with multiple interpretations
- Gap between spec language and code reality
- Scenarios test implementation, not behavior
- Files unchanged while code evolved

**Solutions:**

- Add meaningful assertions to scenarios
- Refine ambiguous steps with concrete examples
- Update language to match domain terms
- Focus scenarios on observable behavior
- Review and update specs with code changes

### Over-Specification

**Symptoms:**

- \>30 scenarios in single file
- Testing every possible input combination
- Duplicated logic across scenarios
- Scenarios coupling to implementation details

**Solutions:**

- Split large files into focused features
- Use Scenario Outlines for input variations
- Extract common steps to reduce duplication
- Focus on business-valuable test cases

### Under-Specification

**Symptoms:**

- Missing error case scenarios
- Happy path only testing
- No edge case coverage
- Gaps in Rule coverage

**Solutions:**

- Run Example Mapping to discover edge cases
- Add at least one error scenario per Rule
- Use boundary value analysis
- Review with testers for missing cases

---

## Refactoring Specifications

### When to Refactor

- File >20 scenarios
- Multiple distinct concerns in one file
- Duplicate logic patterns
- Language evolved but specs haven't
- Scenarios test implementation, not behavior

### How to Refactor

1. **Identify split points** - Group related Rules/Scenarios
2. **Create new feature files** - One per cohesive group
3. **Move content** - Cut and paste Rules with Scenarios
4. **Update feature names** - Follow `module_feature-name` convention
5. **Update tags** - Ensure all scenarios have verification tags
6. **Update step definitions** - Check for shared steps
7. **Run tests** - Verify nothing broke
8. **Commit together** - Single commit for traceability

**Example:**

Split `validation.feature` (40 scenarios) into:

- `format-validation.feature` (10 scenarios)
- `completeness-validation.feature` (10 scenarios)
- `business-rule-validation.feature` (10 scenarios)
- `edge-case-validation.feature` (10 scenarios)

---

## Review Ceremonies

### Weekly Specification Review

**Duration**: 30 minutes
**Attendees**: Developers, QA, Product Owner

| Time | Activity |
|------|----------|
| 10 min | Review new/changed scenarios |
| 10 min | Discuss ambiguous language |
| 5 min | Identify missing coverage |
| 5 min | Refactor verbose scenarios |

### Three Amigos (Before New Features)

**Duration**: 45-60 minutes
**Attendees**: Business, Development, Testing

1. Review existing related specs (15 min)
2. Identify needed updates (10 min)
3. Mini Example Mapping (20 min)
4. Update specifications together (15 min)

---

## Feedback Loops

### From Implementation to Specification

| Discovery | Action |
|-----------|--------|
| Missing acceptance criteria | Add new `Rule:` block |
| Ambiguous steps | Refine with concrete examples |
| Edge cases | Add `@ov` scenarios for error cases |
| Wrong assumptions | Revise preconditions |
| Incomplete verification | Add `Then`/`And` steps |

### From Production to Specification

**Bug-Driven Process:**

1. Write scenario that would have caught bug
2. Verify scenario fails (regression test)
3. Fix code until passes
4. Keep scenario in suite

---

## Automation and Tools

### Health Checks

```bash
# Find large specification files
find specs/ -name "*.feature" -exec wc -l {} \; | awk '$1 > 500 {print}'

# Find old specs (unchanged >6 months)
find specs/ -name "*.feature" -mtime +180

# Count scenarios per feature
grep -r "Scenario:" specs/ | cut -d: -f1 | uniq -c | sort -rn
```

### Validation

```bash
# Validate Gherkin syntax
r2r eac validate specs

# Check control tags
r2r eac validate control-tags

# Verify test tags
r2r eac validate test-tags
```

---

## Best Practices

### DO

- ✅ Review specifications regularly (weekly/monthly/quarterly)
- ✅ Commit spec changes with code changes
- ✅ Refactor immediately when issues found
- ✅ Involve whole team in specification health
- ✅ Treat specs as living, evolving documentation

### DON'T

- ❌ "We'll clean up specs later" - They never get cleaned
- ❌ Write once and forget
- ❌ Let only developers maintain specs
- ❌ Never remove old scenarios
- ❌ Skip regular reviews

---

## Related Documentation

### Discovery

- [Example Mapping](../discovery/example-mapping.md) - Discovering requirements
- [Event Storming](../discovery/event-storming-overview.md) - Domain discovery

### Core Concepts

- [BDD Fundamentals](../concepts/bdd-fundamentals.md) - BDD principles
- [Specifications Evolution](../concepts/specifications-evolution.md) - How specs change
- [Ubiquitous Language](../concepts/ubiquitous-language.md) - Shared vocabulary

### Organization

- [Size Guidelines](../organization/size-guidelines.md) - Rule and scenario limits
- [File Structure](../organization/file-structure.md) - Organizing specifications
