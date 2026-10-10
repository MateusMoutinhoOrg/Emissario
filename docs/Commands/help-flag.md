# `help-flag`

Answer `<command> --help` with that command's help

A middleware: it runs on rung 5, in front of every command line matching
`*`, and hands the line on unless it answers.

Runs in front of every command line. Without --help it hands the line on. With it, it prints the help of the next strict command the line is for — or the general help when there is none — unless that command declares a --help flag of its own, which then reads it.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--help`, `-h` | boolean |  | Print the help of the command this command line is for | — |

| Runs in front of | When |
| --- | --- |
| [`add-backoffice-user`](add-backoffice-user.md) | always |
| [`help`](help.md) | always |
| [`version`](version.md) | always |
| [`start-server`](start-server.md) | always |

Middlewares · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
