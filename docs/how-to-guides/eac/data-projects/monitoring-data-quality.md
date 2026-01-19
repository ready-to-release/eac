# Monitoring Data Quality

## What You'll Learn

Monitor data pipelines, data quality, and pipeline health in production (Stage 11: Live).

## Prerequisites

- Databricks workspace with production pipelines
- Unity Catalog enabled
- Access to job monitoring
- Alert notification channels configured (email, Slack, PagerDuty)

## Monitoring Strategy

### Four Pillars of Data Monitoring

| Category             | What to Monitor                                | Tools                                         |
| -------------------- | ---------------------------------------------- | --------------------------------------------- |
| **Pipeline Health**  | Job success/failure, duration, SLA compliance  | Databricks Jobs, Workflow monitoring          |
| **Data Quality**     | Freshness, completeness, accuracy, consistency | Delta Live Tables expectations, custom checks |
| **Business Metrics** | Row counts, aggregate values, anomalies        | Custom dashboards, SQL queries                |
| **Infrastructure**   | Cluster utilization, costs, performance        | Databricks system tables, billing API         |

## Pipeline Health Monitoring

### Job Run Monitoring

```python
# notebooks/monitor_pipeline_health.py
from databricks.sdk import WorkspaceClient
from datetime import datetime, timedelta

def check_pipeline_health(job_id: int, lookback_hours: int = 24):
    """Monitor job run success rate and duration."""
    w = WorkspaceClient()

    # Get recent runs
    runs = w.jobs.list_runs(
        job_id=job_id,
        completed_only=True,
        start_time_from=int((datetime.now() - timedelta(hours=lookback_hours)).timestamp() * 1000)
    )

    total_runs = 0
    failed_runs = 0
    durations = []

    for run in runs:
        total_runs += 1
        if run.state.result_state != "SUCCESS":
            failed_runs += 1
        durations.append(run.execution_duration / 1000 / 60)  # Convert to minutes

    # Calculate metrics
    success_rate = ((total_runs - failed_runs) / total_runs * 100) if total_runs > 0 else 0
    avg_duration = sum(durations) / len(durations) if durations else 0

    # Log metrics
    spark.createDataFrame([{
        "timestamp": datetime.now(),
        "job_id": job_id,
        "total_runs": total_runs,
        "failed_runs": failed_runs,
        "success_rate": success_rate,
        "avg_duration_minutes": avg_duration
    }]).write.mode("append").saveAsTable("production.monitoring.pipeline_health")

    # Alert if success rate below threshold
    if success_rate < 99:
        send_alert(f"⚠️ Pipeline health degraded: {success_rate:.1f}% success rate")

    return {
        "success_rate": success_rate,
        "avg_duration": avg_duration
    }
```

### SLA Compliance Tracking

```python
# notebooks/track_sla_compliance.py
def check_sla_compliance(table_name: str, max_delay_hours: int = 2):
    """Check if table data meets freshness SLA."""
    # Get latest data timestamp
    latest_df = spark.sql(f"""
        SELECT MAX(event_date) as latest_date
        FROM {table_name}
    """)

    latest_date = latest_df.first().latest_date
    current_date = datetime.now().date()

    delay_hours = (current_date - latest_date).total_seconds() / 3600

    # Log SLA status
    sla_met = delay_hours <= max_delay_hours

    spark.createDataFrame([{
        "timestamp": datetime.now(),
        "table_name": table_name,
        "latest_date": latest_date,
        "delay_hours": delay_hours,
        "sla_met": sla_met
    }]).write.mode("append").saveAsTable("production.monitoring.sla_compliance")

    if not sla_met:
        send_alert(f"🚨 SLA breach: {table_name} data is {delay_hours:.1f} hours old (SLA: {max_delay_hours}h)")

    return sla_met
```

## Data Quality Monitoring

### Freshness Checks

