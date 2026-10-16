# Privacy (LGPD / GDPR Adapt)

## Objective
Make Emissario **private by default** and give the root what they need to run it lawfully under the
LGPD (Brazil), the Marco Civil da Internet, the GDPR (EU) and the laws next to them: the database
encrypted at rest, IPs hashed in the records unless a root asks for them in full, a separate **access
log** with full IPs and source ports that a root turns on (as the Marco Civil requires of Brazilian
companies) and that only named readers open, contact requests and agent summaries off until a root turns
them on, every piece of personal data deleted when its retention ends, data kept by law or on a legal
hold spared from erasure unless a root chooses otherwise, and every risky choice preceded by a notice
that names the law.

Emissario is open source and enterprise first. The defaults are the private and lawful choice, and every
other lawful choice stays open to the root: Emissario warns and lets them do it. It keeps no record of who
chose what, and it never builds evidence against its own user.

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
| Hashing IPs by default is always the safe side | In Brazil, Marco Civil da Internet (Lei 12.965/2014) **art. 15** *requires* a company that runs an internet application for profit to keep access logs (IP + date/time) for **6 months**, confidential, in a controlled and secure environment, and disclosed only on a court order (art. 10). Decreto 8.771/2016 **art. 13** says how: access limited to named people, an authentication that individualises who reads (two-factor is its example), an **inventory of every access** to the logs, and encryption. Using these logs for statistics is another purpose (LGPD art. 6 I–III). | Records hold the hash by default, and a root may keep full IPs in them (`full`, **§6**). A separate **access log** keeps the full IP, the source port and the UTC time for 184 days when a root turns it on, readable only by named readers with two-factor sign-in and the inventory of reads the Decreto asks for (**§6**). The `lgpd` preset turns it on (**§11**). |
| An access log is the IP with its date and time | Behind carrier-grade NAT, common in Brazil, many customers share one IPv4 address. The STJ (3rd Panel, 2019, *TIM v. Google*) ordered an application provider to give the **source port** with the IP, and court orders ask for IP, port, date, time and time zone. | The access log stores the source port and the time in UTC (**§6**). |
| A deletion request deletes everything | Data kept to meet a legal obligation is not erased on request (LGPD art. 16 I; GDPR art. 17(3)(b)), nor data kept for a legal claim (LGPD art. 7 VI; GDPR art. 17(3)(e)). The police, an administrative authority or the Public Prosecutor may ask for access logs to be kept longer than 6 months (Marco Civil art. 15 §2). | By default *Find and erase* keeps the access log during its term and the rows on a **legal hold**, and the cleanup skips held rows. A root can still delete the access log rows, behind a notice that names the duty (**§4**, **§10**). |
| Encrypted data needs no notification | GDPR art. 34(3)(a) spares telling *the people* when the data was unintelligible to the intruder. LGPD art. 48 §3 only weighs encryption when judging how serious the incident is. Res. CD/ANPD 15/2024 asks to tell the ANPD **and the data subjects** within 3 business days when the incident may bring relevant risk or harm, and to **record every incident, even one not communicated, for 5 years** (art. 10). GDPR art. 33(5) also documents every breach. | **§3** and the notices say so, and the checklist asks for an incident register (**§11**). |
| Summaries: "AI can make mistakes" | That is true, and two more facts apply. The whole conversation goes to the model provider, which is a processor and often an international transfer. Since **2 Aug 2026** the EU AI Act (Reg. 2024/1689) **art. 50(1)** requires that people be told when they are talking to an AI. The AI omnibus (Reg. (EU) 2026/1744, in force since 27 Jul 2026) postponed the high-risk rules, not this one, and the free and open-source exemption (art. 2(12)) does not cover art. 50. | The summary notice and a new agent notice cover both, and the agent says it is an AI by default (**§8**). |
| Encrypt "on start server" | `add-backoffice-user` and every other command that opens a database writes to it too. | The key is required by **every command that touches a database** (**§3**). |

### Laws and rules referenced

