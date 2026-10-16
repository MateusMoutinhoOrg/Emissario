# Privacy (LGPD / GDPR Adapt)

## Objective
Make Emissario **private by default** and give the root what they need to run it lawfully under the
LGPD (Brazil), the GDPR (EU) and the laws next to them: the database encrypted at rest, IPs hashed unless
the root asks for them in full, contact requests and agent summaries off until a root turns them on,
every piece of personal data deleted when its retention ends, and every risky choice preceded by a
notice that names the law and is acknowledged on record.

Every privacy choice is made by a root **in the backoffice** (a page and its `/api/admin` JSON twin).
The one exception is the encryption key, which comes from the environment. **Presets** (**§11**) apply a
whole policy in one step (LGPD, GDPR, both, or the company's own), with the same notices. They are
exported to configure other servers the same way, and they show when a server drifts from its policy.

> The notices written below are product information, not legal advice. The operator of the website is
> the **controller** (LGPD art. 5 VI; GDPR art. 4(7)). Emissario helps them comply and cannot comply for
> them. The Privacy page and `docs/Privacy/doc.md` say this. Legal references are as of October 2026.

---

## 1. What the law actually says (and how it changes the first draft)

| First draft said | What the law says | What this plan does |
|---|---|---|
| Storing contacts / IPs **violates** the law | Storing them is lawful when there is a legal basis, transparency, a limited retention, security and a way for people to exercise their rights. The obligations fall on the operator. | The popups list the **obligations** the root takes on, not "violations". |
| Warn when retention is **higher than the law** | Neither the LGPD nor the GDPR sets a number of days. They require "no longer than necessary" (LGPD arts. 15–16; GDPR art. 5(1)(e)). The numbers come from regulators' references and from other laws. | Each retention has a default, an allowed range and a **warning threshold** with its source (**§4**). |
| Hash IPs with a fast hash, "like a password" (`hash(ip + salt + secret)`) | It holds when the hash is **keyed** with a secret kept outside the readable data: whoever has only the database cannot reverse it. Without a key, IPv4's 2³² addresses are all hashed in seconds. A salt must be global to keep the hash deterministic, so salt + secret amount to one key. Whoever holds the key (the server, or whoever breaks into it) can still hash every address or confirm one IP. A stable value also singles a visitor out over time. So it is **pseudonymous** personal data, with retention and rights like any other (GDPR art. 4(5) and recital 26; EDPB Guidelines 01/2025; CJEU C-413/23 P *SRB*, 4 Sep 2025). | `HMAC-SHA256(ip-key, ip)` with a **fixed random key** stored encrypted in the database. The hash is deterministic, so it counts visitors over any period and can match an abuser later. It cannot be reversed without the key. It is blanked when the IP retention ends (**§6**). |
| Hashing IPs by default is always the safe side | In Brazil, Marco Civil da Internet (Lei 12.965/2014) **art. 15** *requires* a company that runs an internet application for profit to keep access logs (IP + date/time) for **6 months**, kept confidential and secured, and disclosed only on a court order (art. 10). | `hashed` remains the default. The Privacy page warns Brazilian companies and offers `full` mode with a 6-month default (**§6**). |
| Summaries: "AI can make mistakes" | That is true, and two more facts apply. The whole conversation goes to the model provider, which is a processor and often an international transfer. Since **2 Aug 2026** the EU AI Act (Reg. 2024/1689) **art. 50(1)** requires that people be told when they are talking to an AI. The 2026 AI omnibus postponed the high-risk rules, not this one. | The summary notice and a new agent notice cover both (**§8**). |
| Encrypt "on start server" | `add-backoffice-user` and every other command that opens a database writes to it too. | The key is required by **every command that touches a database** (**§3**). |

### Laws and rules referenced

| Law | Articles used here |
|---|---|
| **LGPD** — Lei 13.709/2018 (BR) | 3 (scope: data collected in Brazil or offer to people in Brazil) · 5 I, III, VI · 6 III (necessity), VI (transparency), VII (security), X (accountability) · 7 I, V, VI, IX (legal bases) · 9 (information to the data subject) · 10 (legitimate interest: balancing, transparency) · 12 (anonymised data) · 15–16 (end of processing, deletion) · 18 (rights) · 33 (international transfer) · 37 (records of processing) · 41 (encarregado) · 46 §2 (security from the design phase) · 48 (incidents) · 52 (fines: up to 2% of revenue in Brazil, capped at R$ 50 million per infraction) |
| **ANPD resolutions** | CD/ANPD 2/2022 (small-scale agents: simplified records, no encarregado required but a contact channel required) · CD/ANPD 15/2024 (incident communication within **3 business days**) · CD/ANPD 19/2024 (international transfers, standard contractual clauses) · Legitimate interest guide (2024) |
| **Marco Civil da Internet** — Lei 12.965/2014 (BR) | 5 VIII (access log = date/time of use from an IP) · 10 (confidentiality, disclosure only on a court order) · 15 (6 months of access logs for for-profit legal entities) · 16 II (no data beyond the purpose) |
| **GDPR** — Reg. (EU) 2016/679 | 3(2) (applies to a non-EU site that targets people in the EU) · 4(1), 4(5) · 5(1)(b)(c)(e)(f), 5(2) · 6(1)(a)(b)(f) · 11 (no duty to re-identify) · 13 (information at collection) · 15–21 (rights) · 25 (data protection by design and by default) · 27 (EU representative) · 28 (processor contract) · 30 (records) · 32 (security, encryption) · 33 (72 h breach notification) · 34(3)(a) (no need to tell people when the data was encrypted) · 37–39 (DPO, when required) · 44–49 (transfers) · 83 (fines up to €20 million or 4%) |
| **CJEU** | C-582/14 *Breyer* (2016): a dynamic IP is personal data for a website operator · C-413/23 P *EDPS v SRB* (2025): pseudonymised data stays personal for whoever can re-identify it |
| **ePrivacy Directive** 2002/58/EC art. 5(3) | Emissario writes no cookie and runs no script on visitors' devices, so its analytics needs no consent banner. A chat history the website keeps in the browser is strictly necessary for the service the visitor asked for. |
| **EU AI Act** — Reg. (EU) 2024/1689 | art. 50(1): tell people they are interacting with an AI, applicable since 2 Aug 2026 · art. 99(4): fines up to €15 million or 3% |
| **EU–US Data Privacy Framework** | Adequacy decision of July 2023. The General Court upheld it in *Latombe* (T-553/23, 3 Sep 2025), and the appeal C-703/25 P is pending. It covers only US providers that are certified. |
| **Regulators' references** (not law) | CNIL rec. 2021-122: security logs kept **6 months to 1 year**, up to 3 years only with a documented need · CNIL audience measurement: statistics kept **≤ 25 months** · CNIL commercial-management reference: prospect data **≤ 3 years** after the last contact |
| **Pending, watch** | Brazil's AI bill **PL 2338/2023** passed the Senate in December 2024 and is pending in the Chamber. It is not law. The EU **data omnibus** (GDPR and cookie changes) is still being negotiated. Do not build on either. |

---

## 2. Principles
- **Private by default** (GDPR art. 25; LGPD art. 46 §2): an install stores no contact request, no
  summary and no full IP.
- **Fail closed**: a missing or wrong key stops the program, and an unacknowledged notice leaves the
  feature off. Nothing silently downgrades.
- **No backward compatibility**: Emissario is in alpha. A database from an earlier version is not
  migrated: start a new one. Every install starts with the defaults of **§5**.
- **Accountability** (LGPD art. 6 X; GDPR art. 5(2)): each acknowledgement and each privacy setting change
  is recorded with who made it, when, and which notice version they read.
- **Need to know**: a viewer never sees a full IP anywhere. Today contact requests show the whole IP to
  viewers, and this plan changes that.

---

## 3. Encryption at rest

### Key
- Env **`EMISSARIO_ENCRYPT_KEY`**, named like `EMISSARIO_BACKOFFICE_SECRET` and
  `EMISSARIO_AGENT_API_KEY`. It is 32 random bytes in base64 or hex. Anything else is refused: a
  passphrase is not a key.
- New command **`emissario generate-encrypt-key`** prints a fresh key and a warning: *losing this key
  loses every database and every backup*.
- The key is read once and then **removed from the process environment** (`os.Unsetenv`), so the `git`
  that Emissario runs (`osexecrun`) and its hooks never inherit it. It never reaches a log, an error or a
  page.
- Subkeys are derived with HKDF-SHA256 for `storage` and `key-check`. Each purpose has its own subkey, so
  one key never does two jobs.

### Flag `--notencrypt`
For local debugging only. It prints a warning to stderr at every start, and every backoffice page shows a
red banner: *Database not encrypted — debug only*.

### Every command that opens a database
A new cli middleware **`database-encryption`** runs in front of every command line, with priority right
after `database-dir`, in the same way `database-dir` reads `--database`. It reads the key or
`--notencrypt` and checks the database folder before any record is read. Commands that open no database
(`help`, `version`, `generate-encrypt-key`) skip the check.

| Database folder | Key set | `--notencrypt` | Result |
|---|---|---|---|
| empty | yes | – | encrypted. Writes `<database>/encryption.json` |
| empty or plaintext | no | yes | plaintext, with the warning |
| any | no | no | exit 1: *set EMISSARIO_ENCRYPT_KEY (`emissario generate-encrypt-key`) or pass --notencrypt* |
| any | yes | yes | exit 1: contradictory |
| encrypted | the right key | – | runs |
| encrypted | another key | – | exit 1: *key does not match this database* |
| encrypted | no | yes | exit 1: an encrypted database cannot be opened without its key |
| plaintext | yes | – | exit 1: *this database was created with --notencrypt; start a new one* |

`encryption.json` holds `{"version":1,"algorithm":"AES-256-GCM","key_id":"<first 8 bytes of HMAC(key-check, "emissario")>"}`.
It reveals nothing about the key.

### Where it plugs in
Encrypt in the **storage layer under Keep** (`storagedeps` of `github.com/MateusMoutinhoOrg/Keep`), which
every database already goes through. The filters (`contains`, `StartsWith`, ranges) scan records in
memory after `Read`, so they keep working unchanged.

- Do **not** edit `adapters/impls/databasedeps/generated.databasedeps.go`, because it is generated.
  Choose one of these:
  - **(preferred)** Add an `encryptedstorage` adapter to Keep that wraps any storage, release Keep, and
    run `agnos set-dep databasedeps --version <new>`.
  - Add a project adapter for `databasedeps` (`agnos add-adapter`, origin `project`). It builds Keep over
    `remotebinding.New()` with `StorageDeps` wrapped, and `standard` binds it instead of the generated
    one.
- The key reaches the wrapper through a process-level holder that the `database-encryption` middleware
  sets. Until it is set, every storage call fails (fail closed).
- **Format of an encrypted value:** `0xE1 | key_id (4 B) | nonce (12 B, random) | AES-256-GCM ciphertext+tag`.
  The AAD is the storage key segments joined, so a value moved to another field or record fails to
  decrypt.
- Every function of the contract is wrapped:
  - `WriteIfValueEquals` compares **plaintext**. It reads the current ciphertext, decrypts it, compares
    it, then calls the inner CAS with that ciphertext as the old value.
  - `Append`, `InsertAt` and `ReadAt` on an encrypted value decrypt it, apply the change and re-encrypt
    the whole value.
- **What stays in clear** (document it): folder names (database, table and field names, ids), Keep's
  bookkeeping (`last-id`, `size`, `position`, `list/*`, locks) and index file names under
  `keys/<field>/<sha256>`. These hold ids and counts only. **Rule:** personal data is never declared as a
  `key` field, because its unsalted SHA-256 would be a file name.
- In an encrypted database, a value without the `0xE1` header is an error, not data.

### Rotation — `emissario rotate-encrypt-key`
- `EMISSARIO_ENCRYPT_KEY=<new> emissario rotate-encrypt-key --old-key-env <VAR>` re-encrypts from an old
  key to a new one. This is the path after a leak. There is no plaintext → encrypted migration.
- Before it writes, it takes a backup snapshot. It works record by record and can be resumed, and it
  writes `encryption.json` last.

### Backups
Snapshots copy the files as they are on disk, so they stay **encrypted**. A downloaded zip opens only on a
server with the same key. Restoring checks the snapshot's `encryption.json` and refuses one with another
key, or a plaintext one into an encrypted database.

### What encryption does and does not do
It protects a stolen disk, a stolen backup, or a misconfigured volume. In those cases it usually means
people do not have to be notified (GDPR art. 34(3)(a); LGPD art. 48 §3), and it is the measure named in
GDPR art. 32(1)(a) and LGPD art. 46. It does **not** protect against an attacker inside the running
server, which holds the key in memory. Say this in the doc.

---

## 4. Retention: defaults and thresholds

| Data | Default | Allowed | Warning (needs acknowledgement) | Reference for the threshold |
|---|---|---|---|---|
| Contact request | **180 days** | 1–1825 | > 365 | No legal number. Keep it while you answer it and while the relationship lasts (LGPD 15–16; GDPR 5(1)(e)). CNIL: prospects ≤ 3 years after the last contact. 5 years ≈ the Brazilian consumer limitation period (CDC art. 27), as the outer bound for a legal defence (LGPD art. 7 VI) |
| IP, full or hashed (`ip-retention-days`) | **184 days** (the longest 6-month span) | 1–1095 | < 184 in `full` (BR) **and** > 365 (EU) | Marco Civil art. 15: at least 6 months for BR for-profit companies. CNIL 2021-122: 6 months–1 year, up to 3 years only with a documented need. A hash is still personal data (**§1**), so the same limit applies to it |
| Visualization view | 90 (as today) | 1–3650, blank = forever (as today) | > 760 or blank | CNIL audience measurement: ≤ 25 months |
| Agent summary | **90 days** | 1–1825 | > 365 | as contact requests |
| Backup snapshot | blank (kept) | 1–3650 | blank, or longer than the longest retention above | Data that retention removes from the database lives on in the snapshots taken before |

- An `ip` field older than the IP retention is **blanked**, whether it holds a full IP or a hash. The
  record stays until its own retention ends. For example, a contact request kept 180 days with full IPs kept 30 days loses its IP on day 30.
- Retention is counted from `at`, in days of the server's time zone.
- Above the maximum → `400`. Above the warning → the acknowledgement of **§9** with the reference text.

### Retention cleanup
A goroutine in `privacy-start-server` runs every hour. The visualization cleanup that runs today moves
into it. It removes or blanks, in each table: `contact-request`, `visualization-view`, `agent-summary`,
and the `ip` of `agent-usage`, `contact-request` and `visualization-view`. It holds while a backup is
taken or restored, as the writers do. Each run logs only counts to stderr, never values.

**Backups:** the Backups page lists the snapshots that hold data older than the current retentions,
with a delete button. If a backup retention is set, snapshots older than it are removed by the same
goroutine.

---

## 5. Defaults

| Thing | Default |
|---|---|
| Encryption | key required, or `--notencrypt` for debugging (**§3**) |
| `POST /api/contact-requests` | **off** → `503 disabled` |
| IP mode | **`hashed`** |
| `POST /api/agent/summaries` | **off** → `503 summaries-disabled` |
| Agent | off (as today) |
| Full user agent in views | stored only in `full` mode |

---

## 6. IPs

### Modes — `privacy-setting.ip-mode`
| Mode | Stored in `visualization-view.ip`, `contact-request.ip`, `agent-usage.ip` |
|---|---|
| `hashed` (default) | `h:` + hex of the first 16 bytes of `HMAC-SHA256(ip-key, ip)`. IPv4 whole, IPv6 by its `/64`, the same grouping as the rate limits, so an attacker cannot dodge a match by rotating inside their `/64`. Blanked after the IP retention (**§4**) |
| `full` | the IP, blanked after the IP retention (**§4**) |

- **IP key:** 32 random bytes, created on the first start and kept in `privacy-setting.ip-key`. It is
  encrypted by the storage layer like every value, so a stolen disk or backup without
  `EMISSARIO_ENCRYPT_KEY` does not reveal it. It is **not derived** from `EMISSARIO_ENCRYPT_KEY`:
  rotating that key with `rotate-encrypt-key` must not change the hashes.
  - This is the password scheme with its two differences made explicit: the salt is **one for all rows**,
    so the same IP gives the same value, and it is **secret**, kept apart from the hashes.
  - **Do not reuse `passworddeps` (`pbkdf2password`).** Its fresh random salt per hash gives the same IP
    a different value every time, so there is nothing to count or match. Login works only because it
    knows which row to verify, and an IP has no row. Its 600 000 rounds also cost ~0.3 s of CPU per page
    view.
  - The secret is what protects the hash, not the hash function. With the salt readable next to the
    hashes, a salted SHA-256 IP was recovered in **4.7 s** on a 4-core laptop. The stored country
    narrowed the search to `177.0.0.0/8`, and the whole of IPv4 took ~20 min. With a global salt, even
    PBKDF2 is beaten by one table of all 2³² IPs computed once.
  - Combine the secret with `HMAC-SHA256(ip-key, ip)`, not `sha256(ip + salt + secret)`.
- **Rotate IP key** (Privacy page, danger zone, confirmed): new random key. Old hashes can no longer be
  matched to new visits nor confirmed against an IP. This is the way to "forget" linkage early, for
  example after a server compromise. Hashes and blocks made before it stop matching.
- **With `--notencrypt`** the IP key sits in clear in `data/`, so the hashes protect nothing. The debug
  warning says so.
- **Done before hashing, on the raw IP, in memory:** the country lookup, the ignored IPs, and the rate
  limits per minute and per day (in memory, as today).
- **What `hashed` gives and changes, documented in Visualization:**
  - Unique visitors are exact over any period within the retention.
  - The Visitors page groups by hash, with country, device, browser, os, views and first/last seen as
    today.
  - The root's `ip` filter takes a typed IP and filters by its hash over the whole retention. A CIDR
    filter is refused in `hashed` mode: a range cannot be hashed.
  - `agent-usage` per-IP daily limits keep working.
  - The full `user-agent` is not stored. Device, browser and os are. Together with a stable hash, the
    full string adds fingerprinting without serving any feature.
- **Blocking:** the hash is stable, so an abuser seen in Visitors, a contact request or agent usage can be
  matched again later. The blocklist is out of scope here (see **Out of scope**). Volumetric DDoS is
  stopped before Emissario (reverse proxy, firewall, CDN), not by a lookup per request inside it.
- **Viewer masking** (`203.0.113.xxx`, IPv6 /48) applies to `full` mode on **every** page and API,
  contact requests included. A viewer sees a hash as `h:3fa9…`.
- Backoffice users' own IPs (`api-token.last-used-ip`, `ips`) are a security log of staff, not of
  visitors. They are out of the modes and are encrypted with the rest.
- **Brazil note on the Privacy page (both modes):** *If this website is run by a company for profit,
  Marco Civil art. 15 requires access logs (IP, date and time) kept 6 months. Use `full` with at least 184
  days, or keep your reverse proxy's access logs for 6 months.*

---

## 7. Contact requests
- `contact-setting` gains `retention-days` (**§4**). There is no "forever".
- Turning `enabled` on requires the **contacts** notice (**§9**). Saving a retention above 365 requires
  the retention notice.
- The settings page shows: *Your website must say in its privacy notice what the form keeps, why, for how
  long, and how to ask for deletion*, with a link to the draft (**§10**).
- **Export** (root): `GET /admin/root/export-contact-requests` → JSON of the filtered or selected ones.
  This serves access and portability requests (LGPD art. 18 II, V; GDPR arts. 15, 20). Twin:
  `root/export-contact-requests`.

---

## 8. Agent and summaries
- `agent-setting` gains `summaries-enabled` (default `0`) and `summary-retention-days` (default 90).
- `POST /api/agent/summaries` answers `503 summaries-disabled` while it is off.
- Turning summaries on requires the **summaries** notice. Turning the agent on requires the **agent**
  notice (**§9**).
- The agent settings page shows the provider's host and the text *Every message is sent to
  `<host>`*. If the host is not `localhost` or a private address, it adds *this is a transfer to a third
  party*.
- `docs/Agent/doc.md` says that the website's chat UI must tell visitors they are talking to an AI (AI
  Act art. 50(1)), and gives an example line.

