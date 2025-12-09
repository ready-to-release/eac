# Pipeline CI Commands

{{ page_breadcrumb() }}

**Problem**: CI failures need context-rich diagnostic summaries that link to logs, artifacts, and troubleshooting steps for efficient debugging.

**Solution**: Use `pipeline ci summary-link` to generate markdown diagnostics optimized for GitHub's `$GITHUB_STEP_SUMMARY` environment.

## Key Benefits

- Context-aware diagnostics based on failure type
- Direct links to GitHub Actions artifacts and logs
- Troubleshooting guidance embedded in summaries
- Consistent format across all failure scenarios
- Integration with GitHub's native UI

## Quick Start

```bash
# Build failure diagnostic
r2r eac pipeline ci summary-link 12345678 --type build --artifact build-logs

# Test failure diagnostic
r2r eac pipeline ci summary-link 12345678 --type test --artifact test-results

# Container build failure
r2r eac pipeline ci summary-link 12345678 --type container --image myapp:latest

# Release validation failure
r2r eac pipeline ci summary-link 12345678 --type release --workflow release --commit abc123

# Documentation build failure
r2r eac pipeline ci summary-link 12345678 --type docs --artifact docs-site

# Contract deviation
r2r eac pipeline ci summary-link 12345678 --type deviation
```

## Command Reference

### pipeline ci summary-link

Generate diagnostic markdown for CI summaries.

```bash
r2r eac pipeline ci summary-link <run-id> [options]

# Required:
<run-id>               GitHub Actions run ID

# Options:
--type <type>          Failure type (default: "build")
                       Values: build, test, container, release, docs, deviation
--artifact <name>      Artifact name for download command
--image <name>         Container image name (for container type)
--workflow <name>      CI workflow name (for release type)
--commit <sha>         Commit SHA (for release type)

# Examples:
r2r eac pipeline ci summary-link 12345678 --type build --artifact build-logs
r2r eac pipeline ci summary-link 12345678 --type test --artifact test-results
r2r eac pipeline ci summary-link 12345678 --type container --image api:v1.2.3
r2r eac pipeline ci summary-link 12345678 --type release --workflow release --commit abc123
r2r eac pipeline ci summary-link 12345678 --type docs --artifact mkdocs-site
r2r eac pipeline ci summary-link 12345678 --type deviation
```

### pipeline ci dispatch-and-wait

Dispatch a GitHub workflow and wait for completion.

```bash
r2r eac pipeline ci dispatch-and-wait [options]

# Options:
--workflow <name>      Workflow file name to dispatch (e.g., ci-r2r-cli.yaml)
--ref <ref>            Git ref to run workflow on (default: current branch)
--run-id <id>          Existing run ID to wait for (skips dispatch)
--timeout <seconds>    Timeout in seconds (default: 300)
--inputs <json>        Workflow inputs as JSON object

# Examples:
r2r eac pipeline ci dispatch-and-wait --workflow ci-r2r-cli.yaml --ref main
r2r eac pipeline ci dispatch-and-wait --run-id 12345678 --timeout 600
r2r eac pipeline ci dispatch-and-wait --workflow ci-r2r-cli.yaml --inputs '{"version":"1.0.0"}'
```

## Failure Types

### build

Build compilation or execution failures.

```bash
r2r eac pipeline ci summary-link 12345678 --type build --artifact build-logs

# Generates:
# - Link to build logs artifact
# - Common build troubleshooting steps
# - Dependency check commands
# - Module validation suggestions
```

**Use when:**

- `go build` fails
- Compilation errors occur
- Missing dependencies detected
- Module build scripts fail

### test

Test execution failures.

```bash
r2r eac pipeline ci summary-link 12345678 --type test --artifact test-results

# Generates:
# - Link to test results artifact
# - Test debugging guidance
# - How to reproduce locally
# - Test suite information
```

**Use when:**

- Unit tests fail
- Integration tests error
- Test suites don't pass
- Coverage requirements not met

### container

Container image build failures.

