# Verification Types

Verification types classify acceptance tests by **what** they verify, ensuring comprehensive validation at Stage 5.

---

## The Three Types

| Type                      | Abbreviation | Question           | Focus                         |
| ------------------------- | ------------ | ------------------ | ----------------------------- |
| Installation Verification | IV           | Can we deploy it?  | Infrastructure, configuration |
| Operational Verification  | OV           | Does it work?      | Functional requirements       |
| Performance Verification  | PV           | Is it fast enough? | Response times, throughput    |

**All three must pass for production readiness.**

---

## Installation Verification (IV)

Confirms the solution can be deployed and configured correctly.

| Validates             | Examples                                          |
| --------------------- | ------------------------------------------------- |
| Deployment procedures | Scripts execute, IaC provisions correctly         |
| Dependencies          | External services reachable, databases connected  |
| Configuration         | Environment variables applied, secrets loaded     |
| Infrastructure        | Load balancers operational, networking configured |

**Common IV patterns:**

| Pattern               | What it checks                       |
| --------------------- | ------------------------------------ |
| Health check          | `GET /health` returns 200            |
| Database connectivity | `GET /health/db` returns 200         |
| Configuration loaded  | Query config endpoint, verify values |
| Artifact integrity    | Version endpoint matches expected    |

---

## Operational Verification (OV)

Ensures the solution operates as intended.

| Validates               | Examples                               |
| ----------------------- | -------------------------------------- |
| Functional requirements | Features work as specified             |
| User workflows          | Authentication, CRUD, checkout         |
| Business rules          | Calculations, permissions, constraints |
| Error handling          | Meaningful messages, graceful failures |

**Common OV patterns:**

| Pattern           | What it checks                    |
| ----------------- | --------------------------------- |
| Happy path        | Standard workflow succeeds        |
| Error handling    | Invalid input rejected with error |
| Permissions       | Authorization rules enforced      |
| State transitions | Workflow progresses correctly     |

---

## Performance Verification (PV)

Validates the solution meets performance requirements under expected load.

| Validates            | Examples                             |
| -------------------- | ------------------------------------ |
| Response times       | P50, P95, P99 latencies              |
| Throughput           | Requests per second                  |
| Resource utilization | CPU < 70%, Memory < 80%              |
| Scalability          | Performance stable as load increases |

**Typical thresholds:**

| Metric            | Threshold    |
| ----------------- | ------------ |
| P95 response time | < 200ms      |
| P99 response time | < 500ms      |
| Throughput        | > 1000 req/s |
| CPU utilization   | < 70%        |
| Error rate        | < 0.1%       |

**PV test types:**

| Type        | Purpose                               |
| ----------- | ------------------------------------- |
| Load test   | Performance under expected load       |
| Stress test | Behavior under extreme load           |
| Spike test  | Handling sudden traffic increase      |
| Soak test   | Long-running stability (memory leaks) |

---

## Stage 5 Execution

**Execution order (fail fast):**

1. **IV first** - If can't deploy, no point testing functionality
2. **OV next** - If doesn't work, no point testing performance
3. **PV last** - Validate performance of working system

**Tagging in Gherkin:**

```gherkin
@iv
Scenario: Application starts and responds to health checks

@ov
Scenario: User can complete checkout

@pv
Scenario: API responds within 200ms under load
```

---

## Anti-Patterns

| Anti-Pattern             | Problem                                        | Solution                  |
| ------------------------ | ---------------------------------------------- | ------------------------- |
| Only OV tests            | Deployment/performance issues reach production | Ensure IV and PV coverage |
| IV in dev environment    | Unrealistic validation                         | Execute in PLTE           |
| PV with unrealistic load | Issues not detected                            | Use production-like load  |
| Manual verification      | Inconsistent, slow                             | Automate all types        |

---

## External Resources

- [Acceptance Testing (Agile Alliance)](https://www.agilealliance.org/glossary/acceptance/)
- [Performance Testing Guidance](https://learn.microsoft.com/en-us/azure/well-architected/performance-efficiency/performance-test)

---

## Next Steps

- [Test Levels](./test-levels.md) - L0-L4 taxonomy
- [CD Model Stage 5](../cd-model/stages.md#stage-5-acceptance-testing) - Acceptance testing in context
- [Testing Taxonomy](../../specifications/taxonomy/) - Tag contracts
