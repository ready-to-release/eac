# Incident Response

How to respond to production incidents.

## Incident Severity Levels

| Level | Description | Response Time | Examples |
|-------|-------------|---------------|----------|
| P1 - Critical | Service down, data loss | < 15 min | Complete outage, security breach |
| P2 - High | Major feature broken | < 1 hour | Payment processing failing |
| P3 - Medium | Degraded performance | < 4 hours | Slow response times |
| P4 - Low | Minor issue | < 24 hours | UI glitch, typo |

## Incident Response Steps

### Step 1: Detect

**Automated Detection:**

- Monitoring alerts trigger
- Health check failures
- Error rate thresholds exceeded

**Manual Detection:**

- User reports
- Support tickets
- Internal observation

### Step 2: Triage

Assess severity and impact:

| Question | Answer |
|----------|--------|
| How many users affected? | All / Many / Few |
| Is data at risk? | Yes / No |
| Are transactions failing? | Yes / No |
| Is there a workaround? | Yes / No |

Assign severity level based on answers.

### Step 3: Communicate

**For P1/P2:**

```markdown
## Incident Update

**Status:** Investigating
**Impact:** [Description of user impact]
**Start Time:** [Time] UTC
**Next Update:** [Time] UTC

We are aware of issues with [service] and are actively investigating.
```

Update status page and notify stakeholders.

### Step 4: Respond

**Immediate Actions:**

| Action | When |
|--------|------|
| Rollback | Deployment caused issue |
| Disable feature flag | Feature causing issue |
| Scale resources | Capacity issue |
| Restart services | Transient failure |
| Failover | Region/zone failure |

**Do NOT:**

- Make changes without tracking
- Skip communication updates
- Work in isolation

### Step 5: Resolve

When issue is fixed:

1. **Verify** - Confirm metrics returned to normal
2. **Monitor** - Watch for recurrence (30+ minutes)
3. **Update status** - Mark incident resolved
4. **Notify** - Inform all stakeholders

### Step 6: Post-Mortem

Within 48 hours, conduct post-mortem:

```markdown
# Incident Post-Mortem

**Date:** YYYY-MM-DD
**Duration:** [X hours]
**Severity:** P1
**Lead:** [Name]

## Summary

[1-2 sentence description]

## Timeline

| Time | Event |
|------|-------|
| 14:00 | Alert triggered |
| 14:05 | On-call acknowledged |
| 14:15 | Root cause identified |
| 14:30 | Fix deployed |
| 14:45 | Incident resolved |

## Root Cause

[Technical explanation of what went wrong]

## Impact

- [Users affected]
- [Revenue impact if any]
- [Data impact if any]

## What Went Well

- [Positive observations]

## What Went Wrong

- [Areas for improvement]

## Action Items

| Action | Owner | Deadline |
|--------|-------|----------|
| [Action] | [Name] | [Date] |

## Lessons Learned

[Key takeaways for the team]
```

## On-Call Responsibilities

| Responsibility | Action |
|----------------|--------|
| Acknowledge alerts | < 5 min |
| Initial triage | < 10 min |
| Escalate if needed | When blocked |
| Document actions | All changes logged |
| Hand off properly | Brief next person |

## Communication Templates

**Initial Alert:**

> We're investigating reports of [issue]. More updates to follow.

**Identified:**

> We've identified the issue as [cause]. Working on a fix.

**Monitoring:**

> Fix has been deployed. Monitoring for stability.

**Resolved:**

> The issue has been resolved. Services are operating normally.

## Next Steps

- [Rollback Procedures](./rollback-procedures.md) - Execute emergency rollbacks
- [Live Monitoring](../cd-model/cd-model-stages-8-12.md#stage-11-live) - CD Model Stage 11
- [Release Quality Thresholds](../quality-gates/release-gates.md) - Quality gate criteria
- [Deployment Strategies](./deployment-strategies.md) - Deployment patterns and approaches
