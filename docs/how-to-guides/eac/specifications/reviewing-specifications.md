# Reviewing Specifications

How to conduct specification reviews and maintain spec health.

## Review Ceremonies

### Weekly Specification Review

**Duration:** 30 minutes
**Attendees:** Developers, QA, Product Owner

| Time   | Activity                     |
| ------ | ---------------------------- |
| 10 min | Review new/changed scenarios |
| 10 min | Discuss ambiguous language   |
| 5 min  | Identify missing coverage    |
| 5 min  | Refactor verbose scenarios   |

### Three Amigos Session

**Duration:** 45-60 minutes
**Attendees:** Business, Development, Testing

| Time   | Activity                       |
| ------ | ------------------------------ |
| 15 min | Review existing related specs  |
| 10 min | Identify needed updates        |
| 20 min | Mini Example Mapping           |
| 15 min | Update specifications together |

## Review Process

### Step 1: Gather Specs to Review

```bash
# Find recently changed specs
git log --oneline --since="1 week ago" -- specs/

# Find large files (>500 lines)
find specs/ -name "*.feature" -exec wc -l {} \; | awk '$1 > 500'

# Find stale specs (>6 months unchanged)
find specs/ -name "*.feature" -mtime +180
```

### Step 2: Check for Red Flags

For each spec file, verify:

- [ ] Scenarios have meaningful assertions
- [ ] Language matches domain glossary
- [ ] No implementation details exposed
- [ ] Scenarios test behavior, not code
- [ ] File size is manageable (<30 scenarios)

### Step 3: Validate with Stakeholders

- Can Product Owner read and understand?
- Do scenarios match expected behavior?
- Are edge cases covered?
- Is language consistent with business terms?

### Step 4: Document Actions

For each issue found:

1. Create task or ticket
2. Assign owner
3. Set deadline
4. Track to completion

## After Example Mapping

### Immediate (Same Day)

- [ ] Write `specification.feature` with Rules and Scenarios
- [ ] Document Ubiquitous Language terms
- [ ] Create `issues.md` for pink cards
- [ ] Share with stakeholders for review

### Short-term (1-2 Days)

- [ ] Incorporate stakeholder feedback
- [ ] Refine ambiguous scenarios
- [ ] Add missing edge cases
- [ ] Resolve pink card questions

### During Implementation (1 Week)

- [ ] Implement step definitions
- [ ] Discover edge cases via unit testing → Add scenarios
- [ ] Refine vague steps → Update spec
- [ ] Keep synchronized - commit spec with code

### After Implementation

- [ ] Retrospective: Did spec match build?
- [ ] Refactor for clarity
- [ ] Remove redundant scenarios
- [ ] Document lessons learned

## Handling Specification Changes

### Process

1. Update scenario in `specification.feature`
2. Update step definitions in `src/*/tests/`
3. Run tests - verify passing
4. Update production code if behavior changed
5. Commit spec and code together

### Breaking Changes

1. Tag old scenario `@deprecated`
2. Add new scenario
3. Implement new behavior
4. Remove deprecated scenario
5. Deploy

## Feedback Integration

### From Implementation

| Discovery                   | Action                        |
| --------------------------- | ----------------------------- |
| Missing acceptance criteria | Add new `Rule:` block         |
| Ambiguous steps             | Refine with concrete examples |
| Edge cases                  | Add `@ov` scenarios           |
| Wrong assumptions           | Revise preconditions          |
| Incomplete verification     | Add `Then`/`And` steps        |

### From Production (Bug-Driven)

1. Write scenario that would have caught bug
2. Verify scenario fails (regression test)
3. Fix code until passes
4. Keep scenario in suite

## Specification Refactoring

### When to Refactor

- File >20 scenarios
- Multiple distinct concerns in one file
- Duplicate logic patterns
- Language evolved but specs haven't
- Scenarios test implementation, not behavior

### How to Refactor

1. Identify cohesive sub-features
2. Create focused `.feature` files (10-15 scenarios each)
3. Move related scenarios with Rules and @ac tags
4. Delete old monolithic file
5. Update step definitions if needed

## Related

- [Spec Quality Checklist](../../../reference/specifications/spec-quality-checklist.md)
- [Review and Iterate](../../../explanation/specifications/review-and-iterate.md)
- [Splitting Large Features](splitting-large-features.md)
