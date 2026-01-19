# CD Model for Data Projects

## What You'll Learn

How each of the 12 CD model stages applies to Databricks data projects, including notebooks, pipelines, Delta tables, and business logic.

## Overview

The 12-stage CD model applies to data projects with specific considerations for data artifacts. This guide maps each stage to data engineering activities.

**Key Differences for Data Projects:**

- Pipelines process data, creating stateful outputs
- Testing requires realistic data volumes
- Rollback involves data versioning, not just code
- Quality includes data quality, not just code quality
- Multiple artifact types deploy together (notebooks + jobs + schemas)

## Development Stages (1-7)

### Stage 1: Authoring

**Purpose**: Create code, specifications, and configurations on local topic branches.

**Data Activities:**

| Artifact Type      | Activities                                                                    |
| ------------------ | ----------------------------------------------------------------------------- |
| **Notebooks**      | Develop transformation logic locally using Databricks Connect or web IDE      |
| **Specifications** | Write Gherkin specs defining data transformations, schemas, and quality rules |
| **Schemas**        | Define Delta table schemas, partitioning strategies                           |
| **Jobs**           | Design workflow orchestration, dependencies, scheduling                       |
| **Business Rules** | Define segmentation logic, categorization rules, calculations                 |
| **Dashboards**     | Create visualizations, define metrics                                         |

**Tools:**

- **Databricks Connect** - Execute notebooks locally against remote clusters
- **VS Code / PyCharm** - Local IDE with Databricks extension
- **Git** - Version control for all artifacts

**Example** (Customer Segmentation):

```gherkin
# features/customer-segmentation_pipeline.feature
@L2 @ov @control:si-10
Feature: customer-segmentation_pipeline
  Segment customers by purchase behavior for targeted campaigns

  As a marketing analyst
  I want to segment customers by purchase behavior
  So that I can target campaigns effectively

  Background:
    Given Unity Catalog is enabled
    And catalog "production" exists

  Rule: Pipeline aggregates customer features from events

    @ov
    Scenario: Aggregate customer features
      Given customer events in bronze layer
      When I aggregate daily activity metrics
      Then silver layer contains customer_features table
      And table has columns: customer_id, total_purchases, avg_order_value
```

**Environment**: DevBox (local development)

---

### Stage 2: Pre-commit

**Purpose**: Validate changes locally before committing (5-10 minute time-box).

**Data Activities:**

| Validation            | Implementation                                      |
| --------------------- | --------------------------------------------------- |
| **Notebook linting**  | `pylint`, `black`, `mypy` for Python notebooks      |
| **Unit tests**        | Test transformation functions with small DataFrames |
| **Schema validation** | Verify Delta table schemas match specifications     |
| **Bundle validation** | `databricks bundle validate` checks configuration   |
| **Secrets scan**      | Detect hardcoded credentials or API keys            |

**Tools:**

- **pytest** - Unit test framework
- **chispa** - Spark DataFrame assertions
- **databricks CLI** - Bundle validation
- **pre-commit hooks** - Automated validation

**Example:**

```python
# tests/test_customer_features.py
def test_aggregate_features():
    input_df = spark.createDataFrame([
        (1, "2024-01-01", 100.0),
        (1, "2024-01-02", 150.0),
    ], ["customer_id", "date", "amount"])

    result = aggregate_customer_features(input_df)

    assert result.count() == 1
    assert result.first().total_purchases == 250.0
```

**Time Budget**: 5-10 minutes maximum

**Environment**: DevBox + Build Agents (CI validation)

---

### Stage 3: Merge Request

**Purpose**: Peer review and automated CI validation before integration.

**Data Activities:**

| Activity                 | Implementation                                           |
| ------------------------ | -------------------------------------------------------- |
| **Code review**          | Review notebook logic, SQL queries, data transformations |
| **CI pipeline**          | Run unit tests, linting, bundle validation               |
| **Documentation review** | Check specifications, schema docs, DAG diagrams          |
| **Security review**      | Verify access controls, data masking, encryption         |

**Quality Gates:**

- ✅ Peer approval from data engineer
- ✅ All unit tests passing
- ✅ Bundle validation successful
- ✅ No critical security issues
- ✅ Specs updated for schema changes

**Tools:**

- **GitHub Pull Requests** - Code review platform
- **GitHub Actions** - CI automation
- **Asset Bundles** - Configuration validation

**Environment**: Build Agents

---

### Stage 4: Commit

**Purpose**: Integrate validated changes to trunk and trigger full pipeline validation.

**Data Activities:**

