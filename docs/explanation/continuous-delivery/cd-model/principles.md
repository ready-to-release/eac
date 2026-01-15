# Key Principles

The CD Model is built on five core principles that guide all stages and practices.

## Shift-Left Everything

Move validation activities as early as possible in the development process:

| Activity   | Traditional             | CD Model             |
| ---------- | ----------------------- | -------------------- |
| Testing    | Stage 6 (Extended Test) | Stage 2 (Pre-commit) |
| Security   | Stage 11 (Live)         | Stage 2 (Pre-commit) |
| Compliance | Stage 9 (Approval)      | Stage 1 (Authoring)  |

Early validation finds defects when they're cheapest to fix and fastest to remediate.

## Fail Fast

Quality gates at every stage ensure rapid feedback.

**When something fails:**

1. The pipeline stops immediately
2. The team is notified immediately
3. The pipeline is blocking—the issue must be fixed before proceeding

This prevents defects from progressing through the pipeline and accumulating.

!!!! note

     Fail fast! fail early!

## Automation Over Documentation

While documentation remains important, the CD Model prioritizes executable validation:

| Prefer                      | Over                                           |
| --------------------------- | ---------------------------------------------- |
| Automated tests             | Manual test plans                              |
| Vertical end-to-end PLTE's  | Static and Horizontal end-to-end environments  |
| Infrastructure as Code      | Environment documentation, ex. ServiceNow cm's |
| Pipeline configuration      | Deployment procedures/SOPS                     |
| Automated compliance checks | Manual reviews                                 |

This ensures consistency, controls variability, saves resources and reduces manual errors.

## Continuous Integration

All developers integrate their changes to the main branch frequently (up to multiple times per day).

**Benefits:**

- Reduces merge conflicts
- Enables rapid feedback
- Maintains a releasable main branch
- Supports trunk-based development

See [Trunk-Based Development](../workflow/trunk-based-development.md) for detailed practices.

## Traceability

Every change is traceable through the entire pipeline:

- Production code linked to requirements and verification mechanisms
- Test results linked to specific changes
- Deployment history tracked
- Compliance evidence automatically generated

See [Compliance](compliance.md) for traceability requirements and evidence generation.

---

## References

- [Overview](overview.md) - CD Model introduction
- [The 12 Stages](stages.md) - Where principles are applied
- [Compliance](compliance.md) - Traceability and evidence details
