# Go Implementation Guide

{{ page_breadcrumb() }}

> **Implementation-specific guide for Go/Godog BDD and testing**

Complete guide for implementing BDD specifications with Go and Godog.

## In This Section

| Topic | Description |
|-------|-------------|
| [Overview](./overview.md) | Introduction to Go/Godog BDD testing |
| [File Organization](./file-organization.md) | Directory structure and file naming |
| [Test Levels](./test-levels.md) | Build tags and test isolation (L0-L4) |
| [Step Definitions](./step-definitions.md) | Writing and organizing step definitions |
| [Best Practices](./best-practices.md) | Testing patterns and conventions |

## Quick Reference

- Go version: ≥ 1.21
- Framework: Godog for BDD, Go test for unit tests
- Build tags: `//go:build L0` through `//go:build L4`

## Essential Commands

```bash
# Unit tests
go test ./...                    # L0 + L1
go test -tags=L0 ./...          # L0 only
go test -tags=L2 ./...          # L2 integration

# BDD scenarios
godog run                        # All scenarios
godog run --tags=@ov            # Operational verification
godog run --tags=@cli           # CLI features only

# Coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Related Documentation

### Conceptual Understanding

- [Three-Layer Testing Approach](../../concepts/three-layer-approach.md) - Conceptual overview
- [BDD Fundamentals](../../concepts/bdd-fundamentals.md) - BDD fundamentals
- [Testing Taxonomy](../../taxonomy/) - Tag taxonomy concepts

### Organizational

- [Organizing Specifications](../../organization/) - Specification structure
- [Example Mapping](../../discovery/example-mapping.md) - Requirements discovery

{{ diataxis_footer() }}
