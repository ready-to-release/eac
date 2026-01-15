# ML Model Lifecycle

## What You'll Learn

Apply the 12-stage CD model to machine learning model development, testing, deployment, and monitoring using MLflow.

## Prerequisites

- MLflow installed (`pip install mlflow`)
- Databricks workspace with ML Runtime
- Unity Catalog for model registry
- Understanding of CD model stages

## ML Model Through CD Stages

### Stage 1: Authoring - Experiment Tracking

Track experiments during model development:

```python
# notebooks/train_ltv_model.py
import mlflow
import mlflow.sklearn
from sklearn.ensemble import RandomForestRegressor
from sklearn.model_selection import train_test_split

# Set experiment
mlflow.set_experiment("/Users/data-scientist/customer-ltv-experiments")

# Load features
features_df = spark.table("dev.silver.customer_features")
X = features_df.select("total_purchases", "purchase_count", "avg_order_value", "days_since_last").toPandas()
y = features_df.select("actual_ltv").toPandas()

X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

# Start MLflow run
with mlflow.start_run(run_name="rf_baseline") as run:
    # Log parameters
    mlflow.log_param("model_type", "random_forest")
    mlflow.log_param("n_estimators", 100)
    mlflow.log_param("max_depth", 10)

    # Train model
    model = RandomForestRegressor(n_estimators=100, max_depth=10, random_state=42)
    model.fit(X_train, y_train)

    # Evaluate
    from sklearn.metrics import mean_squared_error, r2_score
    predictions = model.predict(X_test)
    rmse = mean_squared_error(y_test, predictions, squared=False)
    r2 = r2_score(y_test, predictions)

    # Log metrics
    mlflow.log_metric("rmse", rmse)
    mlflow.log_metric("r2_score", r2)

    # Log model
    mlflow.sklearn.log_model(
        model,
        "model",
        registered_model_name="customer_ltv_dev"
    )

    print(f"Model RMSE: {rmse:.2f}, R²: {r2:.3f}")
```

### Stage 2: Pre-commit - Model Unit Tests

Test model training logic:

```python
# tests/unit/test_ltv_model.py
import pytest
from src.models.train_ltv import train_ltv_model, prepare_features

def test_feature_preparation():
    """Test feature preparation transforms data correctly."""
    test_data = spark.createDataFrame([
        (1, 100.0, 5, 20.0, 10),
        (2, 200.0, 10, 20.0, 5),
    ], ["customer_id", "total_purchases", "purchase_count", "avg_order_value", "days_since_last"])

    features = prepare_features(test_data)

    assert features.shape[0] == 2
    assert features.shape[1] == 4  # 4 feature columns
    assert not features.isnull().any().any()

def test_model_trains_successfully():
    """Test model training completes without errors."""
    X_train = pd.DataFrame({
        "total_purchases": [100, 200, 150],
        "purchase_count": [5, 10, 7],
        "avg_order_value": [20, 20, 21],
        "days_since_last": [10, 5, 7]
    })
    y_train = pd.Series([150, 300, 200])

    model, metrics = train_ltv_model(X_train, y_train)

    assert model is not None
    assert "rmse" in metrics
    assert "r2_score" in metrics
    assert metrics["rmse"] > 0

def test_model_predictions_shape():
    """Test model predictions return correct shape."""
    model = train_ltv_model(X_train, y_train)[0]
    X_test = X_train[:2]  # Use first 2 rows

    predictions = model.predict(X_test)

    assert len(predictions) == 2
    assert all(pred > 0 for pred in predictions)
```

### Stage 4-5: Commit & Acceptance - Model Validation

Validate model in staging environment:

```python
# notebooks/validate_model_staging.py
import mlflow

# Load model from registry
model_name = "customer_ltv"
model_version = "1"

model_uri = f"models:/{model_name}/{model_version}"
model = mlflow.pyfunc.load_model(model_uri)

# Load staging validation data
validation_df = spark.table("staging.silver.customer_features")
X_val = validation_df.select("total_purchases", "purchase_count", "avg_order_value", "days_since_last").toPandas()
y_val = validation_df.select("actual_ltv").toPandas()

# Make predictions
predictions = model.predict(X_val)

# Validate metrics
from sklearn.metrics import mean_squared_error, r2_score
rmse = mean_squared_error(y_val, predictions, squared=False)
r2 = r2_score(y_val, predictions)

# Log validation results to MLflow
with mlflow.start_run(run_name="staging_validation"):
    mlflow.log_param("environment", "staging")
    mlflow.log_param("model_version", model_version)
    mlflow.log_metric("validation_rmse", rmse)
    mlflow.log_metric("validation_r2", r2)

# Quality gates
assert rmse < 100, f"RMSE too high: {rmse}"
assert r2 > 0.8, f"R² too low: {r2}"

print(f"✅ Validation passed: RMSE={rmse:.2f}, R²={r2:.3f}")
```

