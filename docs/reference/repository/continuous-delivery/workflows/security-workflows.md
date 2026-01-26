# Security Workflows

Reference for security scanning workflows.

## Overview

Security workflows perform automated vulnerability scanning and security analysis on the codebase.

These workflows run independently of CI/CD pipelines to detect security issues early.

## CodeQL Workflow

**File:** `.github/workflows/codeql.yaml`

**Purpose:** Performs static analysis security testing (SAST) using GitHub CodeQL to identify security vulnerabilities and code quality issues.

### Triggers

```yaml
on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main
  schedule:
    - cron: '0 0 * * 1'  # Weekly on Mondays at midnight UTC
```

**Trigger conditions:**

- **Push to main:** Scans every commit merged to main
- **Pull requests:** Scans all PRs to detect issues before merge
- **Weekly schedule:** Full scan every Monday to catch new vulnerabilities

### Permissions

```yaml
permissions:
  actions: read         # Read workflow artifacts
  contents: read        # Checkout code
  security-events: write  # Upload security results to GitHub
```

### Languages Analyzed

CodeQL analyzes the following languages in the repository:

- **Go** - Primary language for modules
- **JavaScript/TypeScript** - VSCode extensions
- **Shell** - Scripts and automation

Configuration is auto-detected from repository content.

### Job: analyze

Runs on `ubuntu-latest` with matrix strategy for each language.

**Strategy:**

```yaml
strategy:
  fail-fast: false
  matrix:
    language: [go, javascript, shell]
```

**Behavior:** Each language runs in parallel as a separate job

### Analysis Steps

#### Step 1: Checkout Repository

```yaml
- name: Checkout repository
  uses: actions/checkout@v6
```

#### Step 2: Initialize CodeQL

```yaml
- name: Initialize CodeQL
  uses: github/codeql-action/init@v3
  with:
    languages: ⟪ matrix.language ⟫
    queries: security-extended
```

**Query Suite:** `security-extended`

- Includes all security queries
- Additional code quality checks
- Best practices validation

**Alternative query suites:**

- `security-and-quality` - Security + quality checks
- `security` - Security checks only

#### Step 3: Autobuild

```yaml
- name: Autobuild
  uses: github/codeql-action/autobuild@v3
```

**Behavior:**

- Automatically detects build system (Go modules, npm, etc.)
- Builds code to enable deep analysis
- For Go: Runs `go build`
- For JavaScript: Runs `npm install` and `npm build`

#### Step 4: Perform CodeQL Analysis

```yaml
- name: Perform CodeQL Analysis
  uses: github/codeql-action/analyze@v3
  with:
    category: "/language:⟪ matrix.language ⟫"
```

**Output:**

- Security alerts uploaded to GitHub Security tab
- SARIF results available for download
- Results visible in pull request checks

### Analysis Categories

Each language creates a separate analysis category:

- `/language:go`
- `/language:javascript`
- `/language:shell`

**Purpose:** Allows filtering and tracking results by language

### Security Alerts

CodeQL findings are reported in multiple locations:

#### Security Tab

Navigate to: **Repository → Security → Code scanning alerts**

**Alert Information:**

- Severity (Critical, High, Medium, Low, Note)
- Description and recommendation
- Affected file and line number
- CWE classification
- First detected date

#### Pull Request Checks

**Check Name:** `CodeQL / Analyze ({language})`

**Status:**

- **Success:** No new alerts introduced
- **Failure:** New security issues detected
- **Warning:** Existing issues in changed code

**Review:**

- Click check for detailed results
- Shows alerts in PR diff
- Allows dismissing false positives

#### Security Advisories

Critical findings may trigger security advisory creation:

**Advisory Information:**

- CVE assignment (if applicable)
- Severity assessment
- Patch recommendations
- Dependency updates

### Alert Management

#### Dismissing Alerts

False positives can be dismissed:

