# `generate-encrypt-key`

Print a new key for EMISSARIO_ENCRYPT_KEY

```bash
emissario generate-encrypt-key [--database <database>] [--help]
```

Prints a fresh key for EMISSARIO_ENCRYPT_KEY — 32 random bytes in base64 — on stdout, and on stderr the warning that losing it loses every database and every backup. It opens no database. Keep the key outside the server, in a secret manager, and inject it as the environment variable: a passphrase is never accepted in its place.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--database` | string | `data` | the folder every database lives under, relative to where the command runs, lower-case letters, digits, - and _ (defaults to data) | [database-dir](database-dir.md) |
| `--help`, `-h` | boolean |  | Print the help of the command this command line is for | [help-flag](help-flag.md) |

| Runs in front of it | When |
| --- | --- |
| [`database-dir`](database-dir.md) | always |
| [`help-flag`](help-flag.md) | always |

```bash
emissario generate-encrypt-key
emissario EMISSARIO_ENCRYPT_KEY=$(emissario generate-encrypt-key) emissario start-server
```

Privacy · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
