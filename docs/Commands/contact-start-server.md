# `contact-start-server`

Start the writer of the contact requests in front of start-server

A middleware: it runs on rung 54, in front of every command line matching
`start-server`, and hands the line on unless it answers.

Runs in front of start-server and answers nothing. It starts the writer of the contact requests, which turns what POST /api/contact-requests queues into records of contact-db, holding them while a backup is taken or restored, so the website keeps sending meanwhile. The route answers 503 not-running in a process that did not start it. The limits are set at /admin/root/set-contact-settings.

| Runs in front of | When |
| --- | --- |
| [`start-server`](start-server.md) | always |

Middlewares · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
