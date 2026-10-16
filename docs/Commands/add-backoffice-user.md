# `add-backoffice-user`

Create a backoffice user with a generated password, printed once

```bash
emissario add-backoffice-user --username <username> --email <email> [--role <role>] [--database <database>] [--notencrypt] [--help]
```

Adds a backoffice user under the same rules as the add form: the username and the email unique across both, regardless of case, and a valid email. The password is generated and printed once, so it never travels on the command line; change it on the edit page. --role defaults to viewer: give the first user --role root.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--username` | string, required |  | the username for the backoffice user | — |
| `--email` | string, required |  | the email for the backoffice user | — |
| `--role` | string, one of root/viewer | `viewer` | the role of the backoffice user: root manages every user, viewer only reads | — |
| `--database` | string | `data` | the folder every database lives under, relative to where the command runs, lower-case letters, digits, - and _ (defaults to data) | [database-dir](database-dir.md) |
| `--notencrypt` | boolean |  | open the databases in clear, without EMISSARIO_ENCRYPT_KEY: for local debugging only, never in production | [database-encryption](database-encryption.md) |
| `--help`, `-h` | boolean |  | Print the help of the command this command line is for | [help-flag](help-flag.md) |

| Runs in front of it | When |
| --- | --- |
| [`database-dir`](database-dir.md) | always |
| [`database-encryption`](database-encryption.md) | always |
| [`help-flag`](help-flag.md) | always |

```bash
emissario add-backoffice-user --username admin --email admin@example.com --role root
```

Backoffice · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
