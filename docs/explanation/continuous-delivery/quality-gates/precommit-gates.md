# Pre-commit Quality Gates
## Introduction

Pre-commit quality gates are the first line of defense in the CD Model, executing at **Stage 2** before code is committed to version control. These gates provide immediate feedback to developers while context is fresh, catching issues when they're easiest and cheapest to fix.

The pre-commit stage operates in two environments:

- **DevBox**: Local developer machine (developer-initiated)
- **Build Agents**: CI/CD pipeline (automated on push)

Both environments run the same checks to ensure consistency, but the DevBox execution provides the fastest possible feedback loop.

---

## Why Pre-commit Gates Matter

### The Cost of Late Detection

Consider a simple bug:

- **Caught at Stage 2** (pre-commit, < 5 minutes): Developer fixes immediately, < 1 minute of work
- **Caught at Stage 3** (merge request, hours later): Developer context switches, reviews PR feedback, fixes and re-submits, 10-15 minutes
- **Caught at Stage 5** (acceptance testing, next day): Developer lost context, must investigate test failure, reproduce locally, fix, wait for full pipeline, 30-60 minutes
- **Caught in Production** (days/weeks later): Incident response, customer impact, emergency fix, full investigation, 2-4 hours+

Pre-commit gates keep issues in the "< 1 minute" category.

### The Psychology of Immediate Feedback

When developers run pre-commit checks locally:

- Code context is fully in working memory
- Intent and reasoning are clear
- Changes are fresh and unforgotten
- Fixing feels like completing the task, not interrupting new work

When issues are found later:

- Context switching disrupts flow
- Developer must reconstruct intent
- Feels like rework rather than completion
- Reduces motivation to maintain quality

Pre-commit gates leverage psychology: fix it now while it's easy, or fix it later when it's hard.

### Building Quality Culture

Teams that embrace pre-commit quality gates develop a culture where:

- Quality is part of the definition of "done"
- Developers own their quality (not relying on QA to catch issues)
- Fast feedback loops are valued
- Clean commits are the norm

---

## The 5-10 Minute Rule

### Why This Time Budget Exists

Pre-commit gates have a strict time budget: **5-10 minutes maximum**. This constraint is not arbitrary - it's based on human attention and workflow patterns.

**Developer Workflow:**

1. Write code (10-60 minutes)
2. Run pre-commit checks (5-10 minutes)
3. Fix any issues (1-5 minutes)
4. Commit
5. Move to next task

If step 2 takes 30 minutes instead of 10, developers will:

- Skip running checks locally
- Context switch to other work while waiting
- Lose momentum and focus
- Batch multiple changes before checking

All of these behaviors reduce effectiveness of pre-commit gates.

### Staying Within Budget

**Strategies to meet the 5-10 minute budget:**

**1. Incremental Execution**:

- Only check files that changed
- Use git diff to identify modified files
- Skip checks on unchanged code

**2. Fast Tests Only**:

- Unit tests only (no integration tests)
- No database or network dependencies
- No external service calls
- In-memory test doubles

**3. Caching**:

- Cache dependency installations
- Cache build artifacts
- Cache test results for unchanged code
- Use watch mode during development

**4. Fail Fast**:

- Run fastest checks first (formatting in seconds)
- Stop on first critical failure
- Don't wait for all checks if one already failed

**5. Parallel Execution**:

- Run independent checks concurrently
- Linting, testing, security scans in parallel
- Maximize use of available CPU cores

**6. Selective Depth**:

- Format all code
- Lint changed files
- Test affected code paths
- Scan new dependencies only

### When Budget is Exceeded

If pre-commit checks consistently take > 10 minutes:

- Move slower checks to Stage 3 (CI pipeline)
- Optimize test suite (remove slow tests, improve setup)
- Invest in faster development machines
- Reconsider scope (are you testing too much at Stage 2?)

---

## Check Categories

Pre-commit gates typically include these categories of checks:

### 1. Code Formatting

**Purpose**: Enforce consistent code style

**Tools**: `gofmt`, `prettier`, `black`, `clang-format`

**Time**: < 10 seconds

**Why at Stage 2**: Formatting is fast and deterministic. No reason to commit badly formatted code.

