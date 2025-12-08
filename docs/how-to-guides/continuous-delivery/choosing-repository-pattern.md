<!-- EDITOR
# Editor: how-to-guides/continuous-delivery/choosing-repository-pattern.md

## Soul

Decision checklist for selecting between monorepo and polyrepo patterns based on team structure, code sharing, and access control needs.

## Sections

1. Decision Checklist
2. Question 1: How often do services change together?
3. Question 2: How much code is shared?
4. Question 3: How many teams?
5. Question 4: What access control do you need?
6. Question 5: Do you need atomic cross-service changes?
7. Quick Decision Matrix
8. Pattern Characteristics
9. Monorepo
10. Polyrepo
11. Anti-Pattern: Technical Boundaries
12. Migration Considerations
13. Related
-->

# Choosing a Repository Pattern

How to select between Monorepo and Polyrepo patterns.

## Decision Checklist

Answer these questions to determine your pattern:

### Question 1: How often do services change together?

- **Frequently (daily/weekly)** → Monorepo
- **Rarely (monthly or less)** → Polyrepo

### Question 2: How much code is shared?

- **Heavy sharing (>30% shared)** → Monorepo
- **Minimal sharing (<10% shared)** → Polyrepo

### Question 3: How many teams?

- **1-3 teams, same timezone** → Monorepo
- **4+ teams, distributed** → Polyrepo

### Question 4: What access control do you need?

- **Everyone can see everything** → Monorepo
- **Team-specific access required** → Polyrepo

### Question 5: Do you need atomic cross-service changes?

- **Yes, frequently** → Monorepo
- **No, services are independent** → Polyrepo

## Quick Decision Matrix

| If you have...                                       | Choose   |
| ---------------------------------------------------- | -------- |
| Small team, shared code, coupled services            | Monorepo |
| Multiple teams, clear boundaries, independent deploy | Polyrepo |
| Need atomic refactoring across services              | Monorepo |
| Need fine-grained access control                     | Polyrepo |

## Pattern Characteristics

### Monorepo

**Directory Structure**:

```text
platform/
├── services/
│   ├── api/
│   ├── web/
│   └── worker/
├── shared/
│   ├── models/
│   └── utils/
└── infrastructure/
```

**Benefits**:

- Single commit for cross-cutting changes
- Easy code sharing
- Unified tooling
- Simplified dependency management

**Tradeoffs**:

- Potentially longer build times
- Coarser access control
- Repository can grow large

### Polyrepo

**Structure**:

```text
organization/
├── api-service/
├── web-service/
├── worker-service/
├── shared-models/
└── infrastructure/
```

**Benefits**:

- Independent team ownership
- Fine-grained access control
- Faster per-repo builds
- Clear service boundaries

**Tradeoffs**:

- Complex dependency management
- Cross-repo changes require coordination
- Tooling duplication

## Anti-Pattern: Technical Boundaries

**DON'T** split by technology:

```text
organization/
├── frontend/     ❌
├── backend/      ❌
├── scripts/      ❌
└── docs/         ❌
```

**DO** split by deployable module:

```text
organization/
├── order-service/     ✅ (includes its frontend, backend, scripts, docs)
├── user-service/      ✅
└── shared-lib/        ✅
```

## Migration Considerations

**Polyrepo → Monorepo**:

- Consolidate related services first
- Maintain git history with `git subtree`
- Update CI/CD for change detection

**Monorepo → Polyrepo**:

- Extract services with clear boundaries
- Version and publish shared code
- Establish contract testing

## Related

- [Repository Comparison](../../explanation/continuous-delivery/architecture/repository-patterns.md)
- [Repository Patterns](../../explanation/continuous-delivery/architecture/repository-patterns.md)
