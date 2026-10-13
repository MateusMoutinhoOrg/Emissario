# Visualization

## Objective
Track every page of the public website that is viewed — **which page, when, and from which ip** — and show
it in the backoffice: search the views by page, by time and by ip, and read the metrics: total views,
unique visitors, views by country, by device, by browser, by operating system, by referrer, and the pages
that were not found.

## Naming
Everything this mechanic adds carries `visualization` in its name — database, tables, packages, routes,
templates, flags, config fields, assets — the way the backoffice names everything `backoffice*`, so a
future mechanic can take the plain names (`view`, `setting`, `--geoip`, `/admin/pages`) without a
collision. Every name below is the one to use.

---

## 1. What Counts as a View
The view is recorded by the `front` route (`sandbox/internal/routes/front/handler.go`), after it answers.

| Counted | Not counted |
|---|---|
| `GET` answered by `front` with an HTML page, status `200` | `HEAD` requests |
| `GET` answered by `front` with the not-found page, status `404` (a **not-found view**) | assets: anything whose Content-Type is not `text/html` (css, js, images, fonts…) |
| | `/admin`, `/api`, `/backoffice/*`, `/health`, `/openapi.json` |
| | ips in the **ignored ips** list of the settings |
| | everything, while **tracking** is off in the settings |

**Bots:** a user agent holding `bot`, `crawler`, `spider`, `slurp`, `curl`, `wget`, `python-requests`,
`go-http-client`, `headless`… is recorded with device `bot`, and left out of every metric and list
unless the filter `bots=1` is given.

**Never slow down or break the website:** the `front` handler only puts the view on an in-memory queue
(10 000 views) and returns; one goroutine writes the queue to the database. A full queue drops the view
and counts it as **dropped**. While a backup is taken or restored (`snapshots.Creating` /
`snapshots.Restoring`), the writer holds the queue and writes it afterwards. A failed write goes to
stderr, never to the visitor.

---

## 2. Database — `visualization-db` (key-prefix `visualizationdb`)
Created with `agnos add-database`, `add-table`, `add-table-field`. It lives in the `--database` folder,
so every backup already holds it.

### Table `visualization-view`
| Field | Type | What |
|---|---|---|
| `at` | integer | Unix seconds of the view |
| `path` | string | the path requested, without the query string, at most 512 characters |
| `page` | string | the canonical page: `/blog/`, `/blog/index.html` → `/blog/`; `/pagina`, `/pagina.html` → `/pagina`; `/` and `/index.html` → `/`. On a `404`, the path requested |
| `status` | integer | `200` or `404` |
| `ip` | string | `props.ClientIp` (behind a proxy, run with `--allow-x-forwarded-for`) |
| `country` | string | ISO 3166-1 alpha-2 (`BR`, `US`), `XX` when unknown — see **3. Country** |
| `device` | string | `desktop`, `mobile`, `tablet`, `bot`, `other` |
| `browser` | string | `Chrome`, `Firefox`, `Safari`, `Edge`, `Opera`, `Samsung Internet`, `Other` |
| `os` | string | `Windows`, `macOS`, `Linux`, `Android`, `iOS`, `ChromeOS`, `Other` |
| `user-agent` | string | the `User-Agent` header, at most 512 characters |
| `referrer` | string | the host of the `Referer` header when it is another site (`google.com`); empty for a direct visit or a link inside the site |

Device, browser and os are worked out from the user agent by hand-written substring rules
(`sandbox/internal/visualization/useragent.go`) — no third-party parser.

### Table `visualization-setting` (one record)
| Field | Type | What |
|---|---|---|
| `enabled` | integer | `1` tracking on (the default), `0` off |
| `retention-days` | integer | how many days a view is kept; `0` keeps them forever; default `90` |
| `ignored-ips` | string | ips and CIDR ranges never recorded, one per line |

**Queries by time:** only `key` fields are indexed, so a query by period reads
`ListVisualizationViews` with `VisualizationViewFilter{AtMin, AtMax}` — bounded by the retention. One
query covers **at most 366 days**. Days are counted in the server's time zone. A query the generated
methods do not cover goes in `methods_custom.go`.

---

## 3. Country (GeoIP)
The country comes from the ip, looked up **locally** in an IP-to-country database — the visitor's ip is
never sent to an external service.

- **Embedded database:** `assets/visualization/dbip-country-lite.mmdb` — DB-IP "IP to Country Lite",
  free, licensed CC BY 4.0. The license asks for credit: show **"IP Geolocation by DB-IP"** with a link
  to `https://db-ip.com` in the footer of the dashboard and in the README.