```python
# notebooks/check_data_freshness.py
def monitor_table_freshness(tables: list[str]):
    """Monitor data freshness across critical tables."""
    results = []

    for table in tables:
        result = spark.sql(f"""
            SELECT
                '{table}' as table_name,
                MAX(event_date) as latest_date,
                COUNT(*) as row_count,
                CURRENT_DATE() as check_date
            FROM {table}
        """).first()

        days_old = (result.check_date - result.latest_date).days

        results.append({
            "table_name": table,
            "latest_date": result.latest_date,
            "days_old": days_old,
            "row_count": result.row_count,
            "status": "OK" if days_old <= 1 else "STALE"
        })

    # Write monitoring results
    spark.createDataFrame(results) \
        .write.mode("append") \
        .saveAsTable("production.monitoring.freshness_checks")

    # Alert on stale tables
    stale_tables = [r for r in results if r["status"] == "STALE"]
    if stale_tables:
        send_alert(f"⚠️ Stale data detected: {', '.join([t['table_name'] for t in stale_tables])}")
```

### Completeness Checks

```python
# notebooks/check_data_completeness.py
def check_data_completeness(table_name: str, required_columns: list[str]):
    """Check for NULL values in required columns."""
    null_checks = []

    for column in required_columns:
        null_count = spark.sql(f"""
            SELECT COUNT(*) as null_count
            FROM {table_name}
            WHERE {column} IS NULL
        """).first().null_count

        total_count = spark.table(table_name).count()
        completeness_pct = ((total_count - null_count) / total_count * 100) if total_count > 0 else 0

        null_checks.append({
            "table_name": table_name,
            "column_name": column,
            "null_count": null_count,
            "total_count": total_count,
            "completeness_pct": completeness_pct,
            "check_date": datetime.now()
        })

        if completeness_pct < 99:
            send_alert(f"⚠️ Data quality issue: {table_name}.{column} is {completeness_pct:.1f}% complete")

    # Write results
    spark.createDataFrame(null_checks) \
        .write.mode("append") \
        .saveAsTable("production.monitoring.completeness_checks")

    return null_checks
```

### Accuracy and Consistency Checks

```python
# notebooks/check_data_accuracy.py
def check_referential_integrity(parent_table: str, child_table: str, key_column: str):
    """Check referential integrity between tables."""
    orphaned_records = spark.sql(f"""
        SELECT COUNT(*) as orphan_count
        FROM {child_table} c
        LEFT JOIN {parent_table} p
            ON c.{key_column} = p.{key_column}
        WHERE p.{key_column} IS NULL
    """).first().orphan_count

    if orphaned_records > 0:
        send_alert(f"🚨 Referential integrity violation: {orphaned_records} orphaned records in {child_table}")

    return orphaned_records == 0

def check_business_rules(table_name: str):
    """Validate business rules on data."""
    violations = []

    # Rule 1: Total purchases should equal sum of individual purchases
    check1 = spark.sql(f"""
        SELECT COUNT(*) as violation_count
        FROM {table_name}
        WHERE ABS(total_purchases - (purchase_count * avg_order_value)) > 0.01
    """).first().violation_count

    if check1 > 0:
        violations.append(f"Total purchases mismatch: {check1} records")

    # Rule 2: Last purchase date should be >= first purchase date
    check2 = spark.sql(f"""
        SELECT COUNT(*) as violation_count
        FROM {table_name}
        WHERE last_purchase < first_purchase
    """).first().violation_count

    if check2 > 0:
        violations.append(f"Invalid date range: {check2} records")

    if violations:
        send_alert(f"🚨 Business rule violations in {table_name}: {', '.join(violations)}")

    return len(violations) == 0
```

### Simple Anomaly Detection

```python
# notebooks/check_data_volume.py
def check_data_volume_changes(table_name: str):
    """Alert on significant data volume changes (simple threshold-based)."""
    # Compare today vs yesterday
    result = spark.sql(f"""
        WITH daily_counts AS (
            SELECT
                event_date,
                COUNT(*) as row_count
            FROM {table_name}
            WHERE event_date >= CURRENT_DATE() - INTERVAL 2 DAYS
            GROUP BY event_date
        )
        SELECT
            MAX(CASE WHEN event_date = CURRENT_DATE() THEN row_count END) as today,
            MAX(CASE WHEN event_date = CURRENT_DATE() - INTERVAL 1 DAY THEN row_count END) as yesterday
        FROM daily_counts
    """).first()

    if result.today is None:
        print(f"⚠️ No data for today in {table_name}")
        return False

    # Alert if today's volume is less than 50% of yesterday
    if result.yesterday and result.today < result.yesterday * 0.5:
        print(f"🚨 Significant data drop in {table_name}: "
              f"Today={result.today}, Yesterday={result.yesterday}")
        return False

    print(f"✅ Data volume normal: Today={result.today}, Yesterday={result.yesterday}")
    return True
```

