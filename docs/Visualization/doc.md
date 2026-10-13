# Visualization

Every page of the public website that is viewed — which page, when, from which ip — recorded by the
server itself and read in the backoffice at `/admin/visualization`. No cookie, no tracking script on
the website, no external analytics service: everything is read from the request the `front` route
answered.

## Run it

```bash
emissario start-server                                          # the embedded DB-IP Lite country database
emissario start-server --visualization-geoip GeoLite2-Country.mmdb   # any other country .mmdb
emissario start-server --addr 127.0.0.1:3000 --allow-x-forwarded-for  # behind a reverse proxy: the visitor's ip
```

Then open `/admin/visualization` (the **Visualization** link of the backoffice). A file
`--visualization-geoip` names that does not open stops the server at start.

## What counts as a view

| Counted | Not counted |
|---|---|
| `GET` answered by `front` with an HTML page, `200` | `HEAD` |
| `GET` answered by `front` with the not-found page, `404` (a not-found view) | assets: any Content-Type but `text/html` |
| | `/admin`, `/api`, `/backoffice/*`, `/health`, `/openapi.json` |
| | ips of the **ignored ips** |
| | everything, while **tracking** is off |

A user agent holding `bot`, `crawler`, `spider`, `slurp`, `curl`, `wget`, `python-requests`,
`go-http-client`, `headless`… — or none — is recorded with device `bot` and left out of every metric
and list unless `bots=1`.

## The queue

`front` hands each answer to `visualization.Enqueue` and returns: an in-memory queue of 10 000 views,
written to the database by one goroutine. A full queue drops the view and counts it as **dropped**.
While a backup is taken or restored (`snapshots.Creating` / `snapshots.Restoring`) the writer holds
the queue and writes it afterwards. A failed write goes to stderr, never to the visitor.

| Piece | What |
|---|---|
| `visualization-start-server` | cli middleware, priority `52`, in front of `start-server`: reads `--visualization-geoip` into `sandbox.Config.VisualizationGeoipFile`, opens the country database, starts the writer and the retention cleanup |
| retention cleanup | every hour (and at start): removes the views older than `retention-days`; nothing when it is `0` |

## Database — `visualization-db` (`./data/visualizationdb`, gitignored)

`visualization-view`:

| Field | What |
|---|---|
| `at` | Unix seconds |
| `path` | the path requested, no query string, 512 characters at most |
| `page` | the canonical page: `/blog/`, `/blog/index.html` → `/blog/`; `/pagina`, `/pagina.html` → `/pagina`; `/`, `/index.html` → `/`. The path requested on a `404` |
| `status` | `200` or `404` |
| `ip` | `props.ClientIp` |
| `country` | ISO 3166-1 alpha-2, `XX` unknown |
| `device` | `desktop`, `mobile`, `tablet`, `bot`, `other` |
| `browser` | `Chrome`, `Firefox`, `Safari`, `Edge`, `Opera`, `Samsung Internet`, `Other` |
| `os` | `Windows`, `macOS`, `Linux`, `Android`, `iOS`, `ChromeOS`, `Other` |
| `user-agent` | the `User-Agent` header, 512 characters at most |
| `referrer` | the host of `Referer` when it is another site; empty for a direct visit or a link inside the site |

`visualization-setting`, one record: `enabled` (`1` on, the default), `retention-days` (`90` by
default, `0` forever), `ignored-ips` (one ip or CIDR range per line).

Device, browser and os come from substring rules in `sandbox/internal/visualization/useragent.go`.
A query by period reads only `at` first (`ListVisualizationViewsBetween`, `methods_custom.go`);
one query covers 366 days at most. Days are counted in the server's time zone (`TZ`), through
`localtimedeps`.

## Country

Looked up **locally** in an IP-to-country `.mmdb` held in memory: the ip is never sent anywhere.
A private, loopback or invalid ip, or one the database does not hold, is `XX`.

