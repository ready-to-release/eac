# Merge Request Quality Gates

## Introduction

Merge request quality gates operate at **Stage 3** of the CD Model, serving as the primary approval checkpoint before code merges to the main branch. Stage 3 combines **automated validation** (repeated pre-commit checks plus integration tests) with **human judgment** (peer review).

This stage is unique because it's the only stage in the CD Model that **always** requires manual human approval - the peer reviewer's sign-off. In the RA pattern, Stage 3 is the first-level approval (code quality). In the CDe pattern, Stage 3 becomes a combined first and second-level approval (code quality AND production approval).

---

## The Dual Nature of Stage 3

### Automated Validation

**What runs automatically**:

- All pre-commit checks (formatting, linting, unit tests, security)
- Integration tests (service-to-service interactions)
- Code coverage analysis
- Static analysis security testing (SAST)
- Dependency vulnerability scanning
- Container image scanning (if applicable)
- Documentation build verification

**Why repeat pre-commit checks**:
Even though Stage 2 already validated these, Stage 3 repeats them because:

- Developers might bypass local checks
- Environment consistency (clean CI environment)
- Creates formal record for audit trail
- Multiple commits might have been pushed
- Base branch might have changed since local check

**Parallel execution**:
While the peer reviewer examines the code, automated checks run in parallel. By the time review is complete, automated validation results are available.

### Human Judgment

**What humans evaluate**:

- Design quality and architecture alignment
- Code clarity and maintainability
- Security implications requiring context
- Business logic correctness
- Test coverage adequacy (not just percentage)
- Documentation completeness
- API contract compatibility
- Error handling appropriateness

**Why humans are essential**:
Automation catches mechanical issues (syntax, style, test failures). Humans catch:

- Subtle logic bugs
- Design smells (over-engineering, tight coupling)
- Security vulnerabilities requiring domain knowledge
- Missing edge cases in tests
- Unclear naming or confusing structure
- Inappropriate abstractions

**The balance**:
Automated checks enforce minimum standards. Human review ensures code is not just correct, but maintainable, secure, and well-designed.

---

## Peer Review Philosophy

### What Makes a Good Review

**Focus on substance over style**:

- ❌ "This variable should be named `userData` not `data`" (linter should catch)
- ✅ "This function has three responsibilities - should we split it?"

**Ask questions, don't demand**:

- ❌ "Change this to use Strategy pattern"
- ✅ "Have you considered the Strategy pattern here? It might make testing easier"

**Understand intent before critiquing**:

- Ask "Why did you approach it this way?" before suggesting alternatives
- Context might justify an unusual approach

**Be specific and actionable**:

- ❌ "This is confusing"
- ✅ "The variable name `x` doesn't indicate it's a timeout duration. Consider `requestTimeout`?"

### Review Scope

**What to review carefully**:

1. **Public APIs**: Changes to interfaces, contracts, exported functions
2. **Security-sensitive code**: Authentication, authorization, data validation, encryption
3. **Error handling**: Are errors properly wrapped, logged, and propagated?
4. **Test coverage**: Are critical paths tested? Are tests meaningful?
5. **Performance implications**: Loops in loops, N+1 queries, unbounded operations

**What to skim**:

1. **Mechanical changes**: Formatting, imports, minor refactoring
2. **Boilerplate**: Standard CRUD operations, config files
3. **Generated code**: Unless generator changed

**What to trust automation for**:

1. **Code style**: Linters enforce this
2. **Test execution**: CI confirms tests pass
3. **Coverage percentage**: Coverage tools report this
4. **Known vulnerabilities**: Security scanners detect these

### The "Two Approvals" Rule

Many teams require **two approving reviews** for sensitive changes:

- Changes to authentication/authorization
- Database schema migrations
- API contract changes
- Security-critical code
- Infrastructure changes

This provides additional safety for high-risk changes.

---

## Automated Checks

Stage 3 runs comprehensive automated validation in the CI/CD pipeline (Build Agents environment).

### Pre-commit Checks (Repeated)

**Why repeat**:

- Enforce for all developers (even those who bypass locally)
- Clean environment (no local configuration drift)
- Formal record for compliance/audit

**What's repeated**:

- Code formatting verification
- Linting (full, not fast mode)
- Unit tests with coverage
- Secret detection
- Dependency vulnerability scanning
- Build verification

### Integration Tests

**New at Stage 3**: Integration tests validate interactions between components.

**Characteristics**:

- Test service-to-service communication
- May use database (test instance)
- May use message queues, caches
- Run in isolated test environment
- Slower than unit tests (seconds to minutes)

**Examples**:

- API endpoint tests (HTTP requests/responses)
- Database query integration
- Message queue pub/sub
- Cache integration
- External service integration (with test doubles)

**Why not at Stage 2**:
Integration tests are too slow for pre-commit (violates 5-10 minute budget) and require infrastructure (database, services) not available on developer laptops.

### Code Coverage Analysis

**What's measured**:

