# `rotate-encrypt-key`

Re-encrypt every database and backup from an old key to the one in EMISSARIO_ENCRYPT_KEY

```bash
emissario rotate-encrypt-key --old-key-env <old-key-env> [--database <database>] [--help]
```

Reseals every database under --database, and every file every snapshot keeps, from the key in the variable --old-key-env names to the one in EMISSARIO_ENCRYPT_KEY: the path after a key leaked. There is no migration from a folder in clear. It takes a snapshot first, works file by file and writes encryption.json last, so a rotation interrupted is run again with both keys and reseals only what is left. Run it with the server stopped. Snapshots stay restorable with the new key; archives downloaded before stay sealed under the old one.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--old-key-env` | string, required |  | the environment variable holding the key the databases are encrypted with now, the one to rotate away from (EMISSARIO_OLD_ENCRYPT_KEY, say) | — |
| `--database` | string | `data` | the folder every database lives under, relative to where the command runs, lower-case letters, digits, - and _ (defaults to data) | [database-dir](database-dir.md) |
| `--help`, `-h` | boolean |  | Print the help of the command this command line is for | [help-flag](help-flag.md) |

| Runs in front of it | When |
| --- | --- |
| [`database-dir`](database-dir.md) | always |
| [`help-flag`](help-flag.md) | always |

```bash
emissario EMISSARIO_OLD_ENCRYPT_KEY=$OLD EMISSARIO_ENCRYPT_KEY=$(emissario generate-encrypt-key) emissario rotate-encrypt-key --old-key-env EMISSARIO_OLD_ENCRYPT_KEY
```

Privacy · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