---

## 9. Notices (the "popups")

The backoffice CSP runs no inline script, so the notice is **rendered by the server**:

1. The root saves a setting that needs a notice without having acknowledged its current version.
2. The page comes back with the notice panel: the text, the laws, a required checkbox *I have read this
   and I am responsible for this processing as controller*, and **Confirm**.
3. The server, not the browser, checks the checkbox. A JSON client cannot skip it:
   `409 {"error":"acknowledgement-required","field":"acknowledge","notice":{"id":"contacts-v1","text":"…"}}`
   until it sends `"acknowledge": ["contacts-v1"]`.
4. The acknowledgement is stored in `privacy-acknowledgement`.

Notices live in `assets/privacy/notices/<id>.md`, versioned in their id (`contacts-v1`). A new version
asks again: a banner shows until the root acknowledges it, and the feature stays on.

### `contacts-v1`
> Turning on contact requests makes this server store what visitors send through your form (names,
> emails, phones, messages) with its time and the visitor's IP (or its hash). **You are the controller of
> this data.** That is lawful as long as you:
> - **tell visitors** in your privacy notice what is kept, why, for how long and how to ask for its
>   deletion (LGPD art. 9; GDPR art. 13);
> - have a **legal basis**, usually answering the request the visitor made (LGPD art. 7 V; GDPR art.
>   6(1)(b)), and do not use it for marketing without consent;
> - keep it **only while needed**: Emissario deletes requests older than the retention you set (LGPD arts.
>   15–16; GDPR art. 5(1)(e));
> - **answer** access and deletion requests (LGPD art. 18; GDPR arts. 15–17) and **report incidents**
>   (ANPD within 3 business days, Res. CD/ANPD 15/2024; EU authority within 72 h, GDPR art. 33);
> - appoint an **EU representative** if you are outside the EU and target people there (GDPR art. 27).
>
> Fines: LGPD up to 2% of revenue in Brazil, capped at R$ 50 million per infraction (art. 52). GDPR up to
> €20 million or 4% of worldwide turnover (art. 83).

