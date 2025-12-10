# Core Workflows

{{ page_breadcrumb() }}

Essential workflows for daily development with r2r and Everything-as-Code. These tutorials teach you the practices you'll use every day as a productive developer.

## Learning Path

Follow these tutorials to master core development workflows:

| Tutorial | Description |
|----------|-------------|
| [TDD with Specifications](./tdd-with-specifications.md) | Write spec → implement steps → write code (the full TDD loop) |
| [Building and Testing Changes](./building-and-testing.md) | Build affected modules, run tests, validate dependencies |
| [Working with Git Worktrees](./working-with-worktrees.md) | Parallel development with isolated workspaces |
| [Making Your First Release](./making-first-release.md) | Generate changelog, validate CI, create release tags |

## What You'll Learn

After completing these tutorials, you'll be able to:

- Practice test-driven development with executable specifications
- Build and test only the modules affected by your changes
- Work on multiple features in parallel using git worktrees
- Prepare and release modules following best practices
- Validate changes before committing and merging

## Prerequisites

Complete [Getting Started](../getting-started/) tutorials first, especially:

- [Quick Start](../getting-started/quick-start.md)
- [Your First Module](../getting-started/first-module.md)
- [Understanding Test Suites](../getting-started/understanding-test-suites.md)

## Core Workflow Pattern

These tutorials teach the daily development cycle:

```text
1. Create feature workspace (worktree)
2. Write specification (Gherkin)
3. Implement with TDD (red → green → refactor)
4. Build and test affected modules
5. Validate before commit
6. Create pull request
7. Prepare release (changelog, tags)
```

## Next Steps

Once you've mastered core workflows, explore [Advanced Practices](../advanced-practices/) for compliance automation, CI/CD integration, and multi-module development.

{{ diataxis_footer() }}