## Simple Quality Checks

### Basic Data Validation

```python
# notebooks/validate_business_rules.py
def validate_segment_business_rules(catalog: str):
    """Validate segmentation business rules are applied correctly."""
    violations = spark.sql(f"""
        SELECT
            'VIP with low purchases' as violation_type,
            COUNT(*) as count
        FROM {catalog}.gold.customer_segments s
        JOIN {catalog}.silver.customer_features f USING (customer_id)
        WHERE s.segment = 'VIP' AND f.total_purchases <= 10000

        UNION ALL

        SELECT
            'Active but not recent' as violation_type,
            COUNT(*) as count
        FROM {catalog}.gold.customer_segments s
        JOIN {catalog}.silver.customer_features f USING (customer_id)
        WHERE s.segment = 'Active' AND f.days_since_last > 30

        UNION ALL

        SELECT
            'Churned but recent' as violation_type,
            COUNT(*) as count
        FROM {catalog}.gold.customer_segments s
        JOIN {catalog}.silver.customer_features f USING (customer_id)
        WHERE s.segment = 'Churned' AND f.days_since_last <= 90
    """).collect()

    has_violations = any(row.count > 0 for row in violations)

    if has_violations:
        for row in violations:
            if row.count > 0:
                print(f"🚨 Rule violation: {row.violation_type} ({row.count} records)")
    else:
        print("✅ All business rules validated successfully")

    return not has_violations
```

### Monitor DLT Expectations

```python
# notebooks/monitor_dlt_expectations.py
def monitor_dlt_expectations(pipeline_id: str):
    """Monitor DLT pipeline expectations."""
    expectations_df = spark.sql(f"""
        SELECT
            dataset,
            name as expectation_name,
            passed_records,
            failed_records,
            CAST(passed_records AS DOUBLE) / (passed_records + failed_records) * 100 as pass_rate
        FROM event_log('{pipeline_id}')
        WHERE event_type = 'flow_progress'
          AND details:flow_progress.data_quality.expectations IS NOT NULL
    """)

    # Alert on low pass rates
    for row in expectations_df.collect():
        if row.pass_rate < 99:
            send_alert(f"⚠️ DLT expectation failing: {row.dataset}.{row.expectation_name} "
                      f"pass rate: {row.pass_rate:.1f}%")
```

## Business Metrics Monitoring

### Custom Dashboards

```sql
-- Create monitoring dashboard views
CREATE OR REPLACE VIEW production.monitoring.daily_metrics AS
SELECT
    DATE(event_date) as metric_date,
    COUNT(DISTINCT customer_id) as active_customers,
    COUNT(*) as total_purchases,
    SUM(amount) as total_revenue,
    AVG(amount) as avg_purchase_value
FROM production.bronze.customer_events
WHERE event_date >= CURRENT_DATE() - INTERVAL 90 DAYS
GROUP BY DATE(event_date)
ORDER BY metric_date DESC;
```

### Threshold-Based Alerts

```python
# notebooks/monitor_business_metrics.py
def monitor_daily_metrics():
    """Monitor key business metrics against thresholds."""
    today_metrics = spark.sql("""
        SELECT *
        FROM production.monitoring.daily_metrics
        WHERE metric_date = CURRENT_DATE()
    """).first()

    alerts = []

    # Check thresholds
    if today_metrics.active_customers < 1000:
        alerts.append(f"Low active customers: {today_metrics.active_customers}")

    if today_metrics.total_revenue < 50000:
        alerts.append(f"Low revenue: ${today_metrics.total_revenue:,.2f}")

    # Compare to 7-day average
    avg_metrics = spark.sql("""
        SELECT AVG(total_purchases) as avg_purchases
        FROM production.monitoring.daily_metrics
        WHERE metric_date >= CURRENT_DATE() - INTERVAL 7 DAYS
          AND metric_date < CURRENT_DATE()
    """).first()

    if today_metrics.total_purchases < avg_metrics.avg_purchases * 0.8:
        alerts.append(f"Purchases down 20% from 7-day average")

    if alerts:
        send_alert(f"🚨 Business metrics alert: {', '.join(alerts)}")
```

