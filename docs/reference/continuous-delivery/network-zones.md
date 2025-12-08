# Network Zone Architecture

Reference for network segregation zones in the CD Model.

## Zone Definitions

| Zone | Name               | Contains                                    | Network Access         |
| ---- | ------------------ | ------------------------------------------- | ---------------------- |
| A    | Development/Test   | DevBox, Build Agents, PLTE, Demo            | No production access   |
| B    | Production         | Production runtime, databases, live traffic | Isolated from Zone A   |
| C    | Deployment Gateway | Deploy Agents only                          | Access to both A and B |

## Zone A - Development/Test

**Components:**

- DevBox (developer laptops)
- Build Agents (CI/CD runners)
- PLTE instances
- Demo environments

**Characteristics:**

- No access to production networks
- Public internet access for package downloads
- Can read from artifact repositories
- Cannot deploy to production

## Zone B - Production (Isolated)

**Components:**

- Production runtime environments
- Production databases and services
- Live user traffic

**Characteristics:**

- Isolated from development/test zones
- Strict ingress/egress controls
- No direct access from Build Agents

## Zone C - Deployment Gateway

**Components:**

- Deploy Agents (production deployment capability)

**Characteristics:**

- Network access to both Zone A (artifact repos) and Zone B (production)
- Segregated credentials (production deployment keys)
- Comprehensive audit logging
- Multi-factor authentication required

## Traffic Flow

```
1. Build Agents (Zone A) → build artifacts → artifact repository
2. Deploy Agents (Zone C) → retrieve artifacts → Production (Zone B)
3. Production (Zone B) never pulls directly from development zones
```

## Platform Implementation

| Platform   | Implementation                                 |
| ---------- | ---------------------------------------------- |
| Azure      | Hub-and-spoke architecture with VNets and NSGs |
| AWS        | VPC with security groups and private subnets   |
| On-premise | Network segmentation with firewalls            |

## Security Benefits

- Build Agents compromised → Production unaffected
- Production credentials never leave Zone C
- Clear audit trail (all production deployments via Deploy Agents)
- Compliance requirement: separation of duties

## When to Use Network Segregation

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

## Related

- [Environments Architecture](../../explanation/continuous-delivery/architecture/environments.md)
- [Deploy Agents](../../explanation/continuous-delivery/cd-model/cd-model-stages-7-12.md#stage-10-production-deployment)
