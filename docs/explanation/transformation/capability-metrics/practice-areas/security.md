# Security & Risk Management

Automated security scanning shifts security left, catching vulnerabilities during development rather than production—reducing risk and enabling faster, safer delivery.

**Impact**: Vulnerabilities found during development (not production), reduced MTTR for security issues, compliance with standards (OWASP, CWE), reduced incidents, faster threat response.

---

## Level 1: Initial

**Manual security reviews.**

- Manual security reviews (late-stage, pre-production)
- No automated vulnerability scanning
- Security issues found in pre-production or production
- MTTR for critical issues: weeks to months
- No security checks in CI/CD
- Reactive security posture

**Advancing to Level 2**: Implement SAST (Static Application Security Testing) in CI, set up basic dependency scanning, configure security checks to run automatically, train team on security basics and common vulnerabilities, establish security issue tracking.

**Resources**: [Security Overview](../../../continuous-delivery/security/index.md) · [Shift-Left Security](../../../continuous-delivery/security/shift-left.md)

---

## Level 2: Managed

**SAST in pipeline.**

- SAST runs automatically in CI on every commit
- Basic dependency vulnerability scanning (SCA)
- Security issues found during development (shift-left begins)
- Some security checks in pipeline (not comprehensive)
- Security findings tracked and prioritized
- MTTR improving but not fast (days to weeks)

**Advancing to Level 3**: Add DAST (Dynamic Application Security Testing), implement secrets detection, add comprehensive SCA (all dependencies including transitive), implement security quality gates (block on critical issues), standardize security practices organization-wide.

**Resources**: [DAST](../../../continuous-delivery/security/dast.md) · [Supply Chain Security](../../../continuous-delivery/security/supply-chain.md) · [Quality Gates](../../../continuous-delivery/quality-gates/index.md)

---

## Level 3: Defined

**Comprehensive security automation.**

- Comprehensive scanning: SAST, DAST, SCA, secrets detection
- Security quality gates enforce standards (block on critical issues)
- Automated security controls in pipeline
- Standardized security practices organization-wide
- Container/infrastructure scanning (if applicable)
- Security compliance automated (OWASP Top 10, CWE)
- MTTR: hours to days for critical issues

**Advancing to Level 4**: Implement security metrics tracking (MTTD, MTTR, vulnerability trends), apply risk-based prioritization (business context), track statistical process control, predict security issues based on patterns, correlate findings with business impact.

**Resources**: [Security Remediation](../../../continuous-delivery/security/remediation.md) · [SAST](../../../continuous-delivery/security/sast.md) · [Measuring Flow](../../../everything-as-code/measuring-and-improving-flow.md)

---

## Level 4: Quantified

**Security metrics and risk-based prioritization.**

- Security metrics tracked comprehensively (MTTD, MTTR, trends)
- Risk-based prioritization (business context and impact)
- Statistical process control applied
- Can predict security issues based on code patterns
- Correlation analysis (vulnerability types vs business impact)
- Data-driven security improvements
- MTTD <1 hour, MTTR <24 hours for critical

**Advancing to Level 5**: Implement proactive threat detection (threat modeling, predictions), continuous security process optimization (A/B testing), experiment with security approaches, share innovations with industry.

**Resources**: [Measuring Flow](../../../everything-as-code/measuring-and-improving-flow.md)

---

## Level 5: Optimizing

**Proactive threat detection.**

- Proactive threat detection (threat modeling, predictive analysis)
- Continuous security process optimization (quarterly experiments)
- AI-assisted vulnerability detection and prioritization
- Industry leadership in security practices
- Community contributions (talks, papers, open source)
- Innovation in security automation
- MTTR near-zero for predicted threats

**Maintaining**: Stay current with security research and emerging threats, active community participation, regular experimentation with measurement, share learnings.

---

## Level Assessment

**You're at a level when**:

- ✅ All characteristics consistently demonstrated organization-wide
- ✅ Capabilities are sustainable (not dependent on heroes)
- ✅ You possess the capability, not just working toward it

**Level Distinctions**:

- **1 → 2**: SAST in pipeline (capability exists)
- **2 → 3**: Comprehensive security suite with automated enforcement
- **3 → 4**: Measure effectiveness, use risk-based prioritization
- **4 → 5**: Proactive threat detection, continuous innovation

**Dependencies**:

- **Depends on**: Version Control Level 2+, CI/CD Level 2+
- **Enables**: Evidence Level 2+, all practices benefit from security integration
- **Blocks**: If weak (Level 1), vulnerabilities undetected until late stages