### `ip-full-v1`
> Full IPs identify visitors. A dynamic IP is personal data (CJEU C-582/14 *Breyer*; LGPD art. 5 I).
> Storing them in views, contact requests and agent usage needs a legal basis, usually a legitimate
> interest in security and statistics with a documented balancing test (LGPD art. 10; GDPR art. 6(1)(f)),
> and the shortest retention that serves it (LGPD art. 6 III; GDPR art. 5(1)(c)).
> - **Brazil:** a company running this website for profit **must** keep access logs (IP, date and time)
>   for 6 months, confidential and secured, and disclose them only on a court order (Marco Civil arts. 10
>   and 15).
> - **EU:** there is no general duty to keep IPs. The CNIL recommends 6 months to 1 year for security
>   logs.
>
> Viewers never see a full IP. Only roots do.

### `summaries-v1`
> Summaries are written by an AI model from the visitor's conversation and stored here. The model is
> told to remove names, contacts, documents, places and dates, and a filter then removes CPF, CNPJ,
> emails, phones and card numbers. **AI makes mistakes.** A summary can still hold personal data, or
> details that identify someone when read together (LGPD art. 12; GDPR recital 26). Treat summaries as
> personal data: they are encrypted, kept for the retention you set, and deleted after it. To write
> one, the **whole conversation is sent to your model provider** (see the agent notice). Tell visitors
> about summaries in your privacy notice.

