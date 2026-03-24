# get documented-commands

<!-- book:cmd get documented-commands -->

## Output Fields

| Field                               | Description                                        |
| ----------------------------------- | -------------------------------------------------- |
| `commands`                          | List of documented commands                        |
| `commands[].command`                | Command name (e.g. `build`, `get modules`)         |
| `commands[].occurrences`            | List of locations where the command appears        |
| `commands[].occurrences[].file`     | Relative path to the markdown file                 |
| `commands[].occurrences[].line`     | Line number                                        |
| `commands[].occurrences[].language` | Code block language (`bash`, `powershell`, `pwsh`) |
| `commands[].occurrences[].snippet`  | The actual command line                            |
| `summary.total_commands`            | Number of unique commands found                    |
| `summary.total_occurrences`         | Total number of command references                 |
| `summary.total_files`               | Number of files containing commands                |

## See Also

- [get valid-commands](./valid-commands.md)
- [show help](../show/help.md)
- [get Commands](../get/index.md)
