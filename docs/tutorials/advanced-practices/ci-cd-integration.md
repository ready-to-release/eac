# CI/CD Integration

{{ page_breadcrumb() }}

**Status:** Placeholder - Content coming soon

**Estimated time:** 60 minutes
**Prerequisites:** [Making Your First Release](../core-workflows/making-first-release.md), GitHub Actions basics

## Planned Content

This tutorial teaches you how to set up automated CI/CD pipelines with quality gates, using GitHub Actions and r2r commands for continuous delivery.

### What You'll Learn

- Set up GitHub Actions workflows for r2r projects
- Implement quality gates at commit, merge, and release stages
- Use `r2r pipeline` commands for orchestration
- Run changed modules only in CI
- Configure test suites for different pipeline stages
- Monitor pipeline status and wait for completion
- Troubleshoot CI failures efficiently

### Tutorial Structure

1. **Understanding the CD model**
   - 12-stage continuous delivery framework
   - Quality gates: pre-commit, merge request, release
   - Test levels at each stage (L0-L2 → L3 → L4)
   - Artifact progression through stages

2. **GitHub Actions setup**
   - Workflow file structure (`.github/workflows/`)
   - Trigger events: push, pull_request, workflow_dispatch
   - Job dependencies and parallelization
   - Secrets and environment variables

3. **Pre-commit quality gate**
   - Local: pre-commit hooks
   - What to validate: format, lint, L0-L2 tests, specs
   - Fast feedback (< 5 minutes)

4. **Commit/Push quality gate**
   ```yaml
   name: Commit Pipeline
   on: push
   jobs:
     validate:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - name: Get changed modules
           run: r2r get changed-modules-ci
         - name: Build changed modules
           run: r2r build $(r2r get changed-modules-ci)
         - name: Test changed modules (commit suite)
           run: r2r test $(r2r get changed-modules-ci)
         - name: Validate all
           run: r2r validate
   ```

5. **Pull request quality gate**
   - Run acceptance suite (L0-L3, IV/OV/PV)
   - Security scanning (SAST, vulnerabilities)
   - Dependency validation
   - Specification validation
   - More thorough (10-20 minutes)

6. **Release quality gate**
   - Production verification suite (L4, PIV)
   - Generate SBOM
   - Compliance scanning
   - Release approval checks

7. **Using pipeline commands**
   - Run pipeline: `r2r pipeline run`
   - Check status: `r2r pipeline status`
   - Dispatch and wait: `r2r pipeline ci dispatch-and-wait`
   - Wait for completion: `r2r pipeline wait`

8. **Optimizing CI performance**
   - Build only changed modules
   - Cache dependencies (Go modules, npm packages)
   - Parallel test execution
   - Incremental builds
   - Matrix builds for multi-platform

9. **Monitoring and debugging**
   - View CI status: `r2r pipeline status`
   - Generate CI summary: `r2r pipeline ci-summary-link`
   - Debug failures: analyze logs, reproduce locally
   - Flaky test detection

### Complete Workflow Example

The tutorial will set up a complete CI/CD pipeline:

**Pre-commit (local):**
```bash
# .git/hooks/pre-commit
r2r validate
```

**Commit pipeline (on push):**
- Build changed modules
- Test changed modules (commit suite)
- Validate contracts and specs

**Pull request pipeline (on PR):**
- Build all affected modules
- Test with acceptance suite
- Security scans (SAST, vulnerabilities, IaC)
- Validate dependencies
- Generate test report

**Release pipeline (on tag):**
- Build release artifacts
- Run production-verification suite
- Generate SBOM
- Compliance checks
- Create GitHub release
- Deploy to staging

### Key Concepts Covered

- Quality gates and their purposes
- Test suite selection per stage
- Incremental builds in CI
- Pipeline orchestration with r2r
- Monitoring and troubleshooting
- Performance optimization

### GitHub Actions Best Practices

- Use `r2r get changed-modules-ci` for incremental builds
- Cache Go modules: `actions/cache@v4`
- Run tests in parallel where possible
- Set appropriate timeouts
- Use matrix builds for multi-platform
- Store artifacts for debugging

### Example: Complete CI Workflow

```yaml
name: CI
on:
  push:
    branches: [main]
  pull_request:

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Install r2r CLI
        run: |
          curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH

      - name: Get changed modules
        id: changed
        run: echo "modules=$(r2r get changed-modules-ci)" >> $GITHUB_OUTPUT

      - name: Build
        if: steps.changed.outputs.modules != ''
        run: r2r build ${{ steps.changed.outputs.modules }}

      - name: Test (commit suite)
        if: steps.changed.outputs.modules != ''
        run: r2r test ${{ steps.changed.outputs.modules }}

      - name: Validate
        run: r2r validate
```

### Troubleshooting CI Failures

Common issues and solutions:

- **Build fails in CI but passes locally**: Check environment differences
- **Tests timeout**: Increase timeout or optimize tests
- **Flaky tests**: Use `r2r show test-timings` to identify
- **Cache issues**: Clear cache and rebuild

### Next Steps

After completing this tutorial, you'll have a complete CI/CD pipeline. Continue to [Multi-Module Development](./multi-module-development.md) to learn how to manage dependencies across multiple modules.

{{ diataxis_footer() }}
