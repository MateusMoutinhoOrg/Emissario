# `database-dir`

Read --database in front of every command

A middleware: it runs on rung 6, in front of every command line matching
`*`, and hands the line on unless it answers.

Runs in front of every command line and answers none of them. It reads --database into sandbox.Config.DatabaseDir, the folder every database's key-prefix is a path inside, so `start-server --database var/app` keeps every database there. The folder is relative to the directory the command runs from, its segments lower-case letters, digits, - and _: the store resolves every key under that directory and spells any other character escaped, so only such a path names the same folder on disk. No command declares the flag itself.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--database` | string | `data` | the folder every database lives under, relative to where the command runs, lower-case letters, digits, - and _ (defaults to data) | — |

| Runs in front of | When |
| --- | --- |
| [`add-backoffice-user`](add-backoffice-user.md) | always |
| [`help`](help.md) | always |
| [`version`](version.md) | always |
| [`start-server`](start-server.md) | always |

Middlewares · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
