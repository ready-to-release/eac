# Cherry-Picking

> **Moving fixes between branches**

Cherry-picking copies a commit from one branch to another, creating a new commit with the same changes.

---

## When to Cherry-Pick

**Primary use case**: Bringing fixes from trunk to release branches.

**Scenario**: Critical bug found after release branch created.

**Preferred Flow**:

1. Fix on trunk first (always)
2. Cherry-pick trunk commit to release branch
3. Create PR for release branch with cherry-picked commit

---

## Why Fix on Trunk First?

Fixing trunk first is the preferred approach because:

- Ensures fix is in next release
- Avoids regressions in future releases
- Maintains trunk as single source of truth
- Follows "fix-forward" principle

---

## Key Principles

| Principle                       | Rationale                                     |
| ------------------------------- | --------------------------------------------- |
| Always fix trunk first          | Prevents regressions in future releases       |
| Cherry-pick via PR              | Maintains audit trail and triggers validation |
| Never skip cherry-pick to trunk | Ensures fix doesn't reappear in next release  |
| Same fix, both branches         | Divergent fixes cause confusion               |

---

## Reference Documentation

For complete git commands and step-by-step procedures, see:

**[Cherry-Picking Reference](../../../reference/eac/workflow/cherry-picking.md)**

Complete guide including:

- Fix on trunk first procedure
- Emergency fix on release branch procedure
- Git commands with examples
- PR creation commands

---

## Related Documentation

- [Branch Types](branch-types.md) - Understanding release branches
- [Branching Strategies](branching-strategies.md) - RA pattern details
- [Commit Messages](commit-messages.md) - Conventional commit format
