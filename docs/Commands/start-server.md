# `start-server`

Start the http server

```bash
emissario start-server [--addr <addr>] [--read-timeout-ms <read-timeout-ms>] [--write-timeout-ms <write-timeout-ms>] [--shutdown-timeout-ms <shutdown-timeout-ms>] [--allow-x-forwarded-for] [--insecure-http] [--database <database>] [--help] [--visualization-geoip <visualization-geoip>] [--fonts <fonts>]
```

Opens the port and serves every route declared under sandbox/internal/routes, until the process is stopped. An interrupt (Ctrl+C) or a termination request stops it gracefully: no new request is taken, and the ones in flight get --shutdown-timeout-ms to finish.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--addr` | string | `3000:4000` | the port the server listens on, or a range of ports it takes the first free one of, with or without a host (8080, 4000:5000, 127.0.0.1:4000:5000, :8080) | — |
| `--read-timeout-ms` | integer | `10000` | how long a request has to arrive, in milliseconds | — |
| `--write-timeout-ms` | integer | `10000` | how long a response has to be written, in milliseconds | — |
| `--shutdown-timeout-ms` | integer | `10000` | how long the requests in flight have to finish once the server is asked to stop, in milliseconds (0 waits for them) | — |
| `--allow-x-forwarded-for` | boolean |  | trust the last entry of X-Forwarded-For as the client ip: turn it on only behind one reverse proxy that appends it (nginx), with the server bound to an address only that proxy reaches (--addr 127.0.0.1:3000) | [backoffice-start-server](backoffice-start-server.md) |
| `--insecure-http` | boolean |  | serve the backoffice over plain http, for local development only: the session cookie drops Secure and no Strict-Transport-Security is sent | [backoffice-start-server](backoffice-start-server.md) |
| `--database` | string | `data` | the folder every database lives under, relative to where the command runs, lower-case letters, digits, - and _ (defaults to data) | [database-dir](database-dir.md) |
| `--help`, `-h` | boolean |  | Print the help of the command this command line is for | [help-flag](help-flag.md) |
| `--visualization-geoip` | string |  | a country database (.mmdb) to look the ip of each page view up in instead of the embedded DB-IP Lite: a newer DB-IP Lite, MaxMind GeoLite2-Country, IPinfo Lite; a file that does not open stops the server at start | [visualization-start-server](visualization-start-server.md) |
| `--fonts` | string | `EmissarioFonts` | the folder the git sources are cloned into, one folder per source named after it, relative to where the server runs (defaults to EmissarioFonts) | [website-source-start-server](website-source-start-server.md) |

| Runs in front of it | When |
| --- | --- |
| [`agent-start-server`](agent-start-server.md) | always |
| [`backoffice-start-server`](backoffice-start-server.md) | always |
| [`database-dir`](database-dir.md) | always |
| [`help-flag`](help-flag.md) | always |
| [`visualization-start-server`](visualization-start-server.md) | always |
| [`website-source-start-server`](website-source-start-server.md) | always |

```bash
emissario start-server
emissario start-server --addr 4000:5000
```

Server · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
