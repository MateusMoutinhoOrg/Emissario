# `POST /admin/root/upload-backup`

Stores a zip archive a download built as a new snapshot

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/admin/root/upload-backup \
  -H 'Content-Type: application/zip' \
  --data-binary @file.bin
```

With every value it reads:

```bash
curl -X POST localhost:3000/admin/root/upload-backup \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host' \
  -H 'Content-Type: application/zip' \
  --data-binary @file.bin
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |
| `origin` | header | text | no | `my-origin` | the origin of the page that sent the request, sent by the browser — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |
| `host` | header | text | no | `my-host` | the host the request was sent to — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |

## Body

Send any data — a file, for example with the header `Content-Type: application/zip`, up to 256 MB. The body is required.

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `application/json`. |
| `400` | Something you sent is missing or has the wrong type or format. The answer's `field` names it. |
| `413` | The body is larger than 256 MB. |
| `415` | The body was not sent with `Content-Type: application/zip`. |

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

For developers: `sandbox/internal/routes/upload_backup/` · Backoffice Backups · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
