# Release Management

## Overview

Release management is the process of planning, scheduling, and controlling software releases.

In the CD Model, it spans **Stage 8 (Start Release)** and **Stage 9 (Release Approval)**.

**Key responsibilities:**

- Document what's changing (release notes, changelog)
- Collect evidence of quality (test results, security scans)
- Decide when to deploy (approval decision)
- Enable recovery (rollback procedures)

### Stage 8: Start Release

- Create release candidate
- Generate release notes
- Update changelog
- Package artifacts

### Stage 9: Release Approval

- Validate production readiness
- Review quality evidence
- Make go/no-go decision (RA) or automated approval (CDe)

---

## In This Section

| Topic                                       | Description                                     |
| ------------------------------------------- | ----------------------------------------------- |
| [Changelog System](./changelog-system.md)   | How to release using the changelog-based system |
| [Approval Patterns](./approval-patterns.md) | RA vs CDe approval workflows                    |
| [Release Evidence](./release-evidence.md)   | What evidence is required for approval          |
| [Release Notes](./release-notes.md)         | How to write effective release notes            |

---

## Next Steps

- [Changelog System](./changelog-system.md) - Start here to learn how releases work
- [Approval Patterns](./approval-patterns.md) - Understand RA vs CDe patterns
- [Quality Gates](../quality-gates/index.md) - Stage-specific quality thresholds
