# `GET /admin/list-backoffice-users`

Lists backoffice users, filtered and paginated

## Try it

Only what is required:

```bash
curl localhost:3000/admin/list-backoffice-users
```

With every value it reads:

```bash
curl 'localhost:3000/admin/list-backoffice-users?search=my-search&role=my-role&page=1&limit=20&notice=my-notice' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `search` | query string | text | no | `my-search` | matches the username or the email, case-insensitive |
| `role` | query string | text | no | `my-role` | root, viewer, or empty for every role |
| `page` | query string | whole number | no — `1` when left out | `1` | the page to show, counted from 1 |
| `limit` | query string | whole number | no — `20` when left out | `20` | how many users a page shows, from 1 to 100 |
| `notice` | query string | text | no | `my-notice` | the outcome code of the last add, edit or remove |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |
| `origin` | header | text | no | `my-origin` | the origin of the page that sent the request, sent by the browser — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |
| `host` | header | text | no | `my-host` | the host the request was sent to — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/html`. |
| `400` | Something you sent is missing or has the wrong type or format. The answer's `field` names it. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

## Runs first

These routes run before this one, on the same request. Any of them may refuse it — with `401`
or `403`, for example — or let it through to this route.

| Route | When |
| --- | --- |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-same-origin`](backoffice_same_origin.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |
| [`backoffice-session-auth`](backoffice_session_auth.md) | depends on the address — `explain-route` gives the exact answer |

---

For developers: `sandbox/internal/routes/list_backoffice_users_page/` · Backoffice Users · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
