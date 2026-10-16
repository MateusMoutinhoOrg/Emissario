# `GET /admin/root/export-access-log`

Exports the access log rows a filter keeps, to its readers

Root readers only, with two-factor sign-in while it is required: the rows the filter of the access log page keeps (ip, from, to, kind), at most 100000, as JSON (format=json, the default) or CSV (format=csv), each with its time in UTC with milliseconds, in the server's time zone, the ip, the source port, the kind and the ref. Every export is inventoried. Disclose it only on a court order (Marco Civil art. 10 §1).

## Try it

Only what is required:

```bash
curl localhost:3000/admin/root/export-access-log
```

With every value it reads:

```bash
curl 'localhost:3000/admin/root/export-access-log?ip=my-ip&from=my-from&to=my-to&kind=my-kind&format=my-format' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `ip` | query string | text | no | `my-ip` | the ip of the filter, or json or csv |
| `from` | query string | text | no | `my-from` | the from of the filter, or json or csv |
| `to` | query string | text | no | `my-to` | the to of the filter, or json or csv |
| `kind` | query string | text | no | `my-kind` | the kind of the filter, or json or csv |
| `format` | query string | text | no | `my-format` | the format of the filter, or json or csv |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-client-port` | header | text | no | `my-x-client-port` | the source port of the visitor's connection, set by the server and never by the client — or by the reverse proxy start-server trusts with --client-port-header — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |
| `origin` | header | text | no | `my-origin` | the origin of the page that sent the request, sent by the browser — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |
| `host` | header | text | no | `my-host` | the host the request was sent to — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/plain`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

## Runs first

These routes run before this one, on the same request. Any of them may refuse it — with `401`
or `403`, for example — or let it through to this route.

| Route | When |
| --- | --- |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-root-guard`](backoffice_root_guard.md) | always |
| [`backoffice-same-origin`](backoffice_same_origin.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |
| [`backoffice-session-auth`](backoffice_session_auth.md) | depends on the address — `explain-route` gives the exact answer |

---

For developers: `sandbox/internal/routes/export_access_log/` · Privacy · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