- **`start-server --visualization-geoip <file.mmdb>`** uses another country database instead: a newer
  DB-IP Lite, MaxMind GeoLite2-Country, IPinfo Lite — any country `.mmdb`. A file that does not open stops
  the server at start with a clear error.
- **New dep `geoipdeps`** — contract `Country(ip string) string` — filled by a new adapter
  `maxmindgeoip` over `github.com/oschwald/maxminddb-golang`, added the way `docs/Adapters/doc.md`
  describes. It is the one name without `visualization`: deps are named after the capability
  (`jwtdeps`, `ratelimitdeps`), and are shared — a future mechanic that needs a country reuses this one.
- A private, loopback or invalid ip, or an ip the database does not know → `XX`.

---

## 4. Start-server
Add `api.VisualizationConfig` (`sandbox/api/visualizationconfig.go`), embedded in `api.Config` like
`api.BackofficeConfig`, with `VisualizationGeoipFile string` — the `--visualization-geoip` file, empty for
the embedded database.

Add a cli middleware `visualization-start-server` (priority `52`, after `website-source-start-server`)
in front of `start-server`. It:

1. reads `--visualization-geoip` into `sandbox.Config.VisualizationGeoipFile` and opens the country
   database (the embedded one when the flag is not given);
2. starts the goroutine that writes the queue (see **1**);
3. starts the retention cleanup: every hour, removes the views older than `retention-days` (nothing when
   it is `0`).

Both goroutines live as long as the process.

---

## 5. Backoffice Pages
Add a **Visualization** link to the top bar of every backoffice page (root and viewer), and a
**Visualization** card on the home page with the views of today and of the last 7 days.

### Who sees what
| | root | viewer |
|---|---|---|
| Dashboard, Pages, Visitors, Views | yes | yes |
| ip of a visitor | full | **masked**: `203.0.113.xxx`; IPv6 only its first 48 bits (`2001:db8:1:xxxx:…`) |
| filter by `ip` | yes | no (`403`) |
| Export, Settings, Delete views | yes | no (`/admin/root`) |

Unique visitors are always counted on the full ip, whoever reads them.

### Common filters
Query string of the Dashboard, Pages, Visitors and Views pages, and the same keys in the JSON body of
their APIs (**6**):

| Filter | What |
|---|---|
| `from`, `to` | `YYYY-MM-DD`, both inclusive; default the last 7 days. Quick links: **Today**, **7 days**, **30 days**, **90 days** |
| `path` | one canonical page, exactly (`/blog/`) |
| `prefix` | the pages starting with it (`/blog/`) |
| `ip` | one ip or one CIDR range (`203.0.113.0/24`) — root only |
| `country` | ISO code (`BR`) |
| `device`, `browser`, `os` | one value of the table above |
| `status` | `200` or `404` |
| `bots` | `1` includes the bots |

Every page shows the active filters as chips, each with an ✕ that removes it.

### Routes of the pages
Named like the backoffice's (`list-backoffice-users-page`, `add-backoffice-user-form`): `-page` a `GET`
that answers html, `-form` a `POST` from a page.

| Route name | Method and path | Folder under `sandbox/internal/routes/backoffice/` |
|---|---|---|
| `visualization-dashboard-page` | `GET /admin/visualization` | `visualization_dashboard_page/` |
| `list-visualization-pages-page` | `GET /admin/list-visualization-pages` | `list_visualization_pages_page/` |
| `list-visualization-visitors-page` | `GET /admin/list-visualization-visitors` | `list_visualization_visitors_page/` |
| `list-visualization-views-page` | `GET /admin/list-visualization-views` | `list_visualization_views_page/` |
| `export-visualization-views` | `GET /admin/root/export-visualization-views` | `root/export_visualization_views/` |
| `set-visualization-settings-page` | `GET /admin/root/set-visualization-settings` | `root/set_visualization_settings_page/` |
| `set-visualization-settings-form` | `POST /admin/root/set-visualization-settings` | `root/set_visualization_settings_form/` |
| `purge-visualization-views-form` | `POST /admin/root/purge-visualization-views` | `root/purge_visualization_views_form/` |

### 5.1 Dashboard — `GET /admin/visualization` (root, viewer)
- **Filter bar** (common filters).
- **Cards:** Total views · Unique visitors (distinct ips) · Views per visitor · Not-found views (`404`) ·
  Bot views (left out of the rest, shown greyed).