| Law | Articles used here |
|---|---|
| **LGPD** — Lei 13.709/2018 (BR) | 3 (scope: data collected in Brazil or offer to people in Brazil) · 5 I, II (sensitive data), III, VI, VII (operador) · 6 I (purpose), III (necessity), VI (transparency), VII (security), X (accountability) · 7 I, II, V, VI, IX (legal bases) · 9 (information to the data subject) · 10 (legitimate interest: balancing, transparency; §3 impact report on request) · 11 (sensitive data) · 12 (anonymised data) · 14 (children and adolescents) · 15–16 (end of processing, deletion; 16 I keeping for a legal obligation) · 18 (rights) · 19 (a full answer within 15 days) · 33 (international transfer) · 37 (records of processing) · 38 (impact report, RIPD) · 39, 42 (operador: follows the controller's instructions, jointly liable) · 41 (encarregado) · 46 §2 (security from the design phase) · 48 (incidents; §3 encryption weighs on severity) · 52 (fines: up to 2% of revenue in Brazil, capped at R$ 50 million per infraction) |
| **ANPD resolutions** | CD/ANPD 2/2022 (small-scale agents: simplified records, no encarregado required but a contact channel required; a large company is not one) · CD/ANPD 15/2024 (incidents: the ANPD **and the data subjects** within **3 business days** when there is relevant risk or harm; every incident recorded for **5 years**) · CD/ANPD 18/2024 (the encarregado, their identity and contact published) · CD/ANPD 19/2024 (international transfers, standard contractual clauses) · CD/ANPD 32/2026 (the EU recognised as adequate) · Legitimate interest guide (2024) |
| **Marco Civil da Internet** — Lei 12.965/2014 (BR) | 5 VIII (access log = date/time of use from an IP) · 10 (confidentiality, disclosure only on a court order) · 13 §§3–4 (a preservation request is confidential, and lapses unless a court is asked within 60 days) · 15 (6 months of access logs for for-profit legal entities; §2 an authority may ask to keep them longer) · 16 II (no data beyond the purpose) |
| **Decreto 8.771/2016** (BR, regulates the Marco Civil) | 13 I (strict access control, exclusive privileges for named users) · 13 II (an authentication that individualises who reads the logs, two-factor as the example) · 13 III (inventory of every access to the logs: when, how long, who, what) · 13 IV (encryption or equivalent) · 13 §2 (delete as soon as the purpose is met or the legal term ends) |
| **STJ** (BR) | 3rd Panel, 2019 (*TIM v. Google*): an application provider gives the **source port** with an IPv4 address, because carrier-grade NAT shares one address among many customers |
| **GDPR** — Reg. (EU) 2016/679 | 3(2) (applies to a non-EU site that targets people in the EU) · 4(1), 4(5) · 5(1)(b)(c)(e)(f), 5(2) · 6(1)(a)(b)(c)(f) · 8 (children) · 9 (special categories) · 11 (no duty to re-identify) · 12(3) (answer within one month) · 13 (information at collection) · 15–21 (rights; 17(3)(b)(e): no erasure of data kept for a legal obligation or a legal claim) · 25 (data protection by design and by default) · 27 (EU representative) · 28 (processor contract) · 30 (records) · 32 (security, encryption) · 33 (72 h breach notification; 33(5) every breach documented) · 34(3)(a) (no need to tell people when the data was encrypted) · 35 (impact assessment) · 37–39 (DPO, when required) · 44–49 (transfers; 45 adequacy) · 83 (fines up to €20 million or 4%) |
| **CJEU** | C-582/14 *Breyer* (2016): a dynamic IP is personal data for a website operator · C-413/23 P *EDPS v SRB* (2025): pseudonymised data stays personal for whoever can re-identify it |
| **ePrivacy Directive** 2002/58/EC art. 5(3) | Emissario writes no cookie and runs no script on visitors' devices, so its analytics needs no consent banner. Art. 5(3) covers any storage on the device, `localStorage` included, not only cookies: a chat history the website keeps in the browser is strictly necessary for the service the visitor asked for, and the notice draft names it (**§10**). |
| **EU AI Act** — Reg. (EU) 2024/1689 | art. 50(1): tell people they are interacting with an AI, applicable since 2 Aug 2026 · art. 2(12): the free and open-source exemption does not cover art. 50 · art. 99(4): fines up to €15 million or 3% · the AI omnibus, Reg. (EU) 2026/1744 (in force 27 Jul 2026), postponed the high-risk rules and left art. 50(1) as it was · Commission guidelines on art. 50 (July 2026) |
| **EU–US Data Privacy Framework** | Adequacy decision of July 2023. The General Court upheld it in *Latombe* (T-553/23, 3 Sep 2025), and the appeal C-703/25 P is pending. It covers only US providers that are certified. |
| **EU–Brazil adequacy** | Commission decision C(2026) 373 of 26 Jan 2026 (GDPR art. 45: Brazil is adequate) and Res. CD/ANPD 32/2026 (LGPD art. 33 I: the EU is adequate). Personal data flows between Brazil and the EU without standard clauses. Reviewed every 4 years. |
| **ECA Digital** — Lei 15.211/2025 (BR) | In force since 17 Mar 2026, for IT products and services aimed at children and adolescents in Brazil or likely to be accessed by them. With LGPD art. 14, it reaches a website likely to be used by minors. |
| **Regulators' references** (not law) | CNIL rec. 2021-122: security logs kept **6 months to 1 year**, up to 3 years only with a documented need · CNIL audience measurement: statistics kept **≤ 25 months** · CNIL commercial-management reference: prospect data **≤ 3 years** after the last contact |
| **Pending, watch** | Brazil's AI bill **PL 2338/2023** passed the Senate in December 2024 and is pending in the Chamber, its vote put off until after the October 2026 elections. It is not law. The EU **data omnibus** (GDPR and cookie changes) is still being negotiated. Do not build on either. |

---

## 2. Principles
- **Private by default** (GDPR art. 25; LGPD art. 46 §2): an install stores no contact request, no
  summary and no full IP.
- **Freedom**: the root is the controller and decides. Every lawful choice a default rules out (full
  IPs, full IPs for viewers, a longer or endless retention, the access log without two-factor or without
  its inventory, erasing rows kept by law) stays possible behind a notice that names the law. Emissario
  warns, and never forbids. The only hard stops keep data from being lost or exposed without anyone
  choosing it: a missing or wrong key (**§3**) and a notice not confirmed.
- **Fail closed**: a missing or wrong key stops the program, and an unacknowledged notice leaves the
  feature off. Nothing silently downgrades.
- **No backward compatibility**: Emissario is in alpha. A database from an earlier version is not
  migrated: start a new one. Every install starts with the defaults of **§5**.
- **Protection of the user, and no evidence against them**: Emissario gives the controller the tools to
  comply (private defaults, notices, presets, the access log, erasure, holds) and keeps nothing that could
  be held against them: no record of who changed a setting, applied a preset, held or erased data, no
  record that a notice was shown or confirmed, no compliance status. It keeps only the state it needs to
  work: the current settings, the active holds, and the ids of erased rows until no backup could bring
  them back. Demonstrating compliance (LGPD art. 6 X; GDPR art. 5(2)) stays with the controller's own
  documents, kept where they choose. The one log of people is the inventory of access log reads that
  Decreto 8.771 art. 13 III asks of Brazilian companies, and a root can turn it off (**§6**).
- **The law outranks erasure, by default**: data kept to meet a legal obligation (the access log,
  **§6**) or on a legal hold is not erased on request nor by the cleanup before its term (LGPD art. 16 I;
  GDPR art. 17(3)(b)(e)), and it is deleted as soon as the term ends (Decreto 8.771 art. 13 §2). A root
  who knows the duty does not apply to them deletes it behind a notice (**§9**).
- **Need to know, by default**: records hold IP hashes, and in `full` mode viewers see IPs masked. A
  root may show full IPs to viewers, behind a notice (**§6**). Today contact requests show the whole IP
  to viewers, and this plan changes that default.

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
- The **controller** generates the key and keeps it, for example in its cloud's secret manager, injected
  as the environment variable. Whoever installs, hosts or supports Emissario for them and can reach its
  data is their **operador** (LGPD arts. 5 VII, 39 and 42; GDPR art. 28): a contract sets the
  instructions, and they keep no copy of the key once the work is delivered. `docs/Privacy/doc.md` says
  so.

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

`encryption.json` holds `{"version":1,"algorithm":"XChaCha20-Poly1305","key_id":"<hex of the first 4 bytes of HMAC(key-check, "emissario")>"}`.
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
- **Format of an encrypted value:** `0xE1 | key_id (4 B, the one of encryption.json) | nonce (24 B, random) | XChaCha20-Poly1305 ciphertext+tag`.
  The AAD is the storage key segments joined, so a value moved to another field or record fails to
  decrypt.
- **Why XChaCha20-Poly1305 and not AES-256-GCM:** GCM with random 96-bit nonces is limited to 2³²
  encryptions per key (NIST SP 800-38D §8.3). Every view writes several values and `Append` re-encrypts
  the whole value, so a busy website gets there, and one repeated nonce breaks both confidentiality and
  integrity. A random 192-bit nonce has no practical limit (`golang.org/x/crypto/chacha20poly1305`,
  `NewX`).
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
It protects a stolen disk, a stolen backup, or a misconfigured volume, and it is the measure named in
GDPR art. 32(1)(a) and LGPD art. 46. When the data was unintelligible to the intruder, GDPR art.
34(3)(a) spares telling the people, and LGPD art. 48 §3 counts it when judging how serious the incident
is. It does not make the incident disappear: it is still recorded for 5 years (Res. CD/ANPD 15/2024
art. 10; GDPR art. 33(5)), and reported when it may still bring relevant risk. It does **not** protect
against an attacker inside the running server, which holds the key in memory. Say this in the doc.

---

## 4. Retention: defaults and thresholds

| Data | Default | Warning (needs the notice) | Reference for the threshold |
|---|---|---|---|
| Contact request | **180 days** | > 365 or blank | No legal number. Keep it while you answer it and while the relationship lasts (LGPD 15–16; GDPR 5(1)(e)). CNIL: prospects ≤ 3 years after the last contact. 5 years ≈ the Brazilian consumer limitation period (CDC art. 27), as the outer bound for a legal defence (LGPD art. 7 VI) |
| Access log, when on (`access-log-retention-days`) | **184 days** (the longest 6-month span) | < 184 (BR), > 365 (EU) or blank | Marco Civil art. 15: at least 6 months for BR for-profit companies, then deleted (Decreto 8.771 art. 13 §2). CNIL 2021-122: 6 months–1 year, up to 3 years only with a documented need |
| IP in records, hash or full (`ip-retention-days`) | **184 days** | > 365 or blank | CNIL 2021-122 as above. A hash is still personal data (**§1**), so it has a limit like any other |
| Visualization view | 90 (as today) | > 760 or blank | CNIL audience measurement: ≤ 25 months |
| Agent summary | **90 days** | > 365 or blank | as contact requests |
| Backup snapshot | blank (kept) | blank, or longer than the longest retention above | Data that retention removes from the database lives on in the snapshots taken before |

- An `ip` field older than the IP retention is **blanked**, hash or full IP. The record stays until its
  own retention ends: a contact request kept 180 days with an IP retention of 30 days loses its IP on
  day 30.
- The access log has its own retention, apart from the records it points to: with the `lgpd` preset a
  view is gone at 90 days and its access log row stays until day 184.
- Rows on a **legal hold** (**§10**) are skipped until the hold ends or is released, then follow their
  retention again.
- Retention is counted from `at`, in days of the server's time zone.
- A retention is a number of days ≥ 1, or blank for forever. Below 1 → `400`. Above the warning, or
  blank, → the notice of **§9** with the reference text. No retention is refused for being long (**§2**).

### Retention cleanup
A goroutine in `privacy-start-server` runs every hour. The visualization cleanup that runs today moves
into it. It removes or blanks, in each table: `contact-request`, `visualization-view`, `agent-summary`,
`access-log`, and the `ip` of `agent-usage`, `contact-request` and `visualization-view`. It skips rows on a
legal hold. It holds while a backup is taken or restored, as the writers do. Each run logs only counts to
stderr, never values.

**Backups:** the Backups page lists the snapshots that hold data older than the current retentions,
with a delete button. If a backup retention is set, snapshots older than it are removed by the same
goroutine.

---

## 5. Defaults

| Thing | Default |
|---|---|
| Encryption | key required, or `--notencrypt` for debugging (**§3**) |
| `POST /api/contact-requests` | **off** → `503 disabled` |
| IP mode | **`hashed`** (**§6**) |
| Full IPs for viewers | masked |
| Access log | **off**. When on: two-factor for its readers and the inventory of reads |
| `POST /api/agent/summaries` | **off** → `503 summaries-disabled` |
| Agent | off (as today) |
| Full user agent in views | stored only in `full` mode |

---

## 6. IPs

### Modes — `privacy-setting.ip-mode`
| Mode | Stored in `visualization-view.ip`, `contact-request.ip`, `agent-usage.ip` |
|---|---|
| `hashed` (default) | `h:` + hex of the first 16 bytes of `HMAC-SHA256(ip-key, ip)`. IPv4 whole, IPv6 by its `/64`, the same grouping as the rate limits, so an attacker cannot dodge a match by rotating inside their `/64`. Blanked after the IP retention (**§4**) |
| `full` (notice `ip-full-v1`) | the IP as it came, IPv6 whole, and the full `user-agent` in views. Blanked after the IP retention (**§4**) |

A mode applies to new rows. *Hash the IPs already stored* (Privacy page, danger zone) turns the full IPs
kept into hashes. `full` does not replace the access log below: it has no source port and goes with its
record.

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
- **Done before hashing, on the raw IP, in memory:** the country lookup, the ignored IPs, the rate
  limits per minute and per day (in memory, as today), and the access log row when it is on.
- **What `hashed` gives and changes, documented in Visualization:**
  - Unique visitors are exact over any period within the retention.
  - The Visitors page groups by hash, with country, device, browser, os, views and first/last seen as
    today.
  - The `ip` filter takes a typed IP and filters by its hash over the whole retention. A CIDR filter is
    refused in `hashed` mode, because a range cannot be hashed. It works in `full` mode and in the
    access log.
  - `agent-usage` per-IP daily limits keep working.
  - The full `user-agent` is not stored, only device, browser and os: with a stable hash, the full
    string adds fingerprinting without serving any feature. `full` mode stores it.
- **Blocking:** the hash is stable, so an abuser seen in Visitors, a contact request or agent usage can be
  matched again later. The blocklist is out of scope here (see **Out of scope**). Volumetric DDoS is
  stopped before Emissario (reverse proxy, firewall, CDN), not by a lookup per request inside it.
- **Who sees full IPs** in `full` mode: roots. Viewers see them masked (`203.0.113.xxx`, IPv6 `/48`) on
  every page and API, unless a root sets `privacy-setting.viewer-ip` to `full` (notice `viewer-ip-v1`).
  A hash shows as `h:3fa9…` to everyone.
- Backoffice users' own IPs (`api-token.last-used-ip`, `ips`) are a security log of staff, not of
  visitors. They are not hashed, since they serve to spot a stolen token, and are encrypted with the
  rest.

### Access log — Marco Civil (`privacy-setting.access-log`, off by default)
When a root turns it on (notice `access-log-v1`, **§9**), every page view the Visualization records
(before the ignored IPs, which filter statistics, not the log) and every call to the public APIs
(`/api/contact-requests`, `/api/agent/*`) writes one `access-log` row (**§12**):

| Field | |
|---|---|
| `at` | UTC, with milliseconds. Pages show it in UTC and in the server's time zone: court orders ask for the date, the time and the time zone |
| `ip` | the full address, IPv6 whole |
| `port` | the **source port** of the visitor's connection (STJ 2019, **§1**) |
| `kind`, `ref` | `view`, `contact-request` or `agent`, and the id of the record the request wrote, so *who sent contact request 123* has an answer |

- **Source port:** read from the connection. Behind a reverse proxy (`--allow-x-forwarded-for`), the
  new flag `--client-port-header <name>` names the header the proxy sets
  (nginx: `proxy_set_header X-Client-Port $remote_port;`), trusted under the same rule as
  X-Forwarded-For. When the port cannot be read, the row keeps it blank and the Privacy page warns
  *the access log has no source port: set --client-port-header*. A load balancer that does not forward
  it keeps it in its own logs, which then follow the same 6 months and the same rules.
- **Clock:** the doc asks for the server's clock synced by NTP.
- **Who reads it** (Decreto 8.771 art. 13):
  - **Named readers (I):** only roots that a root marks `access-log-reader`. Viewers and other roots
    never see it.
  - **Two-factor sign-in (II):** by default a reader must have TOTP on (RFC 6238, 6 digits, 30 s), set
    up on their own account page from a QR code drawn by the server and the base32 secret. Sign-in then
    asks for the code. Turning TOTP on revokes the user's API tokens, so every token of a reader was made
    after a two-factor sign-in. A root resets another user's TOTP, and
    `emissario reset-backoffice-totp --username <u>` does it from the host. A root banner lists the roots
    without TOTP. A root may drop the requirement (`privacy-setting.access-log-totp` = `0`, notice
    `access-log-no-totp-v1`), for example behind a proxy that already asks for single sign-on.
  - **Inventory of accesses (III):** by default every list, search or export of the access log, from a
    page or the API, writes a `privacy-access` row: when it started and ended, who, what
    (`access-log-list`, `access-log-export`…), the filter and how many rows came back. The Decreto asks
    for exactly this, the identity of who read included. It is the only log of people in this plan, it
    covers the access log alone, and it stays in the client's own encrypted database. A root may turn it
    off (`privacy-setting.access-log-inventory` = `0`, notice `access-log-no-inventory-v1`).
  - **Encryption (IV):** the storage layer (**§3**). With `--notencrypt` the debug banner says the access
    log is in clear.
- **Page** `/admin/root/access-log` (readers): filter by IP (exact or CIDR), period and `kind`, **Export**
  (JSON, CSV) to answer a court order, **Hold** the rows shown (**§10**) and **Delete** them. The page
  says *disclose only on a court order (Marco Civil art. 10 §1)*.
- **Kept during its term, by default:** *Find and erase* leaves a row younger than the access log
  retention (LGPD art. 16 I; GDPR art. 17(3)(b)). A root deletes such rows on purpose, from this page
  or from *Find and erase*, with the notice `erase-kept-v1`: the duty may not apply to them (an EU or
  non-profit website). A held row is deleted only after its hold is released. The cleanup deletes rows
  when the retention ends, unless held (**§4**).
- **Turning it off** stops new rows. The rows already kept stay until their retention ends, because
  they were kept under the obligation.
- **Brazil note on the Privacy page, while it is off:** *If this website is run by a company for profit,
  Marco Civil art. 15 requires access logs (IP, source port, date and time) kept 6 months. Turn on the
  access log, or keep your reverse proxy's logs for 6 months under the same rules (Decreto 8.771 art.
  13).*

