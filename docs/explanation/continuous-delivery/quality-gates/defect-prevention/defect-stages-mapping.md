# Stage-by-Stage Defect Prevention

## Introduction

The CD Model's 12 stages each prevent specific classes of defects through targeted validation, creating layers of defense that catch issues progressively earlier in the pipeline.

This guide walks through each stage, explaining:

- What defects this stage prevents
- Detection mechanisms and methods
- Time budgets and automation levels
- Best practices for effective prevention
- Common anti-patterns to avoid

Use this guide when designing or implementing CD pipelines to understand where each validation belongs and why.

---

## Stage 1: Authoring - Discovery and Design Defects

**Purpose**: Validate problems before building solutions

**Environment**: Local (DevBox), collaborative tools

**Defects Prevented**:

- **Building the wrong thing** - Features users don't need or won't adopt
- **Solving a problem nobody has** - Addressing assumed vs. actual pain points
- **Correct problem, wrong solution** - Valid problem but ineffective approach
- **Prioritizing wrong work** - High effort, low value work over high-value opportunities
- **Ambiguous requirements** - Unclear, contradictory, or incomplete specifications

**Prevention Mechanisms**:

| Mechanism                  | Defect Category           | Automation | Tools/Methods                                  |
| -------------------------- | ------------------------- | ---------- | ---------------------------------------------- |
| Problem validation         | Product & Discovery       | Manual     | User research, problem briefs, quantified pain |
| Executable specifications  | Knowledge & Communication | Semi-auto  | Gherkin specs, BDD, example mapping            |
| Architecture design review | Product & Discovery       | Manual     | Design docs, prototypes, trade-off analysis    |
| Priority scoring           | Product & Discovery       | Semi-auto  | WSJF, cost of delay, outcome tracking          |

**Best Practices**:

- Validate problems before designing solutions (problem brief before solution brief)
- Write executable specifications (Gherkin `.feature` files) to clarify requirements
- Use Three Amigos (product, dev, test) to surface ambiguity early
- Apply WSJF (Weighted Shortest Job First) for prioritization based on cost of delay
- Create low-fidelity prototypes to validate solutions before building

**Anti-Patterns**:

- Jumping directly to implementation without problem validation
- Vague acceptance criteria ("works well", "looks good")
- Skipping specifications because "it's obvious"
- Prioritizing by HiPPO (Highest Paid Person's Opinion) vs. data

**CD Model Integration**:

Stage 1 sets the foundation. Poor decisions here cascade through all later stages. Use `eac validate specs` to catch specification quality issues before coding begins.

**Related Documentation**: [BDD Fundamentals](../../../specifications/concepts/bdd-fundamentals.md), [Spec Quality Checklist](../../../specifications/quality/spec-quality-checklist.md)

---

## Stage 2: Pre-commit - Code Quality and Security Defects

**Purpose**: Fast validation before committing to version control

**Environment**: DevBox (local) and Build Agents (CI)

**Time Budget**: 5-10 minutes maximum

**Defects Prevented**:

- **Code style violations** - Inconsistent formatting, naming conventions
- **Simple logic errors** - Syntax errors, type mismatches, basic bugs
- **Hardcoded secrets** - API keys, passwords, tokens in code
- **Vulnerable dependencies** - Known CVEs in third-party libraries
- **Untested code paths** - Missing unit test coverage
- **Race conditions** - Basic concurrency issues (via race detector)
- **Null/missing data assumptions** - Null pointer dereferences
- **Long-lived branches** - Stale branches diverging from main
- **Implicit domain knowledge** - Magic numbers, undocumented business rules
- **Over-engineering** - Unnecessary abstractions, premature optimization

**Prevention Mechanisms**:

| Mechanism              | Defect Category             | Time Budget | Automation | Tools                       |
| ---------------------- | --------------------------- | ----------- | ---------- | --------------------------- |
| Code formatting        | Change & Complexity         | < 10s       | Auto-fix   | gofmt, prettier, black      |
| Static analysis (SAST) | Change & Complexity         | 10-60s      | Automated  | golangci-lint, SonarQube    |
| Secret scanning        | Data & State                | < 30s       | Automated  | gitleaks, truffleHog        |
| Dependency scanning    | Dependency & Infrastructure | 30-60s      | Automated  | Dependabot, Snyk, OWASP DC  |
| Unit tests (L0-L1)     | Testing & Observability     | 1-5 min     | Automated  | go test, pytest, jest       |
| Race detection         | Integration & Boundaries    | 1-5 min     | Automated  | go test -race, TSan         |
| Null safety checks     | Data & State                | 10-30s      | Automated  | NullAway, TypeScript strict |
| Complexity scoring     | Change & Complexity         | < 30s       | Automated  | SonarQube, cyclomatic       |

**Best Practices**:

- Keep total time budget under 10 minutes (developers wait for this)
- Fail fast on critical issues (formatting, secrets, critical security)
- Run checks in parallel where possible
- Use auto-fix for mechanical issues (formatting, import sorting)
- Cache dependencies to speed up test execution
- Use incremental testing (only test changed code)

**Anti-Patterns**:

- Running integration tests at Stage 2 (move to Stage 3)
- Manual approval gates (automate or move to Stage 3)
- No time budget enforcement (developers skip pre-commit if too slow)
- Inconsistent local vs. CI validation (CI catches issues pre-commit should)

**CD Model Integration**:

Stage 2 is the first line of defense. If pre-commit validation is too slow or incomplete, defects slip through to Stage 3 where they're more expensive to fix.

**Related Documentation**: [Pre-commit Quality Gates](../precommit-gates.md), [Pre-commit Setup](../precommit-setup.md)

---

## Stage 3: Merge Request - Integration and Design Defects

**Purpose**: Validate integration and design quality before merging to main branch

**Environment**: Build Agents (CI/CD)

**Time Budget**: Minutes to hours (includes human review)

**Defects Prevented**:

- **Interface mismatches** - API breaking changes, schema incompatibilities
- **Integration failures** - Components don't work together
- **Design issues** - Poor abstractions, tight coupling, missing contracts
- **Ambiguous requirements** - Unresolved edge cases, missing test scenarios
- **Divergent mental models** - Terminology mismatches across teams
- **Unintended side effects** - Changes break existing functionality
- **Tribal knowledge loss** - Knowledge concentrated in single developer
- **Reliance on human review for preventable defects** - Catching issues that should be automated

**Prevention Mechanisms**:

| Mechanism                | Defect Category           | Time Budget | Automation | Tools/Methods                       |
| ------------------------ | ------------------------- | ----------- | ---------- | ----------------------------------- |
| Contract tests           | Integration & Boundaries  | 2-10 min    | Automated  | Pact, Spring Cloud Contract         |
| Integration tests (L2)   | Integration & Boundaries  | 5-30 min    | Automated  | API tests, component integration    |
| Schema validation        | Integration & Boundaries  | < 1 min     | Automated  | OpenAPI, protobuf/buf, JSON Schema  |
| Code coverage thresholds | Testing & Observability   | 1-5 min     | Automated  | 80% line coverage, 70% branch       |
| Peer review              | Knowledge & Communication | Hours       | Manual     | GitHub PR, GitLab MR, code review   |
| Architecture review      | Product & Discovery       | Hours       | Manual     | Design docs, ADRs, trade-off review |
| AI-assisted code review  | Process & Deployment      | < 5 min     | AI-powered | GPT-4, Claude, GitHub Copilot       |
| Mutation testing         | Testing & Observability   | 10-60 min   | Automated  | Stryker, PIT (optional at Stage 3)  |

**Best Practices**:

- Require contract tests for all service boundaries
- Enforce code coverage thresholds (minimum 80% line, 70% branch)
- Use AI-assisted review for semantic issues (logic errors, missing edge cases)
- Reserve human review for design decisions and knowledge transfer
- Automate all mechanical checks (formatting already done in Stage 2)
- Run Stage 2 checks again in CI (trust but verify)

**Anti-Patterns**:

- Manual review as primary defect detection (should be automated)
- "LGTM" approvals without criteria
- Merging with failing tests or low coverage
- Skipping contract tests ("we'll add them later")
- Reviewing code without running it locally

**CD Model Integration**:

Stage 3 is the **first-level approval gate**. In RA pattern, it approves code quality. In CDe pattern, it approves both code quality AND production deployment (second-level approval combined).

**Related Documentation**: [Merge Request Quality Gates](../merge-request-gates.md), [Trunk-Based Development](../../workflow/trunk-based-development.md)

---

## Stage 4: Commit - CI Verification

**Purpose**: Re-verify Stage 2 checks in CI environment

**Environment**: Build Agents (CI)

**Time Budget**: 5-10 minutes

**Defects Prevented**:

- **Environment-specific issues** - Works locally but not in CI
- **Dependency conflicts** - Version mismatches, missing dependencies
- **Build reproducibility** - Non-deterministic builds

**Prevention Mechanisms**:

Stage 4 repeats Stage 2 checks in the CI environment to catch:

- Issues that pass locally but fail in standardized CI environment
- Dependency drift (local cache vs. fresh install)
- Platform-specific issues (Windows dev, Linux CI)

**Best Practices**:

- Use same tools/versions as Stage 2 (consistency)
- Cache dependencies for faster builds
- Fail fast if Stage 4 finds issues Stage 2 missed (indicates pre-commit problem)

**Anti-Patterns**:

- Adding new checks at Stage 4 that weren't in Stage 2 (surprises developers)
- Skipping Stage 4 ("we already ran pre-commit")

**CD Model Integration**:

Stage 4 is a safety net. Ideally, it should always pass if Stage 2 passed. Frequent Stage 4 failures indicate pre-commit validation gaps.

---

## Stage 5: Acceptance Testing - Functional Defects

**Purpose**: Validate functional requirements in production-like environment

**Environment**: PLTE (Production-Like Test Environment)

**Time Budget**: minutes-1 hour

**Defects Prevented**:

- **Functional defects** - Features don't meet acceptance criteria
- **Missing contract tests** - Boundaries lack validation
- **Wrong assumptions about upstream/downstream** - Behavioral contract violations
- **Inconsistent distributed state** - Eventual consistency issues
- **Schema migration failures** - Backward compatibility breaks
- **Concurrency and ordering issues** - Idempotency violations
- **Meets spec but misses user intent** - Technically correct but poor UX

**Prevention Mechanisms**:

| Mechanism                       | Defect Category          | Time Budget | Automation | Tools/Methods                     |
| ------------------------------- | ------------------------ | ----------- | ---------- | --------------------------------- |
| Acceptance tests (L3-L4)        | Testing & Observability  | 15-60 min   | Automated  | BDD scenarios, API tests, E2E     |
| Contract tests (all boundaries) | Integration & Boundaries | 5-30 min    | Automated  | Pact broker, schema registry      |
| Migration dry-runs              | Data & State             | 5-20 min    | Automated  | Test against prod-like data       |
| Chaos experiments               | Integration & Boundaries | 10-30 min   | Automated  | Fault injection, network delays   |
| Idempotency verification        | Data & State             | 5-15 min    | Automated  | Replay tests, duplicate detection |

**Best Practices**:

- Use production-like data (anonymized/synthetic)
- Test all IV (Installation Verification) and OV (Operational Verification) scenarios
- Validate backward compatibility for migrations (expand-then-contract pattern)
- Test behavioral contracts, not just schemas (timeouts, retries, error semantics)
- Run chaos experiments for critical paths (circuit breakers, retries)

**Anti-Patterns**:

- Using toy data that doesn't reflect production complexity
- Skipping contract tests because "we own both sides"
- Testing only happy path (missing error paths and edge cases)
- Assuming Stage 5 will catch everything (Stage 2-3 should catch most issues)

**Example EAC Command**:

```bash
eac test acceptance --environment plte
```

**CD Model Integration**:

Stage 5 validates functional correctness in production-like conditions. This is where IV (Installation Verification) and OV (Operational Verification) tests run.

**Related Documentation**: [Test Levels](../../testing/test-levels.md), [Verification Types](../../testing/verification-types.md)

---

## Stage 6: Extended Testing - Non-Functional Defects

**Purpose**: Validate performance, security, and compliance

**Environment**: PLTE (Production-Like Test Environment)

**Time Budget**: 30 minutes to several hours

**Defects Prevented**:

- **Performance regressions** - Slow queries, memory leaks, high latency
- **Security vulnerabilities** - OWASP Top 10, CVEs, misconfigurations
- **Compliance violations** - Regulatory requirements, audit trail gaps
- **Insufficient monitoring** - Missing observability, no alerts
- **Configuration drift** - IaC vs. actual infrastructure mismatches
- **Unanticipated feature interactions** - Features conflict or degrade
- **Infrastructure differences across environments** - Environment parity issues
- **Network partition handling** - Missing circuit breakers, no retries

**Prevention Mechanisms**:

| Mechanism                | Defect Category             | Time Budget | Automation | Tools/Methods                    |
| ------------------------ | --------------------------- | ----------- | ---------- | -------------------------------- |
| Performance tests (L5)   | Testing & Observability     | 30-120 min  | Automated  | Load tests, stress tests, soak   |
| Security scans (DAST)    | Change & Complexity         | 15-60 min   | Automated  | OWASP ZAP, Burp Suite            |
| Penetration tests (PV)   | Testing & Observability     | Hours       | Manual     | Ethical hacking, red team        |
| Compliance validation    | Testing & Observability     | 10-30 min   | Automated  | Evidence checks, audit trail     |
| IaC drift detection      | Change & Complexity         | 5-15 min    | Automated  | Terraform plan, Pulumi preview   |
| Chaos engineering        | Dependency & Infrastructure | 20-60 min   | Automated  | Gremlin, Litmus, fault injection |
| Observability validation | Testing & Observability     | 5-15 min    | Automated  | Health checks, SLO burn rate     |

**Best Practices**:

- Set performance regression thresholds (e.g., < 5% latency increase)
- Run DAST against deployed application in PLTE
- Validate observability coverage (logs, metrics, traces for all critical paths)
- Test infrastructure as code drift detection
- Run controlled chaos experiments (network partitions, pod failures)
- Verify compliance evidence is being collected

**Anti-Patterns**:

- Running performance tests only before major releases (should be continuous)
- Manual compliance checks (automate evidence collection)
- Skipping observability validation until production issues occur
- Treating Stage 6 as optional ("we'll do it later")

**CD Model Integration**:

Stage 6 completes the validation suite. Combined with Stage 5, it provides comprehensive evidence for Stage 9 (Release Approval).

**Related Documentation**: [Security Integration](../../security/index.md), [DAST](../../security/dast.md)

---

## Stage 7: Exploration - Usability and Experience Defects

**Purpose**: Stakeholder validation and exploratory testing

**Environment**: Demo (production-like or actual production with feature flags off)

**Time Budget**: Hours to days (not blocking)

**Defects Prevented**:

- **Meets spec but misses user intent** - Technically correct but poor UX
- **Correct problem, wrong solution** - Solution doesn't effectively solve problem
- **Unanticipated feature interactions** - User workflows break or confuse

**Prevention Mechanisms**:

| Mechanism           | Defect Category         | Time Budget | Automation | Tools/Methods                     |
| ------------------- | ----------------------- | ----------- | ---------- | --------------------------------- |
| Exploratory testing | Testing & Observability | Hours       | Manual     | Ad-hoc testing, user journeys     |
| Stakeholder demos   | Product & Discovery     | Hours       | Manual     | Sprint reviews, user validation   |
| UX analytics        | Product & Discovery     | Ongoing     | Automated  | FullStory, Hotjar, session replay |

**Best Practices**:

- Use real users or product owners, not just developers
- Test realistic user journeys, not just individual features
- Collect feedback but don't block pipeline (Stage 7 runs in parallel with Stage 8-12)
- Use feature flags to expose features gradually

**Anti-Patterns**:

- Treating Stage 7 as a gate (it's exploratory, not blocking)
- Skipping user validation ("we know what they want")
- Waiting until Stage 7 to get first user feedback (should happen at Stage 1)

**CD Model Integration**:

Stage 7 provides qualitative feedback that informs future iterations. It doesn't block releases but surfaces usability issues early enough to fix cheaply.

**Related Documentation**: [Exploration Stage](../../cd-model/stages.md#stage-7-exploration)

---

## Stage 8: Start Release - Completeness Defects

**Purpose**: Create release candidate and verify completeness

**Environment**: Build Agents (CI)

**Time Budget**: 5-15 minutes

**Defects Prevented**:

- **Incomplete releases** - Missing artifacts, dependencies, configuration
- **Missing documentation** - Changelog, release notes, deployment guide
- **Tagging errors** - Incorrect version numbers, missing tags

**Prevention Mechanisms**:

| Mechanism                   | Defect Category      | Time Budget | Automation | Tools/Methods                     |
| --------------------------- | -------------------- | ----------- | ---------- | --------------------------------- |
| Release artifact validation | Process & Deployment | < 5 min     | Automated  | Checksums, signature verification |
| Changelog generation        | Process & Deployment | < 5 min     | Automated  | Conventional commits, git log     |
| Version tagging             | Process & Deployment | < 1 min     | Automated  | Semantic versioning, git tag      |

**Best Practices**:

- Generate changelogs automatically from conventional commits
- Validate all artifacts are signed and checksummed
- Use semantic versioning (MAJOR.MINOR.PATCH)
- Tag releases consistently (v1.2.3, not "prod-release-jan")

**Anti-Patterns**:

- Manual changelog writing (error-prone, time-consuming)
- Skipping version tagging
- Inconsistent release artifact structure

**CD Model Integration**:

Stage 8 creates the release candidate that Stage 9 will approve (RA) or auto-approve (CDe).

---

## Stage 9: Release Approval - Production Readiness Defects

**Purpose**: Validate production readiness before deployment

**Environment**: PLTE (evidence) + Demo (exploratory validation)

**Time Budget**: RA pattern (hours to days), CDe pattern (seconds - automated)

**Defects Prevented**:

- **Incomplete testing** - Missing test coverage, unevaluated risks
- **Production readiness gaps** - No monitoring, no rollback plan, no runbook
- **Manual CAB delays** - Gatekeeping without evidence-based criteria
- **Inadequate rollback capability** - Can't roll back safely if issues occur

**Prevention Mechanisms**:

| Mechanism                      | Defect Category         | Time Budget | Automation | Pattern     |
| ------------------------------ | ----------------------- | ----------- | ---------- | ----------- |
| Evidence validation            | Testing & Observability | < 5 min     | Automated  | Both RA/CDe |
| Risk scoring                   | Process & Deployment    | < 1 min     | AI-powered | CDe         |
| Manual approval                | Process & Deployment    | Hours-days  | Manual     | RA only     |
| Production readiness checklist | Testing & Observability | 5-15 min    | Automated  | Both RA/CDe |

**Best Practices (CDe Pattern - Automated)**:

- Define objective quality thresholds (100% test pass, < 5% perf regression, zero critical CVEs)
- Automate evidence collection from Stages 5-6
- Use AI-assisted risk scoring based on change diff and deployment history
- Auto-approve low-risk changes, flag high-risk for human review

**Best Practices (RA Pattern - Manual Approval)**:

- Provide evidence dashboard to approvers (test results, security scans, performance)
- Document approval criteria (not "gut feel")
- Time-box approval (SLA: 24 hours for low-risk, 48 hours for high-risk)
- Track approval bottlenecks and improve criteria

**Anti-Patterns**:

- Manual approval without evidence (CAB theater)
- No objective criteria ("we'll know it when we see it")
- Approval delays blocking all releases (should have automated fast track for low-risk)
- Treating all changes equally (low-risk config vs. high-risk architecture change)

**CD Model Integration**:

Stage 9 is the **second-level approval gate** (RA pattern only). In CDe pattern, Stage 3 merge approval implicitly approves production deployment.

**Related Documentation**: [Release Quality Gates](../release-gates.md), [Release Approval Patterns](../../release-management/approval-patterns.md)

---

## Stage 10: Production Deployment - Deployment Defects

**Purpose**: Deploy to production safely

**Environment**: Production

**Time Budget**: 5-30 minutes

**Defects Prevented**:

- **Failed deployments** - Rollout errors, dependency issues, startup failures
- **Inadequate rollback capability** - Can't roll back quickly if needed
- **Configuration drift** - Deployed config differs from IaC

**Prevention Mechanisms**:

| Mechanism               | Defect Category         | Time Budget | Automation | Tools/Methods                      |
| ----------------------- | ----------------------- | ----------- | ---------- | ---------------------------------- |
| Blue/green deployment   | Process & Deployment    | 5-15 min    | Automated  | Traffic switching, health checks   |
| Canary deployment       | Process & Deployment    | 15-60 min   | Automated  | Progressive rollout, auto-rollback |
| Health check validation | Testing & Observability | 1-5 min     | Automated  | Liveness, readiness probes         |
| Automated rollback      | Process & Deployment    | 2-10 min    | Automated  | Revert on health failure           |

**Best Practices**:

- Use blue/green or canary as default (not big-bang deploys)
- Validate health checks pass before declaring success
- Auto-rollback on health check failures
- Practice rollback regularly (monthly game days)
- Use backward-compatible migrations only (expand-then-contract)

**Anti-Patterns**:

- Big-bang deployments (all servers at once)
- No automated rollback (manual rollback takes hours)
- Ignoring health check failures ("we'll monitor it")
- Deploying during peak traffic (increases blast radius)

**CD Model Integration**:

Stage 10 is where deployment happens. Decoupled from release (Stage 12) via feature flags.

**Related Documentation**: [Deployment Strategies](../../deployment/deployment-strategies.md), [Rollback Procedures](../../deployment/rollback-procedures.md)

---

## Stage 11: Live - Monitoring and Observability Defects

**Purpose**: Monitor production behavior and validate SLOs

**Environment**: Production

**Time Budget**: Ongoing (minutes to hours for initial validation)

**Defects Prevented**:

- **Insufficient monitoring** - Can't detect production issues
- **Missed production incidents** - Issues occur but aren't detected
- **Building the wrong thing** - Low adoption signals feature isn't needed

**Prevention Mechanisms**:

| Mechanism            | Defect Category          | Time Budget | Automation | Tools/Methods                      |
| -------------------- | ------------------------ | ----------- | ---------- | ---------------------------------- |
| SLO burn rate alerts | Testing & Observability  | Real-time   | Automated  | Prometheus, Datadog, New Relic     |
| Anomaly detection    | Testing & Observability  | Real-time   | AI-powered | Threshold alerts, ML-based         |
| Distributed tracing  | Integration & Boundaries | Real-time   | Automated  | Jaeger, Zipkin, OpenTelemetry      |
| Adoption dashboards  | Product & Discovery      | Daily       | Automated  | Amplitude, Mixpanel, usage metrics |

**Best Practices**:

- Define SLOs for every user-facing path (latency, availability, error rate)
- Alert on burn rate (20% budget consumed in 1 hour = critical)
- Use distributed tracing to diagnose issues across services
- Monitor adoption metrics to validate feature value

**Anti-Patterns**:

- Alerting on metrics without SLOs (noise, alert fatigue)
- No tracing for distributed systems (blind to cross-service issues)
- Monitoring infrastructure but not user experience
- Not tracking adoption (can't tell if features are valuable)

**CD Model Integration**:

Stage 11 provides continuous feedback on production behavior. Anomalies trigger incident response. Low adoption triggers Stage 1 re-evaluation.

**Related Documentation**: [Incident Response](../../deployment/incident-response.md), [Live Stage](../../cd-model/stages.md#stage-11-live)

---

## Stage 12: Release Toggling - Feature Exposure Defects

**Purpose**: Control feature exposure independently of deployment

**Environment**: Production (features deployed but hidden)

**Time Budget**: Seconds to weeks (gradual rollout)

**Defects Prevented**:

- **Unanticipated feature interactions** - Features conflict when enabled
- **Building the wrong thing** - Can kill features quickly if not adopted
- **Meets spec but misses user intent** - A/B testing shows better alternative

**Prevention Mechanisms**:

| Mechanism           | Defect Category      | Time Budget | Automation | Tools/Methods                 |
| ------------------- | -------------------- | ----------- | ---------- | ----------------------------- |
| Feature flags       | Process & Deployment | Real-time   | Automated  | LaunchDarkly, Split, Unleash  |
| Progressive rollout | Process & Deployment | Hours-weeks | Automated  | Percentage-based, ring-based  |
| A/B testing         | Product & Discovery  | Days-weeks  | Automated  | Optimizely, statistical tests |
| Kill switch         | Product & Discovery  | Seconds     | Manual     | Instant flag disable          |

**Best Practices**:

- Use feature flags for all user-facing changes
- Roll out progressively (1% → 10% → 50% → 100%)
- A/B test alternative solutions
- Kill features that miss adoption thresholds (< 10% after 30 days)
- Monitor feature flag interaction matrix (combinatorial testing)

**Anti-Patterns**:

- Deploying directly to 100% of users (no progressive exposure)
- No kill switch (can't disable quickly if issues occur)
- Leaving flags in code forever (technical debt)
- Not monitoring flag interactions (feature A + B breaks)

**CD Model Integration**:

Stage 12 is the **third-level approval gate** (feature owner controls exposure). It decouples deployment (Stage 10) from release (Stage 12).

**Related Documentation**: [Feature Flags](../../release-toggling/feature-flags.md), [Progressive Exposure](../../release-toggling/progressive-exposure.md)

---

## Next Steps

- **Addressing a specific defect?** See [External Defect Catalog](https://bdfinst.github.io/ai-patterns/defect-detection-and-fixes/) for detailed lookup
- **Exploring AI detection?** Read [AI-Assisted Detection Strategies](ai-detection.md)
- **Implementing quality gates?** Start with [Pre-commit Quality Gates](../precommit-gates.md)
- **Understanding the CD Model?** Review [The 12 Stages](../../cd-model/stages.md)
