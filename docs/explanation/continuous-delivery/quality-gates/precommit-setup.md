# Pre-commit Setup

> **Configuration guide for pre-commit hooks**

How to configure pre-commit hooks for Stage 2 validation, including git hooks, pre-commit framework setup, and optimization strategies.

---

## Overview

Pre-commit hooks provide the fastest feedback loop in the CD Model. When properly configured, they:

- Validate code before it enters version control
- Run automatically on every commit attempt
- Block commits that don't meet quality standards
- Keep the feedback loop tight (5-10 minutes maximum)

---

## Reference Documentation

For complete setup instructions, tool configurations, and scripts, see:

**[Pre-commit Setup Reference](../../../reference/eac/quality-gates/precommit-setup.md)** - Complete implementation guide including:

- Git hook script template
- Pre-commit framework configuration
- Time budget optimization strategies
- Tool reference table
- Emergency bypass options

**[Pre-commit Checks Reference](../../../reference/eac/quality-gates/precommit-checks.md)** - Detailed check categories including:

- Code formatting tools
- Linting configuration
- Unit test execution
- Secret detection
- Dependency scanning
- CI workflow examples

---

## Related Documentation

- [Pre-commit Quality Gates](./precommit-gates.md) - Why pre-commit matters
- [CD Model Stages 1-7](../cd-model/stages.md#development-stages) - See Stage 2 in context
