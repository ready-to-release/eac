# Testing Data Pipelines

## What You'll Learn

Comprehensive testing strategy for data pipelines including unit tests, integration tests, and data quality validation.

## Prerequisites

- pytest installed (`pip install pytest`)
- chispa for Spark DataFrame assertions (`pip install chispa`)
- Local Spark environment or Databricks Connect

## Test Pyramid for Data Pipelines

```text
         /\
        /  \  L3: System Tests (Full pipeline, production-like data)
       /----\
      /      \  L2: Integration Tests (Multi-notebook, test data)
     /--------\
    /          \  L1: Component Tests (Single notebook end-to-end)
   /------------\
  /              \ L0: Unit Tests (Transformation functions)
 /----------------\
```

**Test Distribution:**

- L0 (Unit): 70% of tests, < 1 second each
- L1 (Component): 20% of tests, < 10 seconds each
- L2 (Integration): 8% of tests, < 2 minutes each
- L3 (System): 2% of tests, < 30 minutes each

## L0: Unit Tests

Test individual transformation functions with small, synthetic DataFrames.

### Example: Test Feature Aggregation

```python
# tests/unit/test_aggregate_features.py
import pytest
from pyspark.sql import SparkSession
from pyspark.sql.types import StructType, StructField, LongType, StringType, DecimalType
from chispa.dataframe_comparer import assert_df_equality
from src.silver.aggregate_customer_features import aggregate_customer_features

@pytest.fixture(scope="session")
def spark():
    return SparkSession.builder \
        .master("local[1]") \
        .appName("unit-tests") \
        .getOrCreate()

def test_aggregate_customer_features_basic(spark):
    """Test basic feature aggregation for single customer."""
    # Arrange
    input_data = [
        (1, "2024-01-01", 100.0),
        (1, "2024-01-05", 150.0),
        (1, "2024-01-10", 200.0),
    ]
    input_df = spark.createDataFrame(input_data, ["customer_id", "event_date", "amount"])

    expected_data = [(1, 450.0, 3, 150.0)]
    expected_df = spark.createDataFrame(
        expected_data,
        ["customer_id", "total_purchases", "purchase_count", "avg_order_value"]
    )

    # Act
    result_df = aggregate_customer_features(input_df)

    # Assert
    assert_df_equality(
        result_df.select("customer_id", "total_purchases", "purchase_count", "avg_order_value"),
        expected_df,
        ignore_nullable=True
    )

def test_aggregate_handles_single_purchase(spark):
    """Test aggregation with customer who has only one purchase."""
    input_data = [(1, "2024-01-01", 100.0)]
    input_df = spark.createDataFrame(input_data, ["customer_id", "event_date", "amount"])

    result_df = aggregate_customer_features(input_df)

    assert result_df.count() == 1
    row = result_df.first()
    assert row.customer_id == 1
    assert row.purchase_count == 1
    assert row.avg_order_value == 100.0

def test_aggregate_handles_empty_input(spark):
    """Test aggregation with no input data."""
    schema = StructType([
        StructField("customer_id", LongType()),
        StructField("event_date", StringType()),
        StructField("amount", DecimalType(10, 2))
    ])
    empty_df = spark.createDataFrame([], schema)

    result_df = aggregate_customer_features(empty_df)

    assert result_df.count() == 0

def test_aggregate_multiple_customers(spark):
    """Test aggregation with multiple customers."""
    input_data = [
        (1, "2024-01-01", 100.0),
        (1, "2024-01-05", 200.0),
        (2, "2024-01-03", 150.0),
        (3, "2024-01-07", 300.0),
    ]
    input_df = spark.createDataFrame(input_data, ["customer_id", "event_date", "amount"])

    result_df = aggregate_customer_features(input_df)

    assert result_df.count() == 3
    assert result_df.filter("customer_id = 1").first().purchase_count == 2
    assert result_df.filter("customer_id = 2").first().purchase_count == 1
```

### Example: Test Data Quality Rules