```bash
r2r eac pipeline ci summary-link 12345678 --type container --image api:v1.2.3

# Generates:
# - Container build diagnostics
# - Image name and tag info
# - Docker troubleshooting steps
# - Registry access guidance
```

**Use when:**

- Docker build fails
- Image push errors
- Registry authentication issues
- Multi-stage build problems

### release

Release validation or publishing failures.

```bash
r2r eac pipeline ci summary-link 12345678 --type release --workflow release --commit abc123

# Generates:
# - Release workflow info
# - Commit reference
# - Pre-release validation steps
# - Tag and version guidance
```

**Use when:**

- Release tagging fails
- Version validation errors
- Changelog generation issues
- Release workflow problems

### docs

Documentation build failures.

```bash
r2r eac pipeline ci summary-link 12345678 --type docs --artifact docs-site

# Generates:
# - Docs build diagnostics
# - Markdown validation tips
# - Link checking guidance
# - MkDocs troubleshooting
```

**Use when:**

- MkDocs build fails
- Markdown syntax errors
- Broken internal links
- Navigation structure issues

### deviation

Contract or validation deviations.

```bash
r2r eac pipeline ci summary-link 12345678 --type deviation

# Generates:
# - Contract validation guidance
# - Dependency graph checks
# - Module hierarchy validation
# - How to fix deviations
```

**Use when:**

- Module contracts violated
- Dependency validation fails
- File ownership mismatches
- Contract schema errors

## GitHub Actions Integration

### Basic Integration

```yaml
name: CI Pipeline

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build modules
        id: build
        run: r2r eac build
        continue-on-error: true

      - name: Generate build summary
        if: failure() && steps.build.outcome == 'failure'
        run: |
          r2r eac pipeline ci summary-link ${{ github.run_id }} \
            --type build \
            --artifact build-logs \
            >> $GITHUB_STEP_SUMMARY

      - name: Upload build logs
        if: failure()
        uses: actions/upload-artifact@v3
        with:
          name: build-logs
          path: out/build/**/*.log
```

### Multi-Stage Pipeline

```yaml
name: Full CI/CD Pipeline

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build
        id: build
        run: r2r eac build
        continue-on-error: true

      - name: Build summary
        if: failure() && steps.build.outcome == 'failure'
        run: |
          r2r eac pipeline ci summary-link ${{ github.run_id }} \
            --type build \
            --artifact build-logs \
            >> $GITHUB_STEP_SUMMARY

      - name: Upload build logs
        if: failure()
        uses: actions/upload-artifact@v3
        with:
          name: build-logs
          path: out/build/**/*.log

  test:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Run tests
        id: test
        run: r2r eac test --as-junit > test-results.xml
        continue-on-error: true

      - name: Test summary
        if: failure() && steps.test.outcome == 'failure'
        run: |
          r2r eac pipeline ci summary-link ${{ github.run_id }} \
            --type test \
            --artifact test-results \
            >> $GITHUB_STEP_SUMMARY

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: test-results
          path: |
            test-results.xml
            out/reports/**/*

  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Validate contracts
        id: validate
        run: r2r eac validate
        continue-on-error: true

      - name: Deviation summary
        if: failure() && steps.validate.outcome == 'failure'
        run: |
          r2r eac pipeline ci summary-link ${{ github.run_id }} \
            --type deviation \
            >> $GITHUB_STEP_SUMMARY

  docker:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build container
        id: docker
        run: docker build -t myapp:${{ github.sha }} .
        continue-on-error: true

      - name: Container summary
        if: failure() && steps.docker.outcome == 'failure'
        run: |
          r2r eac pipeline ci summary-link ${{ github.run_id }} \
            --type container \
            --image myapp:${{ github.sha }} \
            >> $GITHUB_STEP_SUMMARY

  release:
    if: github.ref == 'refs/heads/main'
    needs: [build, test, validate]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Create release
        id: release
        run: r2r eac release generate-module-calver r2r-cli
        continue-on-error: true

      - name: Release summary
        if: failure() && steps.release.outcome == 'failure'
        run: |
          r2r eac pipeline ci summary-link ${{ github.run_id }} \
            --type release \
            --workflow release \
            --commit ${{ github.sha }} \
            >> $GITHUB_STEP_SUMMARY
```

