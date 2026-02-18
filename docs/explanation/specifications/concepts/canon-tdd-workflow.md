# Canon TDD Workflow

Reference for Kent Beck's test-driven development process.

## The Five Steps

| Step | Name     | Focus               | Color |
| ---- | -------- | ------------------- | ----- |
| 1    | List     | Behavioral variants | -     |
| 2    | Test     | Write one test      | Red   |
| 3    | Pass     | Make it work        | Green |
| 4    | Refactor | Improve design      | Blue  |
| 5    | Repeat   | Next from list      | -     |

## Step Details

### Step 1: List

**Purpose:** Identify all expected behavioral variants and edge cases.

**Activities:**

- Analyze requirements
- Identify success cases
- Identify failure cases
- Consider edge cases
- Document as test list

**Output:** Written list of tests to implement.

### Step 2: Test (Red)

**Purpose:** Write one failing test from the list.

**Focus:** Interface design - how behavior is invoked from caller perspective.

**Activities:**

- Setup test context
- Invoke the behavior
- Assert expected outcome
- Run test - should fail

**Output:** One failing test.

### Step 3: Pass (Green)

**Purpose:** Make the test pass with minimum code.

**Focus:** Implementation design - internal mechanics.

**Rules:**

- Write only enough code to pass
- No shortcuts (hardcoding values)
- No additional features

**Output:** Passing test.

### Step 4: Refactor (Blue)

**Purpose:** Improve code design without changing behavior.

**Activities:**

- Extract methods
- Rename for clarity
- Remove duplication
- Improve structure

**Constraint:** All tests must still pass.

### Step 5: Repeat

**Purpose:** Continue until test list is empty.

**Note:** New tests discovered during implementation get added to the list.

## Red-Green-Refactor Cycle

```text
Red → Green → Refactor → Red → Green → Refactor → ...
```

| Phase    | Activity           | Tests       |
| -------- | ------------------ | ----------- |
| Red      | Write failing test | 1 failing   |
| Green    | Make test pass     | All passing |
| Refactor | Improve design     | All passing |

## Key Principles

| Principle              | Description                         |
| ---------------------- | ----------------------------------- |
| Red = Interface        | Focus on how behavior is called     |
| Green = Implementation | Focus on making it work             |
| Small steps            | One test at a time                  |
| Test list              | Plan before coding                  |
| Emergent design        | Design improves through refactoring |

## Example Test List

For feature `init-project`:

```text
Test List:
- [ ] Create config in empty directory (success)
- [ ] Create config with custom path (success)
- [ ] Create config when file exists (error)
- [ ] Create config in read-only directory (error)
- [ ] Create config with invalid path (error)
```

## Anti-Patterns

| Anti-Pattern      | Problem               | Fix                        |
| ----------------- | --------------------- | -------------------------- |
| Skipping Red      | Don't know test works | Always see test fail first |
| Gold-plating      | Over-engineering      | Only pass current test     |
| Skipping Refactor | Technical debt        | Refactor after every Green |
| Big steps         | Hard to debug         | One behavior per test      |

## Reference

Based on [Canon TDD by Kent Beck](https://tidyfirst.substack.com/p/canon-tdd).

## Related

- [Three-Layer Approach](./three-layer-approach.md)
- [BDD Fundamentals](./bdd-fundamentals.md)