### Stage 6: Extended Testing - Model Performance Tests

```python
# tests/integration/test_model_performance.py
def test_model_inference_latency():
    """Test model prediction latency meets SLA."""
    import time

    model = mlflow.pyfunc.load_model("models:/customer_ltv/1")
    test_data = generate_test_features(n=1000)

    # Measure latency
    start_time = time.time()
    predictions = model.predict(test_data)
    elapsed = time.time() - start_time

    latency_per_prediction = (elapsed / len(test_data)) * 1000  # milliseconds

    assert latency_per_prediction < 10, f"Latency too high: {latency_per_prediction:.2f}ms"

def test_model_bias_fairness():
    """Test model fairness across customer segments."""
    model = mlflow.pyfunc.load_model("models:/customer_ltv/1")

    # Get predictions by segment
    for segment in ["VIP", "Active", "At-Risk", "Churned"]:
        segment_data = load_segment_data(segment)
        predictions = model.predict(segment_data)

        # Check for bias
        mean_pred = np.mean(predictions)
        std_pred = np.std(predictions)

        # Ensure reasonable distribution
        assert std_pred > 0, f"No variance in predictions for {segment}"
        assert not np.isnan(mean_pred), f"NaN predictions for {segment}"

def test_model_data_drift():
    """Test model performance hasn't degraded due to data drift."""
    model = mlflow.pyfunc.load_model("models:/customer_ltv/Production")

    # Compare against baseline performance
    baseline_rmse = 85.0  # From initial validation
    current_rmse = evaluate_on_recent_data(model)

    drift_pct = ((current_rmse - baseline_rmse) / baseline_rmse) * 100

    assert drift_pct < 10, f"Performance drift detected: {drift_pct:.1f}% increase in RMSE"
```

## MLflow Model Registry

### Register Model

```python
# Register model to Unity Catalog
model_name = "customer_ltv"
model_version = mlflow.register_model(
    model_uri=f"runs:/{run_id}/model",
    name=f"staging.ml_models.{model_name}"
)

# Add description and tags
from mlflow.tracking import MlflowClient
client = MlflowClient()

client.update_model_version(
    name=f"staging.ml_models.{model_name}",
    version=model_version.version,
    description="Random Forest model predicting customer lifetime value"
)

client.set_model_version_tag(
    name=f"staging.ml_models.{model_name}",
    version=model_version.version,
    key="validation_status",
    value="passed"
)
```

### Promotion Workflow

```python
# Stage 9: Release Approval - Promote to Production

def promote_model_to_production(model_name: str, version: str):
    """Promote model from staging to production using Unity Catalog."""
    client = MlflowClient()

    # Verify staging validation passed
    staging_model = f"staging.ml_models.{model_name}"
    version_info = client.get_model_version(staging_model, version)

    validation_tag = version_info.tags.get("validation_status")
    assert validation_tag == "passed", "Model validation must pass before promotion"

    # Unity Catalog uses ALIASES instead of stages
    # Set "champion" alias in staging to mark this version as production-ready
    client.set_registered_model_alias(
        name=staging_model,
        alias="champion",
        version=version
    )
    print(f"✅ Set 'champion' alias on {staging_model} version {version}")

    # Register model to production catalog
    prod_model = f"production.ml_models.{model_name}"

    # Load model from staging and register to production
    model_uri = f"models:/{staging_model}/{version}"
    model_version = mlflow.register_model(
        model_uri=model_uri,
        name=prod_model
    )

    # Set production alias
    client.set_registered_model_alias(
        name=prod_model,
        alias="champion",
        version=model_version.version
    )

    # Add metadata tags
    client.set_model_version_tag(
        name=prod_model,
        version=model_version.version,
        key="promoted_from",
        value=f"{staging_model}/v{version}"
    )

    client.set_model_version_tag(
        name=prod_model,
        version=model_version.version,
        key="promotion_date",
        value=datetime.now().isoformat()
    )

    print(f"✅ Promoted {model_name} v{version} to production as v{model_version.version}")
    print(f"   Use URI: models:/{prod_model}@champion")
```

## Model Deployment

