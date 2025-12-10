# Specification Quality Checklist

{{ page_breadcrumb() }}

Reference for evaluating specification health.

## Health Indicators

### Red Flags (Problems)

| Indicator            | Description                               | Action                    |
| -------------------- | ----------------------------------------- | ------------------------- |
| Stale specs          | Unchanged >6 months while code evolved    | Review and update         |
| Always passing       | Scenarios pass but don't verify behavior  | Add meaningful assertions |
| Implementation focus | Tests "database updated", "cache cleared" | Rewrite in behavior terms |
| Large files          | >30 scenarios in single file              | Split by concern          |
| Technical jargon     | "POST to /api/users", "response 201"      | Use domain language       |
| Ambiguous steps      | Multiple valid interpretations            | Add concrete examples     |
| Gap vs code          | Spec language differs from implementation | Align terminology         |

### Green Indicators (Healthy)

| Indicator           | Description                         | Evidence                   |
| ------------------- | ----------------------------------- | -------------------------- |
| Co-committed        | Specs committed with code changes   | Git history shows together |
| Catch regressions   | Scenarios fail when behavior breaks | Bug caught by tests        |
| Business readable   | Stakeholders understand specs       | Product owner validates    |
| Clear traceability  | spec → step → code linked           | Feature IDs match          |
| Consistent language | Same terms throughout               | Matches glossary           |

## Review Triggers

| Trigger              | Action                       | Frequency      |
| -------------------- | ---------------------------- | -------------- |
| After implementation | Verify spec matches reality  | Per feature    |
| Test failures        | Determine spec vs code issue | As needed      |
| Requirements change  | Update affected specs        | As needed      |
| Regular cadence      | Prevent drift                | Weekly/Monthly |
| Before extensions    | Refresh understanding        | Per feature    |
| Bug discovered       | Add regression scenario      | Per bug        |

## Review Cadence

| Cadence   | Focus             | Activities                                    |
| --------- | ----------------- | --------------------------------------------- |
| Weekly    | Sync with code    | Review failing scenarios, add edge cases      |
| Monthly   | Prevent drift     | Check relevance, verify language, consolidate |
| Quarterly | Major refactoring | Review all specs, split large files, prune    |

## Metrics to Track

| Metric             | Threshold           | Tool               |
| ------------------ | ------------------- | ------------------ |
| File size          | >500 lines          | `wc -l`            |
| Age                | >6 months unchanged | `find -mtime`      |
| Scenario count     | >20 per file        | `grep -c Scenario` |
| Duplicate patterns | Similar steps       | Manual review      |

## Related

- [Review and Iterate](./review-and-iterate.md)
- [Specifications Evolution](../concepts/specifications-evolution.md)
- [Size Guidelines](../organization/size-guidelines.md)

{{ diataxis_footer() }}
