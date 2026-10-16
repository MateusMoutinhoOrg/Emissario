# `POST /api/admin/root/add-ip-block`

Blocks an ip, a CIDR range, a hashed ip or a country

Root only. value is one ip (203.0.113.7), a CIDR range (203.0.113.0/24), a hashed ip (h: and 32 hex digits, as the records keep it) or a country code (BR); an IPv6 address is blocked by its /64. reason is at most 200 characters; days from 1 to 36500, 0 or absent to block until a root removes it. Answers 201 {ip_block}; 400 for a value refused, with the field; 409 for a value blocked already or one matching the ip the request came from.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/api/admin/root/add-ip-block \
  -H 'Content-Type: application/json' \
  -d '{"value":"text"}'
```

With every value it reads:

```bash
curl -X POST localhost:3000/api/admin/root/add-ip-block \
  -H 'authorization: my-authorization' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'Content-Type: application/json' \
  -d '{"days":1,"reason":"text","value":"text"}'
```

More examples:

```bash
curl -X POST localhost:3000/api/admin/root/add-ip-block -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"value":"203.0.113.0/24","reason":"spam","days":30}'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `authorization` | header | text | no | `my-authorization` | an API token created on /admin/list-backoffice-api-tokens, as Bearer <token> — read by [`backoffice-api-token-auth`](backoffice_api_token_auth.md), which runs first |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-client-port` | header | text | no | `my-x-client-port` | the source port of the visitor's connection, set by the server and never by the client — or by the reverse proxy start-server trusts with --client-port-header — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |

## Body

Send JSON with the header `Content-Type: application/json`, up to 64 KB. The body is required.

| Field | What goes there | Required | Rules |
| --- | --- | --- | --- |
| `days` | whole number | no | from 0 to 36500 |
| `reason` | text | no | at most 200 characters |
| `value` | text | yes | at most 64 characters |

Example:

```json
{
  "days": 1,
  "reason": "text",
  "value": "text"
}
```

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `application/json`. |
| `400` | Something you sent is missing or has the wrong type or format. The answer's `field` names it. |
| `413` | The body is larger than 64 KB. |
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
| [`ip-blocker`](ip_blocker.md) | always |

---

For developers: `sandbox/internal/routes/api_add_ip_block/` · IP Blocks API · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
