# IP Blocks

A root blocks an ip, a CIDR range, a hashed ip or a whole country at **IP blocks**
(`/admin/root/ip-blocks`). A request from it is answered `403 {"error":"blocked","field":""}` by the
`ip-blocker` middleware (priority 12), before any route reads it: the website, `/api/contact-requests`,
`/api/agent/*`, `/health` and the sign-in. The blocks live in `blocker-db`
([Databases](../Databases/blocker_db.md)).

## What a block matches

| Kind | Typed as | Matches |
|---|---|---|
| `ip` | `203.0.113.7` | that IPv4 address |
| `range` | `203.0.113.0/24`, `2001:db8::/32` | every address in it. An IPv6 address typed alone is stored as its `/64` |
| `hash` | `h:` + 32 hex digits | the ip whose hash under the current IP key it is, IPv6 by its `/64` ([Privacy](../Privacy/doc.md)) |
| `country` | `BR` (ISO 3166-1 alpha-2) | every ip the country database of the page views places there ([Visualization](../Visualization/doc.md)) |

- The ip is `props.ClientIp`: behind a reverse proxy, `start-server --allow-x-forwarded-for` is required,
  otherwise every visitor is the proxy and a block of it blocks everyone.
- A request a backoffice session or an API token authenticated is never blocked. An `/admin` request
  without a session is sent to `/admin/login` first, which is blocked; `/api/admin` without a token is
  answered `401` first.
- A block that matches the ip the root adds it from is refused `409`: no root blocks themselves.
- A hash block stops matching after **Rotate IP key**: the page and the JSON mark it `stale`.
- A country block matches nothing while no country database is open (the page warns). A VPN or a
  proxy in another country gets through.
- A block may last 1 to 36500 days, or until a root removes it. The hourly retention cleanup deletes
  the ones that ended; until then they match nothing.
- At most 10000 blocks. Each request reads them from memory, read again from the store every 10
  seconds, after each change and after a restore. A store that cannot be read blocks nothing.
- **Turned away** counts the requests each block answered since the server started; it is not stored.
- An ip or a range is kept as typed, also in the hashed IP mode: the block needs it to match. A
  backup holds `blocker-db` like every database.

## Backoffice

| Page / form | Method | Does |
|---|---|---|
| `/admin/root/ip-blocks` | GET | every block, the newest first; the ip form; the country form. `?value=` fills the ip form |
| `/admin/root/add-ip-block` | POST | `value`, `reason` (≤ 200), `days` (blank: until removed) → `303 ?notice=added`; `400` invalid, `409` already blocked or yourself |
| `/admin/root/remove-ip-blocks` | POST | `ids` (≤ 100) → `303 ?notice=removed&count=`; `404` none found |

A root sees a **Block** link beside each ip of **Visualization → Visitors** and of a contact request: it
opens the page with that ip or hash in the form.

## JSON api

Root only, `Authorization: Bearer <token>` ([Backoffice](../Backoffice/doc.md)).

| Route | Body | Answers |
|---|---|---|
| `POST /api/admin/root/list-ip-blocks` | — | `{ip_blocks: [{id, value, kind, reason, created_at, expires_at, expired, stale, hits}], total}` |
| `POST /api/admin/root/add-ip-block` | `{value, reason?, days?}` (`days` 0 or absent: until removed) | `201 {ip_block}`; `400`, `409` |
| `POST /api/admin/root/remove-ip-blocks` | `{ids: [1, 2]}` | `{removed}`; `404` none found |

```bash
curl -X POST localhost:3000/api/admin/root/add-ip-block -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"value":"BR","reason":"no audience there","days":30}'
```

## Command line

```bash
emissario remove-ip-blocks --value 203.0.113.0/24   # the block of one value, typed as the page takes it
emissario remove-ip-blocks --all                    # every block
```

The way back in for a root the sign-in turns away. A running server goes by it within 10 seconds.

## Files

| Path | Holds |
|---|---|
| `sandbox/internal/blocker/` | parsing, adding, removing, the in-memory index and the match |
| `sandbox/internal/databases/blocker_db/` | `ip-block`: `value` (key), `kind`, `reason`, `created-at`, `expires-at` |
| `sandbox/internal/routes/blocker/middleware/ip_blocker/` | the middleware |
| `sandbox/internal/server/backoffice/backofficerender/ip_blocks.go`, `assets/backoffice/ip_blocks.html` | the page |
| `sandbox/internal/commands/backoffice/remove_ip_blocks/` | the command |
