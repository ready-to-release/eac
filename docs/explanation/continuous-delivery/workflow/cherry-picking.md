# Cherry-Picking

## What is Cherry-Picking?

Cherry-picking copies a commit from one branch to another, creating a new commit with the same changes.

**Primary use case:** Bringing fixes from trunk to release branches.

---

## When to Cherry-Pick

**Scenario:** Critical bug found after release branch created.

**Preferred Flow:**

1. Fix on trunk first (always)
2. Cherry-pick trunk commit to release branch
3. Create PR for release branch with cherry-picked commit

---

## Fix on Trunk First (Preferred)

```bash
# 1. Create topic branch from trunk
git checkout main
git pull origin main
git checkout -b fix/critical-login-bug

# 2. Implement fix
# ... make changes ...

# 3. Merge to trunk via PR (Stages 2-4)
git push origin fix/critical-login-bug
gh pr create --title "Fix critical login bug" --body "Resolves #123"
# ... PR approved and merged ...

# 4. Cherry-pick to release branch
git checkout release/10
git pull origin release/10
git cherry-pick <commit-sha-from-main>

# 5. Create PR for release branch
git push origin release/10
gh pr create --base release/10 \
  --title "Cherry-pick: Fix critical login bug" \
  --body "Cherry-picked from main: <commit-sha>"
```

**Why fix on trunk first?**

- Ensures fix is in next release
- Avoids regressions in future releases
- Maintains trunk as single source of truth
- Follows "fix-forward" principle

---

## Emergency: Fix on Release Branch First

**Only when trunk has diverged significantly and fix is urgent:**

```bash
# 1. Create topic branch from release branch
git checkout release/10
git pull origin release/10
git checkout -b fix/release-critical-bug

# 2. Implement minimal fix
# ... make changes ...

# 3. Merge to release branch via PR
git push origin fix/release-critical-bug
gh pr create --base release/10 \
  --title "Emergency fix for critical bug"
# ... PR approved and merged ...

# 4. IMMEDIATELY cherry-pick to trunk
git checkout main
git pull origin main
git cherry-pick <commit-sha-from-release>
git push origin main

# 5. Verify fix works in both branches
```

**Warning:** Fixing on release branch first risks forgetting to cherry-pick to trunk, causing regressions.

---

## Emergency Fixes in Production

Emergency fixes follow the exact same procedure as any other change.

**Fundamental constraint:** The validated path to production is used for all production changes.

No shortcuts. No direct commits. Same stages, expedited if needed.

---

## Key Principles

| Principle | Rationale |
|-----------|-----------|
| Always fix trunk first | Prevents regressions in future releases |
| Cherry-pick via PR | Maintains audit trail and triggers validation |
| Never skip cherry-pick to trunk | Ensures fix doesn't reappear in next release |
| Same fix, both branches | Divergent fixes cause confusion |

---

## Next Steps

- [Branch Types](branch-types.md) - Understanding release branches
- [Branching Strategies](branching-strategies.md) - RA pattern details
