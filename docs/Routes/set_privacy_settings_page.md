# `GET /admin/root/set-privacy-settings`

Shows the Privacy page: the privacy settings, what this server keeps, the legal holds and the notices

Root only. The Privacy page: the preset applied, the encryption at rest, the IP mode, what viewers see of a full IP, the IP retention, the access log (on/off, retention, two-factor for its readers, the inventory of its reads, whether the source port is read) with the Brazil note while it is off, the backup retention, the template of the privacy notice draft, the access log readers, what this server keeps (one row per table of personal data: fields, purpose, a suggested legal basis, retention, rows, oldest), the legal holds with Release, the notices of the settings in use, and the danger zone (hash the IPs stored, rotate the IP key). notice, count and reason tell the outcome of the last action.

## Try it

Only what is required:

```bash
curl localhost:3000/admin/root/set-privacy-settings
```

With every value it reads:

```bash
curl 'localhost:3000/admin/root/set-privacy-settings?notice=my-notice&count=0&reason=my-reason' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `notice` | query string | text | no | `my-notice` | the outcome code of the last action |
| `count` | query string | whole number | no — `0` when left out | `0` | how many records the last action changed |
| `reason` | query string | text | no | `my-reason` | why the last action was refused |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-client-port` | header | text | no | `my-x-client-port` | the source port of the visitor's connection, set by the server and never by the client — or by the reverse proxy start-server trusts with --client-port-header — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |
| `origin` | header | text | no | `my-origin` | the origin of the page that sent the request, sent by the browser — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |
| `host` | header | text | no | `my-host` | the host the request was sent to — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/html`. |
| `400` | Something you sent is missing or has the wrong type or format. The answer's `field` names it. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

## Runs first

These routes run before this one, on the same request. Any of them may refuse it — with `401`
or `403`, for example — or let it through to this route.

| Route | When |
| --- | --- |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-root-guard`](backoffice_root_guard.md) | always |
| [`backoffice-same-origin`](backoffice_same_origin.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |
| [`backoffice-session-auth`](backoffice_session_auth.md) | depends on the address — `explain-route` gives the exact answer |

---

For developers: `sandbox/internal/routes/set_privacy_settings_page/` · Privacy · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
