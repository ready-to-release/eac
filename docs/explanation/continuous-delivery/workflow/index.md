# Workflow Practices

Practical day-to-day development workflow practices that enable Continuous Integration and Continuous Delivery. These articles explain how to work with code, branches, and releases in a trunk-based development model.

---

## [Trunk-Based Development](trunk-based-development.md)

Comprehensive guide to trunk-based development practices enabling Continuous Integration and Continuous Delivery.

Trunk-Based Development (TBD) is the branching strategy that makes Continuous Delivery possible. Instead of long-lived feature branches, TBD emphasizes frequent integration to a single main branch (trunk), enabling rapid feedback and reducing merge conflicts.

**Core Principles:**

1. **Single Source of Truth**: There is only ever one meaningful version of the code (trunk)
2. **Do Not Branch (Or Branch Very Briefly)**: Topic branches live for hours to 2 days maximum
3. **Small, Incremental Changes**: Work in small batches (< 400 lines)
4. **Continuous Integration**: Integrate at least daily, preferably multiple times

**Topics covered:**

- Three branch types: Trunk, topic branches, release branches
- Integration with CD Model stages (1-12)
- Daily development flow (7-step guide)
- Commits and squash merging (one topic branch = one trunk commit)
- Feature hiding strategies: code-level, configuration, feature flags
- Release flows for RA and CDE patterns
- Cherry-picking fixes between branches
- Best practices (DO/DON'T lists)
- Emergency fixes and conflict resolution
- Timeline expectations

**Read this article to understand**: How to work day-to-day with trunk-based development and maintain a clean, releasable trunk.

---

## [Branching Strategies](branching-strategies.md)

Detailed branching flows for Release Approval and Continuous Deployment patterns.

While Trunk-Based Development establishes core principles, this article provides detailed stage-by-stage branching flows for the two implementation patterns.

**Release Approval (RA) Pattern:**

- Uses release branches created at Stage 8
- Topic branches → Trunk → Release branches → Production
- Fixes applied to trunk first, cherry-picked to release
- Suitable for regulated systems
- **Cycle time**: 1-2 weeks

**Continuous Deployment (CDE) Pattern:**

- No release branches (deploy directly from trunk)
- Topic branches → Trunk → Production
- Feature flags provide runtime control
- Suitable for non-regulated systems
- **Cycle time**: 2-4 hours

**Topics covered:**

- Stage-by-stage flow for each pattern (Stages 1-12)
- Release branch lifecycle and management
- Fixing bugs on release branches vs trunk
- Pipeline integration and separation
- Pin and stitch dependency management
- Timeline and best practices comparison
- When to use each pattern

**Read this article to understand**: Detailed branching flows and how code progresses through each CD Model stage in RA and CDE patterns.

---

## How These Articles Work Together

**Trunk-Based Development** establishes:

- Core principles that apply to all patterns
- Branch types and their characteristics
- Daily development practices
- Feature hiding approaches

**Branching Strategies** provides:

- Detailed flows for RA and CDE patterns
- Stage-by-stage progression
- Pattern-specific practices
- Pipeline integration

**Recommended reading order:**

1. Read **Trunk-Based Development** first for principles and practices
2. Read **Branching Strategies** for pattern-specific flows

---

## Integration with CD Model Stages

These workflow practices integrate directly with CD Model stages:

**Stages 1-3: Topic Branch Development**:

- Create topic branch from trunk (Stage 1)
- Run pre-commit checks locally (Stage 2)
- Create merge request for peer review (Stage 3)

**Stage 4: Trunk Integration**:

- Squash-merge topic branch to trunk
- Automated validation (L0-L2 tests, Hybrid E2E)
- Build immutable artifacts

**Stages 5-7: Testing and Validation**:

- Deploy to PLTE (Stage 5)
- Extended testing (Stage 6)
- Exploratory testing in Demo (Stage 7)

**Stages 8-10: Release (Pattern Dependent)**:

- **RA**: Create release branch (Stage 8), manual approval (Stage 9)
- **CDE**: Tag trunk commit (Stage 8), automated approval (Stage 9)
- Deploy to production (Stage 10)

**Stages 11-12: Production**:

- Monitor live environment (Stage 11)
- Feature flag management (Stage 12)

See **[CD Model](../cd-model/index.md)** for complete stage details.

---

## Integration with Other Sections

**[Core Concepts](../core-concepts/index.md)**:

- Trunk is one of the four Unit of Flow components
- Deployable Units are built from trunk
- Understanding these concepts helps contextualize workflow

**[Testing](../testing/index.md)**:

- Test levels (L0-L4) execute at specific workflow stages
- Pre-commit runs L0/L1 (Stage 2)
- Merge request adds L2 (Stage 3)
- Commit stage runs comprehensive suite (Stage 4)

**[Architecture](../architecture/index.md)**:

- Repository patterns (monorepo vs polyrepo) affect branching
- Environment architecture supports workflow stages

**[Security](../security/index.md)**:

- Security scanning integrated throughout workflow
- Pre-commit includes secret scanning (Stage 2)
- Dependency scanning in commit stage (Stage 4)

---

## Best Practices Summary

**For All Developers:**

✅ **DO:**

- Integrate to trunk at least daily
- Keep changes small (< 400 lines)
- Run Stage 2 checks before every push
- Write semantic commit messages
- Pull from trunk frequently (every few hours)
- Delete topic branches immediately after merge

❌ **DON'T:**

- Keep topic branches open > 2 days
- Make large, sweeping changes in single commit
- Push broken code to trunk
- Keep "work in progress" branches
- Create long-lived feature branches

**For RA Pattern:**

✅ **DO:**

- Always fix on trunk first, cherry-pick to release
- Document approvals and evidence
- Archive release branches when superseded

❌ **DON'T:**

- Fix on release branch without cherry-picking to trunk
- Allow non-critical changes on release branches

**For CDE Pattern:**

✅ **DO:**

- Use feature flags for all incomplete features
- Implement kill switches for high-risk features
- Monitor deployment health continuously
- Practice rollback procedures regularly

❌ **DON'T:**

- Deploy without comprehensive automated tests
- Skip feature flags for risky changes

---

## Next Steps

- **New to TBD?** Start with [Trunk-Based Development](trunk-based-development.md)
- **Need pattern details?** Read [Branching Strategies](branching-strategies.md)
- **Want to understand stages?** See [CD Model](../cd-model/index.md)
- **Ready to implement testing?** Explore [Testing](../testing/index.md)

Return to **[Continuous Delivery Overview](../index.md)** for complete navigation.
