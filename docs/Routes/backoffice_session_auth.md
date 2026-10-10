# `ANY /admin/* !(/admin/login)`

Requires a valid backoffice session on /admin, except /admin/login

## Try it

Only what is required:

```bash
curl localhost:3000/admin
```

With every value it reads:

```bash
curl localhost:3000/admin \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host' \
  -b 'backoffice_session=my-backoffice-session'
```

## In the address

| Part | What goes there | Example | Description |
| --- | --- | --- | --- |
| the whole address | text that must not be exactly `/admin/login` | — | the login route is reachable without a session |

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `backoffice_session` | cookie | text | no | `my-backoffice-session` | the session JWT set by POST /admin/login |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |
| `origin` | header | text | no | `my-origin` | the origin of the page that sent the request, sent by the browser — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |
| `host` | header | text | no | `my-host` | the host the request was sent to — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/html`. |

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

---

For developers: `sandbox/internal/routes/backoffice_session_auth/` · Middleware · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
