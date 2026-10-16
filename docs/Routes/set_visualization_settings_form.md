# `POST /admin/root/set-visualization-settings`

Saves the settings of the visualization: tracking on or off, the retention in days and the ignored ips

Root only. enabled=1 turns tracking on, its absence off; a blank retention keeps the views forever, else 1 to 36500 days, past 760 days or blank asking for the retention-view-v1 notice; the ignored ips are one ip or CIDR range per line, 100 at most. A refused value answers the form again under a 400, a save that needs the notice under a 409 with it inside until acknowledge=retention-view-v1 is sent; a save answers 303 back to the page with ?notice=saved.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/admin/root/set-visualization-settings \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d ''
```

With every value it reads:

```bash
curl -X POST 'localhost:3000/admin/root/set-visualization-settings?lang=my-lang' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host' \
  -b 'backoffice_language=my-backoffice-language' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'acknowledge=text&enabled=text&ignored-ips=text&retention-days=text'
```

More examples:

```bash
curl -b "backoffice_session=$SESSION" -X POST localhost:3000/admin/root/set-visualization-settings -H 'Origin: http://localhost:3000' --data 'enabled=1&retention-days=90&ignored-ips=203.0.113.7'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-client-port` | header | text | no | `my-x-client-port` | the source port of the visitor's connection, set by the server and never by the client — or by the reverse proxy start-server trusts with --client-port-header — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `lang` | query string | text | no | `my-lang` | the language a page is asked in, pt-BR — kept in the backoffice_language cookie for the next pages — read by [`backoffice-language`](backoffice_language.md), which runs first |
| `backoffice_language` | cookie | text | no | `my-backoffice-language` | the language chosen before, set by ?lang= — read by [`backoffice-language`](backoffice_language.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |
| `origin` | header | text | no | `my-origin` | the origin of the page that sent the request, sent by the browser — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |
| `host` | header | text | no | `my-host` | the host the request was sent to — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |

## Body

Send form fields, like `name=value&other=value` with the header `Content-Type: application/x-www-form-urlencoded`, up to 16 KB. The body is required.

| Field | What goes there | Required | Rules |
| --- | --- | --- | --- |
| `acknowledge` | list of texts | no |  |
| `enabled` | text | no | at most 8 characters |
| `ignored-ips` | text | no | at most 8192 characters |
| `retention-days` | text | no | at most 16 characters |

Example:

```text
acknowledge=text&enabled=text&ignored-ips=text&retention-days=text
```

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/html`. |
| `400` | Something you sent is missing or has the wrong type or format. The answer's `field` names it. |
| `413` | The body is larger than 16 KB. |
| `415` | The body was not sent with `Content-Type: application/x-www-form-urlencoded`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

## Runs first

These routes run before this one, on the same request. Any of them may refuse it — with `401`
or `403`, for example — or let it through to this route.

| Route | When |
| --- | --- |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-language`](backoffice_language.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-root-guard`](backoffice_root_guard.md) | always |
| [`backoffice-same-origin`](backoffice_same_origin.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |
| [`backoffice-session-auth`](backoffice_session_auth.md) | depends on the address — `explain-route` gives the exact answer |
| [`ip-blocker`](ip_blocker.md) | always |

---

For developers: `sandbox/internal/routes/set_visualization_settings_form/` · Visualization · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
