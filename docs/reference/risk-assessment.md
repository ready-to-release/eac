<!-- EDITOR
# Editor: reference/risk-assessment.md

## Soul

Comprehensive NIST 800-53 risk assessment treating EAC CLI as developer infrastructure in regulated environments, with threat analysis and prioritized control implementation.

## Sections

1. Executive Summary
2. System Description
3. Asset Inventory
4. Threat Analysis
5. Vulnerability Assessment
6. Risk Evaluation
7. Required Security Controls
8. Recommended Control Implementation
9. Residual Risk
10. Compliance Mapping
11. Conclusions and Recommendations
12. Approval and Sign-Off
13. Appendix A: References
-->

# Risk Assessment

**Version**: 1.0
**Date**: 2025-12-04
**Assessment Type**: Initial Risk Assessment
**Scope**: r2r as a developer tooling solution in regulated environments

## Executive Summary

This risk assessment evaluates the EAC CLI, a modular monorepo management tool designed for use by software development teams in regulated industries (healthcare, financial services, government). The CLI is treated as developer infrastructure—similar to Git, Docker, or IDE tooling—rather than production infrastructure.

**Key Findings:**

- **Risk Level**: MODERATE
- **Primary Concerns**: Supply chain security, credential management, audit trail integrity
- **Recommended Control Framework**: NIST 800-53 Rev5 (Developer Tools subset)
- **Compliance Considerations**: SOC 2, HIPAA (developer tools), FedRAMP Low

---

## 1. System Description

### 1.1 Purpose and Function

The EAC CLI is a command-line developer tool that provides:

- Modular monorepo architecture management
- Automated testing and validation workflows
- Security scanning and evidence collection (SBOM, vulnerability scanning, secrets detection)
- AI-assisted specification and documentation generation
- Git workflow automation and worktree management
- Build orchestration with dependency resolution
- OSCAL-based security control documentation

### 1.2 Deployment Context

**Usage Environment:**

- Installed on developer workstations (Windows, macOS, Linux)
- Integrated into CI/CD pipelines (GitHub Actions)
- Used by software engineers, security engineers, and DevOps teams
- Operates within source code repositories (Git)

**Data Classification:**

- **Source Code**: Organization confidential
- **Test Results**: Internal use
- **Security Scan Evidence**: Internal/audit use
- **Developer Credentials**: Highly sensitive (GitHub tokens, AI API keys)
- **Module Contracts**: Internal use

**Regulatory Context:**

- Tool used in development of regulated software systems
- Subject to organizational security policies
- Audit trail requirements for compliance evidence
- Supply chain transparency requirements

---

## 2. Asset Inventory

### 2.1 Information Assets

| Asset | Classification | Criticality | Location |
|-------|---------------|-------------|----------|
| Source Code Repository | Confidential | High | Developer workstation, GitHub |
| Developer Credentials (GITHUB_TOKEN) | Highly Sensitive | Critical | Environment variables, CI secrets |
| AI Provider Keys (OpenAI, Anthropic) | Highly Sensitive | High | `.ai/config.json` |
| Security Scan Evidence | Internal | Medium | `out/security/` directory |
| Test Results | Internal | Medium | `out/test/` directory |
| Module Contracts (modules.yml, environments.yml) | Internal | Medium | `.r2r/eac/` directory |
| OSCAL Security Documentation | Internal | Medium | `docs/risk/` directory |
| Build Artifacts | Internal | Low | `out/build/` directory |

### 2.2 Technical Assets

| Asset | Type | Purpose | Risk Profile |
|-------|------|---------|--------------|
| CLI Binary (Go) | Executable | Command execution | Supply chain, integrity |
| Go Dependencies | Third-party libraries | Functionality | Supply chain vulnerability |
| External Tools (Trivy, Semgrep, ZAP) | Security scanners | Evidence generation | Tool compromise, supply chain |
| MCP Servers | Local services | Command execution, GitHub API | Privilege escalation |
| AI Providers (Claude, GPT) | External API | Code/spec generation | Data leakage, prompt injection |
| Git Worktrees | File system | Parallel development | File system access |

---

## 3. Threat Analysis

### 3.1 Threat Actors

