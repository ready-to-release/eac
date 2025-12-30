# Feature Hiding

## The Problem

When practicing Continuous Integration with small batches, you'll often integrate partial features into trunk.

These incomplete features must not affect production behavior.

**You cannot rely on branches to hide incomplete work** - the work is in trunk, potentially deployed to production.

---

## Hiding Strategies

### Level 1: Code-Level Hiding

Hide new code by simply not activating it:

```go
// New implementation exists but not injected
type NewPaymentProcessor struct {}

// Still using old implementation
func main() {
    processor := OldPaymentProcessor{} // Not injecting new one yet
}
```

**Benefits:** Simple, no infrastructure needed

**Limitation:** Requires deployment to activate

### Level 2: Configuration-Based Activation

Use configuration to control feature activation:

```yaml
# config.yaml
features:
  newCheckoutFlow: false  # Deploy OFF, flip to true later
```

**Benefits:** Activation decoupled from deployment

**Limitation:** Requires redeployment to change config

### Level 3: Feature Flags

Use runtime feature flags for full control:

```go
if featureFlags.IsEnabled("new-checkout-flow", user) {
    return NewCheckoutFlow(user)
} else {
    return OldCheckoutFlow(user)
}
```

**Benefits:**

- Deployment fully decoupled from release
- Enable for percentage of users (gradual rollout)
- Instant rollback without redeployment
- A/B testing capabilities

This is **Stage 12 (Release Toggling)** of the CD Model.

---

## When to Use Each Strategy

| Strategy      | Use When                                                  | Don't Use When                                  |
| ------------- | --------------------------------------------------------- | ----------------------------------------------- |
| Code-Level    | Simple internal changes, refactoring                      | Larger feature sets                             |
| Configuration | Features ready for immediate activation                   | Need gradual rollout or A/B testing             |
| Feature Flags | Customer-facing features, risky changes, gradual rollouts | Simple internal changes (overhead not worth it) |

---

## Branch by Abstraction

For larger or architectural changes:

1. Create abstraction layer
2. Implement new code behind abstraction
3. Gradually migrate calls to new implementation
4. All changes integrated to trunk incrementally
5. Remove old implementation when migration complete

This avoids long-lived branches for architectural changes.

**Example:**

```go
// Step 1: Create abstraction
type PaymentProcessor interface {
    Process(order Order) error
}

// Step 2: Old implementation satisfies interface
type LegacyProcessor struct{}
func (p *LegacyProcessor) Process(order Order) error { /* old logic */ }

// Step 3: New implementation (integrated to trunk, but not used)
type ModernProcessor struct{}
func (p *ModernProcessor) Process(order Order) error { /* new logic */ }

// Step 4: Gradually switch (via config or feature flag)
func NewProcessor(useModern bool) PaymentProcessor {
    if useModern {
        return &ModernProcessor{}
    }
    return &LegacyProcessor{}
}

// Step 5: After migration complete, remove LegacyProcessor
```

---

## Next Steps

- [Trunk-Based Development](trunk-based-development.md) - Core principles
- [CD Model Stages](../cd-model/stages.md#stage-12-release-toggling) - Stage 12 details
- [Release Toggling](../release-toggling/index.md) - Feature flag patterns
