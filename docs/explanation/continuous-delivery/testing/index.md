# Testing

Testing is integrated throughout every stage of the Continuous Delivery Model, not a separate phase after development.

---

## Core Concepts

| Concept                                       | Description                                             |
| --------------------------------------------- | ------------------------------------------------------- |
| [Test Levels](./test-levels.md)               | L0-L4 taxonomy based on execution environment and scope |
| [Verification Types](./verification-types.md) | IV/OV/PV classification for acceptance testing          |

---

## Quick Reference

### Test Levels (L0-L4)

| Level | Name               | Environment  | Scope                        | Dependencies                 |
| ----- | ------------------ | ------------ | ---------------------------- | ---------------------------- |
| L0-L1 | Unit Tests         | DevBox/Agent | Source/Binary                | All test doubles/contracts   |
| L2    | Emulated System    | DevBox/Agent | Deployable artifacts         | All test doubles/contracts   |
| L3    | In-Situ Vertical   | PLTE         | Deployed system (vertical)   | All test doubles/contracts   |
| HE2E  | In-Situ Vertical   | PLTE         | Deployed system (vertical)   | Staged external services     |
| L4    | Production Testing | Production   | Deployed system (horizontal) | Production external services |

### Verification Types (Stage 5)

| Type              | Question           | Focus                         |
| ----------------- | ------------------ | ----------------------------- |
| IV (Installation) | Can we deploy it?  | Infrastructure, configuration |
| OV (Operational)  | Does it work?      | Functional requirements       |
| PV (Performance)  | Is it fast enough? | Response times, throughput    |

### Shift-Left and Shift-Right

- **Shift LEFT (L0-L3):** Fast, deterministic tests with test doubles
- **Shift RIGHT (L4):** Production validation with real services
- **Avoid:** Horizontal pre-production environments (fragile, non-deterministic)

---

## Stage Mapping

| Stage             | Test Levels | Time Budget |
| ----------------- | ----------- | ----------- |
| 2-4 (Development) | L0-L2       | 5-30 min    |
| 5-6 (Acceptance)  | L3          | 1-8 hours   |
| 11-12 (Live)      | L4          | Continuous  |

---

## Next Steps

- [Test Levels](./test-levels.md) - L0-L4 taxonomy and shift-left strategy
- [Verification Types](./verification-types.md) - IV/OV/PV for acceptance testing
- [Testing Taxonomy](../../specifications/taxonomy/index.md) - Tag contracts and test suites