---

## 7. Contact requests
- `contact-setting` gains `retention-days` (**§4**). Blank keeps them forever, with the retention notice.
- Turning `enabled` on requires the **contacts** notice (**§9**). Saving a retention above 365 requires
  the retention notice.
- The settings page shows: *Your website must say in its privacy notice what the form keeps, why, for how
  long, and how to ask for deletion*, with a link to the draft (**§10**).
- A contact request keeps the IP as the IP mode says, a hash by default (**§6**). With the access log
  on, a reader finds its full IP and source port through the `ref` of its access log row.
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
  party: choose terms under which the provider does not use the conversations to train its models or for
  its own purposes*.
- The embedded `assets/agent/system.md` tells the model to say it is an AI in its first answer and
  whenever asked, and never to claim to be a person. The disclosure does not rest on the website alone:
  AI Act art. 50(1) binds whoever designs the system, open source included (art. 2(12)).
- `docs/Agent/doc.md` says that the website's chat UI must also tell visitors they are talking to an AI
  before the chat starts (AI Act art. 50(1)), and its chat example shows the line above the chat box
  (*You are chatting with an AI assistant.*).

---

## 9. Notices (the "popups")

The backoffice CSP runs no inline script, so the notice is **rendered by the server**:

1. The root saves a setting, or runs an action, that needs a notice.
2. The page comes back with the notice panel: the text, the laws, a required checkbox *I have read this
   notice*, and **Confirm**.