| Source | |
|---|---|
| embedded | `assets/visualization/dbip-country-lite.mmdb`, DB-IP *IP to Country Lite*, CC BY 4.0 — [IP Geolocation by DB-IP](https://db-ip.com) |
| `--visualization-geoip <file>` | a newer DB-IP Lite, MaxMind GeoLite2-Country, IPinfo Lite — any country `.mmdb` |

The lookup is the dep `geoipdeps` (`Open(content) → Database{Country(ip), BuiltAt, Kind}`), filled by
the adapter `maxmindgeoip` over `github.com/oschwald/maxminddb-golang`.

## Who sees what

| | root | viewer |
|---|---|---|
| Dashboard, Pages, Visitors, Views | yes | yes |
| ip of a visitor | whole | masked: `203.0.113.xxx`, IPv6 its first 48 bits (`2001:db8:1:xxxx:…`) |
| filter `ip` | yes | `403` |
| Export, Settings, Delete views | yes | no (`/admin/root`) |

Unique visitors are always counted on the whole ip.

## Filters

The query string of every page, and the same keys in the JSON body of every api:

| Key | What |
|---|---|
| `from`, `to` | `YYYY-MM-DD`, both included; the last 7 days by default |
| `path` | one canonical page, exactly |
| `prefix` | the pages starting with it |
| `ip` | one ip or CIDR range — root only |
| `country` | ISO code |
| `device`, `browser`, `os` | one value of the table above |
| `status` | `200` or `404` |
| `bots` | `1` (`true` in JSON) includes the bots |

`400` for a bad date, `from` after `to`, a period over 366 days, a bad ip or range, a value no view
holds, an unknown `sort` or `bucket`; `403` for an ip sent by a viewer.

## Pages

| Page | Route |
|---|---|
| Dashboard: cards, views over time (inline SVG, per hour up to 2 days, per day up to 90, per month above), top 10 pages, countries, devices, browsers, os, referrers, not-found paths, visitors — every row adds its filter | `GET /admin/visualization` |
| Pages: views, unique, not found, last view; `sort=views\|unique\|last\|page` | `GET /admin/list-visualization-pages` |
| Visitors: country of the last view, device, browser and os used most; `sort=views\|last` | `GET /admin/list-visualization-visitors` |
| Views: the raw log, newest first; one exact `ip` opens a visitor summary | `GET /admin/list-visualization-views` |
| Export: CSV, 100 000 rows at most | `GET /admin/root/export-visualization-views` |
| Settings: tracking, retention (1–3650 days, blank forever), ignored ips (100 at most); the country database, views stored, oldest, dropped, your ip | `GET` / `POST /admin/root/set-visualization-settings` |
| Delete views: `before=YYYY-MM-DD` or `all=1` | `POST /admin/root/purge-visualization-views` |

Listings paginate with `page` and `limit` (20, 1 to 100).

## JSON api

`POST`, `Authorization: Bearer <token>`, failures `{"error","field"}`.

| Route | Who | Body | Answers |
|---|---|---|---|
| `/api/admin/visualization-metrics` | all | filters + `bucket` (`hour`, `day`, `month`) + `top` (1–100, 10) | `{summary, bucket, timeline, top}` |
| `/api/admin/list-visualization-pages` | all | filters + `sort`, `page`, `limit` | `{website_pages, total, page, pages, limit}` |
| `/api/admin/list-visualization-visitors` | all | filters + `sort`, `page`, `limit` | `{visitors, total, page, pages, limit}` |
| `/api/admin/list-visualization-views` | all | filters + `page`, `limit` | `{views, total, page, pages, limit}` |
| `/api/admin/root/get-visualization-settings` | root | — | `{settings, geoip, stored, oldest_at, dropped}` |
| `/api/admin/root/set-visualization-settings` | root | `{enabled, retention_days, ignored_ips}` | `{settings}` |
| `/api/admin/root/purge-visualization-views` | root | `{before}` or `{all: true}` | `{removed}` |

```bash
curl -X POST localhost:3000/api/admin/visualization-metrics -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"from":"2026-10-07","to":"2026-10-13","country":"BR"}'
```

Every route, with a request to run, is in [Routes](../Routes/doc.md).

## Personal data (LGPD)

The ip is personal data: views are kept 90 days by default, the ip is masked to a viewer, and a root
deletes the views at any time — before a date or all of them. Every backup of the `--database`
folder holds `visualizationdb` too.

## Where it lives

| Path | What |
|---|---|
| `sandbox/internal/visualization/` | queue and writer, counted rules, canonical page, user agent, referrer, country, filters, metrics, listings, export, settings, retention |
| `sandbox/internal/databases/visualization_db/` | the two tables, `methods_custom.go` |
| `sandbox/api/visualizationconfig.go` | `VisualizationGeoipFile` |
| `sandbox/deps/geoipdeps/`, `adapters/impls/maxmindgeoip/` | the country lookup |
| `sandbox/deps/localtimedeps/`, `adapters/impls/stdlocaltime/` | the server's time zone |
| `assets/visualization/dbip-country-lite.mmdb` | the embedded country database |
| `sandbox/internal/commands/middleware/visualization_start_server/` | the middleware |
| `sandbox/internal/routes/backoffice/*visualization*`, `…/backoffice/root/*visualization*` | the pages |
| `sandbox/internal/routes/api/backoffice/api_*visualization*`, `…/root/api_*visualization*` | the JSON api |
| `sandbox/internal/server/backoffice/backofficerender/visualization_*.go`, `assets/backoffice/visualization_*.html` | the page templates |
| `sandbox/internal/server/backoffice/backofficeapi/visualization.go` | the JSON documents |
| `sandbox/internal/routes/front/handler.go` | hands each answer to the queue |

## Out of scope

Time on page, sessions, a JavaScript tracker, a real-time view, city or region, downloading a newer
country database by itself.
