# emissario

[![Go Reference](https://pkg.go.dev/badge/github.com/MateusMoutinhoOrg/Emissario.svg)](https://pkg.go.dev/github.com/MateusMoutinhoOrg/Emissario)
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
| **IP blocks** | A root blocks one ip, a CIDR range, a hashed ip or a whole country; their requests are answered `403` before the website, the contact form, the agent or the sign-in reads them. [IP Blocks](docs/IpBlocks/doc.md) |
| **Backups** | Snapshots of every database, taken in the background, then downloaded, uploaded and restored from the backoffice. [Backups](docs/Backups/doc.md) |
| **Skills for coding agents** | `drop-front-skills` and `drop-agents-skills` write into the website and the agent repositories the skills Claude Code, Codex and Antigravity load: the public routes the site calls, and the files the agent reads. [Skills](docs/Skills/doc.md) |

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


## Documentation

### CliUsage

Driving the CLI from a terminal - install, commands, flags, exit codes

| Doc | Description |
| --- | --- |
| [CliInstall](docs/CliInstall/doc.md) | Install emissario: a released binary per platform, or a build from source |
| [Commands](docs/Commands/doc.md) | Every command of emissario, generated from the command declarations |
| [CliExamples](docs/CliExamples/doc.md) | Index of every runnable example of the emissario cli |
| [Skills](docs/Skills/doc.md) | Skills for coding agents: drop-front-skills and drop-agents-skills write into the website and the agent repositories what Claude Code, Codex and Antigravity need to know of emissario |

### ServerUsage

Serving http - bring the server up, declare routes, read a body

| Doc | Description |
| --- | --- |
| [ServerUsage](docs/ServerUsage/doc.md) | Serve http from emissario: bring the layer up, declare routes, read a body |
| [FrontUsage](docs/FrontUsage/doc.md) | Serve a website from emissario: bring the front layer up, drop files in assets/front, add pages |
| [Routes](docs/Routes/doc.md) | Every address emissario answers: what to send, a request to run, what comes back, and the OpenAPI document for Postman or Swagger |
| [Backoffice](docs/Backoffice/doc.md) | The admin backoffice of emissario: secret, users, roles, sessions, API tokens and the /api/admin JSON api |
| [Backups](docs/Backups/doc.md) | Backups of every database of emissario from the backoffice: take a snapshot in the background, download, upload and restore one, or build one by hand file by file |
| [Privacy](docs/Privacy/doc.md) | Privacy by default under the LGPD, the Marco Civil da Internet and the GDPR: every database encrypted at rest, IPs hashed, a Marco Civil access log a root turns on, retentions with their references, notices before every risky choice, find and erase, legal holds and presets, at /admin/root/set-privacy-settings |
| [Agent](docs/Agent/doc.md) | The public agent of the website: an anonymous visitor talks to a configurable model through /api/agent/messages, its instructions synced from a git repository, its urls, limits, signed history and anonymous summaries set and read in the backoffice at /admin/agent |
| [Contact Requests](docs/ContactRequests/doc.md) | The contact form of the website: any JSON the website sends to /api/contact-requests, kept as sent with its time, ip and country, read, searched, archived and deleted in the backoffice at /admin/list-contact-requests |
| [IP Blocks](docs/IpBlocks/doc.md) | Turn away the requests of an ip, a CIDR range, a hashed ip or a whole country before any route answers them: set by a root in the backoffice at /admin/root/ip-blocks, its JSON api, and remove-ip-blocks on the command line |
| [Visualization](docs/Visualization/doc.md) | Page views of the public website, recorded on the server: which page, when, from which ip, country, device, browser and os, searched and counted in the backoffice at /admin/visualization and its JSON api |
| [Website Source](docs/WebsiteSource/doc.md) | Serve the public website from a git repository: the root sets the repository, branch and sync time in the backoffice, start-server keeps it pulled, and Sync or its API syncs it on demand |

### LibUsage

Using the project as a Go module - wiring the deps, calling the sandbox

| Doc | Description |
| --- | --- |
| [LibUsage](docs/LibUsage/doc.md) | Use emissario as a Go module: wire the deps, build the sandbox, call its API |
| [PublicApi](docs/PublicApi/doc.md) | Every exported symbol of emissario, generated from the contract sources and their doc comments |
| [LibExamples](docs/LibExamples/doc.md) | Index of every runnable example of emissario as a Go module |

### Architecture

How the project is put together - layers, boundaries, data flow

| Doc | Description |
| --- | --- |
| [Databases](docs/Databases/doc.md) | Every database of emissario, generated from the table declarations |
| [Adapters](docs/Adapters/doc.md) | Contract, adapter and binding: three units, one field of Deps, and who fills it |

### Development

Changing this repository - schema, build mechanics, recipes

| Doc | Description |
| --- | --- |
| [Requirements](docs/Requirements/doc.md) | The two tools this project needs — Go and agnos — installed per platform |
| [Workflow](docs/Workflow/doc.md) | Every change this project takes and the agnos command that makes it |
| [Rules](docs/Rules/doc.md) | Every rule the generators, `verify` and the hand-written files must hold to |
| [Structure](docs/Structure/doc.md) | The project schema: what lives where, what is generated, what verify enforces |

### Reference

Lookup tables - schemas, file formats, generated file listings

| Doc | Description |
| --- | --- |
| [CommandYaml](docs/CommandYaml/doc.md) | Every key of a command's command.yaml and what the generated code does with it |
| [Extensions](docs/Extensions/doc.md) | The generation mechanics this project turns on, and what each one writes |
| [DepList](docs/DepList/doc.md) | Every dep `agnos add-dep` can add, the adapters that fill it, and what backs each one |
| [Databases](docs/Databases/doc.md) | Every database of emissario, generated from the table declarations |
| [GeneratedFiles](docs/GeneratedFiles/doc.md) | Every file agnos writes into this project and whether build overwrites it |
| [RouteYaml](docs/RouteYaml/doc.md) | Every key of a route's route.yaml and what the generated code does with it |

## License

MIT License

Copyright (c) 2026 Mateus Moutinho

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

