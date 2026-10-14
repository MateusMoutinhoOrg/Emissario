# `agent-start-server`

Start the writer of the public agent in front of start-server

A middleware: it runs on rung 53, in front of every command line matching
`start-server`, and hands the line on unless it answers.

Runs in front of start-server and answers nothing. It counts the tokens of today from agent-usage, for the limits of the public agent, and starts its goroutines: the writer, which turns the usage and the summaries POST /api/agent/messages and /api/agent/summaries queue into records of agent-db (holding them while a backup is taken or restored, so the chat keeps answering meanwhile), the retention, which removes every hour the usage older than 90 days, and the watcher, which counts the tokens of today again after a restore. The agent is configured at /admin/root/set-agent-settings; its instructions are the agent source, synced by website-source-start-server into <fonts>/agent.

| Runs in front of | When |
| --- | --- |
| [`start-server`](start-server.md) | always |

Middlewares · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