```bash
# Via GitHub UI
Security → Code scanning → Select alert → Dismiss

# Via API
gh api repos/{owner}/{repo}/code-scanning/alerts/{alert_number} \
  -X PATCH \
  -f state=dismissed \
  -f dismissed_reason="false positive"
```

**Dismissal Reasons:**

- `false positive` - Not actually a security issue
- `won't fix` - Acknowledged but not fixing
- `used in tests` - Test code, not production

#### Tracking Remediation

**Alert States:**

- `open` - Active issue requiring attention
- `dismissed` - Marked as false positive or won't fix
- `fixed` - Resolved in subsequent commit

### Common Findings

#### Go Security Issues

- **SQL Injection:** Unsanitized SQL query construction
- **Command Injection:** Unsafe command execution
- **Path Traversal:** Unvalidated file path operations
- **Hardcoded Credentials:** Secrets in source code
- **Weak Cryptography:** Use of deprecated crypto algorithms

#### JavaScript Security Issues

- **XSS (Cross-Site Scripting):** Unsafe DOM manipulation
- **Prototype Pollution:** Unsafe object merging
- **Path Traversal:** File system access vulnerabilities
- **Regular Expression DoS:** Inefficient regex patterns
- **Hardcoded Secrets:** API keys or tokens in code

#### Shell Security Issues

- **Command Injection:** Unquoted variable expansion
- **Path Traversal:** Unsafe file operations
- **Privilege Escalation:** Unsafe sudo usage
- **Hardcoded Credentials:** Passwords in scripts

### Configuration

CodeQL configuration can be customized via `.github/codeql/codeql-config.yml`:

```yaml
name: "CodeQL Config"

queries:
  - uses: security-extended

paths-ignore:
  - "**/*_test.go"
  - "out/**"
  - "vendor/**"

paths:
  - "go/**"
  - "typescript/**"
  - "scripts/**"
```

**Customization Options:**

- **Query suites:** Change security analysis depth
- **Paths:** Specify files to include/exclude
- **Custom queries:** Add organization-specific rules

### Performance Considerations

**Build optimization:**

- Autobuild may take 5-15 minutes for large codebases
- Parallel language analysis improves speed
- Results cached for unchanged code

**Schedule optimization:**

- Weekly scans balance security with resource usage
- Can increase frequency for high-security requirements

### Debugging CodeQL

#### View Workflow Runs

```bash
# List recent CodeQL runs
gh run list --workflow codeql.yaml --limit 10

# View specific run
gh run view <run-id>

# View logs
gh run view <run-id> --log
```

#### View Security Alerts

```bash
# List code scanning alerts
gh api repos/{owner}/{repo}/code-scanning/alerts

# View specific alert
gh api repos/{owner}/{repo}/code-scanning/alerts/{alert_number}
```

#### Test Locally

CodeQL CLI can be installed for local testing:

```bash
# Install CodeQL CLI
gh extension install github/gh-codeql

# Create database
codeql database create db --language=go

# Run analysis
codeql database analyze db \
  --format=sarif-latest \
  --output=results.sarif \
  security-extended.qls
```

## Additional Security Tools

While CodeQL is the primary security workflow, the repository uses additional security scanning via commands:

### Trivy Scanning

```bash
# Vulnerability scanning
r2r eac scan --scanner vuln

# Secret detection
r2r eac scan --scanner secrets

# IaC scanning
r2r eac scan --scanner iac

# Compliance checking
r2r eac scan --scanner compliance
```

### Semgrep SAST

```bash
# Static analysis
r2r eac scan --scanner sast
```

See module documentation for detailed scan command specifications.

## References

- CodeQL Documentation: [https://codeql.github.com/docs/](https://codeql.github.com/docs/)
- Security Best Practices: [https://docs.github.com/en/code-security](https://docs.github.com/en/code-security)
- Workflow file: `.github/workflows/codeql.yaml`