### Stage 10: Production Deployment - Batch Scoring

```python
# notebooks/batch_score_customers.py
import mlflow
from pyspark.sql.functions import struct

# Load production model using Unity Catalog alias
model_uri = "models:/production.ml_models.customer_ltv@champion"
model = mlflow.pyfunc.spark_udf(
    spark,
    model_uri=model_uri,
    result_type="double"
)

# Load customer features
customers_df = spark.table("production.silver.customer_features")

# Score all customers
# Note: Pass features as a struct to match model's expected input
scored_df = customers_df.withColumn(
    "predicted_ltv",
    model(struct(
        "total_purchases",
        "purchase_count",
        "avg_order_value",
        "days_since_last"
    ))
)

# Write predictions with merge for idempotency
from delta.tables import DeltaTable

target_table = "production.gold.customer_ltv_predictions"

if spark.catalog.tableExists(target_table):
    target = DeltaTable.forName(spark, target_table)
    target.alias("target").merge(
        scored_df.alias("source"),
        "target.customer_id = source.customer_id"
    ).whenMatchedUpdateAll() \
     .whenNotMatchedInsertAll() \
     .execute()
    print(f"✅ Merged predictions for {scored_df.count()} customers")
else:
    scored_df.write.mode("overwrite").saveAsTable(target_table)
    print(f"✅ Scored {scored_df.count()} customers")
```

### Real-Time Model Serving

```yaml
# databricks.yml - Model Serving Endpoint
resources:
  model_serving_endpoints:
    customer_ltv_serving:
      name: "customer-ltv-${bundle.target}"
      config:
        served_entities:
          - entity_name: "${var.catalog}.ml_models.customer_ltv"
            entity_version: "1"
            workload_size: "Small"
            scale_to_zero_enabled: true

        traffic_config:
          routes:
            - served_model_name: "${var.catalog}.ml_models.customer_ltv-1"
              traffic_percentage: 100
```

**Invoke Serving Endpoint:**

```python
import requests
import os

def predict_ltv(customer_features: dict) -> float:
    """Call model serving endpoint for real-time prediction."""
    endpoint_url = f"{os.getenv('DATABRICKS_HOST')}/serving-endpoints/customer-ltv-production/invocations"
    token = os.getenv("DATABRICKS_TOKEN")

    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }

    payload = {
        "dataframe_records": [customer_features]
    }

    response = requests.post(endpoint_url, json=payload, headers=headers)
    response.raise_for_status()

    return response.json()["predictions"][0]

# Example usage
prediction = predict_ltv({
    "total_purchases": 1250.0,
    "purchase_count": 15,
    "avg_order_value": 83.33,
    "days_since_last": 5
})
print(f"Predicted LTV: ${prediction:.2f}")
```

## Stage 11: Live - Model Monitoring

### Monitor Model Performance

```python
# notebooks/monitor_model_performance.py
import mlflow
from datetime import datetime, timedelta

def monitor_model_performance():
    """Monitor production model performance and data drift."""
    # Load production model
    model = mlflow.pyfunc.load_model("models:/production.ml_models.customer_ltv/Production")

    # Get recent predictions and actuals
    results_df = spark.sql("""
        SELECT
            p.customer_id,
            p.predicted_ltv,
            a.actual_ltv,
            p.prediction_date
        FROM production.gold.customer_ltv_predictions p
        JOIN production.gold.customer_actual_ltv a
            ON p.customer_id = a.customer_id
        WHERE p.prediction_date >= current_date() - INTERVAL 7 DAYS
    """)

    # Calculate metrics
    from sklearn.metrics import mean_squared_error, r2_score
    predictions = results_df.select("predicted_ltv").toPandas()
    actuals = results_df.select("actual_ltv").toPandas()

    rmse = mean_squared_error(actuals, predictions, squared=False)
    r2 = r2_score(actuals, predictions)

    # Log monitoring metrics
    with mlflow.start_run(run_name="production_monitoring"):
        mlflow.log_metric("production_rmse", rmse)
        mlflow.log_metric("production_r2", r2)
        mlflow.log_metric("num_predictions", results_df.count())

    # Alert if performance degraded
    baseline_rmse = 85.0
    if rmse > baseline_rmse * 1.1:
        send_alert(f"⚠️ Model performance degraded: RMSE={rmse:.2f} (baseline={baseline_rmse})")

    return {"rmse": rmse, "r2": r2}
```

### Data Drift Detection

