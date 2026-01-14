# Risk Assessment - Executive Summary

Generate an executive summary analyzing overall risk posture. Output ONLY valid JSON with 2 fields: `executive_summary` and `confidence`.

## Output Format

```json
{
  "executive_summary": {
    "overall_risk_posture": "moderate",
    "summary_narrative": "2-3 paragraph narrative (100-1000 chars)",
    "key_findings": ["Finding 1 (20+ chars)", "Finding 2", "Finding 3"],
    "strategic_recommendations": ["Recommendation 1 (30+ chars)", "Recommendation 2"],
    "critical_modules": [{"module": "name", "reason": "Why critical (20+ chars)"}],
    "trends": ["Trend 1 (20+ chars)"]
  },
  "confidence": 0.85
}
```

## Field Requirements

**overall_risk_posture** (required): One of "critical", "high", "moderate", "low" (lowercase)

**summary_narrative** (required): 100-1000 characters describing:

- Scope and overall posture
- Key risks and strengths
- Context and impact

**key_findings** (required): Array of 3-7 strings (each 20+ chars)

- Most impactful systemic risks
- Both negative and positive findings

**strategic_recommendations** (required): Array of 2-5 strings (each 30+ chars)

- Root cause solutions
- Prioritized by impact

**critical_modules** (optional): Array of objects with high-risk modules needing immediate attention

**trends** (optional): Array of strings with cross-cutting patterns

**confidence**: Number 0.0-1.0 based on data completeness

## Constraints

- NO markdown fences (```json)
- NO extra fields (metadata, module_analyses, total_controls, etc.)
- Pure JSON starting with { and ending with }
- All strings use double quotes

Assessment data:
