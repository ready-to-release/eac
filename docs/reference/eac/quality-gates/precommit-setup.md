# Pre-commit Setup

Configuration guide for pre-commit hooks and validation tools.

## Git Hook Script

Create `.git/hooks/pre-commit`:

```bash
#!/bin/sh

# Format check
go fmt ./...
if [ $? -ne 0 ]; then
    echo "Format check failed"
    exit 1
fi

# Lint
golangci-lint run --fast
if [ $? -ne 0 ]; then
    echo "Lint check failed"
    exit 1
fi

# Unit tests
go test -short ./...
if [ $? -ne 0 ]; then
    echo "Unit tests failed"
    exit 1
fi

# Security scan
trivy fs --severity HIGH,CRITICAL .
if [ $? -ne 0 ]; then
    echo "Security scan failed"
    exit 1
fi

echo "All pre-commit checks passed"
exit 0
```

Make executable:

```bash
chmod +x .git/hooks/pre-commit
```

---

## Pre-commit Framework

Install [pre-commit](https://pre-commit.com/):

```bash
pip install pre-commit
```

Create `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/golangci/golangci-lint
    rev: v1.54.0
    hooks:
      - id: golangci-lint

  - repo: local
    hooks:
      - id: go-test
        name: Go Tests
        entry: go test -short ./...
        language: system
        pass_filenames: false

      - id: trivy-scan
        name: Security Scan
        entry: trivy fs --severity HIGH,CRITICAL .
        language: system
        pass_filenames: false
```

Install hooks:

```bash
pre-commit install
```

---

## Time Budget Optimization

**Target**: 5-10 minutes maximum

| Strategy             | Description                        |
| -------------------- | ---------------------------------- |
| Incremental scanning | Only scan changed files            |
| Local caching        | Reuse results from previous runs   |
| Fail fast            | Stop on first critical failure     |
| Parallel execution   | Run independent checks in parallel |

---

## Tool Reference

| Check        | Tool                      | Purpose              |
| ------------ | ------------------------- | -------------------- |
| Format       | `go fmt`, `prettier`      | Code style           |
| Lint         | `golangci-lint`, `eslint` | Code quality         |
| Unit tests   | `go test -short`          | Fast tests only      |
| Secrets      | `trivy`, `gitleaks`       | Credential detection |
| Dependencies | `trivy fs`                | Vulnerability scan   |
| Build        | `go build`                | Compilation check    |

---

## Skipping Hooks (Emergency Only)

```bash
git commit --no-verify -m "emergency fix"
```

**Warning**: Only use for genuine emergencies. CI will still run all checks.

---

## Related Documentation

- [Pre-commit Checks](./precommit-checks.md) - Detailed check categories
- [Pre-commit Quality Gates (Conceptual)](../../../explanation/continuous-delivery/quality-gates/precommit-gates.md) - Why pre-commit matters
- [Security Scanning](../security/index.md) - Security scan commands
