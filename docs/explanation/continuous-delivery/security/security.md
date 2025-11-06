# Security in the CD Model

## Introduction

Security is not a separate phase or stage - it's a continuous practice integrated throughout all 12 stages of the CD Model. This approach, often called "DevSecOps" or "shift-left security," ensures vulnerabilities are detected and addressed as early as possible when they're cheapest and fastest to fix.

This article explains security testing layers, how they integrate with CD Model stages, and focuses on three free, open-source tools: OWASP ZAP, Trivy, and Dependabot.

### Defense in Depth

Defense in depth uses multiple security layers to protect against different threat vectors:

- **SAST**: Static analysis before code runs
- **DAST**: Dynamic testing of running applications
- **Dependency Scanning**: Identifying vulnerable libraries
- **Container Security**: Multi-layer image scanning
- **Runtime Protection**: Monitoring production behavior

No single layer is perfect - multiple overlapping layers provide comprehensive coverage.

### Shift-Left Security

Traditional security approaches test late in the cycle:

- Security testing after development complete
- Vulnerabilities discovered near release
- Expensive to fix, delays release
- Reactive rather than proactive

Shift-left security moves testing earlier:

- Security integrated from Stage 2 (Pre-commit)
- Vulnerabilities caught during development
- Fast feedback, cheap fixes
- Proactive, continuous validation

**Benefits:**

- 10-100x cheaper to fix vulnerabilities early
- Faster releases (no last-minute security blocks)
- Developers learn secure coding practices
- Better security posture overall

### Why Open-Source Tools

This article focuses on three free, open-source tools:

**OWASP ZAP:**

- Free, widely used DAST tool
- Active community and frequent updates
- No licensing costs or restrictions

**Trivy:**

- Free, comprehensive security scanner
- Fast and accurate
- Multi-purpose (containers, filesystems, IaC)

**Dependabot:**

- GitHub-native, free
- Automated dependency updates
- Zero setup for GitHub repositories

These tools provide enterprise-grade security capabilities without licensing costs, making them accessible to all teams.

---

## Security Testing Layers

### SAST (Static Application Security Testing)

**What is SAST:**

Static analysis examines source code without executing it (white-box testing):

- Analyzes code structure and patterns
- Identifies potential vulnerabilities
- Detects insecure coding practices
- Runs before or during build

**When It Runs:**

- **Stage 2 (Pre-commit)**: Fast scans of changed files
- **Stage 3 (Merge Request)**: Comprehensive scan of PR changes
- **Stage 4 (Commit)**: Full codebase scan

**What It Detects:**

- **SQL Injection**: Unsanitized database queries
- **Cross-Site Scripting (XSS)**: Unescaped user input in HTML
- **Hardcoded Secrets**: API keys, passwords in code
- **Insecure Cryptography**: Weak algorithms, bad implementations
- **Path Traversal**: Unsafe file path handling
- **Command Injection**: Unsafe system command execution

**Example (Trivy for SAST):**

```bash
# Scan source code for vulnerabilities
trivy fs --scanners vuln,secret,misconfig .

# Focus on high/critical severity
trivy fs --severity HIGH,CRITICAL --exit-code 1 .
```

**Benefits:**

- Catches issues before code runs
- Fast feedback (seconds to minutes)
- Identifies exact code location
- No runtime environment needed

**Limitations:**

- False positives require tuning
- Can't detect runtime-only issues
- Requires language-specific analyzers
- May miss business logic flaws

### DAST (Dynamic Application Security Testing)

**What is DAST:**

Dynamic analysis tests running applications (black-box testing):

- Interacts with application via APIs/UI
- Simulates attacker behavior
- Tests actual runtime behavior
- No source code access needed

**When It Runs:**

- **Stage 5 (Acceptance Testing)**: Baseline DAST scan
- **Stage 6 (Extended Testing)**: Full comprehensive scan
- **Stage 11 (Production)**: Continuous monitoring scans

**What It Detects:**

- **Authentication Flaws**: Weak auth, session issues
- **Authorization Bypasses**: Privilege escalation
- **Injection Attacks**: SQL, command, LDAP injection at runtime
- **Security Misconfigurations**: Default credentials, open ports
- **Sensitive Data Exposure**: Unencrypted data, information leakage
- **Business Logic Flaws**: Workflow bypasses

**Example (OWASP ZAP for DAST):**

```bash
# Baseline scan (passive, fast)
docker run -t owasp/zap2docker-stable zap-baseline.py \
  -t https://app.example.com \
  -r baseline-report.html

# Full scan (active, comprehensive)
docker run -t owasp/zap2docker-stable zap-full-scan.py \
  -t https://app.example.com \
  -r full-report.html

# API scan with OpenAPI spec
docker run -t owasp/zap2docker-stable zap-api-scan.py \
  -t https://api.example.com \
  -f openapi \
  -d api-spec.yaml
```

