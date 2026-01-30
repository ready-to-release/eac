# Continuous Integration & Delivery

CI/CD automation enables fast feedback, reduces manual errors, and allows on-demand deployment—the technical foundation for delivering changes safely and rapidly.

**Impact**: Deployment time reduced from days to minutes, reduced errors, fast feedback (minutes vs hours), on-demand deployment capability, rapid response to issues.

---

## Level 1: Initial

**Manual builds and deployments.**

- No automated builds
- Manual deployments with runbooks or scripts
- Deployment time: hours to days
- Build process varies by team/developer
- No pipeline, no quality gates

**Advancing to Level 2**: Set up CI server (GitHub Actions, GitLab CI, Jenkins), automate builds on every commit, create basic automated test suite for CI, write deployment scripts, establish quality checks in pipeline.

**Resources**: [CD Model Overview](../../../continuous-delivery/cd-model/overview.md) · [CD Model Stages](../../../continuous-delivery/cd-model/stages.md)

---

## Level 2: Managed

**CI established, semi-automated deployment.**

- Builds run automatically on every commit
- Automated tests run in CI (basic suite)
- Deployment semi-automated (scripts + manual gates)
- Some quality checks in pipeline
- Build failures visible immediately
- Deployment requires manual steps/approvals

**Advancing to Level 3**: Automate deployment to all environments, implement comprehensive quality gates (tests, security, compliance), standardize pipeline across teams, remove manual deployment steps, enable on-demand deployment.

**Resources**: [CD Model](../../../continuous-delivery/cd-model/index.md) · [Deployment Strategies](../../../continuous-delivery/deployment/deployment-strategies.md)

---

## Level 3: Defined

**Full CD pipeline with quality gates.**

- Full CD pipeline (commit → production automated)
- Deployment fully automated to all environments
- Comprehensive quality gates (tests, security scans, compliance checks)
- Standardized pipeline across teams
- **Can deploy on-demand** (technical capability exists)
- Zero-downtime deployments, automated rollback

**Advancing to Level 4**: Implement DORA metrics tracking (deployment frequency, lead time, MTTR, change failure rate), collect pipeline metrics, apply statistical process control, implement predictive analytics, create metrics dashboards.

**Resources**: [Measuring Flow](../../../everything-as-code/measuring-and-improving-flow.md)

---

## Level 4: Quantified

**Pipeline metrics and DORA metrics tracked.**

- DORA metrics tracked precisely (deployment freq, lead time, MTTR, CFR)
- Pipeline metrics collected (build time, test time, bottlenecks)
- Statistical process control applied
- Predictive analytics (predict pipeline failures)
- Data-driven improvements with measured outcomes
- Trend analysis on all metrics

**Advancing to Level 5**: Implement self-optimizing pipeline, experiment with pipeline approaches (A/B test strategies), proactive issue detection, share practices with industry.

**Resources**: [Measuring Flow](../../../everything-as-code/measuring-and-improving-flow.md)

---

## Level 5: Optimizing

**Self-optimizing pipeline.**

- Self-optimizing pipeline (automatically adjusts based on patterns)
- Continuous experimentation with pipeline approaches (quarterly A/B tests)
- Proactive failure detection (catch issues before problems)
- Industry-leading practices (others learn from you)
- Community contributions (talks, papers, open source)

**Maintaining**: Stay current with CI/CD research, active community participation, regular experimentation with measurement, share learnings.

---

## Level Assessment

**You're at a level when**:

- ✅ All characteristics consistently demonstrated organization-wide
- ✅ Capabilities are sustainable (not dependent on heroes)
- ✅ You possess the capability, not just working toward it

**Level Distinctions**:

- **1 → 2**: CI with automated builds and tests (capability exists)
- **2 → 3**: Full CD with on-demand deployment capability
- **3 → 4**: Measure effectiveness, track DORA metrics
- **4 → 5**: Self-optimizing, continuous experimentation

**Dependencies**:

- **Depends on**: Version Control Level 2+, Testing Level 2+
- **Enables**: Evidence Level 3+, Security Level 3+
- **Blocks**: If weak (Level 1-2), deployment remains manual and slow