**Auto-fix**: Most formatting tools can auto-fix issues

**Example**: `gofmt -w .` (formats all Go files)

### 2. Linting

**Purpose**: Catch code quality issues, potential bugs, style violations

**Tools**: `golangci-lint`, `eslint`, `pylint`, `rubocop`

**Time**: 10-60 seconds

**Why at Stage 2**: Linting catches common mistakes (unused variables, undefined references, etc.) immediately

**Scope**: Focus on high-severity issues at Stage 2, defer low-severity to Stage 3

**Example**: `golangci-lint run --fast` (fast mode for quick feedback)

### 3. Unit Tests

**Purpose**: Validate individual units of code in isolation

**Characteristics**:

- No external dependencies (database, network, filesystem)
- Fast (milliseconds per test)
- Deterministic (same input = same output)
- Isolated (tests don't affect each other)

**Time**: 1-5 minutes for full suite

**Why at Stage 2**: Unit tests validate logic changes immediately, before code leaves developer's machine

**Scope**: Run all unit tests, or subset affected by changes

**Example**: `go test ./... -short` (run unit tests, skip integration)

### 4. Secret Detection

**Purpose**: Prevent committing secrets (API keys, passwords, tokens)

**Tools**: `trivy`, `git-secrets`, `truffleHog`, `detect-secrets`

**Time**: 5-30 seconds

**Why at Stage 2**: Once committed, secrets enter version history - extremely hard to remove

**Fail Behavior**: Block commit immediately, require developer to remove secret

**Example**: `trivy fs --scanners secret .`

### 5. Dependency Vulnerability Scanning

**Purpose**: Detect known vulnerabilities in dependencies

**Tools**: `trivy`, `snyk`, `dependabot`, `npm audit`

**Time**: 10-60 seconds

**Why at Stage 2**: Catch critical vulnerabilities before they enter codebase

**Scope**: Critical and high severity only at Stage 2, defer medium/low to Stage 3

**Example**: `trivy fs --severity CRITICAL,HIGH .`

### 6. Build Verification

**Purpose**: Ensure code compiles successfully

**Time**: 30 seconds - 3 minutes

**Why at Stage 2**: No point committing code that doesn't compile

**Scope**: Build changed modules only, not entire project

**Example**: `go build ./...` (compile all packages)

---

## Execution Environments

Pre-commit gates run in two distinct environments with different characteristics:

### DevBox (Local Execution)

**Characteristics**:

- Developer-controlled environment
- Full access to local tools and editors
- Fast iteration (no network latency)
- Developer-initiated (manual or git hook)

**Advantages**:

- Fastest feedback (no CI queue time)
- Developer can immediately fix issues
- Context fully in memory
- Incremental execution during development

**Implementation Options**:

**1. Git Hooks**:

```bash
# .git/hooks/pre-commit
#!/bin/bash
set -e

echo "Running pre-commit checks..."
gofmt -w .
golangci-lint run --fast
go test ./... -short
trivy fs --scanners secret .
```

**2. Make Targets**:

```makefile
.PHONY: precommit
precommit:
    gofmt -w .
    golangci-lint run --fast
    go test ./... -short
    trivy fs --scanners secret .
```

**3. Task Runners**:

- [pre-commit framework](https://pre-commit.com/)
- [husky](https://typicode.github.io/husky/) (Node.js)
- [lefthook](https://github.com/evilmartians/lefthook) (polyglot)

**4. Watch Mode**:

```bash
# Continuous testing during development
go test ./... -short -watch
```

### Build Agents (CI Execution)

**Characteristics**:

- Automated execution on git push
- Clean, consistent environment
- Isolated from developer machine
- Results tracked and reported

**Advantages**:

- Catches issues developers might skip locally
- Consistent across all developers
- Creates audit trail
- Enforces quality even if developer bypasses local checks

**Why Repeat in CI**:
Even though DevBox runs checks, CI repeats them because:

- Developers might skip local checks
- Environment differences (OS, tool versions)
- Ensures consistency for all code entering repository
- Creates formal quality gate record

**Implementation**:

```yaml
# .github/workflows/precommit.yml
name: Pre-commit
on: [push]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Format check
        run: gofmt -l .
      - name: Lint
        run: golangci-lint run --fast
      - name: Unit tests
        run: go test ./... -short -cover
      - name: Secret detection
        run: trivy fs --scanners secret .
```

---

## Implementation Strategies

### Progressive Adoption

Don't implement all pre-commit checks at once. Gradual rollout prevents overwhelming developers:

**Week 1**: Code formatting only

- Auto-fix on save in IDE
- Git hook enforces formatting
- Zero friction (auto-fix handles everything)

**Week 2**: Add linting

- Start with error-level only
- Fix existing violations first
- Add to git hooks

**Week 3**: Add unit tests

- Require tests pass before commit
- Optimize test suite for speed
- Add coverage reporting

**Week 4**: Add security scans

- Secret detection (critical)
- Dependency vulnerabilities (high severity only)

### Making it Easy

**IDE Integration**:

- Format on save
- Show linting errors inline
- Run tests in IDE with one click
- Display test results immediately

**Git Hooks with Bypass**:

```bash
# Allow bypass in emergencies
git commit --no-verify
```

**Clear Error Messages**:

```text
❌ Pre-commit checks failed:

[Formatting] 3 files need formatting:
  - src/main.go
  - src/handler.go
  - src/service.go

Run 'gofmt -w .' to fix automatically.

[Linting] 2 issues found:
  src/main.go:42: unused variable 'result'
  src/handler.go:18: error return not checked

[Tests] 1 test failed:
  TestUserService_CreateUser: expected error, got nil

Fix these issues and try again.
```

### Continuous Improvement

**Monitor Time Budget**:

```bash
# Track how long pre-commit takes
time make precommit
```

**Developer Feedback**:

- Survey: "Do pre-commit checks help or hinder?"
- Measure: Are developers bypassing checks?
- Iterate: Adjust scope based on feedback

**Optimize Regularly**:

- Profile slow tests, optimize or move to Stage 3
- Update tool versions (often faster)
- Review scope (are we checking too much?)

---

## Anti-Patterns

### Anti-Pattern 1: Integration Tests at Stage 2

**Problem**: Running integration tests that require database, external services

**Impact**: Slow feedback (> 10 minutes), breaks time budget

**Solution**: Move integration tests to Stage 3 or Stage 5

### Anti-Pattern 2: No Local Execution

**Problem**: Only running checks in CI, not locally

**Impact**: Developers wait for CI feedback (10-20 minutes), slower iteration

**Solution**: Implement git hooks, make local execution easy

### Anti-Pattern 3: Checks That Can't Be Auto-Fixed

**Problem**: Requiring manual fixes for things machines can fix (formatting)

**Impact**: Wastes developer time, creates friction

**Solution**: Auto-fix everything possible, only flag issues that need human judgment

### Anti-Pattern 4: Overly Strict Linting

**Problem**: Failing on low-severity style preferences

**Impact**: Developers frustrated, bypass checks

**Solution**: Focus on high-severity issues at Stage 2, defer style preferences to Stage 3

### Anti-Pattern 5: No Bypass Mechanism

**Problem**: Forcing checks even in emergencies

**Impact**: Developers find workarounds, checks lose credibility

**Solution**: Allow `--no-verify` bypass, but track usage and investigate frequent bypasses

---

## Best Practices Summary

1. **Keep it fast**: 5-10 minutes maximum, optimize relentlessly
2. **Fail fast**: Run fastest checks first, stop on critical failures
3. **Auto-fix**: Let machines fix formatting, linting (when possible)
4. **Clear feedback**: Show exactly what failed and how to fix
5. **Progressive adoption**: Start simple, add checks gradually
6. **Make it easy**: IDE integration, git hooks, one-command execution
7. **Run locally**: Don't rely solely on CI
8. **Allow bypass**: Emergency escape hatch with `--no-verify`
9. **Monitor and improve**: Track time, gather feedback, optimize
10. **Trust but verify**: Run same checks in CI to catch bypassed issues

---

## Next Steps

- [Merge Request Quality Gates](./merge-request-gates.md) - Stage 3 validation
- [CD Model Stages 1-7](../cd-model/cd-model-stages-1-7.md) - See Stage 2 in full context
- [Testing Strategy](../testing/testing-strategy-overview.md) - How unit tests fit in overall strategy
