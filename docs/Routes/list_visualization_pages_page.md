# `GET /admin/list-visualization-pages`

Lists every canonical page of the public website viewed in a period: its views, unique visitors and last view

Open to every backoffice user, with the filters of the dashboard; prefix searches the pages, sort orders them by views, unique, last or page, and page and limit paginate them, 20 a page.

## Try it

Only what is required:

```bash
curl localhost:3000/admin/list-visualization-pages
```

With every value it reads:

```bash
curl 'localhost:3000/admin/list-visualization-pages?from=my-from&to=my-to&path=my-path&prefix=my-prefix&ip=my-ip&country=my-country&device=my-device&browser=my-browser&os=my-os&status=my-status&bots=my-bots&sort=my-sort&page=1&limit=20' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host'
```

More examples:

```bash
curl -b "backoffice_session=$SESSION" 'localhost:3000/admin/list-visualization-pages?prefix=/blog/&sort=unique'
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
| `sort` | query string | text | no | `my-sort` | views (the default), unique, last or page |
| `page` | query string | whole number | no — `1` when left out | `1` | the page to show, counted from 1 |
| `limit` | query string | whole number | no — `20` when left out | `20` | how many rows a page shows, from 1 to 100 |
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
| [`backoffice-same-origin`](backoffice_same_origin.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |
| [`backoffice-session-auth`](backoffice_session_auth.md) | depends on the address — `explain-route` gives the exact answer |

---

For developers: `sandbox/internal/routes/list_visualization_pages_page/` · Visualization · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
