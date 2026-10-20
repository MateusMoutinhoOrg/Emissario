# Privacy

Emissario is **private by default** and gives the root what they need to run a website under the LGPD
(Brazil), the Marco Civil da Internet and the GDPR (EU): every database encrypted at rest, IPs hashed in
the records, a Marco Civil access log a root turns on, contact requests and agent summaries off until a
root turns them on, every piece of personal data deleted when its retention ends, and every risky choice
preceded by a notice that names the law. Every other lawful choice stays open: Emissario warns, and never
forbids. It keeps no record of who chose what, and never builds evidence against its own user.

> The notices are product information, not legal advice. The operator of the website is the
> **controller** (LGPD art. 5 VI; GDPR art. 4(7)): Emissario helps them comply and cannot comply for
> them. Legal references are as of October 2026.

## Run it

```bash
export EMISSARIO_ENCRYPT_KEY=$(emissario generate-encrypt-key)   # keep it in a secret manager: losing it loses every database and backup
emissario add-backoffice-user --username admin --email admin@example.com --role root
emissario start-server --addr 127.0.0.1:3000 --allow-x-forwarded-for --client-port-header X-Client-Port
```

Then, as a root, **Privacy** in the top bar (`/admin/root/set-privacy-settings`): pick a preset
(**Presets**) or set each value, read the privacy notice draft, and publish your own notice.

## What the law says, and what Emissario does

| Point | The law | Emissario |
|---|---|---|
| Keeping contacts or IPs | Lawful with a legal basis, transparency, a limited retention, security and a way to exercise rights. The obligations fall on the controller | The notices list the obligations the root takes on |
| How long | Neither law sets a number of days: "no longer than necessary" (LGPD arts. 15–16; GDPR art. 5(1)(e)) | Each retention has a default, a warning threshold and its reference (below) |
| Hashing IPs | A hash keyed with a secret cannot be reversed by whoever holds only the data; it is still **pseudonymous** personal data (GDPR art. 4(5), recital 26; CJEU C-413/23 P *SRB*) | `HMAC-SHA256(ip-key, ip)` under a random key, blanked when the IP retention ends |
| Access logs in Brazil | Marco Civil art. 15: a for-profit company keeps IP + date/time 6 months, confidential, disclosed on a court order (art. 10); Decreto 8.771/2016 art. 13: named readers, individual authentication, an inventory of reads, encryption | An access log a root turns on, with the source port and UTC time (STJ 2019, *TIM v. Google*), two-factor readers and an inventory of reads |
| Erasure | Data kept for a legal obligation or a legal claim is not erased on request (LGPD art. 16 I; GDPR art. 17(3)(b)(e)) | *Find and erase* keeps the access log and the held rows by default |
| Incidents | Res. CD/ANPD 15/2024: tell the ANPD **and the data subjects** within 3 business days when there is relevant risk; record every incident 5 years. GDPR arts. 33–34 | The notices and the checklist say so; encryption weighs on severity (LGPD art. 48 §3; GDPR art. 34(3)(a)), it does not make the incident disappear |
| AI | EU AI Act art. 50(1), since 2 Aug 2026: tell people they talk to an AI; the open-source exemption (art. 2(12)) does not cover art. 50 | The agent says it is an AI; the chat UI must say so too ([Agent](../Agent/doc.md)) |

