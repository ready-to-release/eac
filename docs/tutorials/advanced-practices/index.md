# Advanced Practices

{{ page_breadcrumb() }}

Master advanced workflows including compliance automation, CI/CD integration, multi-module development, and CLI extensibility.

**Estimated time:** 3-4 hours total

## Learning Path

These tutorials cover advanced topics for power users:

| Tutorial | Time | Description |
|----------|------|-------------|
| [Compliance-as-Code Basics](./compliance-as-code-basics.md) | 45 min | Tag specs for GxP compliance, generate risk assessments |
| [CI/CD Integration](./ci-cd-integration.md) | 60 min | Set up GitHub Actions with quality gates |
| [Multi-Module Development](./multi-module-development.md) | 45 min | Manage dependencies across multiple modules |
| [Creating Custom Commands](./creating-custom-commands.md) | 60 min | Extend r2r CLI with custom commands |

## What You'll Learn

After completing these tutorials, you'll be able to:

- Implement compliance automation for regulated environments
- Set up automated CI/CD pipelines with quality gates
- Work effectively across module dependencies
- Extend the r2r CLI with custom commands
- Generate audit evidence automatically
- Manage complex monorepo architectures

## Prerequisites

Complete [Core Workflows](../core-workflows/) tutorials first. These advanced tutorials assume you're comfortable with:

- Creating and testing modules
- Writing specifications
- Using git worktrees
- Making releases

## Advanced Workflow Patterns

These tutorials teach sophisticated practices:

```text
- Compliance: Tag specifications → Generate evidence → Audit ready
- CI/CD: Quality gates → Automated deployment → Monitoring
- Multi-module: Dependency management → Incremental builds → Integration testing
- Extensibility: Custom commands → MCP integration → Automation
```

## Who Should Take These?

These tutorials are for:

- **Compliance teams** implementing automation in regulated environments
- **DevOps engineers** setting up CI/CD pipelines
- **Architects** managing complex modular systems
- **Power users** who want to customize and extend the CLI

## Next Steps

After mastering advanced practices, explore [Specialized Topics](../specialized-topics/) for deep dives into specific areas like BDD techniques, architecture documentation, and security scanning.

{{ diataxis_footer() }}
