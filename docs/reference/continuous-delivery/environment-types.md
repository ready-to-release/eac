# Environment Types

Reference for environment types in the CD Model.

## Environment Summary

| Environment   | Purpose                 | Lifecycle             | CD Model Stages |
| ------------- | ----------------------- | --------------------- | --------------- |
| DevBox        | Local development       | Persistent            | 1, 2            |
| Build Agents  | CI/CD automation        | Ephemeral per build   | 2, 3, 4, 8      |
| PLTE          | Production-like testing | Ephemeral per feature | 5, 6            |
| Demo          | Stakeholder validation  | Semi-persistent       | 7               |
| Deploy Agents | Production deployment   | Persistent            | 10, 11          |
| Production    | Live user traffic       | Persistent            | 10, 11, 12      |

## DevBox

**Purpose:** Local development environment for authoring changes.

| Characteristic | Value                  |
| -------------- | ---------------------- |
| Control        | Full developer control |
| Isolation      | Complete isolation     |
| Feedback       | Immediate (seconds)    |
| Impact         | No impact on others    |

**Tools:** IDE, local build tools, unit testing frameworks, security scanners, Git, Docker.

## Build Agents

**Purpose:** Dedicated CI/CD pipeline runners for consistent builds.

| Characteristic | Value                                   |
| -------------- | --------------------------------------- |
| Execution      | Isolated per build                      |
| Configuration  | Consistent across runs                  |
| Credentials    | Read artifact repos, write test results |
| Network        | No production access                    |

## PLTE (Production-Like Test Environment)

**Purpose:** Ephemeral environments emulating production for realistic testing.

| Characteristic | Value                                      |
| -------------- | ------------------------------------------ |
| Infrastructure | Same as production                         |
| Configuration  | Production-like (no prod credentials)      |
| Data           | Realistic, anonymized                      |
| Isolation      | Per feature branch                         |
| Lifecycle      | Created on-demand, destroyed after testing |

## Demo Environment

**Purpose:** Stable environment for stakeholder validation and exploratory testing.

| Characteristic | Value                               |
| -------------- | ----------------------------------- |
| State          | Current main branch                 |
| Lifecycle      | Days to weeks                       |
| Access         | Non-technical stakeholders          |
| Data           | Production-like, no production data |

**Update Cadence:** Typically updated after successful Stage 6.

## Deploy Agents

**Purpose:** Specialized runners with production deployment access.

| Characteristic | Value                                |
| -------------- | ------------------------------------ |
| Network        | Access to production                 |
| Credentials    | Production deployment keys (vaulted) |
| Logging        | Comprehensive audit                  |
| Access Control | Strict, MFA required                 |

## Production

**Purpose:** Live environment serving end users.

| Characteristic | Value                             |
| -------------- | --------------------------------- |
| Traffic        | Live user traffic                 |
| Data           | Real business data                |
| Availability   | High availability required        |
| Monitoring     | Continuous performance monitoring |

## Related

- [Environments Architecture](../../explanation/continuous-delivery/architecture/environments.md)
- [Network Zones](network-zones.md)
- [Deployment Strategies](deployment-strategies.md)