3. The server, not the browser, checks the checkbox. A JSON client cannot skip it:
   `409 {"error":"acknowledgement-required","field":"acknowledge","notice":{"id":"contacts-v1","text":"…"}}`
   until it sends `"acknowledge": ["contacts-v1"]`.
4. The server keeps nothing about it: not who confirmed, not when, not that the notice was shown. The
   notice comes back every time the setting is turned to that value again, and at every use of an
   action (`erase-kept-v1`).

Notices live in `assets/privacy/notices/<id>.md`, versioned in their id (`contacts-v1`). The Privacy page
shows, next to each setting in use, its current notice to read again. A new version is shown there and
asks nothing.

### `contacts-v1`
> Turning on contact requests makes this server store what visitors send through your form (names,
> emails, phones, messages) with its time and the visitor's IP, hashed unless you chose `full`. **You are
> the controller of this data.** That is lawful as long as you:
> - **tell visitors** in your privacy notice what is kept, why, for how long and how to ask for its
>   deletion (LGPD art. 9; GDPR art. 13);
> - have a **legal basis**: steps the visitor asked for before a contract (LGPD art. 7 V; GDPR art.
>   6(1)(b)), or your legitimate interest in answering other messages (LGPD art. 7 IX; GDPR art.
>   6(1)(f)), and do not use it for marketing without consent;
> - watch for **sensitive data** (health, religion, biometrics…) and **minors**: a form that may receive
>   them needs the stricter rules of LGPD arts. 11 and 14 and GDPR arts. 8 and 9, and a website likely
>   used by minors falls under the ECA Digital (Lei 15.211/2025);
> - keep it **only while needed**: Emissario deletes requests older than the retention you set (LGPD arts.
>   15–16; GDPR art. 5(1)(e));
> - **answer** access and deletion requests (LGPD arts. 18–19, a full answer within 15 days; GDPR arts.
>   12(3) and 15–17, within one month);
> - **record every incident** for 5 years, and **report** one that may bring relevant risk to the ANPD
>   and to the people affected within 3 business days (Res. CD/ANPD 15/2024), and to the EU authority
>   within 72 h (GDPR art. 33);
> - name an **encarregado** and publish their identity and contact (LGPD art. 41; Res. CD/ANPD
>   18/2024; only small-scale agents are spared, Res. CD/ANPD 2/2022), and appoint an **EU
>   representative** if you are outside the EU and target people there (GDPR art. 27).
>
> Fines: LGPD up to 2% of revenue in Brazil, capped at R$ 50 million per infraction (art. 52). GDPR up to
> €20 million or 4% of worldwide turnover (art. 83).

### `access-log-v1`
> The access log keeps every visitor's full IP, source port and time. A dynamic IP is personal data
> (CJEU C-582/14 *Breyer*; LGPD art. 5 I).
> - **Brazil:** a company running this website for profit **must** keep it 6 months (Marco Civil art.
>   15), and that **legal obligation** is its legal basis (LGPD art. 7 II). Keep it confidential,
>   disclose it only on a court order (art. 10), and use it for nothing else: statistics and blocking
>   use the records (LGPD art. 6 I–III). An authority may ask you to keep rows longer (art. 15 §2): put
>   them on hold, and keep the request confidential (art. 13 §4).
> - **EU:** there is no general duty to keep IPs. Turn it on only for a security need, under a
>   legitimate interest with a documented balancing test (GDPR art. 6(1)(f)), within the CNIL's 6
>   months to 1 year.
>
> By default Emissario applies Decreto 8.771/2016 art. 13 for you: named readers, two-factor sign-in, an
> inventory of reads, encryption. During its term a row is kept even when its person asks for erasure
> (LGPD art. 16 I; GDPR art. 17(3)(b)).

