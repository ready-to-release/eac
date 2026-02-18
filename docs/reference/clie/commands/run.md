# clie run

Run an extension from the configuration.

## Syntax

```bash
clie run <extension> [args...]
```

## Description

The `run` command executes a configured extension in its Docker container.

All arguments after the extension name are passed directly to the container without
modification — flag parsing is disabled for the `run` command to allow extensions to
receive their own flags.

**Behaviour by argument count:**

- `clie run <extension>` with no further arguments: switches to interactive mode
  (equivalent to `clie interactive <extension>`)
- `clie run <extension> [args...]` with arguments: executes the extension non-interactively
  and exits when the container finishes

**Note:** Most users do not need to use `run` explicitly. When extensions are configured,
CLIE registers each extension name as a direct subcommand alias. This means `clie eac build`
is equivalent to `clie run eac build`. See [Extension Aliases](#extension-aliases) below.

## Flags

| Flag         | Description                                                    |
| ------------ | -------------------------------------------------------------- |
| `-h, --help` | Display help and list available extensions (reads config file) |

Global flags (`--clie-debug`, `--clie-quiet`) are available but must be placed before
the extension name, since all arguments after the extension name are passed to the container.

```bash
clie --clie-debug run eac build   # correct: global flag before run
clie run eac --clie-debug build   # incorrect: --clie-debug passed to container
```

## Examples

### Run a Command in an Extension

```bash
clie run eac build
```

### Pass Arguments and Flags to Extension

```bash
clie run eac test --filter unit
```

### Open Interactive Shell (No Args)

```bash
clie run eac
# equivalent to: clie interactive eac
```

### Enable Debug Logging

```bash
clie --clie-debug run eac build
```

## Extension Aliases

When CLIE starts, it reads `.clie/clie.yml` and registers each configured extension name
as a direct Cobra subcommand. This is why `clie eac build` works — `eac` becomes a registered
command that is an alias for `clie run eac`.

Aliases are created at startup. If the config file is not present or fails to load, the
aliases are not registered and only `clie run <extension>` will work.

**Alias format:**

```bash
clie <extension-name> [args...]      # alias (shorthand)
clie run <extension-name> [args...]  # explicit form
```

Both forms are equivalent and produce identical behaviour.

## See Also

- [interactive command](interactive.md) — Open interactive shell in extension container
- [install command](install.md) — Install extensions
- [Configuration Reference](../configuration.md) — Configure extensions in `.clie/clie.yml`
