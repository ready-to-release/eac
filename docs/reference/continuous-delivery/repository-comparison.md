<!-- EDITOR
# Editor: reference/continuous-delivery/repository-comparison.md

## Soul

Monorepo vs Polyrepo comparison across atomic changes, team autonomy, build times, and deployment independence.

## Sections

1. Factor Comparison
2. Decision Factors
3. CD Model Stage Impact
4. Polyrepo Coordination Requirements
5. Related
-->

# Repository Pattern Comparison

Side-by-side comparison of Monorepo vs Polyrepo patterns.

## Factor Comparison

| Factor                     | Monorepo                      | Polyrepo                    |
| -------------------------- | ----------------------------- | --------------------------- |
| **Atomic Changes**         | ✅ Excellent - single commit  | ❌ Difficult - multiple PRs |
| **Team Autonomy**          | ⚠️ Limited - shared decisions | ✅ Excellent - independent  |
| **Build Times**            | ⚠️ Potentially long           | ✅ Fast per repository      |
| **Dependency Management**  | ✅ Simple - unified           | ⚠️ Complex - versioned      |
| **Code Reuse**             | ✅ Easy - shared directly     | ⚠️ Requires versioning      |
| **Access Control**         | ⚠️ Coarse-grained             | ✅ Fine-grained             |
| **Discoverability**        | ✅ All code in one place      | ⚠️ Spread across repos      |
| **Independent Deployment** | ⚠️ Coordinated releases       | ✅ Independent cycles       |
| **Tooling**                | ✅ Unified                    | ⚠️ Duplicated               |

## Decision Factors

**Choose Monorepo if:**

- Services frequently change together
- Heavy code sharing between modules
- Small to medium team size
- Unified ownership and responsibility
- Need atomic cross-cutting changes

**Choose Polyrepo if:**

- Services have independent lifecycles
- Multiple teams with separate ownership
- Need fine-grained access control
- Services deploy on different schedules
- Clear service boundaries exist

## CD Model Stage Impact

| Stage        | Monorepo                                | Polyrepo                                   |
| ------------ | --------------------------------------- | ------------------------------------------ |
| **Stage 3**  | Larger PRs, single review               | Smaller focused PRs, may need coordination |
| **Stage 4**  | Change detection for selective builds   | Independent builds in parallel             |
| **Stage 5**  | Single PLTE with all services           | Version pinning, contract testing          |
| **Stage 8**  | Single orchestrated release event       | Multiple independent release tags          |
| **Stage 10** | Coordinated deployment or feature flags | Independent deployment schedules           |

## Polyrepo Coordination Requirements

- Contract testing for API compatibility
- Version pinning in PLTE
- Deployment sequencing for backward compatibility
- API versioning for gradual rollout

## Related

- [Repository Patterns](../../explanation/continuous-delivery/architecture/repository-patterns.md)
