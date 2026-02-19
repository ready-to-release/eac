# AI-Assisted Defect Detection

## Introduction

AI (LLMs like GPT-4, Claude, GitHub Copilot) provides semantic analysis to detect defects that traditional tools miss. Most valuable at **Stages 1-3** (shift-left) and **Stage 9** (automated approval).

---

## AI Capabilities by CD Stage

| Stage  | AI Capability                    | Defect Category           | Value  | How AI Helps                                                               |
| ------ | -------------------------------- | ------------------------- | ------ | -------------------------------------------------------------------------- |
| **1**  | Ambiguous requirements detection | Knowledge & Communication | High   | Flags vague terms, generates test scenarios, finds contradictions in specs |
| **1**  | Test scenario generation         | Testing & Observability   | High   | Generates Given/When/Then from requirements, exposes edge cases            |
| **2**  | Implicit knowledge detection     | Knowledge & Communication | High   | Flags magic numbers, undocumented business rules, missing "why" comments   |
| **2**  | Comprehensive test generation    | Testing & Observability   | High   | Generates edge cases, negative inputs, boundary values, error paths        |
| **3**  | Semantic code review             | Change & Complexity       | High   | Detects logic errors, missing validations, design smells beyond syntax     |
| **3**  | Divergent mental models          | Knowledge & Communication | High   | Finds terminology mismatches across services, bounded context violations   |
| **5**  | Integration test generation      | Integration & Boundaries  | Medium | Generates API test scenarios from contracts, cross-service test cases      |
| **9**  | Automated risk scoring           | Process & Deployment      | High   | Analyzes change diff + history to auto-approve low-risk changes            |
| **11** | Anomaly detection                | Testing & Observability   | Medium | ML-based alert thresholds, SLO burn rate prediction, pattern detection     |

---

## Next Steps

- **Understanding defect categories?** See [External Defect Catalog](https://bdfinst.github.io/ai-patterns/defect-detection-and-fixes/)
- **Stage-by-stage prevention?** Read [Stage-by-Stage Guide](defect-stages-mapping.md)
- **Quality gates overview?** Review [Quality Gates Overview](../index.md)
- **Using AI agents as contributors?** See
  [Agentic CD](../../agentic-cd/index.md) and
  [ACD and the CD Model](../../agentic-cd/acd-and-cd-model.md).