| Actor | Motivation | Capability | Likelihood |
|-------|-----------|------------|------------|
| External Attacker | Financial gain, espionage | High (nation-state, ransomware) | Medium |
| Malicious Insider | Sabotage, data theft | High (authorized access) | Low |
| Supply Chain Compromise | Backdoor installation | High (sophisticated) | Medium |
| Accidental Misuse | Human error | Low | High |

### 3.2 Threat Scenarios

#### T1: Malicious Dependency Injection

**Description**: Attacker compromises a Go dependency or external tool (Trivy, Semgrep) to inject malicious code.

**Attack Vector**: Supply chain poisoning via dependency confusion, typosquatting, or compromised maintainer accounts.

**Impact**:

- Code execution on developer workstations
- Credential theft (GitHub tokens, AI keys)
- Source code exfiltration
- Backdoor installation in built artifacts

**Likelihood**: MEDIUM (historical precedent: SolarWinds, codecov, event-stream)

---

#### T2: Credential Leakage via CLI

**Description**: Developer credentials (GITHUB_TOKEN, AI API keys) are exposed through logs, error messages, or insecure storage.

**Attack Vector**:

- Accidental commit of `.ai/config.json` to Git
- Logging credentials in debug output
- Environment variable exposure in CI logs
- Insecure file permissions on credential files

**Impact**:

- Unauthorized repository access
- API abuse (AI provider costs)
- Data exfiltration via GitHub API
- CI/CD pipeline compromise

**Likelihood**: HIGH (common developer mistake)

---

#### T3: AI Prompt Injection

**Description**: Attacker crafts malicious input (risk assessment, commit messages, specifications) that causes AI to generate harmful code or commands.

**Attack Vector**:

- Injecting shell commands in AI-generated commit messages
- Crafting specifications that generate vulnerable code
- Manipulating risk assessments to omit security controls

**Impact**:

- Command injection vulnerabilities
- Security control bypass
- Misleading audit evidence
- Code quality degradation

**Likelihood**: MEDIUM (emerging threat for AI-assisted tools)

---

#### T4: Audit Trail Tampering

**Description**: Security evidence files (SBOM, vulnerability scans, test results) are modified to hide vulnerabilities or compliance failures.

**Attack Vector**:

- Direct file modification in `out/security/` directory
- Git history rewriting to remove evidence
- Timestamp manipulation on evidence files

**Impact**:

- Failed audit due to unreliable evidence
- Undetected vulnerabilities in production
- Compliance violations
- Loss of regulatory approval

**Likelihood**: LOW (requires malicious intent)

---

#### T5: Unauthorized Access to Source Code

**Description**: CLI is compromised or misconfigured, allowing unauthorized parties to access proprietary source code.

**Attack Vector**:

- CLI binary with embedded backdoor
- MCP server privilege escalation
- GitHub token theft via CLI
- Worktree file permission issues

**Impact**:

- Intellectual property theft
- Competitive disadvantage
- Regulatory data breach (if code contains PHI/PII)

**Likelihood**: MEDIUM

---

#### T6: External Tool Compromise

**Description**: Third-party security tools (Trivy, Semgrep, OWASP ZAP) are compromised and provide false security evidence.

**Attack Vector**:

- Man-in-the-middle attack on tool downloads
- Registry poisoning (Docker Hub, npm)
- Tool maintainer account compromise

**Impact**:

- False sense of security (vulnerabilities not detected)
- Compliance evidence unreliable
- Production deployment of vulnerable code

**Likelihood**: LOW-MEDIUM

---

#### T7: CI/CD Pipeline Abuse

**Description**: CLI in CI/CD environment is used to escalate privileges or exfiltrate data.

**Attack Vector**:

- Pull request with malicious CLI usage
- Compromised GitHub Actions workflow
- Secrets exposed via CLI debug output in CI logs

**Impact**:

- Full repository compromise
- CI/CD pipeline takeover
- Production deployment of malicious code

**Likelihood**: MEDIUM

---

## 4. Vulnerability Assessment

### 4.1 Technical Vulnerabilities

| Vulnerability | CVSS | CWE | Description | Affected Component |
|--------------|------|-----|-------------|-------------------|
| VUL-1: Dependency Vulnerabilities | 7.5 (High) | CWE-1035 | Known CVEs in Go dependencies | `go.mod` dependencies |
| VUL-2: Insecure Credential Storage | 8.1 (High) | CWE-522 | AI config stored in plaintext | `.ai/config.json` |
| VUL-3: Command Injection in AI Output | 7.3 (High) | CWE-78 | Unsanitized AI-generated commands | `create commit-message`, `create pr` |
| VUL-4: Path Traversal in File Operations | 6.5 (Medium) | CWE-22 | Insufficient path validation | File read/write operations |
| VUL-5: Unsigned Evidence Files | 5.0 (Medium) | CWE-345 | No cryptographic signature on evidence | `out/security/*` |
| VUL-6: Insufficient Access Control on MCP | 6.8 (Medium) | CWE-284 | MCP servers run with user privileges | MCP server implementation |
| VUL-7: Cleartext Transmission of Secrets | 7.5 (High) | CWE-319 | Environment variables may log/expose | CI/CD integration |

### 4.2 Process Vulnerabilities

| Vulnerability | Impact | Description |
|--------------|--------|-------------|
| PROC-1: No Verification of External Tools | High | Trivy, Semgrep, ZAP not cryptographically verified before use |
| PROC-2: Inconsistent Security Evidence Format | Medium | Evidence files lack standardized structure for audit |
| PROC-3: No Audit Log for CLI Actions | Medium | No record of who ran what command when |
| PROC-4: Lack of Developer Training | Medium | Developers may misuse security features |
| PROC-5: No Incident Response Plan | Medium | No procedure for handling CLI compromise |

---

## 5. Risk Evaluation

### 5.1 Risk Matrix

| Risk ID | Threat | Likelihood | Impact | Risk Level | Priority |
|---------|--------|-----------|--------|------------|----------|
| RISK-1 | Malicious Dependency (T1) | Medium | Critical | **HIGH** | P1 |
| RISK-2 | Credential Leakage (T2) | High | High | **HIGH** | P1 |
| RISK-3 | AI Prompt Injection (T3) | Medium | High | **MEDIUM** | P2 |
| RISK-4 | Audit Tampering (T4) | Low | Critical | **MEDIUM** | P2 |
| RISK-5 | Source Code Access (T5) | Medium | High | **MEDIUM** | P2 |
| RISK-6 | Tool Compromise (T6) | Low | High | **MEDIUM** | P3 |
| RISK-7 | CI/CD Abuse (T7) | Medium | Critical | **HIGH** | P1 |
| RISK-8 | Dependency Vulns (VUL-1) | High | High | **HIGH** | P1 |
| RISK-9 | Insecure Cred Storage (VUL-2) | High | Critical | **HIGH** | P1 |
| RISK-10 | Command Injection (VUL-3) | Medium | High | **MEDIUM** | P2 |

**Overall System Risk**: **MODERATE-HIGH** (requires immediate control implementation)

---

## 6. Required Security Controls

### 6.1 Access Control (AC)

**RISK-2, RISK-5, RISK-7, VUL-6**:

- **AC-2: Account Management** - Control who can install/use the CLI
- **AC-3: Access Enforcement** - Enforce least privilege for CLI operations
- **AC-6: Least Privilege** - MCP servers run with minimal necessary privileges
- **AC-17: Remote Access** - Secure GitHub token usage and API access

**Justification**: Prevent unauthorized use of the CLI and limit blast radius of compromise.

---

### 6.2 Audit and Accountability (AU)

**RISK-4, PROC-3**:

- **AU-2: Audit Events** - Log all CLI command executions with user, timestamp, arguments
- **AU-3: Content of Audit Records** - Include command, module, result, duration
- **AU-9: Protection of Audit Information** - Evidence files cryptographically signed (SHA256)
- **AU-10: Non-repudiation** - Immutable audit trail for compliance evidence

**Justification**: Enable detection of tampering and provide compliance audit trail.

---

### 6.3 Configuration Management (CM)

**PROC-4**:

- **CM-2: Baseline Configuration** - Document approved CLI version and dependencies
- **CM-3: Configuration Change Control** - Review and approve CLI updates
- **CM-7: Least Functionality** - Disable unused commands/features in regulated environments
- **CM-8: Information System Component Inventory** - Track all deployed CLI instances

**Justification**: Maintain known-good state and prevent drift.

---

### 6.4 Identification and Authentication (IA)

