# Build Tags

Build tags control which tests are compiled and run. Every test file should declare a build tag on its first line to assign the test to a level.

---

## Test Levels

| Tag  | Level       | Environment  | Description                                      |
| ---- | ----------- | ------------ | ------------------------------------------------ |
| `L0` | Unit        | DevBox       | Pure unit tests with no I/O or external deps     |
| `L1` | Unit+       | DevBox       | Unit tests with minimal I/O (filesystem, config) |
| `L2` | Integration | Build Agent  | Docker-based integration tests                   |
| `L3` | Acceptance  | PLTE         | Production-like system tests                     |
| `L4` | Production  | Production   | Production verification smoke tests              |

**L0** and **L1** together form the **unit** suite. They run fast and require no infrastructure beyond a development machine.

**L2** tests need Docker and typically run on a build agent in CI.

**L3** tests run against a Production-Like Test Environment (PLTE) and may take hours.

**L4** tests run in production after deployment to verify the release succeeded.

---

## Verification Tags

These tags classify tests by what they verify. They are used alongside level tags.

| Tag  | Name                          | Purpose                                   |
| ---- | ----------------------------- | ----------------------------------------- |
| `ov` | Operational Verification      | Tests that verify features work correctly  |
| `iv` | Installation Verification     | Tests that verify deployment succeeded     |
| `pv` | Performance Verification      | Tests that verify SLA/SLI/SLO compliance  |
| `piv`| Production Install Verification | Required for production-verification suite |

The `ov` tag is the most common verification tag. It marks tests that require operational infrastructure (running services, configuration files, or tools present on the system). In the unit suite, `ov` tests are typically slower and may be excluded from quick local runs.

---

## How to Tag a Test File

Place the `//go:build` directive on the very first line of the file, before the `package` declaration:

```go
//go:build L0

package mypackage

import "testing"

func TestSomething(t *testing.T) {
    // pure unit test
}
```

Combine tags with `&&` when a test belongs to a level and has a verification constraint:

```go
//go:build L1 && ov

package mypackage
```

This test is included in the unit suite (L1) but only when `ov` is also requested.

---

## How Build Tags Map to Test Suites

Test suites are defined in `contracts/core/0.1.0/schemas/defaults/test-suites.yml`. Each suite selects tests by their tags:

| Suite                      | Includes         | Excludes              |
| -------------------------- | ---------------- | --------------------- |
| `unit`                     | `L0` or `L1`    | `L2`, `L3`, `L4`     |
| `integration`              | `L2`             | `L0`, `L1`, `L3`, `L4` |
| `acceptance`               | `L3`             | `L0`, `L1`, `L2`, `L4` |
| `production-verification`  | `L4` and `piv`   | --                    |
| `manual`                   | `Manual`         | `L0`-`L4`            |

Tests tagged `L1 && ov` are compiled only when **both** `L1` and `ov` are passed as build tags. The unit suite passes `L0` and `L1` but does not pass `ov` by default, so these tests run only when `ov` is explicitly included.

---

## Running Tests by Level

Run a specific suite:

```bash
eac test <module> --suite unit
eac test <module> --suite integration
eac test <module> --suite acceptance
```

Run all modules:

```bash
eac test --suite unit
```

---

## Choosing the Right Tag

- **L0** -- No filesystem, no network, no external tools. Fast and deterministic.
- **L1** -- Reads files, uses temp directories, or loads configuration. Still runs locally.
- **L1 && ov** -- Needs local tools or services running (e.g., Docker daemon, git repos). Slower than plain L1.
- **L2** -- Needs Docker containers or external services spun up by CI.
- **L3** -- Needs a deployed environment.
- **L4** -- Runs in production.

When in doubt, prefer **L0**. Move to higher levels only when the test genuinely requires the additional infrastructure.

---

## Other Build Tags

| Tag         | Purpose                                              |
| ----------- | ---------------------------------------------------- |
| `benchmark` | Performance benchmarks (not included in any suite)   |
| `!lite`     | Excluded from lite builds of the `eac` binary        |
| `!windows`  | Excluded on Windows                                  |
| `windows`   | Included only on Windows                             |

---

## Related Documentation

- [Test Suites](./test-suites.md) -- Suite definitions and run commands
- [Test Levels (Explanation)](../../../explanation/specifications/taxonomy/test-levels.md) -- Conceptual overview of L0-L4
- [Test Commands](../commands/test/index.md) -- CLI command reference