### `ip-full-v1`
> Full IPs identify visitors. A dynamic IP is personal data (CJEU C-582/14 *Breyer*; LGPD art. 5 I).
> Keeping them in views, contact requests and agent usage, with the full user agent of each view, needs
> a legal basis, usually a legitimate interest in security, fraud prevention or statistics with a
> documented balancing test (LGPD art. 10; GDPR art. 6(1)(f)), the shortest retention that serves it
> (LGPD art. 6 III; GDPR art. 5(1)(c)), and a line in your privacy notice.
> - A full IP with its time is also an access log (Marco Civil art. 5 VIII): keep it confidential and
>   disclose it only on a court order (art. 10).
> - It does not replace the access log of Marco Civil art. 15: it has no source port and goes with its
>   record. A for-profit company in Brazil still turns the access log on.
> - Viewers see these IPs masked unless you choose otherwise.

### `viewer-ip-v1`
> Viewers will see visitors' full IPs on every page and API. The laws ask you to give personal data only
> to whoever needs it for their job, and to keep it from an indefinite number of people by default (LGPD
> arts. 6 III and 46; GDPR arts. 5(1)(c), 25(2) and 32(4)).

### `access-log-no-totp-v1`
> Decreto 8.771/2016 art. 13 II asks for an authentication that individualises who reads the access
> log, and names two-factor as its example. Without it, one stolen password opens every visitor's IP.
> Use another measure that does the same, such as single sign-on with two factors at your proxy.

### `access-log-no-inventory-v1`
> Decreto 8.771/2016 art. 13 III asks a company that keeps Marco Civil access logs for a detailed
> inventory of every access to them: when, for how long, who and what. Without this one, keep that
> inventory by other means, because the duty stays.

### `erase-kept-v1` (asked at every use)
> You are deleting access log rows inside their term. A company that runs this website for profit in
> Brazil must keep them 6 months (Marco Civil art. 15), and the laws do not require erasing them on a
> person's request (LGPD art. 16 I; GDPR art. 17(3)(b)). Delete them when that duty does not apply to
> you, such as an EU or non-profit website.

### `summaries-v1`
> Summaries are written by an AI model from the visitor's conversation and stored here. The model is
> told to remove names, contacts, documents, places and dates, and a filter then removes CPF, CNPJ,
> emails, phones and card numbers. **AI makes mistakes.** A summary can still hold personal data, or
> details that identify someone when read together (LGPD art. 12; GDPR recital 26). Treat summaries as
> personal data: they are encrypted, kept for the retention you set, and deleted after it. To write
> one, the **whole conversation is sent to your model provider** (see the agent notice). Tell visitors
> about summaries in your privacy notice. *Find and erase* searches summaries too (**§10**).

### `agent-v1`
> Every message a visitor sends goes to the model provider you configured (`<host>`). That provider
> processes it on your behalf: sign its data processing terms (GDPR art. 28), under which it does **not
> use the conversations to train its models** or for its own purposes. Otherwise it is not your
> processor but another controller, and you are sharing the data with it. A provider in another country
> is an international transfer. Between Brazil and the EU, adequacy covers it (Res. CD/ANPD 32/2026;
> Commission decision of 26 Jan 2026). Elsewhere, use the standard clauses of Res. CD/ANPD 19/2024
> (LGPD art. 33) and GDPR arts. 44–49 (the EU–US Data Privacy Framework covers certified US providers
> only). Visitors may type **sensitive data**, like a clinic's patients writing about their health:
> your notice and your legal basis must cover it (LGPD art. 11; GDPR art. 9). A local model (Ollama)
> keeps conversations on your infrastructure. The agent says it is an AI in its first answer, and **your
> chat UI must also tell visitors they are talking to an AI** (EU AI Act art. 50(1), applicable since 2
> August 2026).

### `retention-<data>-v1`
Generated from **§4**: *You set N days. Above T days, the references below ask for a documented reason:
…*. It lists that row's reference and the threshold that was crossed.

---

## 10. Privacy page — `/admin/root/set-privacy-settings` (root)
Linked from the top bar (root) and from the home page.

- **Preset** (**§11**), first on the page: *Preset: LGPD v1*, or *No
  preset (install defaults)*. It lists the settings that differ from it and links to the Presets page.
- **Encryption** (read only): on/off, `key_id`, algorithm.
- **IP mode** (`hashed`/`full`), **full IPs for viewers** (masked/full), **IP retention**, and the
  **access log** (on/off, retention, readers, two-factor, inventory, whether the source port is read),
  with the Brazil note (**§6**). Danger zone, each irreversible and confirmed with `data-confirm`: *Hash
  the IPs already stored* · *Rotate IP key*.
- **What this server keeps**: one row per table holding personal data, with the fields, the purpose, a
  suggested legal basis, the retention, the count and the oldest record. This is the base of the
  records of processing (LGPD art. 37; GDPR art. 30).
- **Privacy notice draft**: text built from the current settings. It covers what is kept, why, for how
  long, the access log and its legal obligation, the model provider's host, that no cookie is set and
  the chat history stays in the visitor's browser (`localStorage`), the rights and the time to answer
  them, and placeholders for the controller, the contact and the encarregado/DPO. The root copies it
  into the website. It is marked *draft — review it*. Its template is `privacy-setting.notice-draft`: `br` (pt-BR: controlador,
  encarregado, ANPD, Marco Civil logs), `eu` (en: controller, DPO, EU representative, supervisory
  authority) or `br-eu` (both, the default). A preset sets it (**§11**).
- **Find and erase a person** (LGPD arts. 18–19; GDPR arts. 15, 17): an IP (hashed with the current IP
  key in `hashed` mode) or a text (email, phone, name). It shows the matches in contact requests, views, agent usage and
  agent summaries, with **Export JSON** and **Delete all** (confirmed). Hashes made before a *Rotate IP
  key* no longer match, and the page says so (GDPR art. 11).
  - It also lists the matching access log rows (to readers) and the held rows, marked *kept*. *Delete
    all* leaves them by default: the access log for its legal obligation (LGPD art. 16 I; GDPR art.
    17(3)(b)), a held row until its hold ends. Ticking *also delete the access log rows* deletes those
    too, with the notice `erase-kept-v1`. A held row needs its hold released first. The page gives no
    hold's reason, because a preservation request is confidential (Marco Civil art. 13 §4).
  - An erasure of contact requests, views, agent usage or summaries keeps the table and the ids deleted
    in `privacy-erasure`, with no personal data and no user, so that a restore can replay it. The entry
    is deleted once no snapshot older than it remains. Deleting access log rows keeps nothing: a restore
    never rolls the access log back (**§12**).
  - **Backups:** an erased row lives on in the snapshots taken before, until the backup retention
    removes them (**§4**), and the page says so. Restoring a snapshot re-applies every erasure kept
    after it was taken, so erased data does not come back (**§12**).
