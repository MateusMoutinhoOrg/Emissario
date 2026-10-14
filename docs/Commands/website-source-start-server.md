# `website-source-start-server`

Read --fonts and start syncing the website source and the agent source in front of start-server

A middleware: it runs on rung 51, in front of every command line matching
`start-server`, and hands the line on unless it answers.

Runs in front of start-server and answers nothing. It reads --fonts into sandbox.Config.FontsDir - the folder every source is cloned into, one folder per source named after it: the website source into <fonts>/frontend, the agent source into <fonts>/agent - and starts one goroutine per source that clones its repository and pulls its branch every sync time, and runs a sync whenever the Sync button or POST /api/admin/root/sync-website-source (or sync-agent-source) asks for one. A root sets each source at /admin/root/set-website-source and /admin/root/set-agent-source. git runs with the credentials of the user running the server; none is asked for or stored.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--fonts` | string | `EmissarioFonts` | the folder the git sources are cloned into, one folder per source named after it, relative to where the server runs (defaults to EmissarioFonts) | — |

| Runs in front of | When |
| --- | --- |
| [`start-server`](start-server.md) | always |

Middlewares · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
