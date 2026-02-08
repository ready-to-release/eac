# Step Definitions

> **Writing and organizing Godog step definitions**

Step definitions connect Gherkin scenarios to Go code, enabling executable specifications.

---

## What Are Step Definitions?

Step definitions are Go functions that implement the behavior described in Gherkin steps:

```gherkin
Given I am in an empty folder     # → iAmInAnEmptyFolder()
When I run "cli init"             # → iRun("cli init")
Then a file should exist          # → aFileShouldExist()
```

---

## Key Concepts

| Concept            | Description                  |
| ------------------ | ---------------------------- |
| ScenarioContext    | Registers steps and hooks    |
| Context Management | Maintains state across steps |
| Regex Patterns     | Flexible step matching       |
| Given/When/Then    | Setup, action, assertion     |

---

## Step Categories

| Type  | Purpose             | Example                     |
| ----- | ------------------- | --------------------------- |
| Given | Setup preconditions | `iAmInAnEmptyFolder()`      |
| When  | Execute actions     | `iRun(command)`             |
| Then  | Assert outcomes     | `theCommandShouldSucceed()` |

---

## Best Practices

- Keep steps reusable across scenarios
- Use regex for flexibility
- Return descriptive errors
- Use context structure for state

---

## Reference Documentation

For complete code templates, patterns, and examples, see:

**[Step Definitions Reference](../../../../reference/eac/testing/go/step-definitions.md)** - Complete guide including:

- Godog test suite setup
- Context management patterns
- Regex step definition patterns
- Common step patterns (Given/When/Then)
- Table handling
- Error handling
- Complete file templates

---

## Related Documentation

- [File Organization](./file-organization.md) - Where to place step definition files
- [Test Levels](./test-levels.md) - Understanding test isolation
- [Best Practices](./best-practices.md) - Testing patterns and conventions
