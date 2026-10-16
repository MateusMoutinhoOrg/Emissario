# `GET /admin/root/export-visualization-views`

Downloads the page views of a period as CSV, newest first, 100 000 at most

Root only. The views of the filters of the dashboard as text/csv, downloaded as visualization-views-<from>-<to>.csv, the time in the server's time zone; a cell a spreadsheet would run as a formula is led by a quote mark.

## Try it

Only what is required:

```bash
curl localhost:3000/admin/root/export-visualization-views
```

With every value it reads:

```bash
curl 'localhost:3000/admin/root/export-visualization-views?from=my-from&to=my-to&path=my-path&prefix=my-prefix&ip=my-ip&country=my-country&device=my-device&browser=my-browser&os=my-os&status=my-status&bots=my-bots&lang=my-lang' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host' \
  -b 'backoffice_language=my-backoffice-language'
```

More examples:

```bash
curl -b "backoffice_session=$SESSION" -o views.csv 'localhost:3000/admin/root/export-visualization-views?from=2026-10-01&to=2026-10-13'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `from` | query string | text | no | `my-from` | the first day of the period, YYYY-MM-DD, included; 6 days before to when absent |
| `to` | query string | text | no | `my-to` | the last day of the period, YYYY-MM-DD, included; today when absent; a period covers 366 days at most |
| `path` | query string | text | no | `my-path` | one canonical page, exactly (/blog/) |
| `prefix` | query string | text | no | `my-prefix` | keeps the pages starting with it (/blog/) |
| `ip` | query string | text | no | `my-ip` | one ip or one CIDR range (203.0.113.0/24): a root only, a viewer is answered 403 |
| `country` | query string | text | no | `my-country` | one ISO 3166-1 alpha-2 code (BR), XX for an unknown country |
| `device` | query string | text | no | `my-device` | desktop, mobile, tablet, bot or other |
| `browser` | query string | text | no | `my-browser` | Chrome, Firefox, Safari, Edge, Opera, Samsung Internet or Other |
| `os` | query string | text | no | `my-os` | Windows, macOS, Linux, Android, iOS, ChromeOS or Other |
| `status` | query string | text | no | `my-status` | 200, or 404 for the not-found views |
| `bots` | query string | text | no | `my-bots` | 1 includes the views of bots, left out of everything otherwise |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-client-port` | header | text | no | `my-x-client-port` | the source port of the visitor's connection, set by the server and never by the client — or by the reverse proxy start-server trusts with --client-port-header — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `lang` | query string | text | no | `my-lang` | the language a page is asked in, pt-BR — kept in the backoffice_language cookie for the next pages — read by [`backoffice-language`](backoffice_language.md), which runs first |
| `backoffice_language` | cookie | text | no | `my-backoffice-language` | the language chosen before, set by ?lang= — read by [`backoffice-language`](backoffice_language.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |
| `origin` | header | text | no | `my-origin` | the origin of the page that sent the request, sent by the browser — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |
| `host` | header | text | no | `my-host` | the host the request was sent to — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/csv`. |

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

---

For developers: `sandbox/internal/routes/export_visualization_views/` · Visualization · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