| Activity             | Implementation                                           |
| -------------------- | -------------------------------------------------------- |
| **Merge to main**    | Squash merge approved changes                            |
| **Build artifacts**  | Package notebooks as wheels/JARs if using separate repos |
| **Trigger CI**       | Run full test suite (L0, L1, L2)                         |
| **Version tracking** | Tag commit SHA, link to work items                       |

**Automated Testing:**

- **L0**: Unit tests (transformation logic)
- **L1**: Component tests (notebook end-to-end with test data)
- **L2**: Integration tests (multi-notebook pipeline)
- **L3**: System tests (if fast enough, otherwise Stage 5)

**Traceability**: Commit SHA links to requirements, test results, and artifacts

**Environment**: Build Agents

---

### Stage 5: Acceptance Testing

**Purpose**: Validate pipeline functionality in production-like environment.

**Data Activities:**

| Activity                | Implementation                                             |
| ----------------------- | ---------------------------------------------------------- |
| **Deploy to PLTE**      | Create ephemeral workspace or use dedicated test workspace |
| **Clone test data**     | Use Delta Lake `CLONE` for production-like data            |
| **Run pipeline**        | Execute full workflow with realistic data volumes          |
| **Verify outputs**      | Check Delta table contents, row counts, schemas            |
| **Data quality checks** | Validate completeness, accuracy, consistency               |

**PLTE Characteristics for Data:**

- **Unity Catalog isolation**: Separate catalog (e.g., `dev`, `staging`)
- **Delta Lake clones**: `CREATE TABLE staging.customers SHALLOW CLONE prod.customers`
- **Realistic data**: Sufficient volume to catch performance issues
- **Ephemeral or persistent**: Choice depends on cost vs setup time

**Verification Types:**

- **IV (Installation Verification)**: Bundle deployed successfully, tables created
- **OV (Operational Verification)**: Pipeline runs without errors, completes in SLA
- **PV (Performance Verification)**: Data processing rates acceptable

**Example:**

```sql
-- Verify segmentation output
SELECT segment, COUNT(*) as customer_count
FROM staging.customer_segments
GROUP BY segment
ORDER BY segment;

-- Expected: ~4 segments with balanced distribution
```

**Environment**: PLTE (Production-Like Test Environment)

---

### Stage 6: Extended Testing

**Purpose**: Deep validation of performance, security, and data quality.

**Data Activities:**

| Test Type                    | Implementation                                       |
| ---------------------------- | ---------------------------------------------------- |
| **Performance testing**      | Load test with production-scale data volumes         |
| **Security scanning**        | DAST with OWASP ZAP, Unity Catalog permission audits |
| **Data quality validation**  | pytest assertions, schema drift detection            |
| **Cross-system integration** | External API connections, event streaming            |
| **Cost analysis**            | Cluster utilization, compute costs, storage growth   |

**Trigger**: Periodic (e.g., nightly) or on-demand, not every commit

**Quality Gates:**

- Performance regression < 5%
- No critical/high security vulnerabilities
- Data quality expectations passing
- Cost within budget thresholds

**Tools:**

- **pytest** - Data quality assertions
- **Databricks monitoring** - Query performance metrics
- **OWASP ZAP** - Security testing
- **Cost dashboards** - DBU consumption tracking

**Environment**: PLTE (longer-lived instances with production-scale data)

---

### Stage 7: Exploration

**Purpose**: Stakeholder validation and exploratory testing.

**Data Activities:**

| Activity                 | Implementation                                 |
| ------------------------ | ---------------------------------------------- |
| **Demo pipelines**       | Show data flows, transformations, lineage      |
| **Dashboard validation** | Stakeholders review metrics, visualizations    |
| **Data quality review**  | Business users validate accuracy, completeness |
| **UAT**                  | Business users test with real scenarios        |
| **Documentation**        | Review data dictionaries, runbooks             |

**What Exploration Catches:**

- Incorrect business logic in transformations
- Missing or unexpected data patterns
- Confusing dashboard layouts
- Performance issues with real query patterns

**Environment**: Demo workspace (trunk-HEAD or release-HEAD)

---

## Release Stages (8-12)

### Stage 8: Start Release

**Purpose**: Initiate formal release process and prepare for production.

**Data Activities:**

| Activity                   | Implementation                                           |
| -------------------------- | -------------------------------------------------------- |
| **Create release branch**  | Branch from main (e.g., `release/v2024.01.15`)           |
| **Tag version**            | Semantic versioning or CalVer: `YYYY.MM.DD`              |
| **Generate release notes** | Document schema changes, new pipelines, breaking changes |
| **Freeze features**        | Only critical fixes allowed after this point             |