- **Views over time:** bars of views and unique visitors — per hour up to 2 days, per day up to 90 days,
  per month above. Drawn on the server as inline SVG in the template (the backoffice CSP runs no inline
  script). An empty period shows *No views in this period*.
- **Top 10 tables**, each row with views, unique visitors and the % of the total as a bar:
  - Pages · Countries (flag and name) · Devices · Browsers · Operating systems · Referrers (empty is
    **Direct**) · Not-found paths · Top visitors (ip masked to a viewer).
- **Every row is a link that adds its filter** to the dashboard: `?path=/blog/` is the detail of one page,
  `?country=BR` the detail of one country, `?device=mobile` of the mobile traffic. Each table has a
  **See all** link to Pages, Visitors or Views with the same filters.
- Footer: *IP Geolocation by DB-IP* while the embedded database is in use.

### 5.2 Pages — `GET /admin/list-visualization-pages` (root, viewer)
Every canonical page viewed in the period, one row each:

| Column | What |
|---|---|
| Page | the canonical page; links to the dashboard `?path=` |
| Views | views in the period |
| Unique visitors | distinct ips |
| Not found | marked when the page answered `404` |
| Last view | when it was last seen |
| | **Views** link → Views `?path=` |

`prefix` searches, `sort=views|unique|last|page` (default `views`), `page` and `limit` paginate
(20 per page, 1 to 100).

### 5.3 Visitors — `GET /admin/list-visualization-visitors` (root, viewer)
One row per ip seen in the period:

| Column | What |
|---|---|
| Ip | masked to a viewer; to a root, links to Views `?ip=` |
| Country | of its last view |
| Device, browser, os | the ones it used most |
| Views | views in the period |
| Pages | distinct pages it saw |
| First seen, last seen | its first and last view in the period |

`sort=views|last` (default `views`), paginated like Pages.

### 5.4 Views — `GET /admin/list-visualization-views` (root, viewer)
The raw log, newest first: time, path, status, ip, country, device, browser, os, referrer — the full
user agent on hover. Every common filter, paginated like Pages.

Filtered by one exact ip (root), the page opens with a **visitor summary**: first and last view, total
views, the pages it saw, the countries and the user agents it came with.

A root also sees an **Export CSV** button with the filters of the page.

### 5.5 Export — `GET /admin/root/export-visualization-views` (root)
The views of the common filters as `text/csv` (`visualization-views-<from>-<to>.csv`), columns as in
Views, at most 100 000 rows, newest first.

### 5.6 Settings — `GET` / `POST /admin/root/set-visualization-settings` (root)
**Fields:**

| Field | Rule |
|---|---|
| Tracking | on / off; default on |
| Retention (days) | 1 to 3650; blank keeps the views forever; default `90` |
| Ignored ips | one ip or CIDR range per line, at most 100 (your own office, a monitor) |

**Read only:** the country database in use (embedded DB-IP Lite or the `--visualization-geoip` file) and its build
date · how many views are stored · the oldest view · views dropped since the server started · **your ip**
(`props.ClientIp`), with a button that adds it to the ignored ips.

**Danger zone:** **Delete views** — every view before a date, or every view. Asks for confirmation
(`data-confirm`).

Saving answers `303` back to the page with `?notice=`.

### 5.7 Delete views — `POST /admin/root/purge-visualization-views` (root)
The action of the Delete views form: `before` (`YYYY-MM-DD`) or `all=1`. Answers `303` to Settings with
`?notice=` and how many views were removed.

---

## 6. JSON API
Under `/api/admin`, `Authorization: Bearer <token>`, every action a `POST` with a JSON body — the twins of
the pages. Failures are the `{"error","field"}` of `sandbox/internal/server/errors/handle_*.go`.

**Common filters** in the body: `{from, to, path, prefix, ip, country, device, browser, os, status, bots}`,
the same rules as **5**.

Route names are `api-` + the last segment of the path, like `api-list-backoffice-users`; folders under
`sandbox/internal/routes/api/backoffice/` (`root/` for the root ones), the name in underscores.

