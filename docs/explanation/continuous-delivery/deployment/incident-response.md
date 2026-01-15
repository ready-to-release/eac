# Incident Response

How to respond to production incidents effectively.

---

## Severity Levels

| Level         | Description             | Ex. SLA/SLI Response Time | Examples                         |
| ------------- | ----------------------- | ------------------------- | -------------------------------- |
| P1 - Critical | Service down, data loss | < 15 min                  | Complete outage, security breach |
| P2 - High     | Major feature broken    | < 1 hour                  | Payment processing failing       |
| P3 - Medium   | Degraded performance    | < 4 hours                 | Slow response times              |
| P4 - Low      | Minor issue             | < 24 hours                | UI glitch, typo                  |

---

## Response Steps

### 1. Detect

- **Automated:** Monitoring alerts, health check failures, error rate thresholds
- **Manual:** User reports, support tickets, internal observation

### 2. Triage

| Question                  | Impact Assessment |
| ------------------------- | ----------------- |
| How many users affected?  | All / Many / Few  |
| Is data at risk?          | Yes / No          |
| Are transactions failing? | Yes / No          |
| Is there a workaround?    | Yes / No          |

### 3. Communicate

Update status page immediately for P1/P2. Include: status, impact, start time, next update time.

### 4. Respond

| Action               | When                    |
| -------------------- | ----------------------- |
| Rollback             | Deployment caused issue |
| Disable feature flag | Feature causing issue   |
| Scale resources      | Capacity issue          |
| Restart services     | Transient failure       |
| Failover             | Region/zone failure     |

### 5. Resolve

1. **Verify** - Confirm metrics returned to normal
2. **Monitor** - Watch for recurrence (30+ minutes)
3. **Update status** - Mark incident resolved
4. **Notify** - Inform all stakeholders

### 6. Post-Mortem

Within 48 hours, document: summary, timeline, root cause, impact, what went well, what went wrong, action items.

---

## On-Call Responsibilities

| Responsibility       | Target            |
| -------------------- | ----------------- |
| Acknowledge alerts   | < 5 min           |
| Initial triage       | < 10 min          |
| Escalate if blocked  | Immediately       |
| Document all actions | Real-time         |
| Hand off properly    | Brief next person |

---

## Communication Templates

- **Investigating:** "We're investigating reports of [issue]. More updates to follow."
- **Identified:** "We've identified the issue as [cause]. Working on a fix."
- **Monitoring:** "Fix has been deployed. Monitoring for stability."
- **Resolved:** "The issue has been resolved. Services are operating normally."

---

## External Resources

- [Google SRE Book - Incident Management](https://sre.google/sre-book/managing-incidents/)
- [PagerDuty Incident Response Guide](https://response.pagerduty.com/)
- [Atlassian Incident Management](https://www.atlassian.com/incident-management)

---

## Next Steps

- [Rollback Procedures](./rollback-procedures.md) - Execute emergency rollbacks
- [CD Model Stage 11](../cd-model/stages.md#stage-11-live) - Live monitoring
- [Deployment Strategies](./deployment-strategies.md) - Deployment patterns
