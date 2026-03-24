# release execute-layers

<!-- book:cmd release execute-layers -->

## Layers JSON Format

```json
[[{"module":"docs","version":"2025.0116.1430","tag":"docs/2025.0116.1430","type":"calver"}],
 [{"module":"clie","version":"1.0.0","tag":"clie/1.0.0","type":"semver"}]]
```

Each inner array is a layer. Layers are processed in order. Modules within a layer are dispatched concurrently.

## Output

- Progress messages for each module dispatch and completion
- Exit code 0 if all releases succeed
- Exit code 1 if any release fails

## See Also

- [release pending](./pending.md)
- [release](../release/index.md)
