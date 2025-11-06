# Security

Security integration throughout all stages of the Continuous Delivery Model. This section explains how to embed security practices and tools into your delivery pipeline using open-source solutions.

---

## [Security in the CD Model](security.md)

Security integration throughout all stages using open-source tools.

Security is not a separate phase or gate - it's integrated throughout the entire CD Model from Stage 2 (Pre-commit) through Stage 11 (Live). This approach, known as "shift-left security," finds vulnerabilities early when they're cheapest and easiest to fix.

**The Shift-Left Security Principle:**

Traditional security approaches perform security testing late in the cycle (before production deployment). By then, vulnerabilities are expensive to fix and may block releases.

The CD Model integrates security at multiple stages:

**Stage 2 (Pre-commit):**

- Secret scanning (prevent credentials from being committed)
- Basic SAST (static analysis)
- Dependency vulnerability scanning
- < 10 minutes to maintain fast feedback

**Stage 4 (Commit):**

- Comprehensive SAST analysis
- Container image scanning
- Dependency analysis with fail thresholds
- Software Bill of Materials (SBOM) generation

**Stage 5 (Acceptance Testing):**

- DAST (dynamic analysis) against running application
- API security testing
- Authentication and authorization validation

**Stage 6 (Extended Testing):**

- Full DAST suite with authenticated scans
- Penetration testing (automated)
- Compliance validation

**Stage 11 (Live):**

- Runtime security monitoring
- Vulnerability response
- Incident detection and response

**Topics covered:**

- Security tool taxonomy (SAST, DAST, SCA, container scanning)
- OWASP ZAP for dynamic application security testing
- Trivy for multi-purpose scanning (containers, dependencies, secrets, IaC)
- Dependabot for automated dependency updates
- Security by stage matrix (what runs where)
- Shift-left security practices
- Blocking strategies (when to fail builds)
- Remediation workflow
- Vulnerability prioritization

**Read this article to understand**: How to integrate security throughout your CD pipeline using open-source tools.

---

## Security Tools Overview

**Primary Open-Source Tools:**

**1. Trivy (Multi-Purpose Scanner)**:

- Container image vulnerability scanning
- Dependency vulnerability scanning (package.json, go.mod, etc.)
- Secret detection in code
- Infrastructure as Code scanning (Terraform, CloudFormation)
- SBOM generation

**Use in**: Stages 2, 4, 5, 6

**2. OWASP ZAP (Dynamic Application Security Testing)**:

- Runtime security testing against deployed application
- API security testing
- Authentication/authorization testing
- Active and passive scanning modes

**Use in**: Stages 5, 6

**3. Dependabot (Dependency Management)**:

- Automated dependency update PRs
- Security advisory monitoring
- Automatic vulnerability detection
- Configurable update strategy

**Use in**: Continuous background process, creates PRs that go through Stages 2-4

**4. golangci-lint / ESLint / Similar SAST Tools**:

- Language-specific static analysis
- Security-focused linting rules
- Code quality and security issues
- Fast execution for pre-commit

**Use in**: Stages 2, 3, 4

---

## Security by Stage

| Stage | Security Activities | Tools | Time Budget |
|-------|-------------------|-------|-------------|
| **2. Pre-commit** | Secret scanning, basic SAST | Trivy, golangci-lint | < 2 min |
| **3. Merge Request** | Full SAST, dependency scan | Trivy, language-specific | 5-10 min |
| **4. Commit** | Container scan, SBOM generation | Trivy | 5 min |
| **5. Acceptance** | DAST passive scan, API testing | OWASP ZAP | 15-30 min |
| **6. Extended** | DAST active scan, authenticated | OWASP ZAP | 30-60 min |
| **11. Live** | Runtime monitoring, incident response | Monitoring tools | Continuous |

---

## Shift-Left Security Benefits

**Early Detection:**

- Find vulnerabilities in Stage 2 (minutes after writing code)
- Not Stage 6 (hours/days later)
- Not Stage 11 (in production)

**Lower Cost:**

- Fixing in Stage 2: 1x cost (minutes of developer time)
- Fixing in Stage 6: 10x cost (blocks release, requires full testing)
- Fixing in Stage 11: 100x cost (production incident, emergency deployment)

**Better Security:**

- Developers learn secure coding practices through immediate feedback
- Security becomes part of workflow, not a blocker
- More vulnerabilities caught before production

**Faster Delivery:**

- No late-stage security surprises
- Security integrated, not bolted on
- Parallel security testing in pipeline

---

## Blocking Strategies

Not all security issues should block the pipeline. The CD Model uses risk-based blocking:

**Stage 2 (Pre-commit) - BLOCK:**

