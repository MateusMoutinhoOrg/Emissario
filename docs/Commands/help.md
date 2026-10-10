# `help`

Display help for a command

```bash
emissario help [Name…] [--database <database>] [--help]
```

When called without arguments, lists every available command grouped by category. When called with a command name, shows detailed usage, arguments, flags, and examples for that command.

| Arg | Type | Default | Description |
| --- | --- | --- | --- |
| `Name` | string, repeatable |  | The command to describe; omit it to list every command |

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--database` | string | `data` | the folder every database lives under, relative to where the command runs, lower-case letters, digits, - and _ (defaults to data) | [database-dir](database-dir.md) |
| `--help`, `-h` | boolean |  | Print the help of the command this command line is for | [help-flag](help-flag.md) |

| Runs in front of it | When |
| --- | --- |
| [`database-dir`](database-dir.md) | always |
| [`help-flag`](help-flag.md) | always |

```bash
emissario help
emissario help start
```

Info · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
