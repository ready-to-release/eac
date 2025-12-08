# Deployment Strategies

## Introduction

A deployment strategy determines **how** new code reaches production and **what happens** if problems are detected. The right strategy balances:

- **Risk**: How much can go wrong?
- **Speed**: How fast can we deploy?
- **Cost**: What infrastructure is required?
- **Complexity**: How hard is it to implement and operate?

No single strategy fits all situations. Mature organizations use different strategies for different services based on risk profile, infrastructure, and business requirements.

---

## Hot Deploy (In-Place Update)

### What It Is

Hot deploy replaces running instances with new versions directly, in place, with a brief service interruption.

**Process**:

1. Stop running application instance
2. Replace application files with new version
3. Start application with new code
4. Verify health checks
5. Repeat for additional instances (if multiple)

**Characteristics**:

- Fastest deployment (seconds to minutes)
- Brief downtime during replacement (< 30 seconds per instance)
- Simple to implement
- Low infrastructure cost (no extra resources)
- Simple rollback (redeploy previous version)

### When to Use

**Best for**:

- **Single-instance applications** (no load balancing)
- **Acceptable brief downtime** (internal tools, batch processing)
- **Fast-starting applications** (startup < 10 seconds)
- **Small deployable modules** (< 100 MB artifacts)
- **Immature infrastructure** (limited resources)

**Not suitable for**:

- User-facing applications requiring zero downtime
- Long startup times (> 1 minute)
- Stateful applications (active sessions lost)
- High-availability requirements

### Example Implementation

```bash
#!/bin/bash
# hot-deploy.sh

echo "Stopping application..."
systemctl stop myapp

echo "Backing up current version..."
cp /opt/myapp/app /opt/myapp/app.backup

echo "Deploying new version..."
cp /artifacts/app-v1.2.0 /opt/myapp/app
chmod +x /opt/myapp/app

echo "Starting application..."
systemctl start myapp

echo "Verifying health..."
sleep 5
curl -f http://localhost:8080/health || {
    echo "Health check failed, rolling back..."
    systemctl stop myapp
    cp /opt/myapp/app.backup /opt/myapp/app
    systemctl start myapp
    exit 1
}

echo "Deployment successful"
```

### Tradeoffs

**Advantages**:

- ✅ Simple to understand and implement
- ✅ Fast (1-2 minutes total)
- ✅ Low infrastructure cost
- ✅ Works with simple infrastructure

**Disadvantages**:

- ❌ Brief downtime (< 30 seconds)
- ❌ All-or-nothing (no gradual rollout)
- ❌ Rollback requires redeployment (1-2 minutes)
- ❌ Stateful sessions lost

---

## Rolling Deployment

### What It Is

Rolling deployment updates instances gradually - one or a few at a time - while others continue serving traffic.

**Process**:

1. Remove instance #1 from load balancer
2. Update instance #1 to new version
3. Verify health checks on instance #1
4. Add instance #1 back to load balancer
5. Repeat for instance #2, #3, etc.

**Characteristics**:

- Zero downtime
- Gradual rollout (percentage increases over time)
- Mixed versions running temporarily
- Built-in health checks stop rollout if issues detected
- Medium complexity

### When to Use

**Best for**:

- **Multiple instances** (≥ 2 instances behind load balancer)
- **Zero downtime required** (user-facing applications)
- **Backward-compatible changes** (new version can coexist with old)
- **Automated health checks** (can detect deployment issues)

**Not suitable for**:

- Single instance applications
- Breaking changes (protocol changes, incompatible APIs)
- Database migrations requiring all instances on same version

### Example Implementation

```bash
#!/bin/bash
# rolling-deploy.sh

INSTANCES=("instance-1" "instance-2" "instance-3")

for instance in "${INSTANCES[@]}"; do
    echo "Deploying to $instance..."

    # Remove from load balancer
    aws elb deregister-instances-from-load-balancer \
        --load-balancer-name my-lb \
        --instances $instance

    # Wait for drain
    sleep 30

    # Deploy new version
    ssh $instance "sudo systemctl stop myapp && \
                   sudo cp /artifacts/app-v1.2.0 /opt/myapp/app && \
                   sudo systemctl start myapp"

    # Wait for startup
    sleep 10

    # Health check
    if ! ssh $instance "curl -f http://localhost:8080/health"; then
        echo "Health check failed on $instance, aborting rollout"
        # Re-register healthy instances
        aws elb register-instances-with-load-balancer \
            --load-balancer-name my-lb \
            --instances $instance
        exit 1
    fi

    # Add back to load balancer
    aws elb register-instances-with-load-balancer \
        --load-balancer-name my-lb \
        --instances $instance

    echo "$instance deployed successfully"
done

echo "Rolling deployment complete"
```

