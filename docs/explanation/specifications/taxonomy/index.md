# Testing Taxonomy

Complete reference for the testing taxonomy and tag system.

---

## Overview

The testing taxonomy defines:

- **Test Levels** (L0-L4): Execution environment and scope
- **Verification Tags** (@ov, @iv, @pv, @piv, @ppv): Type of validation
- **Execution Control** (@skip, @Manual): Test execution behavior
- **Dependencies** (@deps, @depm, @env): Required tooling and environments
- **Controls** (@control): Compliance requirements

---

## In This Section

| Topic                                            | Description                                              |
| ------------------------------------------------ | -------------------------------------------------------- |
| [Test Levels](./test-levels.md)                  | L0-L4 execution environments                             |
| [Verification Tags](./verification-tags.md)      | Operational, installation, performance verification      |
| [Execution Control](./execution-control-tags.md) | Skipping and manual tests                                |
| [Dependency Tags](./dependency-tags.md)          | System, module, and environment dependencies             |
| [Control Tags](./control-tags.md)                | Risk and compliance controls                             |
| [Tag Inheritance](./tag-inheritance.md)          | How tags accumulate and override                         |
| [Test Suites](./test-suites.md)                  | Commit, integration, acceptance, production-verification |

---

## Quick Reference

All tags use lowercase. Required tags for scenarios: verification tag (@ov, @iv, etc.)

### Tag Categories

| Category            | Tags                                     | Required |
| ------------------- | ---------------------------------------- | -------- |
| Test Level          | `@L0`, `@L1`, `@L2`, `@L3`, `@L4`        | No       |
| Verification        | `@ov`, `@iv`, `@pv`, `@piv`, `@ppv`      | Yes      |
| Execution Control   | `@ignore`, `@Manual`                     | No       |
| System Dependencies | `@deps:docker`, `@deps:git`, `@deps:go`  | No       |
| Module Dependencies | `@depm:<module>`                         | No       |
| Risk Controls       | `@control:<id>`, `@controls:<id1>,<id2>` | No       |

---

## Tag Inheritance

Tags accumulate from Feature → Rule → Scenario levels.

```gherkin
@L2 @deps:docker
Feature: Container Tests

  @ov
  Rule: Container operations

    Scenario: Start container
      # Effective tags: @L2, @deps:docker, @ov
```

---

## Related Documentation

- [Three-Layer Approach](../concepts/three-layer-approach.md) - Conceptual overview
- [Organizing Specifications](../organization/index.md) - File structure and naming