- **Legal holds**: a root holds rows selected in contact requests or agent summaries, and a reader in
  the access log, with a reason (*Ofício nº …*, a lawsuit) and an end date. For a Marco Civil art. 15
  §2 request the end date defaults to 60 days, the time the authority has to go to court (art. 13 §3),
  and a court order extends it. Held rows are skipped by the cleanup and by *Find and erase*. A hold is
  released by hand or on its end date.
- **Notices**: the current notice of each setting in use, to read again. Nothing records that it was
  read.
- **Backup retention** (**§4**).

| Route name | Method and path |
|---|---|
| `set-privacy-settings-page` | `GET /admin/root/set-privacy-settings` |
| `set-privacy-settings-form` | `POST /admin/root/set-privacy-settings` |
| `hash-privacy-ips-form` | `POST /admin/root/hash-privacy-ips` |
| `find-privacy-subject-page` | `GET /admin/root/find-privacy-subject` |
| `erase-privacy-subject-form` | `POST /admin/root/erase-privacy-subject` |
| `export-contact-requests` | `GET /admin/root/export-contact-requests` |
| `access-log-page` | `GET /admin/root/access-log` |
| `export-access-log` | `GET /admin/root/export-access-log` |
| `delete-access-log-form` | `POST /admin/root/delete-access-log` |
| `set-access-log-readers-form` | `POST /admin/root/set-access-log-readers` |
| `hold-privacy-records-form` | `POST /admin/root/hold-privacy-records` |
| `release-privacy-hold-form` | `POST /admin/root/release-privacy-hold` |
| `reset-backoffice-totp-form` | `POST /admin/root/reset-backoffice-totp` |
| `set-totp-page`, `set-totp-form` | `GET`, `POST /admin/set-totp` (any user, their own account) |

JSON twins under `/api/admin/root/`, named `api-` + the last path segment, as in Visualization:
`get-privacy-settings`, `set-privacy-settings`, `find-privacy-subject`, `erase-privacy-subject`,
`get-privacy-notice-draft`, `hash-privacy-ips`, `export-contact-requests`, `list-access-log`,
`export-access-log`, `delete-access-log`, `set-access-log-readers`, `hold-privacy-records`, `release-privacy-hold`,
`list-privacy-holds`, `list-privacy-access`, `reset-backoffice-totp`. `set-totp` is under `/api/admin/`.

**Banners** on every backoffice page (root): running `--notencrypt` · snapshots holding data past retention · settings that differ from the active preset ·
a new version of the active preset (**§11**) · the access log on without a source port, without its
two-factor requirement or without its inventory · roots without two-factor sign-in.

---

## 11. Presets — `/admin/root/privacy-presets` (root)

A large customer asks two things of a privacy setup: *prove that this server follows our policy* and *set
up the other servers the same way*. A **preset** is a named, versioned and complete set of values for the
privacy settings, each value with its reference. Applying one **is** saving those settings. It goes
through the same ranges (**§4**) and the same notices (**§9**). It is a shortcut and a
proof, never a way around them. It does not make the operator compliant, and the checklist below lists
what is left to them.

### What a preset sets
| Sets | Never sets |
|---|---|
| `privacy-setting.ip-mode`, `access-log`, `access-log-retention-days`, `ip-retention-days`, `backup-retention-days`, `notice-draft` · `contact-setting.retention-days` · `visualization-setting.retention-days` · `agent-setting.summary-retention-days` | **Whether** data is collected: `contact-setting.enabled`, `agent-setting.enabled` and `agent-setting.summaries-enabled` stay as they are. A preset says how data is kept, and a root still decides whether to keep it (**§2**). The keys (`EMISSARIO_ENCRYPT_KEY`, `ip-key`). Who sees full IPs (`viewer-ip`), the access log readers, its two-factor requirement and its inventory. Rate limits and sizes |

A preset holds **every** setting of the left column, so applying it always gives a known state. A preset
with a missing or unknown setting is refused. The access log is the one collection a preset turns on:
in Brazil it is an obligation, not a choice, and it still asks for its notice.

### Built-in presets — `assets/privacy/presets/<id>.json`
| Setting | `lgpd` v1 — for-profit company in Brazil | `gdpr` v1 — EU | `lgpd-gdpr` v1 — Brazilian company that also serves the EU |
|---|---|---|---|
| IP mode | `hashed`: the access log answers the court orders, so the records stay minimal (LGPD art. 6 III) | `hashed`: minimisation and privacy by default (GDPR arts. 5(1)(c), 25) | `hashed` |
| Access log | **on**: Marco Civil art. 15 requires it, and a hash cannot answer a court order for the IP (art. 10) | off: no general duty in the EU, minimisation and privacy by default (GDPR arts. 5(1)(c), 25) | **on** (Marco Civil art. 15) |
| Access log retention | 184: at least 6 months (Marco Civil art. 15) | 184, used if a root turns it on for security: CNIL 2021-122, 6 months–1 year | 184: the span that meets both |
| IP retention in records | 184: CNIL 2021-122, 6 months–1 year | 184 | 184 |
| Contact request | 180 | 180 | 180 |
| Visualization view | 90 | 90 | 90 |
| Agent summary | 90 | 90 | 90 |
| Backup snapshot | 90: data in a snapshot lives up to its retention + 90 days (184 + 90 < 1 year) | 90 | 90 |
| Notice draft (**§10**) | `br` | `eu` | `br-eu` |
| Notices asked when applied | `access-log-v1` | – | `access-log-v1` |

Neither law sets a number of days (**§1**), so the retentions are the same in all three. The presets
differ where the laws differ: the access log, the notice draft and the checklist. A Brazilian company
that keeps the access logs in its reverse proxy instead (**§6**) saves a custom preset from `lgpd` with
the access log off, and those logs then follow Decreto 8.771 art. 13 on their own.

### Controller checklist
Each preset carries the obligations that Emissario cannot fulfil for the operator. The Presets page lists
the items of the active preset. It is a reminder: Emissario stores no status for them,
done or not, so the list holds nothing that could be read as an admission. The controller keeps their
evidence in their own documents. Items marked `agent` show only while the agent is
on.

| Item | `lgpd` | `gdpr` | `lgpd-gdpr` |
|---|---|---|---|
| Privacy notice published on the website (draft in **§10**) | art. 9 | art. 13 | both |
| Encarregado / DPO named, and their identity and contact published | art. 41, Res. CD/ANPD 18/2024 (only small-scale agents are spared, Res. CD/ANPD 2/2022) | arts. 37–39, when required | both |
| Records of processing (start from *What this server keeps*, **§10**) | art. 37 | art. 30 | both |
| Legitimate interest balancing test for analytics and IP hashes | art. 10, ANPD guide 2024 | art. 6(1)(f) | both |
| Impact report (the agent is a new technology, and the ANPD may ask for one where the basis is legitimate interest) | RIPD, arts. 10 §3 and 38 | DPIA, art. 35 | both |
| A channel for data subjects' requests, answered in time | arts. 18–19 (15 days) | arts. 12(3), 15–22 (one month) | both |
| Incident response plan, and a register of every incident | ANPD and data subjects within 3 business days, register kept 5 years (Res. CD/ANPD 15/2024) | authority within 72 h, every breach documented (arts. 33, 34) | both |
| Access logs disclosed only on a court order, preservation requests kept confidential | Marco Civil arts. 10, 13 §4 and 15 §2 | – | Marco Civil |
| Sensitive data and minors: the form and the chat checked for them | arts. 11, 14; ECA Digital (Lei 15.211/2025) when minors are likely users | arts. 8, 9 | both |
| EU representative, if established outside the EU | – | art. 27 | art. 27 |
| `agent` · Processing terms signed with the model provider, no training on the conversations, transfer checked | art. 33 (the EU is adequate, Res. CD/ANPD 32/2026; elsewhere Res. CD/ANPD 19/2024) | arts. 28, 44–49 (Brazil is adequate) | both |
| `agent` · The chat UI tells visitors they are talking to an AI | good practice (PL 2338/2023 pending) | AI Act art. 50(1) | AI Act art. 50(1) |

