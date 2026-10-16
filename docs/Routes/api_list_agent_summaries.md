# `POST /api/admin/list-agent-summaries`

Answers one page of the summaries of a period, newest first, searched by text

Root and viewer. Answers {summaries: [{id, at, preview, messages, input_tokens, output_tokens}], from, to, total, page, pages, limit}; from and to are YYYY-MM-DD, both included, the last 30 days by default, 366 days at most; q is a text the summary or its id holds. 400 for a bad period.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/api/admin/list-agent-summaries
```

With every value it reads:

```bash
curl -X POST localhost:3000/api/admin/list-agent-summaries \
  -H 'authorization: my-authorization' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'Content-Type: application/json' \
  -d '{"from":"text","limit":1,"page":1,"q":"text","to":"text"}'
```

More examples:

```bash
curl -X POST localhost:3000/api/admin/list-agent-summaries -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"from":"2026-10-01","to":"2026-10-14","q":"refund","page":1,"limit":20}'
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
| `from` | text | no | at most 16 characters |
| `limit` | whole number | no | from 1 to 100 |
| `page` | whole number | no | at least 1 |
| `q` | text | no | at most 200 characters |
| `to` | text | no | at most 16 characters |

Example:

```json
{
  "from": "text",
  "limit": 1,
  "page": 1,
  "q": "text",
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

For developers: `sandbox/internal/routes/api_list_agent_summaries/` · Agent API · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
