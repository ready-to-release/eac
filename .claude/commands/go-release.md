# Release

```text
description: "Prepare and validate a Go CLI module release"
```

You are preparing a Go CLI module for release.

## Process

1. **Run release checklist**:
   - Delegate to go-security-release agent
   - Use Task tool with subagent_type="go-security-release"
   - Or invoke go-cli-release-check skill directly
   - This runs all pre-release validation steps

2. **Check CI status**:
   - MCP `release-check-ci <module>` to verify CI passed
   - MCP `release-await-deps <module>` if dependencies need to pass first
   - MCP `pipeline-status` to view overall CI health

3. **Security scans**:
   - MCP `scan <module>` to run security checks
   - Run `govulncheck ./...` for Go vulnerabilities
   - Review scan results and remediate if needed

4. **Validate build and tests**:
   - MCP `build <module>`
   - MCP `test <module>`
   - MCP `validate-artifacts <module>`

5. **Review changelog and release notes**:
   - MCP `get-changelog <module>`
   - MCP `get-release-notes <module>`
   - Ensure version follows semver
   - Verify all changes documented

6. **Dependency validation**:
   - MCP `validate-dependencies <module>`
   - Run `go mod tidy` and verify no changes
   - Check go.sum integrity

7. **Final approval**:
   - All tests passing
   - Security scans clean
   - Documentation updated
   - Ready to tag release

## Output

Provide release readiness summary:

- ✅ All validation checks passed
- ✅ CI status: passing
- ✅ Security: no issues found
- ✅ Tests: all passing
- ✅ Documentation: up to date
- **Ready to proceed with release** OR **Blocked - action items needed**

## Example Usage

User: `/go:release check if commands module is ready for v1.2.0`
