# `POST /api/admin/list-contact-requests`

Answers one page of the contact requests, newest first: inbox, archived or all, filtered by period, country and a text of the JSON

Root and viewer. Answers {contact_requests: [{id, at, ip, country, archived, size, body, body_text}], total, page, pages, limit}; body is the JSON value sent, read back with its keys sorted and its numbers as doubles (an integer past 2^53 rounded), body_text the text exactly as sent. from and to are YYYY-MM-DD, both included, each optional; country an ISO code (XX unknown); contains a text the JSON holds, any case; status inbox (the default), archived or all. 400 for a bad filter.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/api/admin/list-contact-requests
```

With every value it reads:

```bash
curl -X POST localhost:3000/api/admin/list-contact-requests \
  -H 'authorization: my-authorization' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'Content-Type: application/json' \
  -d '{"contains":"text","country":"text","from":"text","limit":1,"page":1,"status":"inbox","to":"text"}'
```

More examples:

```bash
curl -X POST localhost:3000/api/admin/list-contact-requests -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"from":"2026-10-01","country":"BR","contains":"ana","status":"all","page":1,"limit":20}'
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

Send JSON with the header `Content-Type: application/json`, up to 64 KB. The body is optional.

| Field | What goes there | Required | Rules |
| --- | --- | --- | --- |
| `contains` | text | no | at most 200 characters |
| `country` | text | no | at most 8 characters |
| `from` | text | no | at most 16 characters |
| `limit` | whole number | no | from 1 to 100 |
| `page` | whole number | no | at least 1 |
| `status` | text | no | one of `inbox`, `archived`, `all` |
| `to` | text | no | at most 16 characters |

Example:

```json
{
  "contains": "text",
  "country": "text",
  "from": "text",
  "limit": 1,
  "page": 1,
  "status": "inbox",
  "to": "text"
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
| [`backoffice-api-token-auth`](backoffice_api_token_auth.md) | always |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |

---

For developers: `sandbox/internal/routes/api_list_contact_requests/` · Contact Requests API · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
