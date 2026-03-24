# validate contracts

<!-- book:cmd validate contracts -->

## What It Checks

Validates each contract file with schema validation enabled:

| File               | Validation                                      |
| ------------------ | ----------------------------------------------- |
| `repository.yml`   | Module definitions, component types, paths      |
| `environments.yml` | Environment definitions and semantic validation |
| `testing-tags.yml` | Tag definitions and skip reasons                |
| `test-suites.yml`  | Test suite definitions                          |

Reports the number of validated items per file (modules, environments, tags, suites).

## Common Errors

- **FAILED** -- A YAML file does not conform to its JSON schema. Check the error message for the specific field or value.
- **Semantic validation warning** -- The file is schema-valid but has logical issues (e.g., duplicate monikers).

## See Also

- [validate](./validate.md)