### `agent-v1`
> Every message a visitor sends goes to the model provider you configured (`<host>`). That provider
> processes it on your behalf: sign its data processing terms (GDPR art. 28). If it is outside Brazil
> or the EU, this is an international transfer (LGPD art. 33 and Res. CD/ANPD 19/2024; GDPR arts.
> 44–49; the EU–US Data Privacy Framework covers certified US providers only). A local model (Ollama)
> keeps conversations on your infrastructure. **Your chat UI must tell visitors they are talking to an
> AI** (EU AI Act art. 50(1), applicable since 2 August 2026).

### `retention-<data>-v1`
Generated from **§4**: *You set N days. Above T days, the references below ask for a documented reason:
…*. It lists that row's reference and the threshold that was crossed.

---

## 10. Privacy page — `/admin/root/set-privacy-settings` (root)
Linked from the top bar (root) and from the home page.

- **Preset** (**§11**), first on the page: *Preset: LGPD v1, applied on 2026-10-15 by maria*, or *No
  preset (install defaults)*. It lists the settings that differ from it and links to the Presets page.
- **Encryption** (read only): on/off, `key_id`, algorithm.
- **IP mode** and **IP retention**, with the Brazil note (**§6**). Danger zone, each irreversible and
  confirmed with `data-confirm`: *Hash the IPs already stored* · *Rotate IP key*.
