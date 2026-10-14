# Contact Requests

The contact form of the website. The page sends any JSON to `POST /api/contact-requests`: an object, an
array, a string, a number. The website's own source chooses its shape. Emissario keeps it exactly as
sent, with the time it arrived, the ip it came from and that ip's country, and shows it in the
backoffice at `/admin/list-contact-requests`. Every backoffice user reads and searches the contact
requests. A root archives, unarchives and deletes them, and sets the largest body and the limits per ip.

## Run it

```bash
emissario start-server                                   # contact-start-server starts the writer in front of it
emissario start-server --addr 127.0.0.1:3000 --allow-x-forwarded-for   # behind a reverse proxy
```

The route takes contact requests as soon as the server runs: on, 16384 bytes at most, 5 requests a
minute per ip. A root changes this at **Contacts → Settings** (`/admin/root/set-contact-settings`). Behind a
reverse proxy, `--allow-x-forwarded-for` is required, otherwise every visitor shares the proxy's ip, its
country and its limits.

## Public api

### `POST /api/contact-requests`

There is no authentication. The request must come from the same origin: a page of another site is
refused `403` and there is no CORS. A request with no `Origin` header, from curl or a server, passes.

- **Body:** any JSON value, with `Content-Type: application/json`. Its size is checked against the
  largest body of the settings (16384 bytes by default, 1048576 at most).
- It is stored **as sent**: its keys in their order, its numbers whole, its blanks kept.

Answers `201`:

```json
{"id": "3f9a0c1b2d4e5f60"}
```

### From the website

```js
const response = await fetch("/api/contact-requests", {
  method: "POST",
  headers: {"Content-Type": "application/json"},
  body: JSON.stringify({name, email, message, page: location.pathname})
});
if (response.status === 201) { /* thank the visitor */ }
else if (response.status === 429) { /* too many: Retry-After says when to try again */ }
```

### Errors — `{"error", "field"}`

| Status | `error` | Why |
|---|---|---|
| `400` | `invalid-json` | the body is not JSON |
| `400` | `this route requires a request body` | the body is empty |
| `403` | `origin` | sent by a page of another site |
| `413` | `too-large` | larger than the largest body of the settings (the route itself refuses more than 1 MiB before reading it) |
| `415` | — | a `Content-Type` other than `application/json` |
| `429` | `rate-limit` | past the requests per minute of the ip; `Retry-After: 60` |
| `429` | `ip-daily-limit` | past the requests per day of the ip; `Retry-After` until midnight |
| `503` | `disabled` | a root turned the route off |
| `503` | `restoring` | a backup is being restored; `Retry-After: 5` |
| `503` | `not-running` | the process did not run `start-server`, so no writer runs in it |
| `503` | `busy` | the queue of the writer is full (10 000); `Retry-After: 5` |

## Limits

Both limits count per ip and per process. An IPv6 ip counts by its `/64`. Every request that passes the
origin check counts, including one later refused `400` or `413`.

- **Requests per minute** (`rate-per-minute`, default `5`): a fixed window of a minute.
- **Requests per day** (`ip-daily-limit`, blank for no limit): the day of the server (`TZ`), so it
  starts over at midnight.

The counters live in memory, so a restart forgets them.

## The queue

A request never waits on the database. It is checked, given its id and put on a queue of 10 000,
and one goroutine, the writer, writes it to `contact-db`. A backup can be taken while the website
keeps sending: `backoffice-maintenance` lets `/api/contact-requests` through, and the writer holds the
queue until the snapshot ends, then writes it. While a backup is restored, the route answers `503`.

## Settings — `contact-setting`, one record, root only

| Field | Rule |
|---|---|
| `enabled` | on (the default) or off; off → `503 disabled` |
| `max-bytes` | the largest body, 2 to 1048576 bytes; blank is 16384 |
| `rate-per-minute` | requests one ip may send a minute, 1 to 10000; blank is 5 |
| `ip-daily-limit` | requests one ip may send a day, up to 1000000; blank is no limit |

The route goes by a save from the next request on.

## Database — `contact-db` (`./data/contactdb`, gitignored, in every backup)