### Reusable Workflow

```yaml
# .github/workflows/diagnostic-summary.yml
name: Diagnostic Summary

on:
  workflow_call:
    inputs:
      failure-type:
        required: true
        type: string
      artifact-name:
        required: false
        type: string
      image-name:
        required: false
        type: string

jobs:
  summary:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Generate diagnostic summary
        run: |
          ARGS="--type ${{ inputs.failure-type }}"

          if [ -n "${{ inputs.artifact-name }}" ]; then
            ARGS="$ARGS --artifact ${{ inputs.artifact-name }}"
          fi

          if [ -n "${{ inputs.image-name }}" ]; then
            ARGS="$ARGS --image ${{ inputs.image-name }}"
          fi

          r2r eac pipeline ci summary-link ${{ github.run_id }} $ARGS \
            >> $GITHUB_STEP_SUMMARY

# Use in other workflows:
# - uses: ./.github/workflows/diagnostic-summary.yml
#   with:
#     failure-type: build
#     artifact-name: build-logs
```

## Typical Workflows

### Local Development Testing

```bash
# Simulate CI environment locally
export GITHUB_STEP_SUMMARY=summary.md

# Generate diagnostic
r2r eac pipeline ci summary-link 12345678 --type build --artifact build-logs

# View summary
cat summary.md
```

### Debugging CI Failures

```bash
# 1. Get run ID from GitHub Actions UI
# 2. Generate diagnostic locally
r2r eac pipeline ci summary-link 12345678 --type test --artifact test-results

# 3. Download artifact using gh CLI
gh run download 12345678 -n test-results

# 4. Analyze local artifacts
```

### Custom Failure Types

```bash
# For custom scenarios, use the closest matching type
r2r eac pipeline ci summary-link 12345678 --type build --artifact custom-logs

# Or create custom diagnostics by combining with other commands
r2r eac pipeline ci summary-link 12345678 --type deviation
r2r eac validate >> $GITHUB_STEP_SUMMARY
```

## Output Format

### Standard Markdown Output

The command generates markdown formatted for GitHub's UI:

````markdown
## Build Failure Diagnostics

**Run ID:** 12345678

### Quick Actions

Download build logs:
```bash
gh run download 12345678 -n build-logs
```

### Troubleshooting Steps

1. **Check module contracts**

   ```bash
   r2r eac validate contracts
   ```

2. **Verify dependencies**

   ```bash
   r2r eac validate dependencies
   ```

3. **Review build logs**
   - Navigate to the build-logs artifact
   - Look for compilation errors
   - Check for missing imports

### Common Issues

- **Missing dependencies**: Run `go mod tidy` in affected modules
- **Contract violations**: Update module contracts to match code
- **Version mismatches**: Ensure Go version matches CI environment

### Links

