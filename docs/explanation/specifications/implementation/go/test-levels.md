# Test Levels with Go Build Tags

> **Build tags and test isolation (L0-L4)**

Go build tags control which tests run based on isolation level, enabling fast feedback loops and appropriate test scope.

---

## Test Level Overview

| Level | Build Tag        | Purpose                      | Speed        |
| ----- | ---------------- | ---------------------------- | ------------ |
| L0    | `//go:build L0`  | Fully isolated, zero I/O     | Microseconds |
| L1    | (none - default) | Unit tests with minimal deps | Milliseconds |
| L2    | `//go:build L2`  | Integration with containers  | Seconds      |
| L3    | Gherkin `@L3`    | Pre-production (PLTE)        | Minutes      |
| L4    | Gherkin `@L4`    | Production verification      | Minutes+     |

---

## Why Test Levels?

Test levels enable:

- **Fast feedback loops** - L0/L1 run in milliseconds
- **Appropriate isolation** - Each level has clear boundaries
- **Selective execution** - Run only what's needed
- **CI optimization** - Different stages run different levels

---

## Build Tag to Gherkin Tag Mapping

| Go Build Tag    | Gherkin Tag | Auto-inferred from |
| --------------- | ----------- | ------------------ |
| `//go:build L0` | `@L0`       | N/A                |
| (none)          | `@L1`       | `@ov` (default)    |
| `//go:build L2` | `@L2`       | (explicit only)    |
| N/A (Godog)     | `@L3`       | `@iv`, `@pv`       |
| N/A (Godog)     | `@L4`       | `@piv`, `@ppv`     |

---

## Reference Documentation

For complete code examples, execution commands, and TDD workflow, see:

**[Test Levels Reference](../../../../reference/eac/testing/go/test-levels.md)** - Complete guide including:

- L0-L4 code examples
- Test execution commands
- Gherkin scenario examples
- Canon TDD workflow in Go

---

## Related Documentation

- [File Organization](./file-organization.md) - Directory structure and file naming
- [Testing Taxonomy](../../taxonomy/) - Complete tag reference
- [Three-Layer Approach](../../concepts/three-layer-approach.md) - Conceptual overview