**Release Notes Include:**

- New pipelines or jobs
- Schema changes (added columns, renamed tables)
- Breaking changes (deprecated fields, changed semantics)
- Performance improvements
- Data quality enhancements
- Known limitations

**Example Release Note:**

```markdown
## [2024.01.15] - 2024-01-15

### Added
- feat(pipelines): customer segmentation pipeline with 4 segments
- feat(ml-models): lifetime value prediction model (accuracy: 0.92 RMSE)
- feat(schema): added `customer_segments.predicted_ltv` column (DECIMAL)

### Changed
- refactor(schema): renamed `orders_raw` table to `bronze.orders`
- fix(schema): `customer_features.signup_date` now stores UTC timezone (was local time) - **BREAKING**
```

**Environment**: Build Agents

---

### Stage 9: Release Approval

**Purpose**: Obtain formal approval for production deployment.

**Approval Criteria:**

| Category       | Requirements                                                        |
| -------------- | ------------------------------------------------------------------- |
| **Quality**    | All tests passing, data quality checks pass, performance acceptable |
| **Security**   | Scans pass, Unity Catalog permissions reviewed, no data leakage     |
| **Compliance** | Documentation complete, change review board approval (if required)  |
| **Business**   | Stakeholder sign-off, rollback plan documented, support trained     |

**Variants:**

| Pattern                         | When to Use                                                       |
| ------------------------------- | ----------------------------------------------------------------- |
| **RA (Release Approval)**       | Financial data, regulated industries, high-risk schema changes    |
| **CDe (Continuous Deployment)** | Internal analytics, low-risk transformations, experimental models |

**RA Example**: Change review board approves production schema changes
**CDe Example**: Quality gates auto-approve if all tests pass

**Environment**: Evidence from PLTE + Demo validation

---

### Stage 10: Production Deployment

**Purpose**: Deploy to production with appropriate controls.

**Data Activities:**

| Activity                | Implementation                                       |
| ----------------------- | ---------------------------------------------------- |
| **Deploy Asset Bundle** | `databricks bundle deploy --target prod`             |
| **Schema migration**    | Apply Delta table schema evolution if needed         |
| **Update jobs**         | Swap to new notebook versions, adjust schedules      |
| **Backfill data**       | Re-process historical data if transformation changed |

**Deployment Strategies:**

| Strategy         | Use Case                                                         |
| ---------------- | ---------------------------------------------------------------- |
| **Blue-Green**   | Parallel pipelines, switch at completion (double cost)           |
| **Canary**       | Process subset of data first (10% of customers)                  |
| **Rolling**      | Gradual migration by partition (process 2024 data, then 2023...) |
| **Feature Flag** | Deploy code with transformation disabled, enable gradually       |

**Rollback Planning:**

- Delta Lake time travel: `SELECT * FROM table TIMESTAMP AS OF '2024-01-14'`
- Job versioning: Keep previous notebook version available
- Schema rollback: Document reverse migrations
- Target rollback time: < 15 minutes for critical pipelines

**Example Deployment:**

```bash
# Deploy production bundle
databricks bundle deploy --target prod

# Verify deployment
databricks jobs list --output json | jq '.jobs[] | select(.settings.name == "customer_segmentation")'

# Monitor first run
databricks jobs run-now --job-id 12345
```

**Environment**: Production (via Deploy Agents with restricted access)

---

### Stage 11: Live

**Purpose**: Monitor and validate production operation.

**Data Activities:**

| Monitoring Category  | Metrics                                             |
| -------------------- | --------------------------------------------------- |
| **Infrastructure**   | Cluster utilization, memory, disk, network          |
| **Pipeline health**  | Job success rate, duration, SLA compliance          |
| **Data quality**     | Freshness, completeness, accuracy, consistency      |
| **Business metrics** | Record counts, aggregated values, anomaly detection |
| **Cost**             | DBU consumption, storage growth, compute costs      |

**Alert Thresholds:**

| Metric            | Threshold     | Action                  |
| ----------------- | ------------- | ----------------------- |
| Job failure rate  | < 1%          | Alert on-call engineer  |
| Pipeline duration | < SLA + 20%   | Investigate performance |
| Data freshness    | < 2 hours lag | Alert data team         |
| Row count anomaly | > 2 std dev   | Alert data owner        |

**Incident Response:**