### Tradeoffs

**Advantages**:

- ✅ Zero downtime
- ✅ Gradual rollout (issues affect fewer instances)
- ✅ Automatic halt on health check failures
- ✅ Lower infrastructure cost (no extra instances)

**Disadvantages**:

- ❌ Requires multiple instances
- ❌ Mixed versions during rollout (must be compatible)
- ❌ Slower rollback (must roll forward or roll back each instance)
- ❌ Session affinity issues (if stateful)

---

## Blue-Green Deployment

### What It Is

Blue-green deployment maintains two identical production environments. One serves traffic (Blue), while the other is idle (Green). Deploy to Green, test, then switch traffic from Blue to Green instantly.

**Environments**:

- **Blue**: Current production version, serving live traffic
- **Green**: New version deployed, idle (or receiving test traffic)

**Process**:

1. Blue environment serves 100% production traffic
2. Deploy new version to Green environment
3. Run smoke tests against Green (0% production traffic)
4. Switch load balancer/DNS from Blue to Green (instant cutover)
5. Monitor Green with 100% production traffic
6. Blue becomes idle (ready for next deployment or instant rollback)

**Characteristics**:

- Zero downtime
- Instant traffic switch (seconds)
- Instant rollback (switch back to Blue)
- High infrastructure cost (2x resources)
- Requires infrastructure automation

### When to Use

**Best for**:

- **Instant rollback critical** (high-risk deployments)
- **Large artifacts or slow startup** (can't afford multiple rolling updates)
- **Database migrations** (need to test migration fully before traffic switch)
- **High-stakes releases** (major version updates)
- **Infrastructure as Code** (can automate environment creation)

**Not suitable for**:

- Cost-constrained environments (requires 2x infrastructure)
- Stateful applications without shared state store
- Very large infrastructure (cost prohibitive)

### Example Implementation

```bash
#!/bin/bash
# blue-green-deploy.sh

BLUE_ENV="myapp-blue"
GREEN_ENV="myapp-green"
LB_TARGET_GROUP_ARN="arn:aws:elasticloadbalancing:..."

# Current production is Blue
echo "Current production: $BLUE_ENV"

# Deploy to Green
echo "Deploying to $GREEN_ENV..."
terraform apply -var="environment=green" -var="version=v1.2.0"

# Wait for Green to be healthy
echo "Waiting for $GREEN_ENV health checks..."
aws elbv2 wait target-in-service \
    --target-group-arn $LB_TARGET_GROUP_ARN

# Run smoke tests against Green
echo "Running smoke tests on $GREEN_ENV..."
./smoke-tests.sh https://green.myapp.internal

if [ $? -ne 0 ]; then
    echo "Smoke tests failed, aborting deployment"
    exit 1
fi

# Switch traffic to Green
echo "Switching traffic to $GREEN_ENV..."
aws elbv2 modify-listener \
    --listener-arn $LISTENER_ARN \
    --default-actions Type=forward,TargetGroupArn=$GREEN_TARGET_GROUP_ARN

echo "Traffic switched to $GREEN_ENV"
echo "$BLUE_ENV is now idle (ready for rollback or next deployment)"
```

### Rollback

**Instant rollback**:

```bash
# Switch traffic back to Blue
aws elbv2 modify-listener \
    --listener-arn $LISTENER_ARN \
    --default-actions Type=forward,TargetGroupArn=$BLUE_TARGET_GROUP_ARN
```

Rollback completes in seconds - just switch traffic back.

### Tradeoffs

**Advantages**:

- ✅ Instant rollback (seconds)
- ✅ Zero downtime
- ✅ Test in production-like environment before traffic switch
- ✅ Clean separation (no mixed versions)

**Disadvantages**:

- ❌ High cost (2x infrastructure running during deployment)
- ❌ Database migrations complex (shared database between Blue/Green)
- ❌ Requires infrastructure automation
- ❌ Idle environment waste (Blue sits idle after switch)

---

## Canary Deployment

### What It Is

Canary deployment routes a small percentage of traffic to the new version, monitors metrics, and gradually increases traffic if healthy.

**Process**:

1. Deploy new version alongside current version
2. Route 1-5% traffic to new version (canary)
3. Monitor canary metrics (errors, latency, business KPIs)
4. If healthy: Gradually increase traffic (10% → 25% → 50% → 100%)
5. If unhealthy: Route 0% traffic to canary (instant rollback)

**Characteristics**:

- Zero downtime
- Gradual, metrics-driven rollout
- Early warning with minimal blast radius (1-5% affected)
- Requires robust monitoring and metrics
- Requires traffic routing capability
- Medium infrastructure cost

### When to Use

**Best for**:

- **Need production validation** (can't fully test in staging)
- **Metrics-driven decisions** (error rates, latency, business KPIs)
- **Risk-averse organizations** (gradual increases confidence)
- **A/B testing infrastructure** (already have traffic routing)
- **Large user base** (1% still significant sample size)

**Not suitable for**:

- Small user base (1% is too few users for meaningful metrics)
- Infrastructure without traffic routing capability
- Applications without good observability
- Breaking changes (can't coexist with old version)

### Example Implementation

```yaml
# canary-deployment.yaml (using Kubernetes)

apiVersion: v1
kind: Service
metadata:
  name: myapp
spec:
  selector:
    app: myapp
  ports:
    - port: 80

---
# Stable version (95% traffic)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-stable
spec:
  replicas: 19  # 95% of traffic
  selector:
    matchLabels:
      app: myapp
      version: v1.1.0
  template:
    metadata:
      labels:
        app: myapp
        version: v1.1.0
    spec:
      containers:
      - name: myapp
        image: myapp:v1.1.0

---
# Canary version (5% traffic)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-canary
spec:
  replicas: 1  # 5% of traffic
  selector:
    matchLabels:
      app: myapp
      version: v1.2.0
  template:
    metadata:
      labels:
        app: myapp
        version: v1.2.0
    spec:
      containers:
      - name: myapp
        image: myapp:v1.2.0
```

**Automated progressive rollout**:

```python
# canary-controller.py

def canary_rollout():
    percentages = [5, 10, 25, 50, 100]

    for pct in percentages:
        print(f"Routing {pct}% traffic to canary...")
        update_traffic_split(canary_percent=pct)

        # Monitor for 15 minutes
        time.sleep(900)

        metrics = get_canary_metrics()

        if metrics['error_rate'] > THRESHOLD_ERROR_RATE:
            print("Error rate too high, rolling back")
            update_traffic_split(canary_percent=0)
            return FAILED

        if metrics['p95_latency'] > THRESHOLD_LATENCY:
            print("Latency too high, rolling back")
            update_traffic_split(canary_percent=0)
            return FAILED

        print(f"{pct}% canary healthy, proceeding")

    print("Canary fully rolled out")
    return SUCCESS
```

### Metrics to Monitor

**Technical metrics**:

- Error rate (4xx, 5xx responses)
- P50, P95, P99 latency
- Request throughput
- Resource utilization (CPU, memory)

**Business metrics**:

- Conversion rate
- User engagement
- Revenue impact
- Customer complaints

### Tradeoffs

**Advantages**:

- ✅ Minimal blast radius (1-5% affected initially)
- ✅ Production validation before full rollout
- ✅ Instant rollback (route 0% to canary)
- ✅ Metrics-driven confidence building

**Disadvantages**:

- ❌ Requires traffic routing infrastructure
- ❌ Requires robust observability
- ❌ Complex to implement
- ❌ Slower rollout (hours, not minutes)

---

## Feature Flags with Deployment

### What It Is

Feature flags decouple **deployment** (code reaches production) from **release** (features enabled for users). Deploy code with features disabled, enable via runtime flags.

**Process**:

1. Deploy new code to production (features OFF via flags)
2. Gradually enable features for user segments:
   - Internal users (0.1%)
   - Early adopters (1%)
   - Standard users (10% → 50% → 100%)
3. Monitor metrics per enabled segment
4. Instant disable if issues detected (toggle flag OFF)

**Characteristics**:

- Zero downtime deployments
- Independent feature rollout per feature
- Instant feature disable (no redeployment)
- Gradual rollout per feature
- Enables A/B testing
- Requires feature flag infrastructure

### When to Use

**Best for**:

- **Continuous Deployment (CDe) pattern** (required for automated production deployment)
- **A/B testing** (compare feature variants)
- **Gradual rollouts** (enable features progressively)
- **Kill switches** (instant disable capability)
- **Beta programs** (enable for specific users)

**Not suitable for**:

- Infrastructure changes (can't flag infrastructure)
- Database schema changes (flags don't help)
- Very simple applications (overhead not worth it)

### Example Implementation

**Feature flag service**:

```go
// featureflags/featureflags.go

package featureflags

type Service struct {
    client *flagd.Client
}

func (s *Service) IsEnabled(ctx context.Context, feature string, userID string) bool {
    // Check if feature is enabled for this user
    return s.client.BoolValue(ctx, feature, false, flagd.EvaluationContext{
        "userID": userID,
    })
}
```

**Application code**:

```go
// handler.go

func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r)

    if h.flags.IsEnabled(r.Context(), "new-checkout-flow", userID) {
        // New feature
        h.handleNewCheckout(w, r)
    } else {
        // Old feature
        h.handleOldCheckout(w, r)
    }
}
```

**Feature flag configuration**:

```yaml
# flags.yaml

new-checkout-flow:
  state: ENABLED
  variants:
    on: true
    off: false
  defaultVariant: off
  targeting:
    - if:
        - in:
          - userID
          - ["user-1", "user-2", "user-3"]  # Internal users
      variant: on
    - if:
        - startsWith:
          - email
          - "beta-"  # Beta users
      variant: on
    - if:
        - percentage:
          - flagKey: new-checkout-flow
          - 10  # 10% rollout
      variant: on
```

### Progressive Rollout with Flags

**Stage 12 (Release Toggling)**:

1. **Deploy** code to production (Stage 10) - flags OFF
2. **Enable** for internal users (0.1%)
3. Monitor metrics, if healthy: enable for 1%
4. Monitor metrics, if healthy: enable for 10%
5. Monitor metrics, if healthy: enable for 50%
6. Monitor metrics, if healthy: enable for 100%

Each step is independent of deployment - just toggle configuration.

### Tradeoffs

**Advantages**:

- ✅ Decouple deployment from release
- ✅ Instant feature disable (no redeployment)
- ✅ Gradual rollout per feature
- ✅ A/B testing capability
- ✅ Beta programs (enable for specific users)

**Disadvantages**:

- ❌ Code complexity (if/else branches)
- ❌ Feature flag infrastructure required
- ❌ Technical debt (must remove old code paths)
- ❌ Testing complexity (all flag combinations)

---

## Strategy Selection Framework

### Decision Tree

**Question 1: Can you tolerate any downtime?**

- **No** → Rolling, Blue-Green, Canary, or Feature Flags
- **Yes (< 30s)** → Hot Deploy

**Question 2: Do you need instant rollback?**

- **Yes** → Blue-Green or Canary
- **No** → Rolling or Hot Deploy

**Question 3: Do you have robust monitoring/metrics?**

- **Yes** → Canary (metrics-driven)
- **No** → Blue-Green (instant switch)

**Question 4: Do you need gradual feature rollout?**

- **Yes** → Feature Flags (with any deployment strategy)
- **No** → Any strategy

**Question 5: What's your infrastructure budget?**

- **Limited** → Hot Deploy or Rolling
- **Moderate** → Canary or Feature Flags
- **High** → Blue-Green

### Strategy Comparison

| Strategy       | Downtime | Rollback Speed | Complexity | Cost    | Use Case                       |
| -------------- | -------- | -------------- | ---------- | ------- | ------------------------------ |
| Hot Deploy     | Brief    | Fast (1-2 min) | Low        | Low     | Internal tools, batch jobs     |
| Rolling        | None     | Medium (5 min) | Medium     | Low     | Standard web services          |
| Blue-Green     | None     | Instant        | Medium     | High    | High-risk, large deployments   |
| Canary         | None     | Instant        | High       | Medium  | Production validation needed   |
| Feature Flags  | None     | Instant        | High       | Medium  | Gradual rollouts, A/B testing  |

---

## RA vs CDe Pattern Recommendations

### Release Approval (RA) Pattern

**Recommended strategies**:

- **Blue-Green**: Manual approval at Stage 9, then instant switch to production
- **Rolling**: Manual approval at Stage 9, then gradual rollout

**Why these work**:

- Manual approval provides explicit go/no-go decision
- Clear cut-over point (before or after traffic switch)
- Aligns with manual approval workflow

**Typical flow**:

1. Deploy to Green (or first instances)
2. Test in production (or on subset)
3. **Manual approval (Stage 9)**
4. Switch traffic (Blue-Green) or continue rolling

### Continuous Deployment (CDe) Pattern

**Recommended strategies**:

- **Canary**: Automated with gradual validation
- **Feature Flags**: Deploy continuously, release independently

**Why these work**:

- Automated progression (no manual approval)
- Metrics-driven decisions
- Gradual increases confidence
- Instant rollback on threshold breaches

**Typical flow**:

1. Deploy to production (automated)
2. Route 1% traffic (Canary) or deploy with flags OFF
3. Monitor metrics (automated)
4. Gradually increase (automated if metrics healthy)
5. Full rollout or instant rollback

**CDe requires**:

- Robust automated testing (Stages 5-6)
- Comprehensive monitoring and metrics
- Automated rollback triggers
- Feature flags (for independent release)

---

## Implementation Considerations

### Infrastructure Requirements

**Hot Deploy**:

- Minimal (single server, basic automation)

**Rolling**:

- Load balancer
- Multiple instances (≥ 2)
- Health check endpoints
- Orchestration (Kubernetes, AWS ECS, etc.)

**Blue-Green**:

- Infrastructure as Code (Terraform, CloudFormation)
- Load balancer with traffic switching
- 2x production capacity
- Automated environment provisioning

**Canary**:

- Traffic routing capability (service mesh, load balancer)
- Observability platform (Prometheus, Datadog)
- Automated metrics collection
- Rollout automation (Flagger, Argo Rollouts)

**Feature Flags**:

- Feature flag service (LaunchDarkly, Flagsmith, Unleash)
- Application integration
- Configuration management
- Flag lifecycle management

### Monitoring Requirements

All strategies require:

- Health check endpoints (`/health`, `/ready`)
- Application metrics (errors, latency, throughput)
- Infrastructure metrics (CPU, memory, disk)
- Logging (structured, centralized)

**Canary and Feature Flags additionally require**:

- Per-version metrics (compare canary to stable)
- Business metrics (conversion, revenue)
- Automated alerting on threshold breaches
- Real-time dashboards

### Rollback Procedures

**Every deployment strategy must have documented rollback**:

- Rollback triggers (when to rollback)
- Rollback steps (how to execute)
- Rollback time (how long it takes)
- Database rollback (if applicable)
- Verification (how to confirm rollback success)

---

## Anti-Patterns

### Anti-Pattern 1: No Rollback Plan

**Problem**: Deploy without defined rollback procedure

**Impact**: When issues arise, scramble to figure out how to recover

**Solution**: Document rollback procedure before deploying

### Anti-Pattern 2: All-at-Once Deployments (Big Bang)

**Problem**: Deploy to all instances simultaneously

**Impact**: All users affected if deployment fails

**Solution**: Use gradual rollout (Rolling, Canary, Rings)

### Anti-Pattern 3: Ignoring Canary Metrics

**Problem**: Roll out canary despite elevated error rates

**Impact**: Problems affect all users when fully rolled out

**Solution**: Automated rollback on threshold breaches

### Anti-Pattern 4: Feature Flag Debt

**Problem**: Accumulating old feature flags, never removing

**Impact**: Code complexity, testing burden, technical debt

**Solution**: Remove flags after full rollout (set lifecycle policy)

### Anti-Pattern 5: No Health Checks

**Problem**: Deploying without health check validation

**Impact**: Broken deployments serve traffic, users affected

**Solution**: Implement health checks, verify before adding to load balancer

---

## Best Practices Summary

1. **Match strategy to risk**: Higher risk → More gradual rollout
2. **Always have rollback**: Document and test rollback procedures
3. **Automate deployments**: Reduce human error, increase consistency
4. **Monitor actively**: Watch metrics during and after deployment
5. **Health checks required**: Never serve traffic from unhealthy instances
6. **Start simple**: Hot Deploy → Rolling → Blue-Green → Canary → Feature Flags
7. **CDe requires maturity**: Comprehensive testing, monitoring, feature flags
8. **RA for high-risk**: Manual approval for critical systems
9. **Gradual rollouts**: Minimize blast radius, build confidence
10. **Document everything**: Runbooks, rollback procedures, decision criteria

---

## Next Steps

- [Deployment Rings](./deployment-rings.md) - Progressive user group rollout pattern
- [CD Model Stages 7-12](../cd-model/cd-model-stages-7-12.md) - See deployment in CD Model context
- [Implementation Patterns](../cd-model/implementation-patterns.md) - RA vs CDe pattern selection
- [Environments Architecture](../architecture/environments.md) - Deploy Agents and production environments

## Quick Reference

### Strategy Comparison

| Strategy   | Rollback Speed | Downtime      | Complexity | Resource Cost |
| ---------- | -------------- | ------------- | ---------- | ------------- |
| Hot Deploy | Fast (1-2 min) | Brief (< 30s) | Low        | Low           |
| Rolling    | Fast (< 1 min) | None          | Medium     | Low           |
| Blue-Green | Instant        | None          | Medium     | High (2x)     |
| Canary     | Instant        | None          | High       | Medium        |
| Rings      | Gradual        | None          | High       | Medium        |

---

_[Tutorials](../../../tutorials/) | [How-to Guides](../../../how-to-guides/) | **Explanation** | [Reference](../../../reference/)_

**You are here:** Explanation — understanding-oriented discussion that clarifies concepts.
