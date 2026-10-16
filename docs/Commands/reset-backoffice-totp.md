# `reset-backoffice-totp`

Turn off the two-factor sign-in of a backoffice user

```bash
emissario reset-backoffice-totp --username <username> [--database <database>] [--notencrypt] [--help]
```

Forgets the two-factor secret of a backoffice user and turns two-factor sign-in off: they sign in with their password alone until they set it up again on /admin/set-totp. It is the way back in for a root who lost their phone; a root resets another user's from the users page.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--username` | string, required |  | the username or the email of the backoffice user | — |
| `--database` | string | `data` | the folder every database lives under, relative to where the command runs, lower-case letters, digits, - and _ (defaults to data) | [database-dir](database-dir.md) |
| `--notencrypt` | boolean |  | open the databases in clear, without EMISSARIO_ENCRYPT_KEY: for local debugging only, never in production | [database-encryption](database-encryption.md) |
| `--help`, `-h` | boolean |  | Print the help of the command this command line is for | [help-flag](help-flag.md) |

| Runs in front of it | When |
| --- | --- |
| [`database-dir`](database-dir.md) | always |
| [`database-encryption`](database-encryption.md) | always |
| [`help-flag`](help-flag.md) | always |

```bash
emissario reset-backoffice-totp --username admin
```

Backoffice · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
