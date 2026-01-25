# Pre-commit Check Categories

Detailed reference for each pre-commit check category with tool configurations.

---

## 1. Code Formatting

**Purpose**: Enforce consistent code style

**Tools**: `gofmt`, `prettier`, `black`, `clang-format`

**Time**: < 10 seconds

**Auto-fix**: Most formatting tools can auto-fix issues

```bash
# Go
gofmt -w .

# JavaScript/TypeScript
prettier --write .

# Python
black .
```

---

## 2. Linting

**Purpose**: Catch code quality issues, potential bugs, style violations

**Tools**: `golangci-lint`, `eslint`, `pylint`, `rubocop`

**Time**: 10-60 seconds

```bash
# Go (fast mode for quick feedback)
golangci-lint run --fast

# JavaScript
eslint --max-warnings 0 .

# Python
pylint --fail-under=8.0 src/
```

**Scope**: Focus on high-severity issues at Stage 2, defer low-severity to Stage 3.

---

## 3. Unit Tests

**Purpose**: Validate individual units of code in isolation

**Characteristics**:

- No external dependencies (database, network, filesystem)
- Fast (milliseconds per test)
- Deterministic (same input = same output)
- Isolated (tests don't affect each other)

**Time**: 1-5 minutes for full suite

```bash
# Go (run unit tests, skip integration)
go test ./... -short

# JavaScript
npm test -- --testPathIgnorePatterns=integration

# Python
pytest -m "not integration"
```

---

## 4. Secret Detection

**Purpose**: Prevent committing secrets (API keys, passwords, tokens)

**Tools**: `trivy`, `git-secrets`, `truffleHog`, `detect-secrets`

**Time**: 5-30 seconds

```bash
# Trivy
trivy fs --scanners secret .

# Gitleaks
gitleaks detect --source .

# TruffleHog
trufflehog filesystem .
```

**Fail Behavior**: Block commit immediately, require developer to remove secret.

---

## 5. Dependency Vulnerability Scanning

**Purpose**: Detect known vulnerabilities in dependencies

**Tools**: `trivy`, `snyk`, `dependabot`, `npm audit`

**Time**: 10-60 seconds

```bash
# Trivy (critical and high only at Stage 2)
trivy fs --severity CRITICAL,HIGH .

# npm
npm audit --audit-level=high

# Snyk
snyk test --severity-threshold=high
```

---

## 6. Build Verification

**Purpose**: Ensure code compiles successfully

**Time**: 30 seconds - 3 minutes

```bash
# Go
go build ./...

# TypeScript
tsc --noEmit

# Rust
cargo check
```

---

## Execution Environments

### DevBox (Local Execution)

**Git Hooks**:

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

**Make Targets**:

```makefile
.PHONY: precommit
precommit:
    gofmt -w .
    golangci-lint run --fast
    go test ./... -short
    trivy fs --scanners secret .
```

**Task Runners**:

- [pre-commit framework](https://pre-commit.com/)
- [husky](https://typicode.github.io/husky/) (Node.js)
- [lefthook](https://github.com/evilmartians/lefthook) (polyglot)

**Watch Mode**:

```bash
# Continuous testing during development
go test ./... -short -watch
```

### Build Agents (CI Execution)

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

## Error Message Format

Clear, actionable error messages:

```text
Pre-commit checks failed:

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

---

## Related Documentation

- [Pre-commit Setup](./precommit-setup.md) - Hook configuration
- [Pre-commit Quality Gates (Conceptual)](../../../explanation/continuous-delivery/quality-gates/precommit-gates.md) - Philosophy and time budget
- [Test Suites](../testing/test-suites.md) - Test execution commands
