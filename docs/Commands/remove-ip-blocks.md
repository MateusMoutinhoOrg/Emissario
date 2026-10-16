# `remove-ip-blocks`

Remove IP blocks: the one of a value, or every one

```bash
emissario remove-ip-blocks [--value <value>] [--all] [--database <database>] [--notencrypt] [--help]
```

Removes the IP block of --value, or every one with --all. It is the way back in for a root whose network a block turns away from /admin/login: a request a backoffice session authenticated is never blocked, but a sign-in is.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--value` | string |  | the value of the block: an ip, a CIDR range, a hashed ip or a country code, as the IP blocks page shows it | — |
| `--all` | boolean |  | remove every block | — |
| `--database` | string | `data` | the folder every database lives under, relative to where the command runs, lower-case letters, digits, - and _ (defaults to data) | [database-dir](database-dir.md) |
| `--notencrypt` | boolean |  | open the databases in clear, without EMISSARIO_ENCRYPT_KEY: for local debugging only, never in production | [database-encryption](database-encryption.md) |
| `--help`, `-h` | boolean |  | Print the help of the command this command line is for | [help-flag](help-flag.md) |

| Runs in front of it | When |
| --- | --- |
| [`database-dir`](database-dir.md) | always |
| [`database-encryption`](database-encryption.md) | always |
| [`help-flag`](help-flag.md) | always |

```bash
emissario remove-ip-blocks --value 203.0.113.0/24
emissario remove-ip-blocks --all
```

Backoffice · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
