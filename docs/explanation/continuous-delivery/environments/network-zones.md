# Network Zones

## Introduction

Network segregation enforces security boundaries between environment types through isolated network zones.

This separation ensures that Build Agents cannot access production, and that environment access are isolated to Deploy Agents only.

---

## Purpose

- Prevent unauthorized environment access
- Limit blast radius of security breaches
- Enforce principle of least privilege
- Meet compliance requirements for production isolation

---

## Zone Architecture

### Zone A - Development/Test

- DevBox (developer laptops)
- Build Agents (CI/CD runners)
- PLTE instances
- Demo environments

**Characteristics:**

- No access to production networks
- Public internet access for package downloads
- Can read from artifact repositories
- Cannot deploy to production

### Zone B - Production

- Production runtime environments
- Production databases and services
- Live user traffic

**Characteristics:**

- Isolated from development/test zones
- Strict ingress/egress controls
- No direct access from Build Agents

### Zone C - Deployment Gateway

- Deploy Agents (production deployment capability)

**Characteristics:**

- Network access to both Zone A (artifact repos) and Zone B (production)
- Segregated credentials (production deployment keys)
- Comprehensive audit logging
- Multi-factor authentication required

---

## Traffic Flow

```text
Zone A (Build Agents)
    │
    ▼ publish artifacts
Artifact Repository
    │
    ▼ retrieve artifacts
Zone C (Deploy Agents)
    │
    ▼ deploy
Zone B (Production)
```

1. Build Agents (Zone A) build artifacts → publish to artifact repository
2. Deploy Agents (Zone C) retrieve artifacts → deploy to Production (Zone B)
3. Production (Zone B) never pulls directly from development zones

---

## Implementation

| Platform   | Approach                                                |
| ---------- | ------------------------------------------------------- |
| Azure      | Hub-and-spoke architecture with VNets, subnets and NSGs |
| AWS        | VPC with security groups and private subnets            |
| On-premise | Network segmentation with firewalls                     |

---

## Benefits

- **Containment**: Build Agents compromised → Production unaffected
- **Credential Isolation**: Production credentials never leave Zone C
- **Audit Trail**: All production deployments via Deploy Agents

---

## When to Use

**Required for:**

- Regulated industries (finance, healthcare)
- High-security requirements
- Compliance mandates (SOC 2, ISO 27001)
- Large organizations with separate teams

**Optional for:**

- Small teams with full trust
- Internal tools only
- Non-regulated domains
- Startups in early stages

---

## Next Steps

- [Environment Types](environments.md) - The 6 environment types
- [Deploy Agents](environments.md#deploy-agents) - Production deployment runners
- [CD Model Stages](../cd-model/stages.md) - Stage 10 deployment

## References

- [CD Model Overview](../cd-model/overview.md)
