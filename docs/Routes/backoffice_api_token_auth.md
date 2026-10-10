# `ANY /api/admin/*`

Requires a valid API token on every /api/admin route

Every /api/admin route needs an API token, sent as Authorization: Bearer <token>. Tokens are created and revoked on the backoffice page /admin/list-backoffice-api-tokens; the api never issues one. A token may expire on a date or never, and may be limited to a list of client ips. A missing, unknown, revoked or expired token, or one used from an ip it does not allow, is answered 401 in JSON. After 20 invalid tokens from one client ip in 15 minutes, every token it sends is answered 429 with Retry-After, unchecked, until the 15 minutes pass.

## Try it

Only what is required:

```bash
curl localhost:3000/api/admin
```

With every value it reads:

```bash
curl localhost:3000/api/admin \
  -H 'authorization: my-authorization' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-request-method: my-x-request-method'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `authorization` | header | text | no | `my-authorization` | an API token created on /admin/list-backoffice-api-tokens, as Bearer <token> |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `application/json`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

## Runs first

These routes run before this one, on the same request. Any of them may refuse it — with `401`
or `403`, for example — or let it through to this route.

| Route | When |
| --- | --- |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |

---

For developers: `sandbox/internal/routes/backoffice_api_token_auth/` · Middleware · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