```python
# tests/unit/test_data_quality.py
from pyspark.sql.functions import col

def test_no_negative_amounts(spark):
    """Test that negative amounts are filtered out."""
    input_data = [
        (1, "2024-01-01", 100.0),
        (2, "2024-01-02", -50.0),  # Invalid
        (3, "2024-01-03", 200.0),
    ]
    input_df = spark.createDataFrame(input_data, ["customer_id", "event_date", "amount"])

    # Apply filter
    cleaned_df = input_df.filter(col("amount") > 0)

    assert cleaned_df.count() == 2
    assert cleaned_df.filter("customer_id = 2").count() == 0

def test_deduplication(spark):
    """Test that duplicate records are removed."""
    input_data = [
        (1, "2024-01-01", 100.0, "evt1"),
        (1, "2024-01-01", 100.0, "evt1"),  # Duplicate
        (1, "2024-01-02", 150.0, "evt2"),
    ]
    input_df = spark.createDataFrame(input_data, ["customer_id", "event_date", "amount", "event_id"])

    deduped_df = input_df.dropDuplicates(["customer_id", "event_date", "event_id"])

    assert deduped_df.count() == 2
```

## L1: Component Tests

Test entire notebooks end-to-end with test data.

### Example: Test Bronze Ingestion Notebook

```python
# tests/component/test_ingest_notebook.py
import pytest
from databricks.connect import DatabricksSession
from pyspark.sql import DataFrame

@pytest.fixture(scope="session")
def spark():
    return DatabricksSession.builder.remote(
        host="https://staging.cloud.databricks.com",
        token=os.getenv("DATABRICKS_TOKEN")
    ).getOrCreate()

def test_bronze_ingestion_end_to_end(spark):
    """Test complete bronze ingestion process."""
    # Arrange - Create test source data
    test_data_path = "s3://test-bucket/customer-events-test/"
    test_data = [
        (1, "2024-01-01 10:00:00", "purchase", 100.0, "evt1"),
        (1, "2024-01-02 11:00:00", "purchase", 150.0, "evt2"),
        (2, "2024-01-01 12:00:00", "purchase", 200.0, "evt3"),
    ]
    test_df = spark.createDataFrame(
        test_data,
        ["customer_id", "event_timestamp", "event_type", "amount", "event_id"]
    )
    test_df.write.mode("overwrite").parquet(test_data_path)

    # Act - Run notebook
    spark.conf.set("catalog", "test")
    spark.conf.set("source_path", test_data_path)

    dbutils.notebook.run(
        "./src/bronze/ingest_customer_events",
        timeout_seconds=300,
        arguments={}
    )

    # Assert - Verify output table
    result_df = spark.table("test.bronze.customer_events")

    assert result_df.count() == 3
    assert result_df.filter("customer_id = 1").count() == 2
    assert result_df.schema.fieldNames() == [
        "customer_id", "event_date", "amount", "event_id", "source_file"
    ]

    # Cleanup
    spark.sql("DROP TABLE IF EXISTS test.bronze.customer_events")
```

### Example: Test Silver Transformation Notebook

```python
# tests/component/test_aggregate_notebook.py
def test_silver_aggregation_end_to_end(spark):
    """Test complete silver layer aggregation."""
    # Arrange - Create bronze test data
    spark.sql("CREATE SCHEMA IF NOT EXISTS test.bronze")
    bronze_data = [
        (1, "2024-01-01", 100.0, "evt1"),
        (1, "2024-01-05", 150.0, "evt2"),
        (2, "2024-01-02", 200.0, "evt3"),
    ]
    bronze_df = spark.createDataFrame(
        bronze_data,
        ["customer_id", "event_date", "amount", "event_id"]
    )
    bronze_df.write.mode("overwrite").saveAsTable("test.bronze.customer_events")

    # Act - Run aggregation notebook
    spark.conf.set("catalog", "test")
    dbutils.notebook.run(
        "./src/silver/aggregate_customer_features",
        timeout_seconds=300
    )

    # Assert
    result_df = spark.table("test.silver.customer_features")

    assert result_df.count() == 2

    customer_1 = result_df.filter("customer_id = 1").first()
    assert customer_1.total_purchases == 250.0
    assert customer_1.purchase_count == 2

    # Cleanup
    spark.sql("DROP TABLE IF EXISTS test.silver.customer_features")
    spark.sql("DROP TABLE IF EXISTS test.bronze.customer_events")
```

## L2: Integration Tests

Test multiple notebooks together with realistic test data.

### Example: Test Complete Pipeline

