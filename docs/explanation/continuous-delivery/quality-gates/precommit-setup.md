# Pre-commit Setup

How to configure pre-commit hooks for Stage 2 validation.

## Pre-commit Hook Script

Create `.git/hooks/pre-commit` or use a pre-commit framework:

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

## Make Hook Executable

```bash
chmod +x .git/hooks/pre-commit
```

## Using Pre-commit Framework

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

## Time Budget

**Target**: 5-10 minutes maximum

Optimization strategies:

- **Incremental scanning**: Only scan changed files
- **Local caching**: Reuse results from previous runs
- **Fail fast**: Stop on first critical failure
- **Parallel execution**: Run independent checks in parallel

## Checks to Include

| Check        | Tool                      | Purpose              |
| ------------ | ------------------------- | -------------------- |
| Format       | `go fmt`, `prettier`      | Code style           |
| Lint         | `golangci-lint`, `eslint` | Code quality         |
| Unit tests   | `go test -short`          | Fast tests only      |
| Secrets      | `trivy`, `gitleaks`       | Credential detection |
| Dependencies | `trivy fs`                | Vulnerability scan   |
| Build        | `go build`                | Compilation check    |

## Skipping Hooks (Emergency Only)

```bash
git commit --no-verify -m "emergency fix"
```

**Warning**: Only use for genuine emergencies. CI will still run all checks.

## Related

- [Pre-commit Quality Gates](./precommit-gates.md)
- [CD Model Stages 1-6](../cd-model/cd-model-stages-1-6.md)

---

_[Tutorials](../../../tutorials/) | [How-to Guides](../../../how-to-guides/) | **Explanation** | [Reference](../../../reference/)_

**You are here:** Explanation — understanding-oriented discussion that clarifies concepts.