**RISK-2, VUL-2, VUL-7**:

- **IA-5: Authenticator Management** - Secure storage of GitHub tokens and AI keys
- **IA-5(7): No Embedded Unencrypted Credentials** - Credentials never in source code
- **IA-5(1): Password-Based Authentication** - Use keyring/credential manager for secrets

**Justification**: Protect highly sensitive credentials from exposure.

---

### 6.5 System and Information Integrity (SI)

**RISK-1, RISK-6, RISK-8, VUL-1**:

- **SI-2: Flaw Remediation** - Regular updates for CLI and dependencies
- **SI-3: Malicious Code Protection** - Scan CLI binary and dependencies for malware
- **SI-4: Information System Monitoring** - Monitor for anomalous CLI usage patterns
- **SI-7: Software, Firmware, and Information Integrity** - Verify CLI binary signature, SBOM
- **SI-10: Information Input Validation** - Sanitize AI-generated output before execution

**Justification**: Defend against supply chain attacks and malicious code injection.

---

### 6.6 Supply Chain Risk Management (SR)

**RISK-1, RISK-6, VUL-1, PROC-1**:

- **SR-3: Supply Chain Controls** - Verify provenance of CLI and dependencies
- **SR-4: Provenance** - Maintain SBOM for CLI and all dependencies
- **SR-5: Acquisition Strategies** - Use official releases, verify signatures
- **SR-6: Supplier Assessments** - Evaluate security of external tools (Trivy, Semgrep)
- **SR-11: Component Authenticity** - Verify checksums/signatures of external tools

**Justification**: Mitigate supply chain compromise risk.

---

### 6.7 System and Communications Protection (SC)

**VUL-7**:

- **SC-8: Transmission Confidentiality** - Use TLS for all external API calls
- **SC-12: Cryptographic Key Establishment** - Secure generation/storage of signing keys
- **SC-13: Cryptographic Protection** - Use industry-standard algorithms for evidence signing
- **SC-28: Protection of Information at Rest** - Encrypt credential files at rest

**Justification**: Protect data in transit and at rest.

---

### 6.8 Risk Assessment (RA)

**ALL RISKS**:

