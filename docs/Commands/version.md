# `version`

Print the installed version

```bash
emissario version [--database <database>] [--help]
```

Prints the current version of the installed binary and exits.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--database` | string | `data` | the folder every database lives under, relative to where the command runs, lower-case letters, digits, - and _ (defaults to data) | [database-dir](database-dir.md) |
| `--help`, `-h` | boolean |  | Print the help of the command this command line is for | [help-flag](help-flag.md) |

| Runs in front of it | When |
| --- | --- |
| [`database-dir`](database-dir.md) | always |
| [`help-flag`](help-flag.md) | always |

```bash
emissario version
```

Info · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