- **What this server keeps**: one row per table holding personal data, with the fields, the retention,
  the count and the oldest record. This is the base of the records of processing (LGPD art. 37; GDPR art.
  30).
- **Privacy notice draft**: text built from the current settings. It covers what is kept, why, for how
  long, the model provider's host, that no cookie is set, the rights, and placeholders for the
  controller, the contact and the encarregado/DPO. The root copies it into the website. It is marked
  *draft — review it*. Its template is `privacy-setting.notice-draft`: `br` (pt-BR: controlador,
  encarregado, ANPD, Marco Civil logs), `eu` (en: controller, DPO, EU representative, supervisory
  authority) or `br-eu` (both, the default). A preset sets it (**§11**).
- **Find and erase a person** (LGPD art. 18; GDPR arts. 15, 17): an IP (hashed with the current IP key
  in `hashed` mode) or a text (email, phone). It shows the matches in contact requests, views and agent
  usage, with **Export JSON** and **Delete all** (confirmed). Hashes made before a *Rotate IP key* no
  longer match, and the page says so (GDPR art. 11).
- **History**: the acknowledgements (`privacy-acknowledgement`) and the privacy setting changes, preset
  applications and checklist ticks (`privacy-change`), newest first.
- **Backup retention** (**§4**).

| Route name | Method and path |
|---|---|
| `set-privacy-settings-page` | `GET /admin/root/set-privacy-settings` |
| `set-privacy-settings-form` | `POST /admin/root/set-privacy-settings` |
| `hash-privacy-ips-form` | `POST /admin/root/hash-privacy-ips` |
| `find-privacy-subject-page` | `GET /admin/root/find-privacy-subject` |
| `erase-privacy-subject-form` | `POST /admin/root/erase-privacy-subject` |
| `export-contact-requests` | `GET /admin/root/export-contact-requests` |

