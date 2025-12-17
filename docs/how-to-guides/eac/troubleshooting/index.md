# Troubleshooting
Solutions to common problems when working with EAC.

## Coming Soon

This section will provide troubleshooting guides for:

### Build Issues

- **Build Fails with Compilation Errors**
  - Understanding error messages
  - Common Go compilation issues
  - Dependency resolution problems
  - Module path issues

- **Build Takes Too Long**
  - Identifying bottlenecks
  - Parallel build optimization
  - Caching strategies

- **Artifacts Not Generated**
  - Verifying build pipeline
  - Checking artifact definitions
  - Output directory permissions

### Test Failures

- **Tests Failing After Changes**
  - Using `test debug` command
  - Reading test output
  - Identifying root causes
  - Step definition mismatches

- **Intermittent Test Failures**
  - Race conditions
  - Timing issues
  - External dependencies

- **Test Suite Not Running**
  - Suite configuration issues
  - Tag filter problems
  - Step definition discovery

### Dependency Problems

- **Module Dependency Errors**
  - Circular dependencies
  - Missing dependencies
  - Version conflicts
  - Using `validate dependencies`

- **Go Module Issues**
  - `go mod tidy` failures
  - Replace directives
  - Local module references

### Validation Failures

- **Contract Validation Errors**
  - Understanding validation messages
  - Fixing module contract issues
  - Resolving file ownership conflicts

- **Specification Validation Fails**
  - Gherkin syntax errors
  - Tag contract violations
  - Missing step definitions

### MCP Server Issues

- **MCP Server Not Connecting**
  - Checking .mcp.json configuration
  - Verifying server is running
  - Authentication problems
  - Port conflicts

- **Commands Not Available**
  - Server health check
  - Partial connection issues
  - Fallback to CLI mode

### AI Integration Issues

- **AI Commands Failing**
  - API key configuration
  - Rate limiting
  - Network connectivity
  - Provider-specific issues

- **Generated Output Quality**
  - Improving prompts
  - Context management
  - Model selection

### Git and Workspace Issues

- **Worktree Problems**
  - Branch conflicts
  - Workspace synchronization
  - Merge conflicts

- **Commit Message Generation Fails**
  - Staged changes requirements
  - History analysis issues

### Performance Issues

- **Slow Command Execution**
  - Profiling commands
  - Repository size optimization
  - Caching strategies

- **Memory Usage**
  - Large repository handling
  - Concurrent operations

## Quick Diagnostics

Common diagnostic commands:

```bash
# Check overall repository health
go run ./go/eac/commands validate

# View configuration
go run ./go/eac/commands show config

# Check module dependencies
go run ./go/eac/commands validate dependencies

# Debug test failures
go run ./go/eac/commands test debug

# Check CI status
go run ./go/eac/commands pipeline status
```

## Related Guides

- [Validate Before Commit](../commands/build-test-validate/validate-before-commit.md)
- [Debug Test Failures](../commands/build-test-validate/debug-test-failures.md)
- [Check Dependencies](../commands/build-test-validate/check-dependencies.md)

## Getting Help

If you can't find a solution:

1. Check the command help: `go run ./go/eac/commands help <command>`
2. Review the reference documentation
3. Check project issues on GitHub
4. Consult the development team
