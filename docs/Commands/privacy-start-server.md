# `privacy-start-server`

Start the access log writer and the privacy cleanup in front of start-server

A middleware: it runs on rung 55, in front of every command line matching
`start-server`, and hands the line on unless it answers.

Runs in front of start-server and answers nothing. It adds the privacy settings with a new random IP key on a new database, starts the writer of the access log — which keeps every page view and every call to the public APIs with the full IP, the source port and the time in UTC once a root turns it on at /admin/root/set-privacy-settings — and the hourly cleanup that deletes what is past its retention. Behind one reverse proxy (--allow-x-forwarded-for), the port the server sees is the proxy's own: --client-port-header names the header the proxy sets the visitor's port in, trusted under the same rule as X-Forwarded-For (nginx: proxy_set_header X-Client-Port $remote_port).

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--client-port-header` | string |  | the header the reverse proxy in front sets to the source port of the visitor's connection (X-Client-Port, with nginx: proxy_set_header X-Client-Port $remote_port), which the access log keeps; read only with --allow-x-forwarded-for | — |

| Runs in front of | When |
| --- | --- |
| [`start-server`](start-server.md) | always |

Middlewares · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
