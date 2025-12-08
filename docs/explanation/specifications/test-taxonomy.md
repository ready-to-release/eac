# Test Taxonomy

## Overview

The testing taxonomy defines how tests are categorized, executed, and linked to compliance requirements across the CD Model stages.

## Why Test Levels Matter

The five test levels (L0-L4) represent a progression from fast, isolated unit tests to production validation:

- **L0-L1**: Maximum speed, minimum dependencies - catch bugs early
- **L2**: Emulated system tests - validate component behavior
- **L3**: PLTE validation - verify deployment and integration
- **L4**: Production testing - confirm real-world behavior

This progression balances **speed** (fast feedback) with **fidelity** (production-like conditions).

## Why Verification Tags Matter

Verification tags (`@ov`, `@iv`, `@pv`, `@piv`, `@ppv`) categorize the *type* of validation:

- **Operational Verification**: Does the business logic work?
- **Installation Verification**: Did deployment succeed?
- **Performance Verification**: Does it meet SLAs?

This enables targeted test execution at each CD Model stage.

## Why Risk Control Tags Matter

OSCAL-based `@control:` tags link scenarios to standardized compliance requirements, enabling:

- Automated evidence collection
- Audit trail generation
- Gap analysis reporting

## Complete Reference

For complete tag definitions, syntax, and examples, see:

**[Tag Reference](../../reference/specifications/tag-reference.md)**

## Related

- [Testing Strategy Overview](../continuous-delivery/testing/testing-strategy-overview.md)
- [Gherkin Concepts](gherkin-concepts.md)
- [Risk Controls](risk-controls.md)
