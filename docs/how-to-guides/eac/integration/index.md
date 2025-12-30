# Integration Guides

Learn how to integrate EAC with your development tools and workflows.

## Coming Soon

This section will provide integration guides for:

### CI/CD Integration

- **GitHub Actions Integration**
  - Setting up EAC commands in workflows
  - Parallel build and test execution
  - Caching strategies for faster builds
  - Using `pipeline ci` for orchestration

- **Other CI Platforms**
  - GitLab CI/CD integration
  - Jenkins integration
  - Azure DevOps pipelines
  - Generic CI/CD patterns

- **Release Automation**
  - Automated changelog generation
  - Version tagging workflows
  - Release verification
  - Multi-module releases

### IDE Integration

- **VS Code Setup**
  - Recommended extensions
  - Task configuration for EAC commands
  - Debugging Go tests
  - Gherkin syntax support
  - Snippet libraries

- **GoLand/IntelliJ IDEA**
  - External tools configuration
  - Run configurations
  - Test runners
  - File watchers

- **Generic Editor Setup**
  - Command palette integration
  - Build task configuration
  - Terminal integration

### Git Hooks

- **Pre-Commit Hooks**
  - Running validation before commit
  - Format checking
  - Linting
  - Quick test execution
  - Example hook scripts

- **Commit-Msg Hooks**
  - Commit message validation
  - Conventional commits enforcement
  - AI-assisted message generation

- **Pre-Push Hooks**
  - Full test suite execution
  - Dependency validation
  - Build verification

### Claude Code Integration

- **Best Practices with AI Assistance**
  - Using EAC commands with Claude Code
  - Specification-driven development
  - Test-first workflows
  - Code review automation

- **Agent Configuration**
  - Setting up agent.md
  - Custom instructions
  - Tool preferences

- **MCP Server Usage**
  - Leveraging MCP commands in Claude Code
  - Troubleshooting MCP connections
  - Fallback strategies

### Documentation Generation

- **MkDocs Integration**
  - Documentation site setup
  - Auto-generation from code
  - API documentation
  - Publishing strategies

- **Architecture Diagrams**
  - Structurizr integration
  - Diagram generation workflow
  - Version control for diagrams

### Testing Tools Integration

- **Godog Integration**
  - Step definition organization
  - Custom formatters
  - Parallel execution
  - Coverage reporting

- **Test Reporting**
  - JUnit XML output
  - Coverage reports
  - Test timing analysis
  - Trend tracking

### Security Scanning Integration

- **Trivy Integration**
  - Container scanning
  - Vulnerability detection
  - IaC scanning
  - Secret detection

- **Semgrep Integration**
  - SAST configuration
  - Custom rules
  - CI/CD integration

- **Compliance Tools**
  - OSCAL integration
  - Control mapping
  - Assessment results

## Quick Integration Examples

Common integration patterns:

### GitHub Actions Example

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: Run EAC Pipeline
        run: go run ./go/eac/commands pipeline ci
```

### Pre-Commit Hook Example

```bash
#!/bin/bash
# .git/hooks/pre-commit
go run ./go/eac/commands validate || exit 1
```

## Related Guides

While this section is being developed, see:

- [Development Workflow](../commands/development-workflow/index.md)
- [Build, Test and Validate](../commands/build-test-validate/index.md)
- [Release Management](../commands/release-management/index.md)