- [View full run logs](https://github.com/owner/repo/actions/runs/12345678)
- [Download build-logs artifact](https://github.com/owner/repo/actions/runs/12345678#artifacts)

````

### GitHub UI Rendering

The markdown appears in the Actions summary tab with:

- Collapsible sections
- Syntax-highlighted code blocks
- Clickable links to runs and artifacts
- Clear visual hierarchy

## Best Practices

### Do's

- **Use specific types**: Choose the most accurate `--type` for context-relevant guidance
- **Include artifacts**: Always specify `--artifact` when logs are uploaded
- **Add to summary**: Append to `$GITHUB_STEP_SUMMARY` in GitHub Actions
- **Continue on error**: Use `continue-on-error: true` to ensure summary generation
- **Upload artifacts**: Upload relevant logs before generating summary

### Don'ts

- **Don't skip run ID**: Always provide the GitHub Actions run ID
- **Don't guess types**: Use default "build" if unsure rather than wrong type
- **Don't duplicate summaries**: Only generate one summary per failure
- **Don't hardcode values**: Use GitHub Actions context variables
- **Don't ignore artifacts**: Upload logs even if summary generation fails

## Integration Patterns

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Run local validation
r2r eac validate

if [ $? -ne 0 ]; then
  echo "Validation failed. See diagnostics:"
  r2r eac pipeline ci summary-link 0 --type deviation
  exit 1
fi
```

### Make Integration

```makefile
.PHONY: ci-test

ci-test:
  @r2r eac test || \
    (r2r eac pipeline ci summary-link $(RUN_ID) --type test --artifact test-results && exit 1)
```

### Docker Integration

```dockerfile
# Dockerfile for CI runner
FROM golang:1.21

# Install r2r CLI
RUN go install github.com/ready-to-release/eac/go/r2r/cli@latest

# Generate diagnostics on build failure
RUN r2r eac build || \
    r2r eac pipeline ci summary-link ${RUN_ID} --type build --artifact build-logs
```

## Troubleshooting

| Problem                | Solution                                               |
| ---------------------- | ------------------------------------------------------ |
| Summary not appearing  | Check `$GITHUB_STEP_SUMMARY` is set and writable       |
| Markdown not rendering | Verify valid markdown syntax in output                 |
| Links broken           | Ensure run ID is correct and run exists                |
| Artifact links 404     | Upload artifacts before generating summary             |
| Wrong diagnostics      | Use correct `--type` for failure scenario              |
| Missing context        | Provide all relevant flags (--artifact, --image, etc.) |
| Permissions error      | Check GitHub token has read access to runs             |

## Advanced Usage

### Custom Templates

```bash
# Generate base diagnostic
r2r eac pipeline ci summary-link 12345678 --type build --artifact logs > summary.md

# Append custom sections
cat >> summary.md <<EOF

## Custom Diagnostics

Additional project-specific guidance here.
EOF

# Write to GitHub summary
cat summary.md >> $GITHUB_STEP_SUMMARY
```

### Multiple Failure Types

```bash
# For complex failures, combine multiple diagnostics
echo "## Multi-Stage Failure Analysis" >> $GITHUB_STEP_SUMMARY
echo "" >> $GITHUB_STEP_SUMMARY

r2r eac pipeline ci summary-link $RUN_ID --type build --artifact build-logs >> $GITHUB_STEP_SUMMARY
echo "" >> $GITHUB_STEP_SUMMARY

r2r eac pipeline ci summary-link $RUN_ID --type test --artifact test-results >> $GITHUB_STEP_SUMMARY
```

### Conditional Diagnostics

```yaml
- name: Smart diagnostic generation
  if: failure()
  run: |
    if [ -f "out/build/errors.log" ]; then
      TYPE="build"
      ARTIFACT="build-logs"
    elif [ -f "test-results.xml" ]; then
      TYPE="test"
      ARTIFACT="test-results"
    else
      TYPE="deviation"
      ARTIFACT=""
    fi

    r2r eac pipeline ci summary-link ${{ github.run_id }} \
      --type $TYPE \
      ${ARTIFACT:+--artifact $ARTIFACT} \
      >> $GITHUB_STEP_SUMMARY
```

## Summary

**Basic usage:**

1. Identify failure type (build, test, container, release, docs, deviation)
2. Get GitHub Actions run ID from context
3. Run: `r2r eac pipeline ci summary-link <run-id> --type <type> [options]`
4. Append output to `$GITHUB_STEP_SUMMARY`

**GitHub Actions integration:**

1. Use `continue-on-error: true` on steps that may fail
2. Add conditional summary step with `if: failure()`
3. Upload artifacts before generating summary
4. Use `${{ github.run_id }}` for run ID
5. Append to `$GITHUB_STEP_SUMMARY` for UI display

**Supported failure types:**

- `build` - Compilation and build failures
- `test` - Test execution failures
- `container` - Docker/container build issues
- `release` - Release validation problems
- `docs` - Documentation build errors
- `deviation` - Contract and validation failures

The pipeline CI commands enhance GitHub Actions workflows with rich, context-aware diagnostic summaries that accelerate debugging and resolution.

{{ diataxis_footer() }}
