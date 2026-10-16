# `POST /api/admin/root/set-agent-settings`

Saves the provider and the limits of the public agent

Root only. A blank or absent api_key keeps the one saved, clear_api_key removes it; a limit of 0 is no limit. Answers {settings} as get-agent-settings does; 400 for a value refused, with the field.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/api/admin/root/set-agent-settings \
  -H 'Content-Type: application/json' \
  -d '{"base_url":"text","daily_limit":1,"enabled":true,"ip_daily_limit":1,"max_tokens":1,"model":"text","rate_per_minute":1}'
```

With every value it reads:

```bash
curl -X POST localhost:3000/api/admin/root/set-agent-settings \
  -H 'authorization: my-authorization' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'Content-Type: application/json' \
  -d '{"api_key":"text","base_url":"text","clear_api_key":true,"daily_limit":1,"enabled":true,"ip_daily_limit":1,"max_tokens":1,"model":"text","rate_per_minute":1}'
```

More examples:

```bash
curl -X POST localhost:3000/api/admin/root/set-agent-settings -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"enabled":true,"base_url":"https://api.openai.com/v1","api_key":"sk-...","model":"gpt-4o-mini","max_tokens":1024,"daily_limit":200000,"ip_daily_limit":20000,"rate_per_minute":10}'
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
| `api_key` | text | no | at most 4096 characters |
| `base_url` | text | yes | at most 2048 characters |
| `clear_api_key` | `true` or `false` | no |  |
| `daily_limit` | whole number | yes | at least 0 |
| `enabled` | `true` or `false` | yes |  |
| `ip_daily_limit` | whole number | yes | at least 0 |
| `max_tokens` | whole number | yes | from 1 to 1000000 |
| `model` | text | yes | at most 200 characters |
| `rate_per_minute` | whole number | yes | from 1 to 10000 |

Example:

```json
{
  "api_key": "text",
  "base_url": "text",
  "clear_api_key": true,
  "daily_limit": 1,
  "enabled": true,
  "ip_daily_limit": 1,
  "max_tokens": 1,
  "model": "text",
  "rate_per_minute": 1
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

---

For developers: `sandbox/internal/routes/api_set_agent_settings/` · Agent API · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
