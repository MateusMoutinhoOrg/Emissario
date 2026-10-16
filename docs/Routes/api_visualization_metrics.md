# `POST /api/admin/visualization-metrics`

Answers the metrics of the page views of a period: the totals, the views over time and the top pages, countries, devices, browsers, operating systems, referrers, not-found paths and visitors

Open to every backoffice user. Every field is optional: the period defaults to the last 7 days, bucket to hour up to 2 days, day up to 90 and month above, top to 10. Answers {summary, timeline, top}; unique visitors are counted on the whole ip, and the ips of top.visitors are masked to a viewer. A bad filter is answered 400, an ip sent by a viewer 403.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/api/admin/visualization-metrics
```

With every value it reads:

```bash
curl -X POST localhost:3000/api/admin/visualization-metrics \
  -H 'authorization: my-authorization' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'Content-Type: application/json' \
  -d '{"bots":true,"browser":"text","bucket":"hour","country":"text","device":"text","from":"text","ip":"text","os":"text","path":"text","prefix":"text","status":1,"to":"text","top":1}'
```

More examples:

```bash
curl -X POST localhost:3000/api/admin/visualization-metrics -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"from":"2026-10-07","to":"2026-10-13","country":"BR","bucket":"day","top":10}'
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
| `bucket` | text | no | one of `hour`, `day`, `month` |
| `country` | text | no |  |
| `device` | text | no |  |
| `from` | text | no |  |
| `ip` | text | no |  |
| `os` | text | no |  |
| `path` | text | no |  |
| `prefix` | text | no |  |
| `status` | whole number | no |  |
| `to` | text | no |  |
| `top` | whole number | no | from 1 to 100 |

Example:

```json
{
  "bots": true,
  "browser": "text",
  "bucket": "hour",
  "country": "text",
  "device": "text",
  "from": "text",
  "ip": "text",
  "os": "text",
  "path": "text",
  "prefix": "text",
  "status": 1,
  "to": "text",
  "top": 1
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

---

For developers: `sandbox/internal/routes/api_visualization_metrics/` · Visualization API · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