| Table | Fields |
|---|---|
| `contact-request` | `ref` (the id, 16 hex characters), `at` (unix seconds), `ip`, `country` (ISO 3166-1 alpha-2, `XX` unknown), `body` (the JSON as sent), `size` (bytes), `archived` (`0` inbox, `1` archived) |
| `contact-setting` | the settings above |

The country comes from the ip, looked up locally in the country database of
[Visualization](../Visualization/doc.md#country) (the embedded DB-IP Lite, or `--visualization-geoip`).
A private or loopback ip is `XX`.

## Filters

Query string of the list, and the same keys in the JSON body of `list-contact-requests`:

| Filter | What |
|---|---|
| `status` | `inbox` (the default), `archived` or `all`: the tabs of the list |
| `from`, `to` | `YYYY-MM-DD`, both included, each optional (no bound when blank); days of the server's time zone. Quick links: **Today**, **7 days**, **30 days**, **90 days**, **All time** |
| `country` | an ISO code (`BR`), `XX` for unknown |
| `contains` | a text the JSON holds, any case, at most 200 characters: a key, a name, an email. A character the website sent as an escape (`ã`) is found as itself (`ã`) |
| `page`, `limit` | 20 per page, 1 to 100 |

The list is newest first. Every active filter shows as a chip with an ✕, and the country of a row is a link
that keeps the list to that country.

## Who sees what

| | root | viewer |
|---|---|---|
| List, search, one contact request whole (its JSON, time, ip and country) | yes | yes |
| The ip | whole | whole |
| Archive, move back to the inbox, delete (the ones selected, or every one before a date) | yes | no |
| Settings | yes | no (`/admin/root`) |

| Page | JSON twin (`POST /api/admin/…`, `Authorization: Bearer`) |
|---|---|
| `GET /admin/list-contact-requests` | `list-contact-requests` `{from, to, country, contains, status, page, limit}` |
| `GET /admin/get-contact-request?id=` | `get-contact-request` `{id}` |
| `POST /admin/root/archive-contact-requests` | `root/archive-contact-requests` `{ids, archived}` |
| `POST /admin/root/remove-contact-requests` | `root/remove-contact-requests` `{ids}` or `{before}` |
| `GET`/`POST /admin/root/set-contact-settings` | `root/get-contact-settings`, `root/set-contact-settings` |

The JSON api answers each contact request as `{id, at, ip, country, archived, size, body, body_text}`:

- `body` is the JSON value read back, with its keys sorted and its numbers as doubles, so an integer past
  2^53 is rounded.
- `body_text` is the text exactly as sent.

Every request and answer is in [Routes](../Routes/doc.md).

## Personal data (LGPD)

A contact request holds personal data: what the visitor typed (a name, an email, a phone) and the ip. It
stays until a root deletes it. There is no retention: delete the ones answered, or every one before a
date from **Settings → Danger zone**. Say in the website's privacy notice what the form keeps and why.
The JSON never leaves the server. The country is looked up locally.

## Where it lives

| Path | What |
|---|---|
| `sandbox/internal/contact/` | settings, the limits and the checks of the route, the queue and its writer, filters, listing, archive, removal, the layout of a JSON as sent |
| `sandbox/internal/databases/contact_db/` | the two tables; the queries by time in `methods_custom.go` |
| `sandbox/internal/routes/contact/contact_requests/` | `POST /api/contact-requests` |
| `sandbox/internal/routes/backoffice/*contact*`, `…/root/*contact*` | the pages |
| `sandbox/internal/routes/api/backoffice/api_*contact*`, `…/root/api_*contact*` | the JSON api |
| `sandbox/internal/server/backoffice/backofficerender/contact_*.go`, `assets/backoffice/contact_*.html` | the page templates |
| `sandbox/internal/server/backoffice/backofficeapi/contact.go` | the JSON documents |
| `sandbox/internal/commands/middleware/contact_start_server/` | starts the writer in front of `start-server` |

## Out of scope

- A form UI: the website builds its own.
- Notifications of a new contact request by email or webhook.
- A JSON schema the body must follow.
- CORS: the form posts from the website's own origin.