| Route name | Method and path | Who | Body | Answers |
|---|---|---|---|---|
| `api-visualization-metrics` | `POST /api/admin/visualization-metrics` | root, viewer | filters + `bucket` (`hour`, `day`, `month`; chosen from the period when absent) + `top` (1 to 100, default 10) | `{summary, timeline, top}` |
| `api-list-visualization-pages` | `POST /api/admin/list-visualization-pages` | root, viewer | filters + `sort`, `page`, `limit` | `{website_pages, total, page, pages, limit}` |
| `api-list-visualization-visitors` | `POST /api/admin/list-visualization-visitors` | root, viewer | filters + `sort`, `page`, `limit` | `{visitors, total, page, pages, limit}` |
| `api-list-visualization-views` | `POST /api/admin/list-visualization-views` | root, viewer | filters + `page`, `limit` | `{views, total, page, pages, limit}` |
| `api-get-visualization-settings` | `POST /api/admin/root/get-visualization-settings` | root | `{}` | `{settings, geoip, stored, oldest_at, dropped}` |
| `api-set-visualization-settings` | `POST /api/admin/root/set-visualization-settings` | root | `{enabled, retention_days, ignored_ips}` | `{settings}` |
| `api-purge-visualization-views` | `POST /api/admin/root/purge-visualization-views` | root | `{before: "YYYY-MM-DD"}` or `{all: true}` | `{removed}` |

**Documents:**

```json
{
  "summary": {"views": 1520, "unique_visitors": 610, "views_per_visitor": 2.49, "not_found": 12, "bots": 340},
  "timeline": [{"start": 1760313600, "views": 210, "unique": 95}],
  "top": {
    "pages":     [{"value": "/", "views": 800, "unique": 500, "percent": 52.6}],
    "countries": [], "devices": [], "browsers": [], "os": [],
    "referrers": [], "not_found": [], "visitors": []
  }
}
```

| List | Each item |
|---|---|
| `website_pages` | `{page, views, unique, not_found, last_at}` |
| `visitors` | `{ip, country, device, browser, os, views, pages, first_at, last_at}` — ip masked to a viewer |
| `views` | `{id, at, path, page, status, ip, country, device, browser, os, referrer, user_agent}` — ip masked to a viewer |
| `settings` | `{enabled, retention_days, ignored_ips: [..]}` |
| `geoip` | `{source: "embedded" \| "<file>", built_at}` |

**Errors:** `400` an invalid filter (bad date, `from` after `to`, a period over 366 days, a bad CIDR,
an unknown `sort` or `bucket`) · `403` a viewer sending `ip`, or calling `/root` · `401` / `429` as every
`/api/admin` route.

Every route of **5** and **6** carries `private: true`, category `Visualization` (pages) or
`Visualization API` (JSON), and an example, like the other backoffice routes.

---

## 7. Where It Lives
| Path | What |
|---|---|
| `sandbox/internal/visualization/` | queue and writer, the counted / not counted rules, canonical page, user agent rules, filters and ip masking, metrics, settings, retention |
| `sandbox/internal/databases/visualization_db/` | tables `visualization-view` and `visualization-setting` |
| `sandbox/api/visualizationconfig.go` | `api.VisualizationConfig`: `VisualizationGeoipFile` |
| `sandbox/deps/geoipdeps/`, `adapters/impls/maxmindgeoip/` | the country lookup |
| `assets/visualization/dbip-country-lite.mmdb` | the embedded country database |
| `sandbox/internal/commands/middleware/visualization_start_server/` | `--visualization-geoip`, the writer, the retention cleanup |
| `sandbox/internal/routes/backoffice/*visualization*`, `sandbox/internal/routes/backoffice/root/*visualization*` | the pages (**5**, *Routes of the pages*) |
| `sandbox/internal/routes/api/backoffice/api_*visualization*`, `sandbox/internal/routes/api/backoffice/root/api_*visualization*` | the JSON api (**6**) |
| `sandbox/internal/server/backoffice/backofficerender/visualization_*.go`, `assets/backoffice/visualization_*.html` | the page templates |
| `sandbox/internal/server/backoffice/backofficeapi/visualization.go` | the JSON documents |
| `sandbox/internal/routes/front/handler.go` | hands each view to the queue |
| `docs/Visualization/` | the doc, like `docs/WebsiteSource/` |

---

## Important
- **Everything is server side**, read from the request: no cookie, no tracking script on the public
  website, no external analytics service.
- **The ip is personal data (LGPD):** 90 days of retention by default, masked to a viewer, and a root can
  delete the views at any time.
- Every yaml through the agnos commands (`add-database`, `add-table-field`, `add-route`,
  `add-command --middleware`, the dep and adapter commands); no generated file edited by hand; finish
  with `agnos build`.
- The website keeps working the same with tracking off, with the queue full, or with the country
  database missing a range.

## Out of Scope
- Time on page, sessions, a JavaScript tracker.
- Real-time view.
- City or region.
- Automatic download of a newer country database.