**Benefits:**

- Tests actual application behavior
- Finds runtime-only vulnerabilities
- Language-agnostic
- Tests real attack scenarios

**Limitations:**

- Slower than SAST (minutes to hours)
- Requires running application
- May miss code-level issues
- Can have false positives

### Dependency Scanning

**What It Is:**

Identifies known vulnerabilities in third-party dependencies:

- Checks against vulnerability databases (CVE, NVD)
- Analyzes dependency manifests (go.mod, package.json, requirements.txt)
- Detects outdated or vulnerable versions
- Monitors for newly disclosed vulnerabilities

**When It Runs:**

- **Continuous**: Dependabot monitors constantly
- **Stage 2 (Pre-commit)**: Scan changed dependencies
- **Stage 3 (Merge Request)**: Scan all dependencies
- **Stage 4 (Commit)**: Full dependency audit

**What It Detects:**

- **Known CVEs**: Published vulnerabilities in dependencies
- **Outdated Dependencies**: Old versions with known issues
- **License Compliance**: Incompatible or problematic licenses
- **Supply Chain Risks**: Compromised or malicious packages

**Example (Trivy for Dependencies):**

```bash
# Scan Go dependencies
trivy fs --scanners vuln --skip-dirs vendor go.mod

# Scan Node.js dependencies
trivy fs --scanners vuln package-lock.json

# Scan Python dependencies
trivy fs --scanners vuln requirements.txt
```

**Benefits:**

- Automated vulnerability detection
- Continuous monitoring
- Easy remediation (update dependency)
- Prevents supply chain attacks

**Limitations:**

- Only detects known vulnerabilities
- Zero-day vulnerabilities not caught
- Requires up-to-date vulnerability database
- May flag false positives

### Container Security

**What It Is:**

Multi-layer scanning of container images:

- **OS Layer**: Base image vulnerabilities
- **Application Layer**: Application dependencies
- **Configuration Layer**: Misconfigurations and secrets
- **Runtime Layer**: Behavioral anomalies

**When It Runs:**

- **Stage 4 (Commit)**: Scan newly built images
- **Stage 5 (Acceptance Testing)**: Validate deployment images
- **Stage 10 (Deployment)**: Final scan before production
- **Continuous**: Registry scanning for new vulnerabilities

**What It Detects:**

- **OS Vulnerabilities**: Outdated base image packages
- **Application Dependencies**: Vulnerable libraries in image
- **Misconfigurations**: Running as root, exposed ports
- **Secrets in Layers**: Hardcoded credentials in image layers
- **Malware**: Known malicious patterns

**Example (Trivy for Containers):**

```bash
# Scan container image
trivy image --severity HIGH,CRITICAL myapp:latest

# Scan with detailed output
trivy image --severity HIGH,CRITICAL --format json myapp:latest

# Scan and fail build on critical
trivy image --exit-code 1 --severity CRITICAL myapp:latest

# Scan image layers
trivy image --scanners vuln,secret,config myapp:latest
```

**Benefits:**

- Comprehensive multi-layer scanning
- Catches issues before deployment
- Fast (seconds per image)
- Integrates with registries

**Limitations:**

- Only scans container contents
- Can't detect runtime behavior
- False positives for base images
- Requires up-to-date vulnerability DB

---

## Open-Source Security Tools

### OWASP ZAP (Dynamic Testing)

OWASP ZAP (Zed Attack Proxy) is a free, open-source DAST tool for finding vulnerabilities in web applications.

**Scanning Modes:**

**Baseline Scan (5-10 minutes):**

- Passive scanning only
- Spiders the application
- Identifies obvious issues
- Safe for production

```bash
docker run -t owasp/zap2docker-stable zap-baseline.py \
  -t https://app.example.com \
  -r baseline-report.html \
  -J baseline-report.json
```

**Full Scan (hours):**

- Active scanning with attacks
- Comprehensive coverage
- More invasive
- Use in test environments only

```bash
docker run -t owasp/zap2docker-stable zap-full-scan.py \
  -t https://app.example.com \
  -r full-report.html \
  -J full-report.json
```

**API Scan:**

- Scans APIs using OpenAPI/Swagger specs
- Validates endpoints
- Tests authentication

```bash
docker run -t owasp/zap2docker-stable zap-api-scan.py \
  -t https://api.example.com \
  -f openapi \
  -d api-spec.yaml \
  -r api-report.html
```

**Integration Points:**

- **Stage 5 (Acceptance Testing)**: Baseline scan
- **Stage 6 (Extended Testing)**: Full scan
- **Stage 11 (Production)**: Baseline scan for monitoring

**CI/CD Integration Example:**