### Applying a preset
1. **Preview** (`preview-privacy-preset-page`): every setting with its current value, the preset's
   value and the reference. Equal values are greyed out. The preview also says what happens to stored
   data:
   - a shorter retention: *N contact requests, M views, K snapshots… will be deleted at the next
     cleanup*, counted now with the same rules as the cleanup (**§4**);
   - access log turned on: *from now on every view and public API call writes an access log row*;
   - access log turned off: *no new rows; the N rows kept stay until their retention ends* (**§6**);
   - `full` → `hashed`: new IPs are hashed, and stored full IPs stay until their retention ends or until
     *Hash the IPs already stored* (**§10**);
   - `hashed` → `full`: only new IPs are stored in full, and stored hashes stay hashes.
2. **Apply**, confirmed with `data-confirm` when something will be deleted. Every value is validated
   (**§4**) before any is written. The notices that the new values need are shown together in one panel.
   The JSON twin answers the `409` of **§9** with a list, `"notices":[…]`, and expects all of them in
   `acknowledge`.
3. **State**: `privacy-setting.preset` = `lgpd@1`, and nothing else: not who applied it, not when.

**Automation** (provisioning many servers) goes through the JSON twin with a root API token:
`apply-privacy-preset {"preset":"lgpd@1","acknowledge":["access-log-v1"]}`.

### Active preset and drift
- A preset is a starting point, not a lock. A setting it covers can still be changed on its own page.
  When one differs, the Privacy page and a root banner say *2 settings differ from LGPD v1*, with the
  list, and offer **Re-apply** or **Detach**. Detaching makes the server *Custom*.
- `GET /api/admin/root/check-privacy-preset` gives the state in one call, for an operator that watches
  many servers:
  `{"preset":"lgpd@1","conform":false,"differences":[{"setting":"ip-retention-days","preset":184,"current":90}],"encrypted":true}`

### Custom presets
A company's DPO usually has a policy of their own: a shorter retention, `full` IPs for fraud review,
the access log off with the reverse proxy keeping the logs, more checklist items.
- **Save current settings as preset**: a name and a description. The checklist is copied from the active
  preset (`lgpd-gdpr` when there is none), and the root can add items (text and reference).
- **Export** (`export-privacy-preset`) downloads the JSON. **Import** (paste or upload) checks it against
  the format, the ranges (**§4**) and the known drafts, then stores it without applying it.
- Versions never change. Editing a preset saves `v2`, so the history and the drift check always point to
  fixed values. Built-in presets are read only.
- A custom preset that is not active can be deleted.
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
    "ip-mode":                   {"value": "hashed", "reference": "LGPD art. 6 III …"},
    "access-log":                {"value": true,     "reference": "Marco Civil art. 15 …"},
    "access-log-retention-days": {"value": 184,      "reference": "…"},
    "ip-retention-days":         {"value": 184,      "reference": "…"},
    "contact-retention-days":    {"value": 180,      "reference": "…"},
    "view-retention-days":       {"value": 90,       "reference": "…"},
    "summary-retention-days":    {"value": 90,       "reference": "…"},
    "backup-retention-days":     {"value": 90,       "reference": "…"},
    "notice-draft":              {"value": "br"}
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

JSON twins under `/api/admin/root/`: `list-privacy-presets`, `preview-privacy-preset`,
`apply-privacy-preset`, `save-privacy-preset`, `import-privacy-preset`, `export-privacy-preset`,
`delete-privacy-preset`, `detach-privacy-preset`, `check-privacy-preset`.

---

## 12. Database — `privacy-db` (key-prefix `privacydb`)
| Table | Fields |
|---|---|
| `privacy-setting` (one record) | `ip-mode` (`hashed`/`full`), `viewer-ip` (`masked`/`full`), `access-log` (`0`/`1`), `access-log-retention-days`, `access-log-totp` (`0`/`1`, default `1`), `access-log-inventory` (`0`/`1`, default `1`), `ip-key` (bytes, 32), `ip-key-rotated-at`, `ip-retention-days`, `backup-retention-days`, `notice-draft` (`br`/`eu`/`br-eu`), `preset` (`<id>@<version>`, blank = none) |
| `privacy-preset` (custom only, the built-ins are in `assets/`) | `preset-id`, `version`, `name`, `json` (the whole preset), `created-at` |
| `access-log` | `at` (UTC, ms), `ip`, `port`, `kind` (`view`/`contact-request`/`agent`), `ref`. No field is a `key` (**§3** rule): searches scan |
| `privacy-access` (only while `access-log-inventory` is on) | `at`, `ended-at`, `user-id`, `username`, `action`, `filter` (JSON), `rows` |
| `privacy-hold` | `reason`, `until`, `table`, `ids` (JSON). Released → deleted |
| `privacy-erasure` | `at`, `table`, `ids` (JSON). Only for the tables a restore rolls back, deleted once no older snapshot remains |

No table records who changed a setting, applied a preset, held or erased data, nor that a notice was
shown or a checklist item done (**§2**).
`user-id` and `username` appear only in `privacy-access`, the inventory the Decreto asks for.

A restore never rolls back what happened after the snapshot: it keeps the live `access-log`,
`privacy-access`, `privacy-hold` and `privacy-erasure`, then re-applies the erasures kept after the
snapshot was taken.

New fields: `contact-setting.retention-days` · `agent-setting.summaries-enabled`,
`agent-setting.summary-retention-days` · on the backoffice user: `totp-secret`, `totp-enabled-at`,
`access-log-reader`.

---

