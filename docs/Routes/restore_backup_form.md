# `POST /admin/root/restore-backup/{Id:integer}`

Starts putting a snapshot back over every database, after a pre-restore snapshot of them, and answers at once

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/admin/root/restore-backup/1
```

With every value it reads:

```bash
curl -X POST localhost:3000/admin/root/restore-backup/1 \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'include-backoffice=true'
```

## In the address

| Part | What goes there | Example | Description |
| --- | --- | --- | --- |
| `{Id:integer}` | whole number | `1` |  |

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |
| `origin` | header | text | no | `my-origin` | the origin of the page that sent the request, sent by the browser — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |
| `host` | header | text | no | `my-host` | the host the request was sent to — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |

## Body

Send form fields, like `name=value&other=value` with the header `Content-Type: application/x-www-form-urlencoded`, up to 4 KB. The body is optional.

| Field | What goes there | Required | Rules |
| --- | --- | --- | --- |
| `include-backoffice` | `true` or `false` | no |  |

Example:

```text
include-backoffice=true
```

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/html`. |
| `400` | Something you sent is missing or has the wrong type or format. The answer's `field` names it. |
| `404` | `{Id:integer}` is not a whole number, so this route does not answer the address. |
| `413` | The body is larger than 4 KB. |
| `415` | The body was not sent with `Content-Type: application/x-www-form-urlencoded`. |

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

For developers: `sandbox/internal/routes/restore_backup_form/` · Backoffice Backups · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