```python
# tests/integration/test_pipeline_integration.py
def test_complete_segmentation_pipeline(spark):
    """Test bronze → silver → gold pipeline flow."""
    # Arrange - Setup test catalog
    spark.sql("CREATE CATALOG IF NOT EXISTS test_integration")
    spark.sql("USE CATALOG test_integration")
    spark.sql("CREATE SCHEMA IF NOT EXISTS bronze")
    spark.sql("CREATE SCHEMA IF NOT EXISTS silver")
    spark.sql("CREATE SCHEMA IF NOT EXISTS gold")

    # Create realistic test dataset (100 customers, 1000 events)
    test_events = generate_test_customer_events(n_customers=100, n_events_per_customer=10)
    test_events.write.mode("overwrite").saveAsTable("bronze.customer_events")

    # Act - Run pipeline notebooks in sequence
    spark.conf.set("catalog", "test_integration")

    # Step 1: Ingest (skip for integration test, using pre-loaded data)
    # Step 2: Aggregate features
    dbutils.notebook.run("./src/silver/aggregate_customer_features", 300)

    # Step 3: Segment customers
    dbutils.notebook.run("./src/gold/segment_customers", 300)

    # Assert - Verify pipeline output
    segments_df = spark.table("gold.customer_segments")

    assert segments_df.count() == 100, "Should have segment for each customer"

    # Check segment distribution
    distribution = segments_df.groupBy("segment").count().collect()
    assert len(distribution) == 4, "Should have 4 segments"

    for row in distribution:
        pct = row['count'] / 100.0
        assert 0.15 <= pct <= 0.40, f"Segment {row.segment} has {pct:.1%}, expected 15-40%"

    # Check data quality
    assert segments_df.filter("customer_id IS NULL").count() == 0
    assert segments_df.filter("segment IS NULL").count() == 0
    assert segments_df.count() == segments_df.select("customer_id").distinct().count()

    # Cleanup
    spark.sql("DROP SCHEMA test_integration.gold CASCADE")
    spark.sql("DROP SCHEMA test_integration.silver CASCADE")
    spark.sql("DROP SCHEMA test_integration.bronze CASCADE")

def generate_test_customer_events(n_customers: int, n_events_per_customer: int) -> DataFrame:
    """Generate realistic test data."""
    import random
    from datetime import datetime, timedelta

    events = []
    for customer_id in range(1, n_customers + 1):
        start_date = datetime.now() - timedelta(days=90)
        for i in range(n_events_per_customer):
            event_date = start_date + timedelta(days=random.randint(0, 89))
            amount = random.uniform(10, 500)
            events.append((customer_id, event_date.strftime("%Y-%m-%d"), amount, f"evt_{customer_id}_{i}"))

    return spark.createDataFrame(events, ["customer_id", "event_date", "amount", "event_id"])
```

## L3: System Tests

Test complete pipeline with production-scale data in PLTE.

### Example: System Test in Staging

```python
# tests/system/test_production_scale.py
def test_pipeline_with_production_scale_data(spark):
    """Test pipeline with production-scale data in staging environment."""
    # This test runs in staging environment with cloned production data

    # Arrange - Clone production data to staging
    spark.sql("""
        CREATE OR REPLACE TABLE staging.bronze.customer_events
        SHALLOW CLONE production.bronze.customer_events
    """)

    # Act - Run complete pipeline
    job_run = dbutils.jobs.run_now(job_id=<staging_job_id>)
    run_id = job_run['run_id']

    # Wait for completion (with timeout)
    dbutils.jobs.wait_for_run(run_id, timeout_seconds=1800)

    run_output = dbutils.jobs.get_run_output(run_id)
    assert run_output['state']['life_cycle_state'] == 'TERMINATED'
    assert run_output['state']['result_state'] == 'SUCCESS'

    # Assert - Verify output quality
    segments_df = spark.table("staging.gold.customer_segments")

    # Volume checks
    assert segments_df.count() > 10000, "Should have segments for >10k customers"

    # Performance checks
    assert run_output['execution_duration'] < 1800000, "Should complete in < 30 min"

    # Data quality checks
    quality_results = run_data_quality_checks(segments_df)
    assert quality_results['completeness'] > 0.99
    assert quality_results['uniqueness'] == 1.0
```

## Data Quality Testing

### Basic Data Quality Checks

