# Testing Strategy

Comprehensive testing strategy integrated throughout all stages of the Continuous Delivery Model.

---

## Overview

The testing strategy uses a taxonomy of test levels (L0-L4) based on execution environment and scope, enabling a shift-left/shift-right approach that avoids fragile horizontal pre-production environments.

**Test Levels:**

- **L0-L1**: Unit Tests (DevBox/Agent, Stages 2-4)
- **L2**: Emulated System Tests (DevBox/Agent, Stages 3-4)
- **L3**: In-Situ Vertical Tests (PLTE, Stages 5-6)
- **L4**: Testing in Production (Production, Stages 11-12)

**Tag Usage**: See **[Tag Reference](../../specifications/tag-reference.md)** for test level tags, verification tags, system dependencies, and test suites.

---

## Available Articles

## [Testing Strategy Overview](testing-strategy-overview.md)

Complete test taxonomy and shift-left/shift-right strategy.

**Topics**: Test levels (L0-L4), determinism vs domain coherency, Horizontal E2E anti-pattern, shift-left/shift-right strategy, test pyramid, ATDD/BDD/TDD integration.

---

## [Testing Integration with CD Model](testing-strategy-integration.md)

How test levels map to CD Model stages and integrate with ATDD/BDD/TDD.

**Topics**: Test level environment mapping, process isolation (in-process, cross-process, in-situ), stage-by-stage mapping, time-boxing, quality gates.

---

## Quick Reference

**Stage Mapping:**

- **Stages 2-4**: L0-L2 (pre-commit suite, 5-30 min)
- **Stages 5-6**: L3 (acceptance suite, 1-8 hours)
- **Stages 11-12**: L4 (production-verification, continuous)

**Test Distribution** (Test Pyramid):

- **L0-L1**: 70-80% (hundreds to thousands)
- **L2**: 15-20% (dozens to hundreds)
- **L3**: 5% (5-20 critical scenarios)
- **L4**: Continuous (synthetic monitoring + exploratory)

**Key Principle**: Test as early as possible (shift-left L0-L3), as fast as possible, then validate in production (shift-right L4). Avoid fragile horizontal pre-production environments.

---

## Related Documentation

**Testing:**

- [Testing Strategy Overview](testing-strategy-overview.md) - Complete taxonomy
- [Testing Integration with CD Model](testing-strategy-integration.md) - Stage mapping
- [Tag Reference](../../specifications/tag-reference.md) - All testing taxonomy tags

**CD Model:**

- [CD Model Overview](../cd-model/index.md) - All 12 stages
- [Stages 1-6](../cd-model/cd-model-stages-1-6.md) - Development stages (L0-L3)
- [Stages 7-12](../cd-model/cd-model-stages-7-12.md) - Release stages (L4)

**Specifications:**

- [Three-Layer Testing Approach](../../specifications/three-layer-approach.md) - ATDD/BDD/TDD integration
- [Gherkin File Organization](../../specifications/gherkin-concepts.md) - Writing specifications

**Architecture:**

- [Environments](../architecture/environments.md) - PLTE for L3 tests

Return to **[Continuous Delivery Overview](../index.md)** for complete navigation.
