# `POST /api/admin/list-visualization-pages`

Lists every canonical page of the public website viewed in a period, paginated

Open to every backoffice user. Answers {website_pages, total, page, pages, limit}, each page {page, views, unique, not_found, last_at}.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/api/admin/list-visualization-pages
```

With every value it reads:

```bash
curl -X POST localhost:3000/api/admin/list-visualization-pages \
  -H 'authorization: my-authorization' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'Content-Type: application/json' \
  -d '{"bots":true,"browser":"text","country":"text","device":"text","from":"text","ip":"text","limit":1,"os":"text","page":1,"path":"text","prefix":"text","sort":"views","status":1,"to":"text"}'
```

More examples:

```bash
curl -X POST localhost:3000/api/admin/list-visualization-pages -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"prefix":"/blog/","sort":"views","page":1,"limit":20}'
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

Send JSON with the header `Content-Type: application/json`, up to 1 MB. The body is optional.

| Field | What goes there | Required | Rules |
| --- | --- | --- | --- |
| `bots` | `true` or `false` | no |  |
| `browser` | text | no |  |
| `country` | text | no |  |
| `device` | text | no |  |
| `from` | text | no |  |
| `ip` | text | no |  |
| `limit` | whole number | no |  |
| `os` | text | no |  |
| `page` | whole number | no |  |
| `path` | text | no |  |
| `prefix` | text | no |  |
| `sort` | text | no | one of `views`, `unique`, `last`, `page` |
| `status` | whole number | no |  |
| `to` | text | no |  |

Example:

```json
{
  "bots": true,
  "browser": "text",
  "country": "text",
  "device": "text",
  "from": "text",
  "ip": "text",
  "limit": 1,
  "os": "text",
  "page": 1,
  "path": "text",
  "prefix": "text",
  "sort": "views",
  "status": 1,
  "to": "text"
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
| [`backoffice-api-token-auth`](backoffice_api_token_auth.md) | always |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |
| [`ip-blocker`](ip_blocker.md) | always |

---

For developers: `sandbox/internal/routes/api_list_visualization_pages/` · Visualization API · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
