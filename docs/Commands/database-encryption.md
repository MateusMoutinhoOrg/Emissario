# `database-encryption`

Open every database encrypted with EMISSARIO_ENCRYPT_KEY, or in clear with --notencrypt

A middleware: it runs on rung 7, in front of every command line matching
`* !(help|version|generate-encrypt-key|rotate-encrypt-key)`, and hands the line on unless it answers.

Runs in front of every command line that may open a database, and answers none of them. It reads the key every database is encrypted with from the EMISSARIO_ENCRYPT_KEY environment variable — 32 random bytes in base64 or hex, which `emissario generate-encrypt-key` prints — removes it from the process environment, so git and its hooks never inherit it, and checks it against <database>/encryption.json before any record is read. A new database folder is encrypted with that key and gets its encryption.json; a folder encrypted with another key, a key missing, or a folder in clear the command refuses, with exit 1. --notencrypt opens a folder in clear instead, for local debugging only: it warns at every start, and every backoffice page says so. help, version and generate-encrypt-key open no database and skip it; rotate-encrypt-key reads both keys itself.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--notencrypt` | boolean |  | open the databases in clear, without EMISSARIO_ENCRYPT_KEY: for local debugging only, never in production | — |

| Runs in front of | When |
| --- | --- |
| [`add-backoffice-user`](add-backoffice-user.md) | always |
| [`start-server`](start-server.md) | always |

Middlewares · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