- **RA-3: Risk Assessment** - Annual review of CLI security posture
- **RA-5: Vulnerability Scanning** - Continuous scanning of dependencies (CLI's own SBOM)
- **RA-7: Risk Response** - Document mitigation plans for identified risks

**Justification**: Continuous improvement and adaptation to evolving threats.

---

## 7. Recommended Control Implementation

### 7.1 Immediate Actions (P1 - High Priority)

| Control | Implementation | Effort | Timeline |
|---------|---------------|--------|----------|
| **SR-4: SBOM Generation** | Use `scan sbom` command to document CLI's own dependencies | Low | Week 1 |
| **RA-5: Dependency Scanning** | Implement `scan vuln` on CLI itself in CI/CD | Low | Week 1 |
| **SR-6: Secrets Detection** | Add `scan secrets` to pre-commit hooks | Low | Week 1 |
| **IA-5: Secure Credential Storage** | Implement keyring integration for `.ai/config.json` | Medium | Week 2-3 |
| **AU-9: Evidence Signing** | Add SHA256 signature to all security evidence files | Medium | Week 2-3 |
| **AC-6: MCP Privilege Reduction** | Review and minimize MCP server permissions | Low | Week 2 |

### 7.2 Short-Term Actions (P2 - Medium Priority)

| Control | Implementation | Effort | Timeline |
|---------|---------------|--------|----------|
| **SI-10: AI Output Validation** | Sanitize AI-generated commit messages and code | High | Month 2 |
| **AU-2: Command Audit Logging** | Implement audit log for CLI operations | Medium | Month 2 |
| **SR-11: Tool Verification** | Verify checksums of Trivy, Semgrep, ZAP before execution | Medium | Month 2-3 |
| **CM-2: Configuration Baseline** | Document approved CLI configuration for regulated use | Low | Month 2 |

### 7.3 Long-Term Actions (P3 - Lower Priority)

| Control | Implementation | Effort | Timeline |
|---------|---------------|--------|----------|
| **SI-7: Binary Signing** | Implement code signing for CLI releases | High | Quarter 2 |
| **AC-2: RBAC for Commands** | Restrict sensitive commands by user role | High | Quarter 2 |
| **AU-10: Non-repudiation** | Blockchain or immutable log for audit evidence | Very High | Quarter 3 |

---

## 8. Residual Risk

After implementation of recommended controls:

| Risk ID | Current Risk | Residual Risk | Acceptance |
|---------|-------------|---------------|------------|
| RISK-1 | HIGH | LOW | Acceptable with SBOM + scanning |
| RISK-2 | HIGH | LOW | Acceptable with keyring integration |
| RISK-3 | MEDIUM | MEDIUM | Acceptable with user training |
| RISK-4 | MEDIUM | LOW | Acceptable with signing |
| RISK-5 | MEDIUM | LOW | Acceptable with AC controls |
| RISK-6 | MEDIUM | LOW | Acceptable with verification |
| RISK-7 | HIGH | MEDIUM | Acceptable with audit logging |
| RISK-8 | HIGH | LOW | Acceptable with continuous scanning |
| RISK-9 | HIGH | LOW | Acceptable with encryption |
| RISK-10 | MEDIUM | LOW | Acceptable with validation |

**Overall Residual Risk**: **LOW-MEDIUM** (acceptable for dev tool in regulated environment)

---

## 9. Compliance Mapping

### 9.1 SOC 2 Type II

- **CC6.1** (Logical Access) → AC-2, AC-3, AC-6, IA-5
- **CC6.6** (Encryption) → SC-8, SC-28
- **CC7.2** (Monitoring) → AU-2, SI-4
- **CC8.1** (Change Management) → CM-2, CM-3

### 9.2 HIPAA (Developer Tools Context)

While the CLI doesn't directly handle PHI, development tools must be secured:

- **§164.308(a)(3)** (Workforce Security) → AC-2, AC-3
- **§164.308(a)(5)** (Security Awareness) → PROC-4
- **§164.312(a)(1)** (Access Control) → AC-6, IA-5
- **§164.312(c)(1)** (Integrity) → AU-9, SI-7

### 9.3 FedRAMP Low

Applicable if CLI is used in government systems development:

- **AC-2, AC-3, AC-6** (Access Control Family)
- **AU-2, AU-3, AU-9** (Audit and Accountability Family)
- **SI-2, SI-3, SI-7** (System Integrity Family)
- **SR-3, SR-4, SR-5** (Supply Chain Risk Management)

---

## 10. Conclusions and Recommendations

### 10.1 Key Findings

1. **Acceptable Risk Profile**: With controls implemented, CLI is suitable for regulated environments
2. **Supply Chain is Top Risk**: Dependency and tool compromise pose greatest threat
3. **Credential Management Critical**: GITHUB_TOKEN and AI keys must be protected
4. **AI Features Require Caution**: Output validation essential to prevent injection attacks
5. **Audit Trail Necessary**: Evidence integrity critical for compliance

### 10.2 Recommendations

**For Immediate Deployment:**

1. Implement P1 controls (SBOM, scanning, secrets detection, credential security)
2. Establish baseline CLI configuration for regulated use
3. Conduct developer security training on CLI usage
4. Deploy with read-only mode for initial evaluation

**For Production Use:**

1. Complete P2 controls (AI validation, audit logging, tool verification)
2. Establish incident response plan for CLI compromise
3. Integrate with organizational SIEM for monitoring
4. Regular (quarterly) security reviews

**For Continuous Improvement:**

1. Track emerging AI security threats and update controls
2. Participate in Go security community (vulnerability disclosure)
3. Contribute security improvements upstream to open-source dependencies
4. Annual penetration testing of CLI in simulated attack scenarios

---

## 11. Approval and Sign-Off

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Risk Assessor | [Name] | _______________ | 2025-12-04 |
| Security Architect | [Name] | _______________ | __________ |
| Engineering Manager | [Name] | _______________ | __________ |
| Compliance Officer | [Name] | _______________ | __________ |

**Next Review Date**: 2026-06-04 (6 months)

---

## Appendix A: References

- NIST SP 800-53 Rev5: Security and Privacy Controls
- NIST SP 800-218: Secure Software Development Framework (SSDF)
- OWASP Top 10 for Large Language Model Applications
- CIS Software Supply Chain Security Guide
- CISA Software Bill of Materials (SBOM) Guidance