```yaml
# GitHub Actions example
- name: OWASP ZAP Baseline Scan
  run: |
    docker run -v $(pwd):/zap/wrk/:rw -t owasp/zap2docker-stable \
      zap-baseline.py -t ${{ secrets.APP_URL }} -r zap-report.html

- name: Upload ZAP Report
  uses: actions/upload-artifact@v3
  with:
    name: zap-report
    path: zap-report.html
```

**Configuration:**

- Custom scan rules
- Authentication configuration
- Context files for complex apps
- Tuning for false positives

### Trivy (Multi-Purpose Scanner)

Trivy is a fast, comprehensive security scanner supporting multiple targets.

**Capabilities:**

**Container Image Scanning:**

```bash
# Scan image for vulnerabilities
trivy image myapp:latest

# Scan with severity filtering
trivy image --severity HIGH,CRITICAL --exit-code 1 myapp:latest

# Output as JSON for parsing
trivy image --format json -o results.json myapp:latest
```

**Filesystem/Dependency Scanning:**

```bash
# Scan current directory
trivy fs .

# Scan specific manifest files
trivy fs --scanners vuln go.mod

# Scan for secrets and misconfigurations
trivy fs --scanners secret,config .
```

**IaC Scanning (Terraform, Kubernetes):**

```bash
# Scan Terraform files
trivy config ./terraform/

# Scan Kubernetes manifests
trivy config ./k8s/

# Scan with custom policies
trivy config --policy ./policies ./infrastructure/
```

**Secret Detection:**

```bash
# Scan for hardcoded secrets
trivy fs --scanners secret .

# Scan specific file types
trivy fs --scanners secret --skip-dirs vendor .
```

**Integration Points:**

- **Stage 2 (Pre-commit)**: Fast secret scan
- **Stage 3 (Merge Request)**: Dependency scan
- **Stage 4 (Commit)**: Full image scan
- **Stage 5 (Acceptance Testing)**: IaC validation
- **Stage 10 (Deployment)**: Final image scan

**CI/CD Integration Example:**

```yaml
# GitHub Actions example
- name: Run Trivy vulnerability scanner
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: 'myapp:${{ github.sha }}'
    format: 'sarif'
    output: 'trivy-results.sarif'
    severity: 'HIGH,CRITICAL'
    exit-code: '1'

- name: Upload Trivy results to GitHub Security
  uses: github/codeql-action/upload-sarif@v2
  with:
    sarif_file: 'trivy-results.sarif'
```

**Severity Thresholds:**

```bash
# Block on critical only
trivy image --severity CRITICAL --exit-code 1 myapp:latest

# Warn on high, block on critical
trivy image --severity HIGH,CRITICAL myapp:latest

# Scan all, informational only
trivy image --exit-code 0 myapp:latest
```

### Dependabot (Dependency Management)

Dependabot is GitHub's native, free automated dependency update tool.

**Features:**

- Automated pull requests for vulnerable dependencies
- Security-only or all updates
- Multi-ecosystem support (Go, npm, Docker, GitHub Actions, etc.)
- Version compatibility checking
- Automated testing before merge

**Configuration Example:**

```yaml
# .github/dependabot.yml
version: 2
updates:
  # Go dependencies
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10

  # Docker base images
  - package-ecosystem: "docker"
    directory: "/"
    schedule:
      interval: "weekly"

  # GitHub Actions
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
```

**Security-Only Updates:**

```yaml
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "daily"
    open-pull-requests-limit: 5
    # Only security updates
    labels:
      - "security"
    # Group security updates
    groups:
      security-updates:
        patterns:
          - "*"
```

**Grouping Strategies:**

```yaml
# Group patch updates together
updates:
  - package-ecosystem: "gomod"
    groups:
      patch-updates:
        update-types:
          - "patch"
```

**Integration Points:**

- **Continuous**: Monitors constantly, creates PRs
- **Stage 3 (Merge Request)**: PRs reviewed and merged

**Benefits:**

- Zero manual effort for monitoring
- Automated PRs reduce friction
- Integrated with GitHub workflows
- Free for all GitHub repositories

---

## Security by Stage Matrix

| Stage | SAST | DAST | Dependency | Container | Primary Tools | Duration |
|-------|------|------|------------|-----------|---------------|----------|
| **Pre-commit (2)** | ✓ | - | ✓ | - | Trivy (secrets, deps) | < 2 min |
| **Merge Request (3)** | ✓ | - | ✓ | ✓ | Trivy, Dependabot | < 5 min |
| **Commit (4)** | ✓ | - | ✓ | ✓ | Trivy (full scan) | < 10 min |
| **Acceptance (5)** | ✓ | ✓ | ✓ | ✓ | Trivy, OWASP ZAP (baseline) | < 15 min |
| **Extended (6)** | ✓ | ✓ | ✓ | ✓ | OWASP ZAP (full scan) | 1-4 hours |
| **Deployment (10)** | - | - | ✓ | ✓ | Trivy (final check) | < 5 min |
| **Production (11)** | - | ✓ | - | - | OWASP ZAP (monitoring) | Continuous |

