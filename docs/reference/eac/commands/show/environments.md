# show environments

<!-- book:cmd show environments -->

## Output Sections

1. **Environment Contracts** - Header with total count, followed by a table with columns: Moniker, Name, Level, Type, System Dependencies.
2. **Summary by Level** - Counts for each environment level:
   - L0 (Very Fast Unit)
   - L1 (Fast Unit)
   - L2 (Local/Docker)
   - L3 (PLTE)
   - L4 (Production)
3. **Summary by Type** - Counts for each environment type (e.g., devbox, ci, docker).

## See Also

- [get environments](../get/environments.md) - JSON output
- [test](../test/test.md)
- [show Commands](../show/index.md)