JSON twins under `/api/admin/root/`, named `api-` + the last path segment, as in Visualization:
`get-privacy-settings`, `set-privacy-settings`, `hash-privacy-ips`, `find-privacy-subject`,
`erase-privacy-subject`, `get-privacy-notice-draft`, `list-privacy-acknowledgements`,
`export-contact-requests`.

**Banners** on every backoffice page (root): running `--notencrypt` · a feature on without its current
acknowledgement · snapshots holding data past retention · settings that differ from the active preset ·
a new version of the active preset (**§11**).

---

## 11. Presets — `/admin/root/privacy-presets` (root)

A large customer asks two things of a privacy setup: *prove that this server follows our policy* and *set
up the other servers the same way*. A **preset** is a named, versioned and complete set of values for the
privacy settings, each value with its reference. Applying one **is** saving those settings. It goes
through the same ranges (**§4**), the same notices (**§9**) and the same record. It is a shortcut and a
proof, never a way around them. It does not make the operator compliant, and the checklist below lists
what is left to them.

### What a preset sets
| Sets | Never sets |
|---|---|
| `privacy-setting.ip-mode`, `ip-retention-days`, `backup-retention-days`, `notice-draft` · `contact-setting.retention-days` · `visualization-setting.retention-days` · `agent-setting.summary-retention-days` | **Whether** data is collected: `contact-setting.enabled`, `agent-setting.enabled` and `agent-setting.summaries-enabled` stay as they are. A preset says how data is kept, and a root still decides whether to keep it (**§2**). The keys (`EMISSARIO_ENCRYPT_KEY`, `ip-key`). Rate limits and sizes |

A preset holds **every** setting of the left column, so applying it always gives a known state. A preset
with a missing or unknown setting is refused.

### Built-in presets — `assets/privacy/presets/<id>.json`
| Setting | `lgpd` v1 — for-profit company in Brazil | `gdpr` v1 — EU | `lgpd-gdpr` v1 — Brazilian company that also serves the EU |
|---|---|---|---|
| IP mode | **`full`**: Marco Civil art. 15 requires the access log, and a hash cannot answer a court order for the IP (art. 10) | `hashed`: minimisation and privacy by default (GDPR arts. 5(1)(c), 25) | **`full`** (Marco Civil art. 15) |
| IP retention | 184: at least 6 months (Marco Civil art. 15) | 184: CNIL 2021-122, 6 months–1 year | 184: the span that meets both |
| Contact request | 180 | 180 | 180 |
| Visualization view | 90 | 90 | 90 |
| Agent summary | 90 | 90 | 90 |
| Backup snapshot | 90: data in a snapshot lives up to its retention + 90 days (184 + 90 < 1 year) | 90 | 90 |
| Notice draft (**§10**) | `br` | `eu` | `br-eu` |
| Notices asked when applied | `ip-full-v1` | – | `ip-full-v1` |

Neither law sets a number of days (**§1**), so the retentions are the same in all three. The presets
differ where the laws differ: the IP mode, the notice draft and the checklist. A Brazilian company that
keeps the access logs in its reverse proxy instead (**§6**) saves a custom preset from `lgpd` with
`hashed`.

### Controller checklist
Each preset carries the obligations that Emissario cannot fulfil for the operator. The Presets page lists
the items of the active preset. A root marks each one done with an **evidence** text (where the document
is, a ticket, a link), and each tick is a `privacy-change` row. Emissario does not verify the items, and
the page says *your own record, not a certification*. Items marked `agent` show only while the agent is
on.

| Item | `lgpd` | `gdpr` | `lgpd-gdpr` |
|---|---|---|---|
| Privacy notice published on the website (draft in **§10**) | art. 9 | art. 13 | both |
| Encarregado / DPO named, and their contact published | art. 41 (small-scale agents: a contact channel, Res. CD/ANPD 2/2022) | arts. 37–39, when required | both |
| Records of processing (start from *What this server keeps*, **§10**) | art. 37 | art. 30 | both |
| Legitimate interest balancing test for IPs and analytics | art. 10, ANPD guide 2024 | art. 6(1)(f) | both |
| Incident response plan | ANPD within 3 business days (Res. CD/ANPD 15/2024) | authority within 72 h (art. 33) | both |
| Access logs disclosed only on a court order | Marco Civil art. 10 | – | Marco Civil art. 10 |
| EU representative, if established outside the EU | – | art. 27 | art. 27 |
| `agent` · Processing terms signed with the model provider, transfer checked | art. 33, Res. CD/ANPD 19/2024 | arts. 28, 44–49 | both |
| `agent` · The chat UI tells visitors they are talking to an AI | – | AI Act art. 50(1) | AI Act art. 50(1) |

### Applying a preset
1. **Preview** (`preview-privacy-preset-page`): every setting with its current value, the preset's
   value and the reference. Equal values are greyed out. The preview also says what happens to stored
   data:
   - a shorter retention: *N contact requests, M views, K snapshots… will be deleted at the next
     cleanup*, counted now with the same rules as the cleanup (**§4**);
   - `full` → `hashed`: new IPs are hashed, and stored full IPs stay until their retention ends or until
     *Hash the IPs already stored* (**§10**);
   - `hashed` → `full`: only new IPs are stored in full, and stored hashes stay hashes.