## Infrastructure Monitoring

### Cluster Utilization

```python
# notebooks/monitor_cluster_utilization.py
def monitor_cluster_costs():
    """Monitor DBU consumption and costs."""
    cost_df = spark.sql("""
        SELECT
            DATE(usage_start_time) as usage_date,
            cluster_id,
            SUM(usage_quantity) as total_dbus,
            AVG(usage_quantity) as avg_dbus_per_hour
        FROM system.billing.usage
        WHERE usage_date >= CURRENT_DATE() - INTERVAL 7 DAYS
        GROUP BY DATE(usage_start_time), cluster_id
    """)

    # Alert on high costs
    for row in cost_df.collect():
        if row.total_dbus > 1000:  # Threshold
            send_alert(f"⚠️ High DBU usage: Cluster {row.cluster_id} used {row.total_dbus} DBUs on {row.usage_date}")
```

## Alerting and Incident Response

### Alert Configuration

```python
# utils/alerting.py
import requests

def send_alert(message: str, severity: str = "warning"):
    """Send alert via multiple channels."""
    # Slack webhook
    slack_webhook = dbutils.secrets.get("alerts", "slack-webhook")
    requests.post(slack_webhook, json={"text": message})

    # Email (via SendGrid/SES)
    send_email(
        to=["data-team@company.com"],
        subject=f"[{severity.upper()}] Data Pipeline Alert",
        body=message
    )

    # PagerDuty for critical alerts
    if severity == "critical":
        trigger_pagerduty_incident(message)
```

### Incident Response Runbook

```python
# notebooks/incident_response.py
def handle_pipeline_failure(job_id: int, run_id: int):
    """Automated incident response for pipeline failures."""
    w = WorkspaceClient()

    # Get failure details
    run = w.jobs.get_run(run_id)
    error_message = run.state.state_message

    # Determine root cause
    if "OutOfMemoryError" in error_message:
        # Increase cluster size and retry
        send_alert("🔧 Auto-remediation: Increasing cluster size and retrying")
        # Update job config and retry

    elif "FileNotFoundException" in error_message:
        # Data source issue - alert data team
        send_alert("🚨 Data source missing - manual intervention required", severity="critical")

    else:
        # Unknown error - alert on-call engineer
        send_alert(f"🚨 Pipeline failure: {error_message}", severity="critical")
```

## Monitoring Dashboard

### Create Centralized Dashboard

```sql
-- Create unified monitoring view
CREATE OR REPLACE VIEW production.monitoring.health_dashboard AS
SELECT
    'Pipeline Health' as category,
    job_id,
    success_rate,
    avg_duration_minutes as value,
    CASE
        WHEN success_rate < 99 THEN 'CRITICAL'
        WHEN success_rate < 99.5 THEN 'WARNING'
        ELSE 'OK'
    END as status
FROM production.monitoring.pipeline_health
WHERE timestamp >= CURRENT_TIMESTAMP() - INTERVAL 1 HOUR

UNION ALL

SELECT
    'Data Freshness' as category,
    table_name as job_id,
    NULL as success_rate,
    delay_hours as value,
    CASE
        WHEN NOT sla_met THEN 'CRITICAL'
        ELSE 'OK'
    END as status
FROM production.monitoring.sla_compliance
WHERE timestamp >= CURRENT_TIMESTAMP() - INTERVAL 1 HOUR;
```

## Next Steps

- [CI/CD Pipeline for Databricks](./cicd-pipeline-databricks.md) - Automate monitoring setup
- [Testing Data Pipelines](./testing-data-pipelines.md) - Prevent issues with comprehensive testing
