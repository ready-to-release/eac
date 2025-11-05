# Testing Strategy

Comprehensive testing strategy integrated throughout all stages of the Continuous Delivery Model. These articles explain the test taxonomy, how test levels map to stages, and the shift-left approach that enables fast, reliable delivery.

---

## [Testing Strategy Overview](testing-strategy-overview.md)

Understanding test taxonomy and the shift-left testing approach.

Testing is integrated throughout every stage of the CD Model, not treated as a separate phase after development. This article establishes the complete test taxonomy and explains why multiple test levels are necessary.

**The Test Taxonomy:**

**L0-L2: Local/Agent Tests**:

- **Execution environment**: Developer workstation or CI agent
- **Test scope**: Unit-under-test only
- **External dependencies**: All replaced with test doubles
- **Trade-off**: High determinism, low coherency with domain language
- **Stages**: 2, 3, 4
- **Purpose**: Fast, reliable validation of logic in isolation

**L3: In-Situ Vertical Tests**:

- **Execution environment**: PLTE (cloud/production-like environment)
- **Test scope**: Single deployable unit boundaries only (vertical testing)
- **External dependencies**: All replaced with test doubles
- **Trade-off**: Moderate determinism, moderate coherency
- **Stages**: 5, 6
- **Purpose**: Validate deployable unit behavior in production-like infrastructure

**L4: Production Horizontal Tests**:

- **Execution environment**: Production
- **Test scope**: Cross-service interactions (horizontal testing)
- **External dependencies**: May use test doubles (e.g., test-double payment service)
- **Trade-off**: Lower determinism, high coherency with domain language
- **Stages**: 11, 12
- **Purpose**: Validate real-world cross-service behavior in production

**Topics covered:**

- Test taxonomy based on execution environment and scope
- Detailed explanation of each category (L0-L2, L3, L4)
- Determinism vs coherency trade-off
- Out-of-category anti-pattern (horizontal pre-production environments)
- Shift-left and shift-right strategy
- Tools and frameworks for each level

**Read this article to understand**: The test taxonomy based on execution constraints and how to avoid the anti-pattern of fragile horizontal pre-production environments.

---

## [Testing Integration with CD Model](testing-strategy-integration.md)

How test levels integrate with CD Model stages.

This article maps test levels to specific stages, explains process isolation strategies, and shows how testing integrates with ATDD/BDD/TDD methodologies.

**Test Level Environment Mapping:**

**L0-L2 Tests:**

- Execute in DevBox (Stage 2) and Build Agents (Stages 3-4)
- Unit-under-test only with test doubles for all external dependencies
- Fast, deterministic feedback (5-30 minutes)

**L3 Tests:**

- Execute in PLTE (Stages 5-6)
- Single deployable unit boundaries (vertical testing)
- Test doubles for all external services
- Validates infrastructure and deployment

**L4 Tests:**

- Execute in Production (Stages 11-12)
- Cross-service interactions (horizontal testing)
- Real production behavior with optional test doubles
- Synthetic monitoring and exploratory testing

**Topics covered:**

- Test level environment mapping
- Process isolation explained (in-process vs cross-process vs full system)
- Test levels by CD Model stage (detailed mapping)
- Time-boxing per stage
- Integration with ATDD/BDD/TDD methodologies
- Test pyramid in practice
- Anti-patterns to avoid (ice cream cone, hourglass)

**Read this article to understand**: Where and when to execute each test level, and how to structure your testing strategy.

---

## The Shift-Left and Shift-Right Strategy

**Traditional Approach (Anti-Pattern):**

- Horizontal pre-production integration environments
- Multiple teams' pre-prod services linked together
- Highly fragile and non-deterministic
- Slow feedback, difficult debugging

**Shift-Left and Shift-Right Approach:**

- **Shift LEFT (L0-L3)**: Fast, deterministic tests with test doubles on local/CI/PLTE
- **Shift RIGHT (L4)**: Real production validation with monitoring and feature flags
- **Avoid the middle**: Skip horizontal pre-production environments
- Fast feedback (L0-L3) + real validation (L4)

**Cost of Finding Defects by Stage:**

| Stage | Find Defect | Relative Cost |
|-------|-------------|---------------|
| Stage 2 (Pre-commit) | Minutes | 1x |
| Stage 4 (Commit) | 30 min | 5x |
| Stage 6 (Extended) | Hours | 10x |
| Stage 11 (Live) | Days | 100x |

**Key Principle**: Test as early as possible, as fast as possible, as much as possible at each level.

---

## Test Pyramid

The test pyramid guides the quantity of tests at each level:

```text
        L4 (few)
      /           \
    L3 (some)
   /               \
  L2 (more)
 /                   \
L1 (many)
L0 (most) ----------
```