2. **Apply**, confirmed with `data-confirm` when something will be deleted. Every value is validated
   (**§4**) before any is written. The notices that the new values need are shown together in one panel.
   The JSON twin answers the `409` of **§9** with a list, `"notices":[…]`, and expects all of them in
   `acknowledge`.
3. **Record**: `privacy-setting.preset` = `lgpd@1`, with `preset-applied-at` and `preset-applied-by`,
   and one `privacy-change` row with source `preset:lgpd@1` and the values before and after.

**Automation** (provisioning many servers) goes through the JSON twin with a root API token:
`apply-privacy-preset {"preset":"lgpd@1","acknowledge":["ip-full-v1"]}`. The acknowledgement is recorded
under the token's user, like any other.

### Active preset and drift
- A preset is a starting point, not a lock. A setting it covers can still be changed on its own page.
  When one differs, the Privacy page and a root banner say *2 settings differ from LGPD v1*, with the
  list, and offer **Re-apply** or **Detach**. Detaching makes the server *Custom* and is recorded.
- `GET /api/admin/root/check-privacy-preset` gives the state in one call, for an operator that watches
  many servers:
  `{"preset":"lgpd@1","conform":false,"differences":[{"setting":"ip-retention-days","preset":184,"current":90}],"checklist":{"done":5,"total":7},"encrypted":true,"pending-notices":[]}`

### Custom presets
A company's DPO usually has a policy of their own: a shorter retention, `hashed` with the reverse proxy
keeping the logs, more checklist items.
- **Save current settings as preset**: a name and a description. The checklist is copied from the active
  preset (`lgpd-gdpr` when there is none), and the root can add items (text and reference).
- **Export** (`export-privacy-preset`) downloads the JSON. **Import** (paste or upload) checks it against
  the format, the ranges (**§4**) and the known drafts, then stores it without applying it.
- Versions never change. Editing a preset saves `v2`, so the history and the drift check always point to
  fixed values. Built-in presets are read only.
- A custom preset that is not active can be deleted. Its values stay in the `privacy-change` rows of the
  history.
- A custom preset cannot add notices, settings or drafts. Those are code (**§9**, **§10**).

### Format
```json
{
  "format": 1,
  "id": "lgpd",
  "version": 1,
  "name": "LGPD — for-profit company in Brazil",
  "description": "…",
  "jurisdictions": ["BR"],
  "values": {
    "ip-mode":                {"value": "full", "reference": "Marco Civil art. 15 …"},
    "ip-retention-days":      {"value": 184,    "reference": "…"},
    "contact-retention-days": {"value": 180,    "reference": "…"},
    "view-retention-days":    {"value": 90,     "reference": "…"},
    "summary-retention-days": {"value": 90,     "reference": "…"},
    "backup-retention-days":  {"value": 90,     "reference": "…"},
    "notice-draft":           {"value": "br"}
  },
  "checklist": [{"id": "encarregado", "text": "…", "reference": "LGPD art. 41", "when": "always"}],
  "changes": "First version."
}
```
`when` is `always` or `agent`. Unknown keys are refused.

### New versions of the built-in presets
When the law or a regulator's reference changes (for example PL 2338/2023 or the EU data omnibus, **§1**),
a release ships a new version (`lgpd@2`) with `changes` filled in. A server on `lgpd@1` keeps its
values and shows the banner *LGPD v2 is available* with the differences. Nothing is applied
automatically: the law changes, and the root decides.

| Route name | Method and path |
|---|---|
| `privacy-presets-page` | `GET /admin/root/privacy-presets` |
| `preview-privacy-preset-page` | `GET /admin/root/preview-privacy-preset` |
| `apply-privacy-preset-form` | `POST /admin/root/apply-privacy-preset` |
| `save-privacy-preset-form` | `POST /admin/root/save-privacy-preset` |
| `import-privacy-preset-form` | `POST /admin/root/import-privacy-preset` |
| `export-privacy-preset` | `GET /admin/root/export-privacy-preset` |
| `delete-privacy-preset-form` | `POST /admin/root/delete-privacy-preset` |
| `detach-privacy-preset-form` | `POST /admin/root/detach-privacy-preset` |
| `set-privacy-checklist-form` | `POST /admin/root/set-privacy-checklist` |

JSON twins under `/api/admin/root/`: `list-privacy-presets`, `preview-privacy-preset`,
`apply-privacy-preset`, `save-privacy-preset`, `import-privacy-preset`, `export-privacy-preset`,
`delete-privacy-preset`, `detach-privacy-preset`, `check-privacy-preset`, `set-privacy-checklist`.

---

## 12. Database — `privacy-db` (key-prefix `privacydb`)
| Table | Fields |
|---|---|
| `privacy-setting` (one record) | `ip-mode` (`hashed`/`full`), `ip-key` (bytes, 32), `ip-key-rotated-at`, `ip-retention-days`, `backup-retention-days`, `notice-draft` (`br`/`eu`/`br-eu`), `preset` (`<id>@<version>`, blank = none), `preset-applied-at`, `preset-applied-by` |
| `privacy-acknowledgement` | `at`, `user-id`, `username`, `notice` (`contacts-v1`…), `values` (JSON of what was saved) |
| `privacy-change` | `at`, `user-id`, `username`, `source` (`manual`, `preset:<id>@<version>`, `detach`, `checklist`), `before` (JSON), `after` (JSON). Every privacy setting change writes one, from any page (**§2**) |
| `privacy-preset` (custom only, the built-ins are in `assets/`) | `preset-id`, `version`, `name`, `json` (the whole preset), `created-at`, `created-by` |
| `privacy-checklist` | `item`, `done`, `evidence`, `at`, `user-id`, `username` |