```python
# tests/quality/test_segment_quality.py
from pyspark.sql.functions import col

def test_customer_segments_quality(spark):
    """Test data quality for customer segments using simple assertions."""
    segments_df = spark.table("staging.gold.customer_segments")

    # Check no nulls in key columns
    null_customer_ids = segments_df.filter(col("customer_id").isNull()).count()
    assert null_customer_ids == 0, f"Found {null_customer_ids} null customer_ids"

    null_segments = segments_df.filter(col("segment").isNull()).count()
    assert null_segments == 0, f"Found {null_segments} null segments"

    # Check unique customer_ids
    total_count = segments_df.count()
    unique_count = segments_df.select("customer_id").distinct().count()
    assert total_count == unique_count, f"Duplicate customer_ids found: {total_count - unique_count}"

    # Check valid segment values
    valid_segments = ["VIP", "Active", "At-Risk", "Churned"]
    invalid_segments = segments_df.filter(~col("segment").isin(valid_segments)).count()
    assert invalid_segments == 0, f"Found {invalid_segments} invalid segments"

    # Check all segments exist (business rule validation)
    segment_counts = {row.segment: row.count for row in segments_df.groupBy("segment").count().collect()}
    for segment in valid_segments:
        assert segment in segment_counts, f"Missing segment: {segment}"
```

### Schema Validation

```python
# tests/quality/test_schema_validation.py
from pyspark.sql.types import StructType, StructField, LongType, StringType, DateType

def test_customer_segments_schema(spark):
    """Verify customer_segments table schema matches specification."""
    expected_schema = StructType([
        StructField("customer_id", LongType(), nullable=False),
        StructField("segment", StringType(), nullable=False),
        StructField("assigned_date", DateType(), nullable=False),
    ])

    actual_df = spark.table("staging.gold.customer_segments")

    # Compare schemas
    assert actual_df.schema == expected_schema, (
        f"Schema mismatch:\n"
        f"Expected: {expected_schema}\n"
        f"Actual: {actual_df.schema}"
    )

    # Additional validation: column names
    expected_columns = {"customer_id", "segment", "assigned_date"}
    actual_columns = set(actual_df.columns)
    assert actual_columns == expected_columns, f"Column mismatch: expected {expected_columns}, got {actual_columns}"
```

## Test Data Management

### Synthetic Data Generation

```python
# tests/fixtures/data_generators.py
from faker import Faker
import random

def generate_customer_events(n_customers=100, days=90):
    """Generate synthetic customer event data."""
    fake = Faker()
    events = []

    for customer_id in range(1, n_customers + 1):
        n_events = random.randint(1, 20)
        for _ in range(n_events):
            event_date = fake.date_between(start_date=f'-{days}d', end_date='today')
            amount = round(random.uniform(10, 500), 2)
            events.append({
                'customer_id': customer_id,
                'event_date': event_date.isoformat(),
                'amount': amount,
                'event_id': fake.uuid4()
            })

    return spark.createDataFrame(events)
```

### Production Data Masking

```python
# tests/fixtures/data_masking.py
def mask_production_data(source_table: str, dest_table: str):
    """Create masked copy of production data for testing."""
    masked_df = (spark.table(source_table)
        .withColumn("customer_id", hash(col("customer_id")).cast("bigint"))
        .withColumn("email", concat(lit("test_"), hash(col("email")).cast("string")))
        .limit(10000)  # Sample for performance
    )

    masked_df.write.mode("overwrite").saveAsTable(dest_table)
```

## Running Tests

### Local Execution

```bash
# Run all tests
pytest tests/ -v

# Run specific test level
pytest tests/unit/ -v
pytest tests/integration/ -v

# Run with coverage
pytest tests/ --cov=src --cov-report=html

# Run parallel
pytest tests/unit/ -n auto
```

### CI Pipeline

```yaml
# .github/workflows/test.yml
- name: Run Unit Tests
  run: pytest tests/unit/ -v --junitxml=test-results/unit.xml

- name: Run Component Tests
  run: |
    databricks configure --token <<< "$DATABRICKS_TOKEN"
    pytest tests/component/ -v --junitxml=test-results/component.xml

- name: Run Integration Tests
  run: pytest tests/integration/ -v --junitxml=test-results/integration.xml
  if: github.event_name == 'push' && github.ref == 'refs/heads/main'
```

## Next Steps

- [CI/CD Pipeline for Databricks](./cicd-pipeline-databricks.md) - Automate test execution
- [Environment Management](./environment-management.md) - Setup PLTE for testing
- [Monitoring Data Quality](./monitoring-data-quality.md) - Production data quality monitoring