**Recommended Distribution:**

- **L0-L2**: 95% of tests (hundreds to thousands) - Fast, deterministic validation
- **L3**: 5% of tests (5-20 critical vertical scenarios) - Infrastructure validation in PLTE
- **L4**: Continuous (synthetic monitoring + exploratory) - Production horizontal validation

**Why This Shape:**

✅ **Fast Feedback**: Most tests (L0) run fastest (milliseconds)
✅ **Reliability**: Lower levels are more stable (fewer flaky tests)
✅ **Cost Efficiency**: Cheaper to write and maintain unit tests
✅ **Comprehensive Coverage**: Each level serves a different purpose

---

## Integration with ATDD/BDD/TDD

The testing strategy integrates three methodologies:

**TDD (Test-Driven Development) → L0-L2:**

- Focus: Unit and integration tests with test doubles
- Purpose: Drive design and validate logic in isolation
- Process: Red → Green → Refactor
- Tools: Go test, Jest, pytest

**BDD (Behavior-Driven Development) → L3:**

- Focus: Vertical tests in PLTE
- Purpose: Validate deployable unit behavior in cloud infrastructure
- Process: Gherkin scenarios → Step implementations with test doubles
- Tools: Godog, Cucumber, SpecFlow

**ATDD (Acceptance Test-Driven Development) → L4:**

- Focus: Production horizontal tests
- Purpose: Validate acceptance criteria in production
- Process: Define criteria → Implement → Validate in production
- Tools: Synthetic monitoring, exploratory testing, production observability

See **[Three-Layer Testing Approach](../../specifications/three-layer-approach.md)** for detailed integration.

---

## Testing by Stage

| Stage | Test Levels | Time Budget | Quality Gates |
|-------|-------------|-------------|---------------|
| **1. Authoring** | Manual validation | N/A | Developer judgment |
| **2. Pre-commit** | L0-L2 | 5-10 min | 100% pass, coverage ≥ threshold |
| **3. Merge Request** | L0-L2 | 15-30 min | 100% pass, peer approval |
| **4. Commit** | L0-L2 | 15-30 min | 100% pass, artifacts built |
| **5. Acceptance** | L3 (vertical) | 1-2 hours | IV, OV, PV validated |
| **6. Extended** | L3 + perf/sec | 2-8 hours | Comprehensive validation |
| **7. Exploration** | Manual prep for L4 | Days | Scenarios defined |
| **8-10. Release** | Regression subset | Minutes | No critical failures |
| **11-12. Live** | L4 (horizontal) | Continuous | Synthetic monitoring, SLA adherence |

---

## Integration with Other Sections

**[Core Concepts](../core-concepts/index.md)**:

- Deployable Units are validated through all test levels
- Unit of Flow includes testing at each stage

**[CD Model](../cd-model/index.md)**:

- Test levels execute at specific stages (1-12)
- Quality gates prevent progression until tests pass
- Evidence collection (IV, OV, PV) in Stage 5

**[Workflow](../workflow/index.md)**:

- Pre-commit runs L0/L1 before push (Stage 2)
- Merge requests require L0-L2 passing (Stage 3)
- Trunk commits trigger comprehensive suite (Stage 4)

**[Architecture](../architecture/index.md)**:

- Environments determine which test levels can execute
- PLTE required for L3 tests
- Demo environment for L4 tests

**[Security](../security/index.md)**:

- Security testing integrated at multiple levels
- SAST in L0/L1, DAST in L3
- Dependency scanning throughout

---

## Best Practices

**For All Test Levels:**

✅ **DO:**

- Follow the test pyramid distribution
- Time-box stages to enforce pyramid
- Write tests before or alongside code
- Keep tests independent and deterministic
- Run lower levels in parallel
- Fail fast (stop on first failure for rapid feedback)

❌ **DON'T:**

- Create ice cream cone (too many E2E tests)
- Skip levels (creates gaps in coverage)
- Let tests become flaky (fix or remove)
- Test implementation details (test behavior)
- Duplicate coverage across levels

**Specific Practices:**

**L0-L2**: Mock all external dependencies, run on local/CI agent in < 10 minutes total
**L3**: Limit to 5-20 critical vertical scenarios, test doubles for external services, validate infrastructure
**L4**: Continuous synthetic monitoring, exploratory testing in production, use feature flags for control

---

## Next Steps

- **New to testing strategy?** Start with [Testing Strategy Overview](testing-strategy-overview.md)
- **Need stage mapping?** Read [Testing Integration with CD Model](testing-strategy-integration.md)
- **Want to understand stages?** See [CD Model](../cd-model/index.md)
- **Need environments?** Explore [Architecture](../architecture/index.md)

Return to **[Continuous Delivery Overview](../index.md)** for complete navigation.
