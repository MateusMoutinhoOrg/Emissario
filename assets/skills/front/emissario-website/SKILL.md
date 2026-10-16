---
name: emissario-website
description: Use when building or changing this website, which Emissario serves from its frontend/ folder - adding pages, assets or a 404 page, wiring a contact form to POST /api/contact-requests, or an AI chat to POST /api/agent/messages.
---

# Emissario website

This repository is a website served by [Emissario](https://github.com/MateusMoutinhoOrg/Emissario).
Emissario clones it, serves its `frontend/` folder as static files, and answers a few routes of its
own beside them: a contact form and an AI chat that the pages of this site call.

## What is served

- **Only `frontend/`.** Nothing outside it is reachable: the rest of the repository (sources, a
  README, the build setup) stays private.
- **Static files, as committed.** Emissario runs no build step, no server-side code and no package
  install. A site built with a bundler or a static generator must commit its built output into
  `frontend/`.
- **Publishing is a `git push`.** The server pulls the branch on its sync time (or when a root presses
  Sync), and serves the new files from then on.

| Request | Answered with |
|---|---|
| `/` | `frontend/index.html` |
| a path ending in `/` (`/about/`) | `frontend/about/index.html` |
| any other path (`/style/app.css`, `/about`) | that file when it exists, else the same path plus `.html` (`/about` → `frontend/about.html`) |
| nothing matches | `404` with `frontend/not-found.html`, or a built-in page when the site has none |

- Only `GET`. A path with `..`, an empty segment or a dotfile (`.env`, `.git/`) is never served;
  `.well-known/` is the one dot-folder that is.
- `Content-Type` comes from the file extension, so give every file its real extension. The server sends
  `X-Content-Type-Options: nosniff`, so a wrong type is not guessed.
- `Cache-Control: no-cache` on every file: browsers revalidate, so there is no need to version file
  names to see a change.
- `X-Frame-Options: SAMEORIGIN`: the pages cannot be framed by another site.

## Paths the server keeps

These are answered by Emissario, never by a file of `frontend/`. Do not create files or links that need
them to be pages of the site:

| Path | What |
|---|---|
| `/admin`, `/admin/…`, `/api/admin/…` | the backoffice, for the people who run the site |
| `/api/contact-requests` | the contact form, see [references/contact-form.md](references/contact-form.md) |
| `/api/agent/messages`, `/api/agent/summaries` | the AI chat, see [references/agent-chat.md](references/agent-chat.md) |
| `/backoffice/…` | a script of the backoffice |
| `/health`, `/openapi.json` | for monitoring and tooling |

## Calling the server from a page

- **Same origin only.** Call the routes with a relative URL (`fetch("/api/contact-requests")`) from a
  page of this site. There is no CORS: a page of another site, or a page opened from another domain
  in development, is refused `403 {"error":"origin"}`.
- Every error is JSON: `{"error": "<code>", "field": "<the part of the body, or empty>"}`. Show the
  visitor a plain message for each code the reference lists, never the raw JSON.
- A visitor whose IP a root blocked is answered `403 {"error":"blocked"}` on every path, pages
  included.
- Both features can be turned off by the people who run the site. They then answer `503 disabled`,
  so a page must still work, and say so, without them.

## Good to know

- **Page views are already counted.** Emissario records them on the server, with no script and no
  cookie: the site needs nothing for that.
- **Everything under `frontend/` is public.** No route above needs an API key, so no page holds one.

## References

Read the one you need, when you need it:

- [references/contact-form.md](references/contact-form.md): `POST /api/contact-requests`, its answers
  and a form that sends it.
- [references/agent-chat.md](references/agent-chat.md): `POST /api/agent/messages` and
  `/api/agent/summaries`, the signed history, the AI disclosure, and a chat client.

These files are written by `emissario drop-front-skills`. Run it again after upgrading Emissario
rather than editing them: a run rewrites this folder.
