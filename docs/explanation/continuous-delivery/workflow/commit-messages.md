# Commit Messages

> **Semantic commit messages for traceability**

Conventional commits provide structured messages that enable automated changelog generation and maintain clear project history.

---

## Why Conventional Commits?

Semantic commit messages enable:

- **Automated changelog generation** - Tools can extract features, fixes, and breaking changes
- **Clear project history** - Anyone can understand what changed and why
- **Semantic versioning** - Commit types inform version bumps (feat = minor, fix = patch)
- **Traceability** - Link commits to issues and requirements

---

## Basic Format

```text
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Key Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

---

## Related Documentation

- [Traceability Requirements](../cd-model/stages.md#stage-4-commit) - Stage 4 traceability
- [Pre-commit Setup](../quality-gates/precommit-setup.md) - Pre-commit hooks
- [Trunk-Based Development](./trunk-based-development.md) - Development workflow
- [Branching Strategies](./branching-strategies.md) - Branch naming
