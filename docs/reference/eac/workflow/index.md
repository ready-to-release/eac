# Workflow Reference

Git workflow commands and conventions.

## Quick Reference

```bash
# Create topic branch
git checkout main
git pull origin main
git checkout -b topic/user/feature-name

# Cherry-pick fix to release
git checkout release/10
git cherry-pick <commit-sha>
gh pr create --base release/10

# Conventional commit
git commit -m "feat(api): add user authentication"
```

---

## In This Section

| Topic                                   | Description                                    |
| --------------------------------------- | ---------------------------------------------- |
| [Commit Messages](./commit-messages.md) | Conventional commit format and examples        |
| [Cherry-Picking](./cherry-picking.md)   | Git commands for moving fixes between branches |

---

## Related Documentation

- [Trunk-Based Development (Conceptual)](../../../explanation/continuous-delivery/workflow/trunk-based-development.md) - Core principles
- [Branch Types (Conceptual)](../../../explanation/continuous-delivery/workflow/branch-types.md) - Branch naming conventions
- [Branching Strategies (Conceptual)](../../../explanation/continuous-delivery/workflow/branching-strategies.md) - RA vs CDe flows
