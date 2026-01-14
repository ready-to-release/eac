---
name: go-cli-release-check
description: Pre-release validation checklist for CLI modules
---

# Go CLI Release Check Skill

This skill orchestrates comprehensive pre-release validation to ensure modules are ready for release.

## When to Use This Skill

- Before tagging a release
- Need to verify release readiness
- Want automated validation checklist
- Preparing for production deployment

## Workflow

### Step 1: Check CI Status (go-security-release)

**Action**: Verify all CI checks passed

- Use MCP `release-check-ci <module>` to verify CI status
- Check all tests passing on CI
- Verify build successful
- If dependencies need to pass first, use MCP `release-await-deps <module>`

**Blockers**:

- ❌ Any CI check failing
- ❌ Tests failing on CI
- ❌ Build errors

### Step 2: Run Security Scans (go-security-release)

**Action**: Execute security scanning

- MCP `scan <module>` for comprehensive security scan
- Run `govulncheck ./...` for Go vulnerability checking
- Check for hardcoded secrets (grep for API keys, passwords)
- Scan dependencies for known vulnerabilities

**Blockers**:

- ❌ High/Critical vulnerabilities found
- ❌ Hardcoded secrets detected
- ❌ Vulnerable dependencies without fixes

### Step 3: Verify Build Artifacts

**Action**: Ensure clean build

- MCP `build <module>` to build the module
- MCP `validate-artifacts <module>` to check build state
- Verify all expected artifacts created
- Check binary sizes reasonable

**Blockers**:

- ❌ Build fails
- ❌ Missing artifacts
- ❌ Compilation errors

### Step 4: Review Changelog

**Action**: Validate release documentation

- MCP `get-changelog <module>` to review changelog entries
- MCP `get-release-notes <module>` to view release notes
- Verify version follows semver (MAJOR.MINOR.PATCH)
- Ensure all changes documented
- Check for breaking changes noted

**Blockers**:

- ❌ Missing changelog entry for this version
- ❌ Invalid version format
- ❌ Breaking changes not documented

### Step 5: Validate Dependencies

**Action**: Check module contracts

- MCP `validate-dependencies <module>` to check contracts
- Run `go mod tidy` and verify no changes
- Run `go mod verify` to check go.sum integrity
- Check for circular dependencies

**Blockers**:

- ❌ Dependency validation fails
- ❌ `go mod tidy` makes changes
- ❌ Circular dependencies detected
- ❌ go.sum integrity check fails

### Step 6: Run Full Test Suite

**Action**: Execute all tests

- MCP `test <module>` for unit tests
- Run integration tests if available
- Run Gherkin specs tests
- Verify no flaky tests
- Check test coverage acceptable

**Command**:

```bash
# Unit tests
go test ./...

# With race detector
go test -race ./...

# With coverage
go test -cover ./...
```

**Blockers**:

- ❌ Any test failures
- ❌ Race conditions detected
- ❌ Coverage significantly decreased

### Step 7: Validate Documentation

**Action**: Check documentation current

- Verify CLI help text accurate
- Check README up to date
- Verify how-to guides current
- Ensure API docs correct (if library)

**Blockers**:

- ❌ Outdated documentation
- ❌ Missing documentation for new features

### Step 8: Final Checklist

**Action**: Comprehensive final review

**Must ALL be ✅**:

- ✅ All tests passing (unit, integration, specs)
- ✅ Security scans clean (or issues accepted)
- ✅ CI checks passing
- ✅ Build successful
- ✅ Changelog complete and accurate
- ✅ Version valid (follows semver)
- ✅ Dependencies validated
- ✅ Documentation updated
- ✅ No hardcoded secrets
- ✅ `go mod tidy` makes no changes
- ✅ Race detector clean

## Output

The skill produces a release readiness report:

```markdown
## Release Readiness Report

**Module**: <module-name>
**Version**: <version>
**Date**: <date>

### CI Status
✅ All checks passing
✅ Build: Success
✅ Tests: 245 passed, 0 failed

### Security
✅ SAST scan: Clean
✅ Dependency audit: No vulnerabilities
✅ Secret scan: No secrets found
✅ govulncheck: Clean

### Build & Artifacts
✅ Build: Success
✅ Artifacts: All present
✅ Binary sizes: Normal

### Release Documentation
✅ Changelog: v1.2.0 entry complete
✅ Release notes: Generated
✅ Version: Valid semver
✅ Breaking changes: None

### Dependencies
✅ Contracts: Valid
✅ go.sum: Verified
✅ go mod tidy: No changes

### Testing
✅ Unit tests: 245 passed
✅ Integration tests: 42 passed
✅ Spec tests: 18 passed
✅ Race detector: Clean
✅ Coverage: 84% (maintained)

### Documentation
✅ CLI help: Current
✅ README: Updated
✅ How-to guides: Current

### Recommendation
**✅ APPROVED FOR RELEASE v1.2.0**

No blocking issues found. Module is ready for release.
```

Or if issues found:

```markdown
### Recommendation
**❌ BLOCKED - Not ready for release**

Blocking issues:
1. ❌ Tests failing: TestValidation (line 42)
2. ❌ Security: 1 high vulnerability in dependency X
3. ❌ Changelog: Missing entry for v1.2.0

Action required:
1. Fix failing test
2. Update dependency X to v2.1.0
3. Document changes in CHANGELOG.md

Re-run this check after addressing issues.
```

## When to Run

**Always run before**:

- Creating git tags
- Publishing releases
- Deploying to production
- Announcing releases

**Recommended frequency**:

- Before every release (mandatory)
- After major changes
- Weekly for active development

## Integration Points

This skill uses:

- **go-security-release agent**: For security and CI checks
- **MCP commands**: For validation operations
- **Standard Go tools**: For building and testing

## Example Usage

**User request**: "Check if the `commands` module is ready for v1.5.0"

**Skill execution**:

1. Verify CI passing for commands module
2. Run security scans
3. Build and validate artifacts
4. Review changelog for v1.5.0 entry
5. Validate all dependencies
6. Run full test suite
7. Check documentation
8. Generate release readiness report

**Outcome**: Either ✅ APPROVED or ❌ BLOCKED with specific action items

## Success Criteria

Release is approved ONLY if:

- All 8 steps pass without blockers
- All checkboxes in final checklist are ✅
- No high/critical security issues
- All tests passing
- Documentation current

This skill ensures every release meets quality standards and is production-ready.
