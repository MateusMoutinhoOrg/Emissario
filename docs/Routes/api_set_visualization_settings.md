# `POST /api/admin/root/set-visualization-settings`

Sets the settings of the visualization: tracking on or off, the retention in days and the ignored ips

Root only. retention_days 0 keeps the views forever, else 1 to 36500; past 760 days, or 0, it answers 409 acknowledgement-required with the retention-view-v1 notice until its id is in acknowledge. ignored_ips holds ips and CIDR ranges, 100 at most. Answers {settings}.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/api/admin/root/set-visualization-settings \
  -H 'Content-Type: application/json' \
  -d '{"enabled":true,"ignored_ips":["text"],"retention_days":1}'
```

With every value it reads:

```bash
curl -X POST localhost:3000/api/admin/root/set-visualization-settings \
  -H 'authorization: my-authorization' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'Content-Type: application/json' \
  -d '{"acknowledge":["text"],"enabled":true,"ignored_ips":["text"],"retention_days":1}'
```

More examples:

```bash
curl -X POST localhost:3000/api/admin/root/set-visualization-settings -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"enabled":true,"retention_days":90,"ignored_ips":["203.0.113.7"]}'
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

Send JSON with the header `Content-Type: application/json`, up to 1 MB. The body is required.

| Field | What goes there | Required | Rules |
| --- | --- | --- | --- |
| `acknowledge` | list of texts | no |  |
| `enabled` | `true` or `false` | yes |  |
| `ignored_ips` | list of texts | yes | at most 100 items |
| `retention_days` | whole number | yes | from 0 to 36500 |

Example:

```json
{
  "acknowledge": [
    "text"
  ],
  "enabled": true,
  "ignored_ips": [
    "text"
  ],
  "retention_days": 1
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
| [`ip-blocker`](ip_blocker.md) | always |

---

For developers: `sandbox/internal/routes/api_set_visualization_settings/` · Visualization API · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