## 13. Where it lives
| Path | What |
|---|---|
| `sandbox/internal/privacy/` | IP key and HMAC, access log (rows, readers, page), retention rules and cleanup, legal holds, notices and the versions read, notice draft, find/erase and its replay after a restore, access inventory, presets (load, validate, preview, apply, drift, checklist) |
| `sandbox/internal/totp/` | TOTP (RFC 6238) and its QR code, used by the backoffice sign-in |
| `sandbox/internal/encryption/` | key parsing, HKDF, value format, `encryption.json` check |
| Keep `encryptedstorage`, or `adapters/impls/<encrypted-keep>/` | the storage wrapper (**§3**) |
| `sandbox/internal/commands/middleware/database_encryption/` | key / `--notencrypt` in front of every command |
| `sandbox/internal/commands/middleware/privacy_start_server/` | the retention goroutine, `--client-port-header` |
| `sandbox/internal/commands/…/generate_encrypt_key/`, `…/rotate_encrypt_key/`, `…/reset_backoffice_totp/` | the three new commands |
| `assets/agent/system.md` | the rule that the agent says it is an AI (**§8**) |
| `sandbox/internal/databases/privacy_db/` | **§12** |
| `assets/privacy/notices/*.md`, `assets/privacy/notice-drafts/{br,eu,br-eu}.md` | notice texts, draft templates |
| `assets/privacy/presets/{lgpd,gdpr,lgpd-gdpr}.json` | built-in presets (**§11**); a new version is a new file, the old one stays |
| `assets/backoffice/privacy_*.html`, `backofficerender/privacy_*.go`, `backofficeapi/privacy.go` | pages and JSON, presets included |
| `docs/Privacy/doc.md` | the doc, with the law tables of **§1**, the operador note and the limits of encryption of **§3**, the access log setup (proxy header, NTP) of **§6** and the presets of **§11** |

Update `docs/ContactRequests`, `docs/Visualization`, `docs/Agent`, `docs/Backups` and `docs/Backoffice`
(two-factor sign-in), and the README
quick start, which needs `EMISSARIO_ENCRYPT_KEY=$(emissario generate-encrypt-key)`. Store the key safely,
because losing it loses the data.

---

## 14. Acceptance
- `ls`/`cat` on `data/` shows no IP, email, contact body, summary or api key in clear. A test greps the
  data folder after inserting known values.
- Every value in an encrypted `data/` starts with `0xE1`, the `key_id` of `encryption.json` and a
  24-byte nonce, and opens with XChaCha20-Poly1305.
- Each row of the start-up table (**§3**) has a test, including the wrong key and both key and flag set.
- `git` started by the sync has no `EMISSARIO_ENCRYPT_KEY` in its environment.
- With the default settings: contact requests → `503 disabled`, summaries → `503 summaries-disabled`,
  views hold `h:` hashes and no user agent, and the access log is off and empty.
- The same IP always gives the same hash, also after a restart and after `rotate-encrypt-key` rotates
  `EMISSARIO_ENCRYPT_KEY`. After *Rotate IP key* it gives another. Two IPv6 addresses in one `/64` give
  one hash. `ip-key` never appears in clear in an encrypted `data/`.
- An IP hash older than the IP retention is blanked by the cleanup.
- `set-privacy-settings {ip-mode: full}` without `acknowledge` → `409` listing `ip-full-v1`. With it, new
  views hold the full IP and user agent, roots see them, and viewers see `203.0.113.xxx` until
  `viewer-ip` is `full` (notice `viewer-ip-v1`). *Hash the IPs already stored* turns them into `h:`
  hashes.
- `set-contact-settings {enabled:true}` without `acknowledge` → `409`. With it → `200`, and no database
  holds anything about the confirmation. Turning it off and on again asks again.
- No table of `privacy-db` but `privacy-access` has a `user-id` or `username` field. A test lists them.
- Retention: a record past retention is gone after one cleanup run. An IP past the IP retention is
  blanked and its record kept. A blank retention (forever) or one above the warning is saved after its
  notice, and `0` → `400`.
- Access log:
  - With it on, a view, a contact request and an agent call each write one row: `at` in UTC, the full
    IP, the source port from the connection, or from `--client-port-header` behind a proxy, and the
    `ref` of their record.
  - With the `lgpd` preset, a view 100 days old is gone and its access log row is still there. At 185
    days the row is gone too.
  - The access log answers `403` to a viewer, to a root who is not a reader, and, while two-factor is
    required, to a reader without TOTP and to a token made before TOTP was turned on. With
    `access-log-totp` `0` (after its notice), a reader without TOTP gets `200`.
  - With the inventory on, every list, search and export writes one `privacy-access` row. With it off,
    none.
- *Find and erase* finds a text in agent summaries, deletes the matches, leaves the access log rows and
  the held rows and lists them as kept, and writes one `privacy-erasure` row with no user. With *also
  delete the access log rows* it asks `erase-kept-v1` every time and deletes them too, never a held row,
  and leaves no row anywhere about that deletion.
- A held row survives a cleanup past its retention. After its release it is gone at the next cleanup.
- Restoring a snapshot taken before an erasure re-applies it, and keeps the live access log and record
  tables (**§12**).
- The embedded `assets/agent/system.md` holds the rule that the agent says it is an AI.
- An encrypted backup restores on the same key and is refused on another.
- Presets:
  - `apply-privacy-preset {"preset":"lgpd@1"}` without `acknowledge` → `409` listing `access-log-v1`, and
    nothing is written. With it → `200`, every setting equals the preset, `privacy-setting.preset` is
    `lgpd@1`, and nothing records who applied it or when.
  - `gdpr@1` on a fresh install needs no notice.
  - Applying never changes `contact-setting.enabled`, `agent-setting.enabled` or
    `agent-setting.summaries-enabled`.
  - An imported preset with one value out of range, a missing setting or an unknown key is refused, and
    nothing is written.
  - Changing `ip-retention-days` by hand after applying → `check-privacy-preset` gives `conform:false`
    with that difference. *Re-apply* → `conform:true`. *Detach* → `preset` blank.
  - The counts in the preview equal what the next cleanup deletes.
  - Export on one server, import and apply on another → the same settings on both.
  - A built-in preset cannot be edited or deleted, a saved version cannot be changed in place, and the
    active preset cannot be deleted.

## Important
- Every yaml through agnos commands (`add-database`, `add-table-field`, `add-route`,
  `add-command --middleware`, `add-adapter`). No generated file is edited by hand. Finish with
  `agnos build`.
- Never log a decrypted value, a key, an IP or a contact body. Errors name the field, not the value.
- Suggested order: encryption → `privacy-db` and notices → IP hashes → access log and TOTP → contacts →
  agent and summaries → retention, holds and backups → Privacy page → presets → docs.

## Out of scope
- A cookie banner or consent manager: Emissario sets no cookie on visitors.
- A self-service portal for visitors' requests: the root answers them through **§10**.
- A blocklist by IP hash (*Block* on a visitor, contact request or usage row → `403` for that hash). The
  stable hash of **§6** makes it possible later.
- KMS, HSM or a key server: the key comes from the environment.
- SSO (SAML, OIDC) for the backoffice: TOTP covers the access log readers.
- An incident register inside Emissario: the checklist reminds the operator to keep one (**§11**).
- A compliance status kept by Emissario (items done, dates, who): the controller keeps their own
  evidence, where they choose (**§2**).
- Choosing a preset automatically by country: a root chooses it (**§11**).
- Built-in presets, notices and drafts for laws other than BR/EU (UK GDPR, US state laws). A custom
  preset can hold their values, but not their notices.
- Locking a server to a preset (refusing changes that leave it). The drift check and
  `check-privacy-preset` show those changes instead of blocking them.
- Legal advice: the notices inform, and the operator decides.