Laws used: LGPD (Lei 13.709/2018) · ANPD resolutions CD/ANPD 2/2022, 15/2024, 18/2024, 19/2024, 32/2026 ·
Marco Civil da Internet (Lei 12.965/2014) and Decreto 8.771/2016 · GDPR (Reg. 2016/679) · ePrivacy
Directive art. 5(3) · EU AI Act (Reg. 2024/1689) · EU–Brazil adequacy (Commission decision C(2026) 373;
Res. CD/ANPD 32/2026) · ECA Digital (Lei 15.211/2025) · CNIL rec. 2021-122 and audience-measurement
guidance (references, not law). PL 2338/2023 (Brazil's AI bill) and the EU data omnibus are pending: no
default rests on them.

## Defaults

| Thing | Default |
|---|---|
| Encryption | key required, or `--notencrypt` for debugging |
| `POST /api/contact-requests` | off → `503 disabled` |
| IP mode | `hashed` |
| Full IPs for viewers | masked (`203.0.113.xxx`, IPv6 `/48`) |
| Access log | off; when on, two-factor for its readers and the inventory of reads |
| `POST /api/agent/summaries` | off → `503 summaries-disabled` |
| Agent | off |
| Full user agent of a view | stored only in the `full` IP mode |

## Encryption at rest

`EMISSARIO_ENCRYPT_KEY` is 32 random bytes in base64 or hex (`emissario generate-encrypt-key`); a
passphrase is refused. It is read once and removed from the process environment, so the `git` Emissario
runs never inherits it, and it never reaches a log or a page. The encryption is **Keep's own**: every
database is opened with `Props.EncryptKey` set to a key derived from it (HKDF-SHA256), so Keep seals
every value with AES-256-GCM (`12-byte nonce | ciphertext | 16-byte tag`) and names every index entry by
`SHA-256(EncryptKey ‖ 0 ‖ value)`. Every command that opens a database checks the key against
`<database>/encryption.json` — `{"version":2,"algorithm":"AES-256-GCM","key_id":"<8 hex>"}`
([database-encryption](../Commands/database-encryption.md)). A folder encrypted by Emissario v0.0.6
(version 1, XChaCha20-Poly1305) is refused: start a new one. Snapshots stay encrypted and restore only
on the same key.

**Rotating the key** — `rotate-encrypt-key --old-key-env <VAR>`, with the server stopped
([rotate-encrypt-key](../Commands/rotate-encrypt-key.md)):

| Step | Does | `encryption.json` |
|---|---|---|
| 1 | a `pre-rotation` snapshot, the databases opened under the old key | `rotating_to` names the new key |
| 2 | every file resealed under the new key; every index entry removed — its name is a digest of the old key | |
| 3 | under the new key, every database reindexed by Keep's `Repair`, every snapshot resealed and its index entries left out | `key_id` names the new key alone |

One interrupted is run again with both keys and reseals only what is left; until then every other
command refuses the folder. A restore reindexes every database once it has written the files, so a
snapshot a rotation resealed comes back whole.

**Who holds the key.** The controller generates it and keeps it — a secret manager injecting the
variable. Whoever installs, hosts or supports Emissario for them and can reach its data is their
**operador** (LGPD arts. 5 VII, 39, 42; GDPR art. 28): a contract sets the instructions, and they keep no
copy of the key once the work is delivered.

**What it does not do.** It protects a stolen disk, a stolen backup or a misconfigured volume (GDPR art.
32(1)(a); LGPD art. 46). It does **not** protect against an attacker inside the running server, which
holds the key in memory. An incident is still recorded for 5 years and reported when it may bring
relevant risk. In clear stay folder names (databases, tables, fields, ids, positions) and how many
records a table holds. Index file names are keyed digests, which no one without the key can check against
a guessed value; personal data is still never a `key` field. A value is not bound to the place it lives
at: whoever can write to the disk can move one ciphertext to another key of the same database.

`--notencrypt` keeps a folder in clear for local debugging: a warning at every start and a red banner
on every page. The IP key then sits in clear too, so the hashes protect nothing.

## IPs

| Mode (`privacy-setting.ip-mode`) | `visualization-view.ip`, `contact-request.ip`, `agent-usage.ip` |
|---|---|
| `hashed` (default) | `h:` + 16 bytes of `HMAC-SHA256(ip-key, ip)`, IPv4 whole, IPv6 by its `/64` |
| `full` (notice `ip-full-v1`) | the IP as it came, and the full user agent of a view |

The country, the ignored IPs, the rate limits and the access log read the raw IP in memory, before it is
hashed. The IP key is 32 random bytes in `privacy-setting`, encrypted like every value and never derived
from `EMISSARIO_ENCRYPT_KEY`, so rotating that key keeps every hash. The danger zone of the Privacy page
**hashes the IPs already stored** and **rotates the IP key**: after a rotation the old hashes no longer
match any IP (GDPR art. 11). A typed IP filters by its hash; a CIDR range needs full IPs.

## Access log (Marco Civil)

Off by default; turning it on asks for `access-log-v1`. Every page view the Visualization records and
every call to `/api/contact-requests` and `/api/agent/*` writes one `access-log` row: `at` (UTC, ms), the
full `ip`, the source `port`, the `kind` and the `ref` of the record it wrote.

- **Source port behind a proxy:** `--client-port-header <name>`, trusted only with
  `--allow-x-forwarded-for` (nginx: `proxy_set_header X-Client-Port $remote_port;`). Without it the Privacy
  page warns. A load balancer that does not forward it keeps the port in its own logs, under the same rules.
- **Clock:** keep the server's clock synced by NTP; pages show UTC and the server's time zone.
- **Readers (Decreto 8.771 art. 13):** only roots a root ticks on the Privacy page (I); signed in with
  TOTP while `access-log-totp` is on, their API tokens made after it (II); every list, search and export
  inventoried in `privacy-access` while `access-log-inventory` is on (III); encrypted (IV). Turning either
  off asks for its notice.
- **Page** `/admin/root/access-log`: filter by IP or CIDR, period and kind; export JSON or CSV for a court
  order; hold or delete the rows shown. Deleting rows inside their term asks `erase-kept-v1` every time.
- Turning it off stops new rows; the rows kept stay until their retention ends.

## Retention

| Data | Default | Notice past | Reference |
|---|---|---|---|
| Contact request | 180 days | 365 or blank | LGPD 15–16; GDPR 5(1)(e); CNIL: prospects ≤ 3 years |
| Access log | 184 days | below 184, above 365, or blank | Marco Civil art. 15; CNIL 2021-122 |
| IP in the records | 184 days | 365 or blank | CNIL 2021-122; a hash is still personal data |
| View | 90 days | 760 or blank | CNIL audience measurement ≤ 25 months |
| Agent summary | 90 days | 365 or blank | as contact requests |
| Snapshot | blank (kept) | blank, or past the longest above | data a retention removes lives on in older snapshots |

One cleanup in `privacy-start-server` runs every hour: it deletes the rows past their retention, blanks the
`ip` of the views, contact requests and agent usage past the IP retention (the record stays), releases
expired holds, removes the snapshots past the backup retention and forgets the erasures older than every
snapshot. Retention counts whole days of the server's time zone from `at`. A held row is skipped. It waits
while a backup runs, and logs counts alone.

## Notices

A notice is the text a root reads before a risky choice, in `assets/privacy/notices/<id>.md`, versioned
in its id. The server renders it — the backoffice runs no inline script — inside the form with a required
checkbox; the JSON twin answers `409 {"error":"acknowledgement-required","field":"acknowledge","notice",
"notices"}` until every id is in `acknowledge`. **Nothing is recorded**: not who confirmed, nor when, nor
that it was shown; it comes back every time the setting is turned to that value again.

| Id | Asked when |
|---|---|
| `contacts-v1` | contact requests turned on |
| `agent-v1`, `summaries-v1` | the agent, its summaries turned on |
| `ip-full-v1`, `viewer-ip-v1` | full IPs; full IPs for viewers |
| `access-log-v1`, `access-log-no-totp-v1`, `access-log-no-inventory-v1` | the access log on; without two-factor; without its inventory |
| `erase-kept-v1` | every deletion of access log rows inside their term |
| `retention-<data>-v1` | a retention past its warning, or blank (generated from the table above) |

## The Privacy page — `/admin/root/set-privacy-settings`

The preset applied, the encryption, the settings form (IP mode, viewer IPs, IP retention, the access log
and its retention, two-factor and inventory, backup retention, notice draft template) with the Brazil note
while the access log is off, the access log readers, **What this server keeps** (one row per table of
personal data: fields, purpose, a suggested legal basis, retention, rows, oldest — the base of the records
of processing, LGPD art. 37, GDPR art. 30), the **privacy notice draft** (`br` pt-BR, `eu` English, or
`br-eu`, built from the settings, with placeholders for the controller, contact and encarregado/DPO), the
legal holds, the notices in use, and the danger zone.

**Find and erase** (`/admin/root/find-privacy-subject`, LGPD arts. 18–19; GDPR arts. 15, 17): an IP or a
text, searched in contact requests, views, agent usage, summaries and, for a reader, the access log.
Export JSON answers access and portability requests; Delete all leaves the access log rows and the held
rows, listed as kept (a hold's reason is never shown, Marco Civil art. 13 §4). An erasure keeps only the
table and the ids in `privacy-erasure`; every restore applies them again, so erased data does not come back.

**Legal holds:** a root holds a contact request or a summary from its page, a reader access log rows, with
a reason and an end date (60 days by default, Marco Civil art. 13 §3). Held rows are skipped by the
cleanup and by Find and erase until released.

**Backups:** the Backups page marks snapshots older than the shortest retention. A restore keeps
`privacydb` — the settings, the access log, its inventory, holds and erasures — and `backofficedb` live
unless *include-backoffice* is ticked ([Backups](../Backups/doc.md)).

## Presets — `/admin/root/privacy-presets`

A preset is a named, versioned and complete set of values for the IP mode, the access log and its
retention, the IP, contact request, view, summary and backup retentions and the notice draft, each with
its reference. Applying it goes through the same ranges and notices, and keeps
`privacy-setting.preset = <id>@<version>` — nothing about who applied it. It never turns on or off contact
requests, the agent or its summaries, and never sets the keys, viewer IPs, the readers, two-factor or the
inventory.

| | `lgpd@1` | `gdpr@1` | `lgpd-gdpr@1` |
|---|---|---|---|
| IP mode | hashed | hashed | hashed |
| Access log | **on** (Marco Civil art. 15) | off | **on** |
| Retentions (access log, IP, contact, view, summary, backup) | 184, 184, 180, 90, 90, 90 | same | same |
| Notice draft | `br` | `eu` | `br-eu` |
| Notices asked on a fresh install | `access-log-v1` | none | `access-log-v1` |

Each preset carries the **controller checklist** — what Emissario cannot do: publish a privacy notice,
name an encarregado/DPO, keep records of processing, a legitimate interest balancing test, an impact
report for the agent, a channel for requests, an incident plan and register, court orders only, sensitive
data and minors, an EU representative, processing terms with the model provider, the AI disclosure. It is
a reminder: Emissario stores no status for it.

**Preview** shows every setting now and in the preset, what the next cleanup would delete (counted now by
the same rules), and the notices. **Save current settings as preset** and **Import** make custom presets,
checked whole (a setting missing, out of range or unknown is refused); a version never changes. **Export**
downloads the JSON. A setting changed on its own page is drift: a banner and
`check-privacy-preset` show it, **Re-apply** or **Detach** answer it. A newer built-in version shows a
banner and is never applied by itself. Automation, with a root API token:

```bash
curl -X POST localhost:3000/api/admin/root/apply-privacy-preset -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"preset":"lgpd@1","acknowledge":["access-log-v1"]}'
curl -X POST localhost:3000/api/admin/root/check-privacy-preset -H "Authorization: Bearer $TOKEN"
# {"conform":true,"differences":[],"encrypted":true,"newer":null,"preset":"lgpd@1"}
```

## Routes

| Page | JSON twin (`POST /api/admin/root/…`) |
|---|---|
| `GET`/`POST /admin/root/set-privacy-settings` | `get-privacy-settings`, `set-privacy-settings`, `get-privacy-notice-draft` |
| `POST /admin/root/hash-privacy-ips`, `/rotate-privacy-ip-key` | `hash-privacy-ips`, `rotate-privacy-ip-key` |
| `GET /admin/root/find-privacy-subject`, `POST /erase-privacy-subject`, `GET /export-privacy-subject` | `find-privacy-subject`, `erase-privacy-subject` |
| `GET /admin/root/access-log`, `/export-access-log`, `POST /delete-access-log` | `list-access-log`, `export-access-log`, `delete-access-log`, `list-privacy-access` |
| `POST /admin/root/set-access-log-readers`, `/reset-backoffice-totp`; `GET`/`POST /admin/set-totp` | `set-access-log-readers`, `reset-backoffice-totp`; `/api/admin/set-totp` |
| `POST /admin/root/hold-privacy-records`, `/release-privacy-hold` | `hold-privacy-records`, `release-privacy-hold`, `list-privacy-holds` |
| `GET /admin/root/privacy-presets`, `/preview-privacy-preset`, `/export-privacy-preset`; `POST /apply-…`, `/save-…`, `/import-…`, `/delete-…`, `/detach-privacy-preset` | `list-privacy-presets`, `preview-…`, `apply-…`, `save-…`, `import-…`, `export-…`, `delete-…`, `detach-…`, `check-privacy-preset` |

Every request and answer is in [Routes](../Routes/doc.md).

## Where it lives

| Path | What |
|---|---|
| `sandbox/internal/encryption/` | `encryption.json`, the start-up table, index entries and the reindex |
| `adapters/impls/keepencryption/`, `adapters/impls/encryptedkeep/` | the key: read, derived into Keep's `EncryptKey`, set on every `Props`; the file-level reseal of a rotation |
| `sandbox/internal/privacy/` | settings and IP key, IP hashes, retention rules, notices, access log writer, holds |
| `sandbox/internal/privacyadmin/` | banners, retention cleanup, what this server keeps, notice draft, find and erase, presets |
| `sandbox/internal/databases/privacy_db/` | `privacy-setting`, `privacy-preset`, `access-log`, `privacy-access`, `privacy-hold`, `privacy-erasure` ([privacy_db](../Databases/privacy_db.md)) |
| `sandbox/internal/commands/middleware/database_encryption/`, `…/privacy_start_server/` | the key in front of every command; the access log, the cleanup and `--client-port-header` in front of `start-server` |
| `assets/privacy/notices/`, `assets/privacy/notice-drafts/`, `assets/privacy/presets/` | notices, draft templates, built-in presets (one file per version) |
| `assets/backoffice/privacy_*.html`, `access_log.html`, `set_totp.html` | the pages |

## Out of scope

A cookie banner (Emissario sets no cookie on visitors) · a self-service portal for visitors' requests ·
a blocklist by IP hash · KMS, HSM or a key server · SSO for the backoffice · an incident register or a
compliance status inside Emissario · choosing a preset by country · presets for laws other than BR/EU
(a custom preset can hold their values, not their notices) · locking a server to a preset · legal advice.
