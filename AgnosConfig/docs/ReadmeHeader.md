# {{.ProjectName}}

[![Go Reference](https://pkg.go.dev/badge/{{.Module}}.svg)](https://pkg.go.dev/{{.Module}})
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-blue)](go.mod)

**Emissario is a self-hosted website server in a single binary.** Point it at a git repository and it
serves the site and keeps it in sync. An admin backoffice comes with it: page-view analytics with no
cookies, a contact form inbox, a public AI chat agent, backups, and the privacy controls the LGPD and the
GDPR ask for. It runs no database server, sends
nothing to an analytics service and installs no runtime. The only thing the host needs is `git`.

## What it does

| Feature | |
| --- | --- |
| **Website from git** | Serves the `frontend/` folder of a git repository and pulls its branch on a schedule, or on demand from a button or the API. Publishing the site is a `git push`. [Website Source](docs/WebsiteSource/doc.md) |
| **Backoffice** | A login at `/admin` with `root` and `viewer` users, sessions, API tokens, and a JSON twin of every page under `/api/admin`. [Backoffice](docs/Backoffice/doc.md) |
| **Analytics without tracking** | The server records each page view itself (page, time, country, device, browser, OS) with no cookie, no tracking script and no third-party service. [Visualization](docs/Visualization/doc.md) |
| **Contact form** | `POST /api/contact-requests` keeps any JSON the site sends, with its time, IP and country, limited per IP. The backoffice reads, searches and archives it. [Contact Requests](docs/ContactRequests/doc.md) |
| **AI agent** | A public chat API backed by any OpenAI-compatible model (OpenAI, OpenRouter, Groq, Ollama…), with its instructions kept in a git repository. The conversation stays in the visitor's browser. [Agent](docs/Agent/doc.md) |
| **Privacy by default** | Every database encrypted at rest, visitor IPs hashed, contact requests and summaries off until turned on, retentions deleted every hour, a Marco Civil access log, find and erase, legal holds, and LGPD/GDPR presets — each risky choice behind a notice that names the law. [Privacy](docs/Privacy/doc.md) |
| **Backups** | Snapshots of every database, taken in the background, then downloaded, uploaded and restored from the backoffice. [Backups](docs/Backups/doc.md) |

Every database is embedded and lives in one folder, `./data` by default.

## Quick start

Install the binary for your platform ([CliInstall](docs/CliInstall/doc.md)), then:

```bash
export EMISSARIO_ENCRYPT_KEY=$(emissario generate-encrypt-key)   # store it safely: losing it loses every database and backup
emissario add-backoffice-user --username admin --email admin@example.com --role root   # prints the password once
emissario start-server --insecure-http --addr 127.0.0.1:3000                            # plain http, local only
```

Open `http://127.0.0.1:3000/admin/login`, sign in, and set the repository under **Website source**.
In production, drop `--insecure-http`, serve over https, and set `EMISSARIO_BACKOFFICE_SECRET` so that
sessions survive a restart ([Backoffice](docs/Backoffice/doc.md)). Every command that opens a database needs
`EMISSARIO_ENCRYPT_KEY`; `--notencrypt` keeps the data in clear, for local debugging only. Then set
the privacy policy under **Privacy** ([Privacy](docs/Privacy/doc.md)).

The country of every page view ([Visualization](docs/Visualization/doc.md)) is looked up in the DB-IP *IP to Country Lite* database embedded in the binary, licensed CC BY 4.0: [IP Geolocation by DB-IP](https://db-ip.com).