1. **Detect** - Monitoring alert or user report
2. **Triage** - Assess impact (data loss? incorrect results?)
3. **Respond** - Rollback pipeline, pause job, or hotfix
4. **Communicate** - Notify stakeholders via status page
5. **Resolve** - Fix root cause, re-process data if needed
6. **Post-mortem** - Document incident, improve monitoring

**Tools:**

- **Databricks Job Monitoring** - Run history, error rates
- **Delta Live Tables** - Built-in expectations and alerts
- **Custom dashboards** - Business-specific data quality metrics

**Environment**: Production

---

### Stage 12: Release Toggling

**Purpose**: Control feature availability using feature flags (optional).

**Data Activities:**

| Use Case                  | Implementation                                     |
| ------------------------- | -------------------------------------------------- |
| **Transformation toggle** | Conditionally apply new aggregation logic          |
| **Schema evolution**      | Gradually add columns, migrate data                |
| **A/B testing**           | Run two transformation versions, compare results   |
| **Canary analysis**       | Process subset of data with new logic              |
| **Dark launch**           | Deploy pipeline but don't update downstream tables |

**Flag Types:**

| Type             | Example                                           |
| ---------------- | ------------------------------------------------- |
| **Boolean**      | `use_new_segmentation_algo: true`                 |
| **Percentage**   | `new_pipeline_rollout_pct: 25` (25% of customers) |
| **Cohort-based** | `enabled_for_regions: ["us-west", "eu-central"]`  |
| **Time-based**   | `enable_after: "2024-02-01T00:00:00Z"`            |

**Implementation Patterns:**

Pattern 1 - Conditional transformation:

```python
# Read feature flag
use_new_algo = spark.conf.get("feature.new_segmentation", "false") == "true"

if use_new_algo:
    segments = apply_new_segmentation(customers_df)
else:
    segments = apply_old_segmentation(customers_df)
```

Pattern 2 - Parallel pipelines:

```yaml
# Asset Bundle with feature flag
resources:
  jobs:
    customer_seg_v1:
      enabled: ${var.use_legacy_pipeline}
    customer_seg_v2:
      enabled: ${var.use_new_pipeline}
```

**Flag Lifecycle:**

1. **Create**: Deploy with flag OFF
2. **Enable**: Gradually roll out (1% → 10% → 50% → 100%)
3. **Validate**: Monitor data quality, performance
4. **Stabilize**: Iterate based on results
5. **Remove**: Delete flag after 30-90 days (avoid tech debt)

**Environment**: Production (runtime control)

---

## Summary: Quick Reference

| Stage            | Data Artifact      | Key Activities                       | Duration   |
| ---------------- | ------------------ | ------------------------------------ | ---------- |
| 1. Authoring     | All                | Write specs, develop locally         | hours-days |
| 2. Pre-commit    | All                | Unit tests, linting, validation      | 5-10 min   |
| 3. Merge Request | All                | Code review, CI checks               | hours      |
| 4. Commit        | All                | Build artifacts, trigger tests       | 5-10 min   |
| 5. Acceptance    | Pipelines, Models  | PLTE deployment, integration tests   | 15-60 min  |
| 6. Extended      | Pipelines, Models  | Performance, security, quality tests | hours      |
| 7. Exploration   | Dashboards, Models | Stakeholder validation, UAT          | ongoing    |
| 8. Start Release | All                | Version, release notes               | minutes    |
| 9. Approval      | All                | RA: manual approval, CDe: auto       | mins-days  |
| 10. Deployment   | All                | Production deployment                | 5-30 min   |
| 11. Live         | All                | Monitor quality, performance, cost   | ongoing    |
| 12. Toggling     | Transformations    | Feature flags, gradual rollout       | as needed  |

## Variant Selection: RA vs CDe

**Use Release Approval (RA) when:**

- Regulated industry (finance, healthcare, government)
- Schema changes affect downstream systems
- High-risk transformations (PII, financial calculations)
- Coordinated releases with other teams
- Manual data validation required

**Use Continuous Deployment (CDe) when:**

- Internal analytics with low business risk
- Experimental models or features
- Well-tested transformations with strong quality gates
- Team has mature testing practices
- Fast feedback loops critical

**Hybrid Approach:**

- CDe for non-production environments
- RA for production with automated approval for low-risk changes
- Risk-based routing (low-risk → auto, high-risk → manual review)

## Next Steps

- [Develop Data Pipeline with Specs](./develop-data-pipeline-with-specs.md) - End-to-end walkthrough
- [Testing Data Pipelines](./testing-data-pipelines.md) - Comprehensive testing strategy
- [CI/CD Pipeline for Databricks](./cicd-pipeline-databricks.md) - Automation implementation