- Line coverage: % of code lines executed by tests
- Branch coverage: % of conditional branches tested
- Function coverage: % of functions called by tests

**Typical thresholds**:

- Minimum: 80% line coverage
- Goal: 90%+ line coverage
- Critical code: 100% coverage (authentication, payment, security)

**Why at Stage 3**:
Coverage analysis requires full test suite execution. Results inform peer reviewer: "Are critical paths tested?"

**Coverage is necessary but not sufficient**:

- 100% coverage with poor assertions = false confidence
- Reviewers must examine test quality, not just percentage

### Static Application Security Testing (SAST)

**Tools**: `semgrep`, `gosec`, `bandit`, `eslint-plugin-security`

**What's detected**:

- SQL injection vulnerabilities
- XSS vulnerabilities
- Insecure deserialization
- Hardcoded secrets (redundant with Stage 2, but deeper scan)
- Insecure cryptography usage
- Command injection risks

**Why at Stage 3**:
SAST tools can be slow (minutes) and noisy (many false positives). Stage 3 is appropriate for comprehensive analysis.

**Handling false positives**:

- Review and suppress with justification
- Document why flagged code is safe
- Track suppressions for periodic review

### Dependency Scanning

**Deeper than Stage 2**:

- All severity levels (not just critical/high)
- Transitive dependencies
- License compliance
- Deprecated packages

**Actions**:

- **Critical/High**: Block merge
- **Medium**: Warning, allow merge with plan to address
- **Low**: Informational only

### Documentation Verification

**What's checked**:

- Documentation builds without errors
- API documentation updated (OpenAPI, Swagger)
- README reflects new features
- Changelog entry added
- Inline code documentation for public APIs

**Why at Stage 3**:
Documentation is part of the deliverable. Don't merge incomplete work.

---

## RA vs CDe Pattern Differences

Stage 3 approval has profoundly different meanings in the two CD Model patterns:

### Release Approval (RA) Pattern

**Stage 3 = First-Level Approval Only**

When a peer reviewer approves in RA pattern, they're saying:

- ✅ "This code is well-designed and correct"
- ✅ "Tests adequately cover the changes"
- ✅ "Security has been considered"
- ❌ NOT saying: "This is ready for production"

**Production approval happens later**:

- Stage 9 (Release Approval): Release manager evaluates production readiness
- Separate decision: code quality vs. production readiness

**Approval scope**:

- Code quality
- Design and architecture
- Test coverage
- Security considerations

**Typical cycle**:

- Code merged at Stage 3
- Validation in Stages 5-6
- **Manual approval at Stage 9** before production

### Continuous Deployment (CDe) Pattern

**Stage 3 = Combined First and Second-Level Approval**

When a peer reviewer approves in CDe pattern, they're saying:

- ✅ "This code is well-designed and correct"
- ✅ "Tests adequately cover the changes"
- ✅ "Security has been considered"
- ✅ **"This is ready for production deployment"**

**Production approval happens now**:

- By merging, the reviewer approves production deployment
- Stage 9 becomes automated (no human approval)

**Approval scope**:

- Everything from RA pattern, PLUS
- Production deployment approval
- Business risk assessment
- Release timing consideration (is now a good time to deploy?)

**Typical cycle**:

- Code merged at Stage 3 = production approved
- Automated validation in Stages 5-6
- **Automated progression to production** (2-4 hours)

**Why this works**:

- Comprehensive automated testing (Stages 5-6) provides confidence
- Feature flags decouple deployment from release (Stage 12)
- Team trusts automated validation
- Rollback is fast and automated

**Requirements for CDe pattern**:

1. **Robust automated testing** (unit, integration, acceptance, performance, security)
2. **Feature flags** for runtime control of feature exposure
3. **Automated rollback** on threshold breaches
4. **Team culture** that trusts automation and peer review

---

## Code Coverage Thresholds

### Why Coverage Matters

Code coverage indicates how much of your code is executed by tests. High coverage reduces (but doesn't eliminate) the risk of undetected bugs.

**Benefits**:

- Confidence that logic is tested
- Catch regressions (changes that break existing functionality)
- Force developers to consider test cases
- Create safety net for refactoring

**Limitations**:

- Coverage ≠ quality (can have 100% coverage with meaningless assertions)
- Some code is hard to test (UI, infrastructure)
- Chasing 100% can lead to diminishing returns

### Setting Thresholds

**Typical thresholds**:

- **80%**: Minimum for merge approval (catches untested major paths)
- **90%**: Goal for mature codebases
- **100%**: Only for critical code (payment, security, compliance)

**Why 80%**:

- Balances thoroughness with pragmatism
- Catches major gaps without perfection paralysis
- Achievable without excessive effort

**Why not 100%**:

- Diminishing returns (last 5% is hardest)
- Some code is not worth testing (trivial getters/setters)
- Over-focus on percentage misses test quality

### Coverage Requirements by Code Type

**Critical code (100% required)**:

- Authentication and authorization
- Payment processing
- Encryption and security
- Data validation at trust boundaries

**Business logic (90%+ required)**:

- Core domain logic
- Calculation engines
- Workflow state machines
- Complex algorithms

**Standard code (80%+ required)**:

- API handlers
- Service layer
- Repository layer
- Utility functions

**Infrastructure code (70%+ acceptable)**:

- Configuration loading
- Logging setup
- Framework boilerplate
- Generated code

### Measuring Quality Beyond Coverage

**Look for**:

- **Assertion quality**: Do tests actually check behavior?
- **Edge case coverage**: Null, empty, boundary conditions tested?
- **Error path coverage**: Failures tested, not just happy paths?
- **Test isolation**: Tests don't depend on each other?

**Red flags**:

- Tests with no assertions
- Tests that only check "no exception thrown"
- Tests that test implementation, not behavior
- Flaky tests (sometimes pass, sometimes fail)

---

## Implementation Strategies

### Merge Request Template

Provide a template that guides submitter and reviewer:

```markdown
## Changes
<!-- Describe what changed and why -->

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
<!-- Describe testing performed -->
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review performed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] Tests added for new functionality
- [ ] All tests passing
- [ ] No new warnings

## Security Considerations
<!-- Any security implications? -->

## Performance Impact
<!-- Any performance considerations? -->
```

### Automated Review Comments

Use bots to provide automated feedback:

- Coverage decreased: "Warning: Coverage dropped from 85% to 82%"
- Large PR: "This PR has 800+ lines changed. Consider splitting."
- Missing tests: "No test files modified. Did you add tests?"
- Breaking change: "This modifies a public API. Update changelog?"

### Review Time Budget

**Small PRs (< 200 lines)**: 15-30 minutes
**Medium PRs (200-500 lines)**: 30-60 minutes
**Large PRs (> 500 lines)**: Consider splitting or schedule dedicated review time

**Why size matters**:

- Large PRs are harder to review thoroughly
- Reviewers rush through, missing issues
- Better: smaller, focused PRs reviewed carefully

### Review Rotation

**Distribute review load**:

- Rotate reviewers to spread knowledge
- Avoid bottlenecks (one person reviewing everything)
- Build team expertise

**Specialize when needed**:

- Security changes → security experts
- Database migrations → database experts
- Performance-critical code → performance experts

---

## Anti-Patterns

### Anti-Pattern 1: Rubber-Stamp Reviews

**Problem**: Approving without actually reviewing ("LGTM" in 30 seconds on 500-line PR)

**Impact**: Defeats purpose of peer review, quality issues reach production

**Solution**: Set expectations for review depth, track review time, require meaningful feedback

### Anti-Pattern 2: Nitpicking Style

**Problem**: Focusing on style preferences instead of substance

**Impact**: Wastes time, frustrates developers, misses real issues

**Solution**: Automate style checks (linters, formatters), focus review on design and logic

### Anti-Pattern 3: Review Bottlenecks

**Problem**: One person reviews everything, becomes bottleneck

**Impact**: PRs wait days for review, slows entire team

**Solution**: Distribute reviews, rotate reviewers, set SLA (review within 24 hours)

### Anti-Pattern 4: Massive PRs

**Problem**: 2000-line changes in one PR

**Impact**: Impossible to review thoroughly, issues slip through

**Solution**: Encourage smaller, focused PRs (< 500 lines), break features into increments

### Anti-Pattern 5: Ignoring Automated Checks

**Problem**: Approving despite failing tests or security scans

**Impact**: Quality gates bypassed, problems reach main branch

**Solution**: Block approval until automated checks pass, investigate failures

### Anti-Pattern 6: Review After Merge

**Problem**: Merging without review, "reviewing" afterwards

**Impact**: Defeats purpose of pre-merge review, main branch polluted

**Solution**: Enforce branch protection (require approval before merge)

---

## Best Practices Summary

1. **Automated checks in parallel**: Run while human reviews
2. **Focus review on substance**: Design, logic, security - not style
3. **Small, focused PRs**: < 500 lines when possible
4. **Review time budget**: 15-60 minutes based on PR size
5. **Meaningful feedback**: Specific, actionable comments
6. **Trust automation**: Let tools handle mechanical issues
7. **Coverage + quality**: Check percentage AND test quality
8. **RA vs CDe awareness**: Understand what approval means in your pattern
9. **Distribute reviews**: Avoid bottlenecks, spread knowledge
10. **Enforce gates**: Don't approve if automated checks fail

---

## Next Steps

- [Pre-commit Quality Gates](./precommit-gates.md) - Stage 2 validation
- [Release Quality Gates](./release-gates.md) - Stage 9 validation
- [CD Model Stages 1-6](../cd-model/cd-model-stages-1-6.md) - See Stage 3 in full context
- [Implementation Patterns](../cd-model/implementation-patterns.md) - RA vs CDe approval differences

{{ diataxis_footer() }}