- ❌ Secrets detected in code
- ❌ Critical vulnerabilities in direct dependencies
- ✅ Allow: Low/medium vulnerabilities (fix in follow-up)

**Stage 4 (Commit) - BLOCK:**

- ❌ High/critical vulnerabilities in container images
- ❌ Critical dependency vulnerabilities
- ❌ Secrets detected
- ✅ Allow: Medium/low vulnerabilities with mitigation plan

**Stage 6 (Extended Testing) - BLOCK:**

- ❌ High/critical DAST findings (SQLi, XSS, etc.)
- ❌ Authentication bypass vulnerabilities
- ✅ Allow: Low-risk findings with documented acceptance

**Stage 11 (Live) - RESPOND:**

- Monitor and respond to security incidents
- Emergency patching process
- Incident response procedures

---

## Vulnerability Remediation Workflow

**1. Detection** (Automated)

- Security scan identifies vulnerability
- Trivy/ZAP/Dependabot reports finding
- Finding includes: CVE, severity, affected component

**2. Triage** (Automated + Manual)

- Automated: Compare against severity thresholds
- Manual: Review exploitability in your context
- Decision: Block, warn, or accept risk

**3. Remediation** (Developer Action)

- Update dependency to patched version
- Apply code fix for vulnerability
- Add security control (WAF rule, input validation)
- Document risk acceptance (if cannot fix immediately)

**4. Verification** (Automated)

- Re-run security scans
- Verify vulnerability resolved
- Update vulnerability tracking

**5. Prevention** (Process Improvement)

- Update security rules to catch similar issues
- Improve developer training
- Enhance pre-commit checks

---

## Integration with Other Sections

**[Core Concepts](../core-concepts/index.md)**:

- Security applies to all Deployable Units
- Immutable artifacts scanned once, deployed many times

**[CD Model](../cd-model/index.md)**:

- Security integrated at multiple stages (2, 4, 5, 6, 11)
- Quality gates include security thresholds
- Evidence collection includes security scan results

**[Workflow](../workflow/index.md)**:

- Pre-commit includes secret scanning (Stage 2)
- Security scans run automatically on trunk commits (Stage 4)
- Dependabot creates PRs that flow through normal workflow

**[Testing](../testing/index.md)**:

- DAST integrated with L3 end-to-end tests (Stage 5)
- Security tests part of acceptance criteria
- Security regression tests in automated suite

**[Architecture](../architecture/index.md)**:

- PLTE environment used for DAST testing
- Build Agents execute security scans
- Production security monitoring and alerting

---

## Compliance and Audit

Security integration supports compliance requirements:

**Evidence Generation:**

- Security scan reports archived
- SBOM generated for each release
- Vulnerability tracking and remediation documented

**Audit Trail:**

- When security scans executed
- What vulnerabilities found
- How vulnerabilities remediated
- Who approved risk acceptances

**Regulatory Support:**

- Demonstrates continuous security validation
- Provides evidence for audits
- Shows security integrated throughout SDLC

---

## Best Practices

**Security Tooling:**

✅ **DO:**

- Integrate security scans in multiple stages (shift-left)
- Use open-source tools when possible (transparency, cost)
- Automate security scanning (consistency)
- Provide fast feedback (< 10 min in Stage 2)
- Generate SBOMs for all releases

❌ **DON'T:**

- Perform security testing only before production
- Block on all vulnerabilities (prioritize by risk)
- Rely on manual security reviews alone
- Skip security in Stage 2 (too slow)

**Vulnerability Management:**

✅ **DO:**

- Triage based on exploitability in your context
- Fix critical/high vulnerabilities immediately
- Document risk acceptance decisions
- Monitor for new vulnerabilities continuously
- Use Dependabot for automated updates

❌ **DON'T:**

- Ignore vulnerability reports
- Accept all risks without documentation
- Delay security updates indefinitely
- Disable security scans to "unblock" releases

**Developer Experience:**

✅ **DO:**

- Provide clear, actionable security feedback
- Include remediation guidance in reports
- Train developers on secure coding
- Make security tools easy to run locally
- Celebrate security improvements

❌ **DON'T:**

- Provide overwhelming, noisy reports
- Blame developers for security issues
- Make security someone else's problem
- Skip developer training

---

## Next Steps

- **Learn about security integration:** Read [Security in the CD Model](security.md)
- **Understand stages:** See [CD Model](../cd-model/index.md)
- **Set up environments:** Read [Architecture](../architecture/index.md)
- **Integrate testing:** Explore [Testing](../testing/index.md)

Return to **[Continuous Delivery Overview](../index.md)** for complete navigation.