```python
# notebooks/detect_data_drift.py
from scipy.stats import ks_2samp

def detect_feature_drift():
    """Detect drift in input features."""
    # Training data distribution
    training_df = spark.table("production.ml_models.customer_ltv_training_data")

    # Recent production data
    recent_df = spark.table("production.silver.customer_features") \
        .filter("feature_date >= current_date() - INTERVAL 7 DAYS")

    features = ["total_purchases", "purchase_count", "avg_order_value", "days_since_last"]

    drift_detected = False
    for feature in features:
        training_dist = training_df.select(feature).toPandas()[feature]
        recent_dist = recent_df.select(feature).toPandas()[feature]

        # Kolmogorov-Smirnov test
        statistic, p_value = ks_2samp(training_dist, recent_dist)

        mlflow.log_metric(f"drift_{feature}_ks_statistic", statistic)
        mlflow.log_metric(f"drift_{feature}_p_value", p_value)

        if p_value < 0.05:
            print(f"⚠️ Drift detected in {feature}: p={p_value:.4f}")
            drift_detected = True

    if drift_detected:
        send_alert("Data drift detected - model retraining may be needed")

    return drift_detected
```

## A/B Testing Models

### Canary Deployment

```yaml
# databricks.yml - Canary deployment with traffic split
resources:
  model_serving_endpoints:
    customer_ltv_serving:
      config:
        served_entities:
          - entity_name: "production.ml_models.customer_ltv"
            entity_version: "1"  # Current production model
            workload_size: "Small"

          - entity_name: "production.ml_models.customer_ltv"
            entity_version: "2"  # New candidate model
            workload_size: "Small"

        traffic_config:
          routes:
            - served_model_name: "customer_ltv-1"
              traffic_percentage: 90  # 90% to current model
            - served_model_name: "customer_ltv-2"
              traffic_percentage: 10  # 10% to new model (canary)
```

### Compare Model Versions

```python
def compare_model_versions():
    """Compare performance of two model versions in production."""
    # Get predictions from both models
    results_df = spark.sql("""
        SELECT
            customer_id,
            prediction_date,
            model_version,
            predicted_ltv,
            actual_ltv,
            ABS(predicted_ltv - actual_ltv) as error
        FROM production.gold.model_predictions
        WHERE prediction_date >= current_date() - INTERVAL 7 DAYS
    """)

    # Compare metrics
    for version in [1, 2]:
        version_df = results_df.filter(f"model_version = {version}")
        rmse = version_df.selectExpr(f"SQRT(AVG(POW(error, 2))) as rmse").first().rmse

        mlflow.log_metric(f"v{version}_rmse", rmse)
        print(f"Model v{version} RMSE: {rmse:.2f}")

    # Decide on promotion
    v1_rmse = results_df.filter("model_version = 1").selectExpr("SQRT(AVG(POW(error, 2))) as rmse").first().rmse
    v2_rmse = results_df.filter("model_version = 2").selectExpr("SQRT(AVG(POW(error, 2))) as rmse").first().rmse

    if v2_rmse < v1_rmse * 0.95:  # At least 5% improvement
        print(f"✅ Model v2 outperforms v1 - promote to 100% traffic")
    else:
        print(f"⚠️ Model v2 not significantly better - rollback")
```

## Automated Retraining

```python
# notebooks/automated_retraining.py
def should_retrain_model() -> bool:
    """Determine if model should be retrained based on performance metrics."""
    # Check recent performance
    recent_metrics = monitor_model_performance()

    # Check data drift
    drift_detected = detect_feature_drift()

    # Retrain if performance degraded or drift detected
    return recent_metrics["rmse"] > 95 or drift_detected

def retrain_model():
    """Retrain model with latest data."""
    # Load latest training data
    training_df = spark.table("production.silver.customer_features")

    # Train new model
    with mlflow.start_run(run_name="automated_retrain"):
        model, metrics = train_ltv_model(training_df)

        # Register new version
        mlflow.sklearn.log_model(
            model,
            "model",
            registered_model_name="staging.ml_models.customer_ltv"
        )

    print(f"✅ Retrained model with RMSE: {metrics['rmse']:.2f}")

# Schedule automated retraining
if should_retrain_model():
    retrain_model()
```

## Next Steps

- [Monitoring Data Quality](./monitoring-data-quality.md) - Monitor model predictions and data quality
- [CI/CD Pipeline for Databricks](./cicd-pipeline-databricks.md) - Automate model deployment
- [Testing Data Pipelines](./testing-data-pipelines.md) - Test model training pipelines
