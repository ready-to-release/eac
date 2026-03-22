# Command Categories

Browse EAC commands organized by function. Each category groups commands that serve a similar purpose.

<!-- book:categories-index -->

## Finding Commands by Task

### Development Tasks

**I want to...**

- **...see what modules exist**: [show modules](../commands/show/modules.md)
- **...check dependencies**: [show dependencies](../commands/show/dependencies.md)
- **...create a workspace**: [work create](../commands/work/create.md)
- **...write specifications**: [create spec](../commands/create/spec.md)
- **...generate diagrams**: [create design](../commands/create/design.md)

### Quality and Validation

**I want to...**

- **...scan for secrets**: [scan](../commands/scan/scan.md) with `--scanner secrets`
- **...check for vulnerabilities**: [scan](../commands/scan/scan.md) with `--scanner vuln`

### Building and Testing

**I want to...**

- **...build a module**: [build](../commands/build/build.md)
- **...run tests**: [test](../commands/test/test.md)
- **...run a test suite**: [test suite](../commands/test/suite.md)
- **...see test results**: [show test-summary](../commands/show/test-summary.md)

### Committing and PRs

**I want to...**

- **...commit changes**: [work commit](../commands/work/commit.md) or [get commit-message](../commands/get/commit-message.md)
- **...create a PR**: [create pr](../commands/create/pr.md)
- **...see changed files**: [show files-changed](../commands/show/files-changed.md)
- **...get changed modules**: [get changed-modules](../commands/get/changed-modules.md)

### Release and Deployment

**I want to...**

- **...check for release changes**: [release pending](../commands/release/pending.md)
- **...generate changelog**: [release changelog](../commands/release/changelog.md)
- **...create a release**: [release this](../commands/release/this.md)
- **...check CI status**: [release check-ci](../commands/release/check-ci.md)
- **...run pipelines**: [pipeline run](../commands/pipeline/run.md)

### Documentation

**I want to...**

- **...serve docs locally**: [serve docs](../commands/serve/docs.md)
- **...view architecture diagrams**: [serve design](../commands/serve/design.md)
- **...manage templates**: [templates commands](./templates.md)

## Category Patterns

### Information Retrieval (get/show)

Most information commands come in pairs:

| get (JSON)         | show (Formatted)    | Information      |
| ------------------ | ------------------- | ---------------- |
| `get modules`      | `show modules`      | Module contracts |
| `get dependencies` | `show dependencies` | Dependency graph |
| `get files`        | `show files`        | File ownership   |
| `get config`       | `show config`       | Configuration    |
| `get tests`        | `show tests`        | Test information |

**Rule**: Use `get` for automation, `show` for interactive and templated markdown use.

### Quality Gates (validate/scan/test)

Quality commands all validate different aspects:

- **validate**: Repository structure and contracts
- **scan**: Security issues and compliance
- **test**: Functional correctness

**Use in**: Pre-commit hooks, CI pipelines, release gates

### Workflow (work/pipeline/release)

Workflow commands manage different stages:

- **work**: Local development (git worktrees)
- **pipeline**: CI/CD execution
- **release**: Version management

## See Also

- [Command Reference Index](../index.md) - Main command reference
- [Command Taxonomy](../overview/command-taxonomy.md) - How commands are organized
- [Naming Conventions](../overview/naming-conventions.md) - Command naming rules
- [Output Formats](../overview/output-formats.md) - JSON vs formatted output
