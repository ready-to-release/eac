# Command Categories

Browse EAC commands organized by function. Each category groups commands that serve a similar purpose.

<!-- book:categories-index -->

## Finding Commands by Task

### Development Tasks

**I want to...**

- **...see what modules exist**: [show modules](../show/modules.md)
- **...check dependencies**: [show dependencies](../show/dependencies.md)
- **...create a workspace**: [work create](../work/create.md)
- **...write specifications**: [create spec](../create/spec.md)
- **...generate diagrams**: [create design](../create/design.md)

### Quality and Validation

**I want to...**

- **...scan for secrets**: [scan](../scan/scan.md) with `--scanner secrets`
- **...check for vulnerabilities**: [scan](../scan/scan.md) with `--scanner vuln`

### Building and Testing

**I want to...**

- **...build a module**: [build](../build/build.md)
- **...run tests**: [test](../test/test.md)
- **...run a test suite**: [test suite](../test/suite.md)
- **...see test results**: [show test-summary](../show/test-summary.md)

### Committing and PRs

**I want to...**

- **...commit changes**: [work commit](../work/commit.md) or [get commit-message](../get/commit-message.md)
- **...create a PR**: [create pr](../create/pr.md)
- **...see changed files**: [show files-changed](../show/files-changed.md)
- **...get changed modules**: [get changed-modules](../get/changed-modules.md)

### Release and Deployment

**I want to...**

- **...check for release changes**: [release pending](../release/pending.md)
- **...generate changelog**: [release changelog](../release/changelog.md)
- **...create a release**: [release this](../release/this.md)
- **...check CI status**: [release check-ci](../release/check-ci.md)
- **...run pipelines**: [pipeline run](../pipeline/run.md)

### Documentation

**I want to...**

- **...serve docs locally**: [serve docs](../serve/docs.md)
- **...view architecture diagrams**: [serve design](../serve/design.md)
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