---

## Shift-Left Security Practices

**IDE Integration:**

- Install Trivy CLI locally
- Run scans before commit
- Immediate feedback in development

**Pre-Commit Hooks:**

```bash
#!/bin/sh
# .git/hooks/pre-commit

# Fast secret scan (< 1 min)
trivy fs --scanners secret --exit-code 1 .

# Fast dependency scan (< 1 min)
trivy fs --scanners vuln --severity CRITICAL --exit-code 1 .
```

**Local Container Scanning:**

```bash
# Before pushing image
trivy image --severity HIGH,CRITICAL myapp:latest
```

**Developer Education:**

- Secure coding training
- Understanding common vulnerabilities (OWASP Top 10)
- How to interpret security scan results
- How to remediate vulnerabilities

---

## Blocking Strategy

**Critical/High Severity: Block Pipeline**:

```yaml
# Fail build on critical/high
trivy image --severity HIGH,CRITICAL --exit-code 1 myapp:latest
```

**Medium Severity: Warn, Require Review**:

```yaml
# Generate warning, don't fail
trivy image --severity MEDIUM --exit-code 0 myapp:latest
```

**Low Severity: Informational Only**:

- Log findings
- Track in backlog
- Address during maintenance

**Example CI/CD Configuration:**

```yaml
security-scan:
  script:
    # Block on critical/high
    - trivy image --severity HIGH,CRITICAL --exit-code 1 $IMAGE

    # Warn on medium (doesn't fail)
    - trivy image --severity MEDIUM --format json -o medium-findings.json $IMAGE

  artifacts:
    reports:
      container_scanning: medium-findings.json
```

---

## Vulnerability Remediation Workflow

**1. Detection:**

- Automated scanning finds vulnerability
- Alert sent to team
- Issue created in tracking system

**2. Triage (within 24 hours):**

- Assess severity and impact
- Determine exploitability
- Identify affected systems

**3. Prioritize:**

- **P0 (Critical)**: Fix within 24 hours
- **P1 (High)**: Fix within 7 days
- **P2 (Medium)**: Fix within 30 days
- **P3 (Low)**: Address in next sprint

**4. Remediate:**

- Update dependency to patched version
- Apply security patch
- Implement workaround if patch unavailable
- Add compensating controls

**5. Verify:**

- Re-run security scans
- Validate fix effective
- Ensure no regression

**6. Document:**

- Record in vulnerability tracking system
- Update knowledge base
- Share learnings with team

---

## Best Practices

**Regular Tool Updates:**

- Update Trivy vulnerability database daily
- Keep OWASP ZAP updated
- Enable Dependabot for all repositories

**Tune False Positives:**

- Maintain suppression files
- Document why issues are suppressed
- Review suppressions regularly

**Security Metrics:**

- Track Mean Time To Detect (MTTD)
- Track Mean Time To Remediate (MTTR)
- Monitor vulnerability counts over time
- Measure scan coverage

**Evidence Collection:**

- Store scan reports as artifacts
- Link findings to commits
- Maintain audit trail for compliance
- Generate compliance reports

**Container Security Best Practices:**

- Use specific version tags (not `latest`)
- Run containers as non-root user
- Use multi-stage builds for minimal attack surface
- Scan before every deployment
- Remove unnecessary packages

---

## Summary

Security in the CD Model is continuous and multi-layered:

**Security Layers:**

- **SAST**: Static analysis (Trivy)
- **DAST**: Dynamic testing (OWASP ZAP)
- **Dependency Scanning**: Vulnerability monitoring (Trivy, Dependabot)
- **Container Security**: Multi-layer scanning (Trivy)

**Open-Source Tools:**

- **OWASP ZAP**: Free DAST tool for web applications
- **Trivy**: Multi-purpose security scanner
- **Dependabot**: Automated dependency updates

**Integration:**

- Security testing at every stage (2-11)
- Shift-left for early detection
- Automated blocking on critical findings
- Continuous monitoring in production

By integrating these free, open-source tools throughout the CD Model, teams achieve enterprise-grade security without licensing costs.

## Next Steps

- [CD Model Overview](../cd-model/cd-model-overview.md) - See security in context
- [Stages 1-6](../cd-model/cd-model-stages-1-6.md) - Security in development
- [Stages 7-12](../cd-model/cd-model-stages-7-12.md) - Security in release
- [Testing Strategy Integration](../testing/testing-strategy-integration.md) - Integrate security testing

## References

- [OWASP ZAP Documentation](https://www.zaproxy.org/docs/)
- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
- [Dependabot Documentation](https://docs.github.com/en/code-security/dependabot)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
