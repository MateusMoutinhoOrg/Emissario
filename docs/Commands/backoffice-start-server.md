# `backoffice-start-server`

Read the backoffice secret and settings in front of start-server

A middleware: it runs on rung 50, in front of every command line matching
`start-server`, and hands the line on unless it answers.

Runs in front of start-server and answers nothing. It reads the secret that signs the backoffice sessions from the EMISSARIO_BACKOFFICE_SECRET environment variable, at least 32 characters and 10 different ones (openssl rand -hex 32); shorter or more repetitive, the server does not start. Unset, a random secret is generated for the run and a warning printed: it lives only in memory, so every session ends when the server restarts and no other instance accepts them. The secret is never a flag, which every user of the machine reads, nor a file, which can end up committed with the code.

The session cookie is Secure and Strict-Transport-Security is sent, so the backoffice is served over https, by a reverse proxy that terminates TLS; --insecure-http turns both off, for local development over plain http.

Behind one reverse proxy, bind the server to an address only the proxy reaches (--addr 127.0.0.1:3000) and pass --allow-x-forwarded-for: the client ip is then the last entry of X-Forwarded-For, the one the proxy appended. In nginx: proxy_set_header Host $host; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--addr` | string |  | the address start-server listens on, read to warn when X-Forwarded-For is trusted on every interface | — |
| `--allow-x-forwarded-for` | boolean |  | trust the last entry of X-Forwarded-For as the client ip: turn it on only behind one reverse proxy that appends it (nginx), with the server bound to an address only that proxy reaches (--addr 127.0.0.1:3000) | — |
| `--insecure-http` | boolean |  | serve the backoffice over plain http, for local development only: the session cookie drops Secure and no Strict-Transport-Security is sent | — |

| Runs in front of | When |
| --- | --- |
| [`start-server`](start-server.md) | always |

Middlewares · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
