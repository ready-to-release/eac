# Live

## Introduction

"Live" is the destination where released software becomes available to consumers.

What "Live" means depends on the module type: a production environment for services, or an artifact registry for libraries.

![Live Legend](../../../assets/branching/legend-live.drawio.png){width=150}

---

## Definition

**Live:** The destination where released software serves consumers.

Live takes different forms depending on the [deployable module](deployable-modules.md) type:

| Module Type      | Live Destination                      | Consumers        |
| ---------------- | ------------------------------------- | ---------------- |
| Services         | Production environment                | End users        |
| Libraries        | Package registry (NuGet, npm, PyPI)   | Developers       |
| CLI tools        | Package manager or release page       | Users/automation |
| Container images | Container registry (Docker Hub, GHCR) | Orchestrators    |
| Documentation    | Published site (GitHub Pages)         | Readers          |

---

## For Deployed Services

Services run continuously in production environments serving real users.

**Characteristics:**

- Real user traffic and business data
- High availability requirements
- Continuous monitoring and alerting
- Incident response procedures

**Key Capabilities:**

- **Monitoring**: Application performance, error rates, resource utilization
- **Observability**: Distributed tracing, log aggregation, metrics dashboards
- **Feature Flags**: Runtime control over feature availability
- **Rollback**: Redeploy previous version if issues detected

**CD Model Stages:**

- Stage 10 (Production Deployment): Deploy to production
- Stage 11 (Live): Monitor health and performance
- Stage 12 (Release Toggling): Enable/disable features via flags

---

## For Published Artifacts

Libraries and tools are published to registries where consumers pull specific versions.

**Characteristics:**

- Immutable versioned releases
- Multiple versions available simultaneously
- Consumers control which version they use
- Deprecation process for old versions

**Key Capabilities:**

- **Semantic Versioning**: Signal compatibility (breaking vs non-breaking)
- **Download Metrics**: Track adoption and usage
- **Deprecation Notices**: Warn consumers of outdated versions
- **Security Advisories**: Alert consumers to vulnerabilities

**CD Model Stages:**

- Stage 10 (Production Deployment): Publish to registry
- Stage 11 (Live): Track downloads and adoption
- Stage 12 (Release Toggling): Mark versions as latest/deprecated

---

## Feedback Loops

Live environments generate feedback that influences future development:

```text
Live Environment
    │
    ├─► Monitoring alerts ─► Investigation ─► Bug fixes to Trunk
    ├─► Performance data ─► Optimization ─► Improvements to Trunk
    ├─► User behavior ─► Product insights ─► Features to Trunk
    └─► Incidents ─► Post-mortems ─► Process improvements
```

**For Services:**

- Error rates trigger alerts
- Performance degradation initiates optimization
- Feature flag data informs product decisions
- Incidents result in fixes committed to trunk

**For Artifacts:**

- Download trends inform prioritization
- Issue reports drive bug fixes
- Consumer feedback shapes roadmap
- Security scans trigger patch releases

---

## Rollback Strategies

When issues occur in live, rollback provides recovery:

**Services:**

| Strategy          | How                                  | Speed   |
| ----------------- | ------------------------------------ | ------- |
| Redeploy previous | Deploy last known good version       | Minutes |
| Feature flag      | Disable problematic feature          | Seconds |
| Blue-green switch | Route traffic to previous deployment | Seconds |
| Database rollback | Restore data (if applicable)         | Varies  |

**Artifacts:**

| Strategy           | How                              | Consumer Impact        |
| ------------------ | -------------------------------- | ---------------------- |
| Yank/unpublish     | Remove bad version from registry | Breaks builds using it |
| Patch release      | Publish fixed version            | Consumers must upgrade |
| Pin recommendation | Advise pinning to safe version   | Manual consumer action |

---

## Next Steps

- [Pipeline](pipeline.md) - How software reaches Live
- [Deployable Modules](deployable-modules.md) - What gets deployed
- [CD Model Stages](../cd-model/stages.md) - Stages 10-12 detail
- [Deployment Strategies](../deployment/deployment-strategies.md) - How to deploy safely

## References

- [CD Model Overview](../cd-model/overview.md)
- [Environments](../environments/environments.md)