New fields: `contact-setting.retention-days` · `agent-setting.summaries-enabled`,
`agent-setting.summary-retention-days`.

---

## 13. Where it lives
| Path | What |
|---|---|
| `sandbox/internal/privacy/` | modes, IP key and HMAC, retention rules and cleanup, notices and acknowledgements, notice draft, find/erase, presets (load, validate, preview, apply, drift, checklist) |
| `sandbox/internal/encryption/` | key parsing, HKDF, value format, `encryption.json` check |
| Keep `encryptedstorage`, or `adapters/impls/<encrypted-keep>/` | the storage wrapper (**§3**) |
| `sandbox/internal/commands/middleware/database_encryption/` | key / `--notencrypt` in front of every command |
| `sandbox/internal/commands/middleware/privacy_start_server/` | the retention goroutine |
| `sandbox/internal/commands/…/generate_encrypt_key/`, `…/rotate_encrypt_key/` | the two new commands |
| `sandbox/internal/databases/privacy_db/` | **§12** |
| `assets/privacy/notices/*.md`, `assets/privacy/notice-drafts/{br,eu,br-eu}.md` | notice texts, draft templates |
| `assets/privacy/presets/{lgpd,gdpr,lgpd-gdpr}.json` | built-in presets (**§11**); a new version is a new file, the old one stays |
| `assets/backoffice/privacy_*.html`, `backofficerender/privacy_*.go`, `backofficeapi/privacy.go` | pages and JSON, presets included |
| `docs/Privacy/doc.md` | the doc, with the law tables of **§1** and the presets of **§11** |

Update `docs/ContactRequests`, `docs/Visualization`, `docs/Agent` and `docs/Backups`, and the README
quick start, which needs `EMISSARIO_ENCRYPT_KEY=$(emissario generate-encrypt-key)`. Store the key safely,
because losing it loses the data.

---

## 14. Acceptance
- `ls`/`cat` on `data/` shows no IP, email, contact body, summary or api key in clear. A test greps the
  data folder after inserting known values.
- Each row of the start-up table (**§3**) has a test, including the wrong key and both key and flag set.
- `git` started by the sync has no `EMISSARIO_ENCRYPT_KEY` in its environment.
- With the default settings: contact requests → `503 disabled`, summaries → `503 summaries-disabled`,
  views hold `h:` hashes and no user agent.
- The same IP always gives the same hash, also after a restart and after `rotate-encrypt-key` rotates
  `EMISSARIO_ENCRYPT_KEY`. After *Rotate IP key* it gives another. Two IPv6 addresses in one `/64` give
  one hash. `ip-key` never appears in clear in an encrypted `data/`.
- An IP hash older than the IP retention is blanked by the cleanup.
- `set-contact-settings {enabled:true}` without `acknowledge` → `409`. With it → `200` and one
  `privacy-acknowledgement` row.
- Retention: a record past retention is gone after one cleanup run. A full IP past the IP retention is
  blanked and its record kept.
- A viewer never receives a full IP from any page or API.
- An encrypted backup restores on the same key and is refused on another.
- Presets:
  - `apply-privacy-preset {"preset":"lgpd@1"}` without `acknowledge` → `409` listing `ip-full-v1`, and
    nothing is written. With it → `200`, every setting equals the preset, `privacy-setting.preset` is
    `lgpd@1`, and there is one `privacy-change` row and one `privacy-acknowledgement` row.
  - `gdpr@1` on a fresh install needs no notice.
  - Applying never changes `contact-setting.enabled`, `agent-setting.enabled` or
    `agent-setting.summaries-enabled`.
  - An imported preset with one value out of range, a missing setting or an unknown key is refused, and
    nothing is written.
  - Changing `ip-retention-days` by hand after applying → `check-privacy-preset` gives `conform:false`
    with that difference. *Re-apply* → `conform:true`. *Detach* → `preset` blank and a `privacy-change`
    row.
  - The counts in the preview equal what the next cleanup deletes.
  - Export on one server, import and apply on another → the same settings on both.
  - A built-in preset cannot be edited or deleted, a saved version cannot be changed in place, and the
    active preset cannot be deleted.

## Important
- Every yaml through agnos commands (`add-database`, `add-table-field`, `add-route`,
  `add-command --middleware`, `add-adapter`). No generated file is edited by hand. Finish with
  `agnos build`.
- Never log a decrypted value, a key, an IP or a contact body. Errors name the field, not the value.
- Suggested order: encryption → `privacy-db` and notices → IP modes → contacts → agent and summaries →
  retention and backups → Privacy page → presets → docs.

## Out of scope
- A cookie banner or consent manager: Emissario sets no cookie on visitors.
- A self-service portal for visitors' requests: the root answers them through **§10**.
- A blocklist by IP hash (*Block* on a visitor, contact request or usage row → `403` for that hash). The
  stable hash of **§6** makes it possible later.
- KMS, HSM or a key server: the key comes from the environment.
- Choosing a preset automatically by country: a root chooses it (**§11**).
- Built-in presets, notices and drafts for laws other than BR/EU (UK GDPR, US state laws). A custom
  preset can hold their values, but not their notices.
- Locking a server to a preset (refusing changes that leave it). The drift check and
  `check-privacy-preset` show those changes instead of blocking them.
- Legal advice: the notices inform, and the operator decides.
