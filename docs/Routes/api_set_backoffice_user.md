# `POST /api/admin/root/set-backoffice-user`

Edits a backoffice user; a missing or blank password keeps the current one

A new password ends every session of the user and revokes every API token of theirs, the token this request carries included when a root changes their own password.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/api/admin/root/set-backoffice-user \
  -H 'Content-Type: application/json' \
  -d '{"email":"text","id":1,"role":"root","username":"text"}'
```

With every value it reads:

```bash
curl -X POST localhost:3000/api/admin/root/set-backoffice-user \
  -H 'authorization: my-authorization' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-request-method: my-x-request-method' \
  -H 'Content-Type: application/json' \
  -d '{"email":"text","id":1,"password":"text","role":"root","username":"text"}'
```

More examples:

```bash
curl -X POST localhost:3000/api/admin/root/set-backoffice-user -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"id":2,"username":"ana","email":"ana@example.com","role":"root"}'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `authorization` | header | text | no | `my-authorization` | an API token created on /admin/list-backoffice-api-tokens, as Bearer <token> — read by [`backoffice-api-token-auth`](backoffice_api_token_auth.md), which runs first |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |

## Body

Send JSON with the header `Content-Type: application/json`, up to 1 MB. The body is required.

| Field | What goes there | Required | Rules |
| --- | --- | --- | --- |
| `email` | text | yes |  |
| `id` | whole number | yes |  |
| `password` | text | no |  |
| `role` | text | yes | one of `root`, `viewer` |
| `username` | text | yes |  |

Example:

```json
{
  "email": "text",
  "id": 1,
  "password": "text",
  "role": "root",
  "username": "text"
}
```

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `application/json`. |
| `400` | Something you sent is missing or has the wrong type or format. The answer's `field` names it. |
| `413` | The body is larger than 1 MB. |
| `415` | The body was not sent with `Content-Type: application/json`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

## Runs first

These routes run before this one, on the same request. Any of them may refuse it — with `401`
or `403`, for example — or let it through to this route.

| Route | When |
| --- | --- |
| [`backoffice-api-root-guard`](backoffice_api_root_guard.md) | always |
| [`backoffice-api-token-auth`](backoffice_api_token_auth.md) | always |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |

---

For developers: `sandbox/internal/routes/api_set_backoffice_user/` · Backoffice Users API · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
