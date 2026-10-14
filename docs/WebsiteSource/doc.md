# WebsiteSource

The public website is a git repository, served from its `frontend/` folder. A root sets the repository,
branch and sync time in the backoffice; `start-server` clones it into `<fonts>/frontend` and keeps
pulling the branch. **Sync** — a button and its JSON twin — pulls it on demand.

## Run it

```bash
emissario start-server --fonts EmissarioFonts   # the default; clones into ./EmissarioFonts/frontend
```

Then, as a root, open `/admin/root/set-website-source` (the **Website source** link of the backoffice).

| Field | Rule |
|---|---|
| Repository URL | `https://`, `http://`, `ssh://`, `git://`, `file://` or `user@host:path`; no `::` remote helper, no leading `-`, no password or token in it |
| Branch | letters, digits and `. _ / -`, as git accepts a branch name (`main`, `release/v2`) |
| Sync time | one or more `<n>s`, `<n>m`, `<n>h` (`30s`, `10m`, `1h30m`), 10s to 720h; defaults to `10m` in the form; blank stops the automatic syncing |

## Sync on demand

| Surface | Answers |
|---|---|
| **Sync** button of the page, `POST /admin/root/sync-website-source` | `303` back to the page, which shows *syncing…* and refreshes every 2 s until the sync ends |
| `POST /api/admin/root/sync-website-source` | `202 {"status":"syncing"}`; `409` no website source saved yet; `503` no sync loop in this process |

```bash
curl -X POST localhost:3000/api/admin/root/sync-website-source -H "Authorization: Bearer $TOKEN"
```

Both are root only, and both run a sync even when the site is up to date, and with a blank sync
time. The sync runs in the loop, so two syncs never run at once; asking while one is pending
asks once.

## Credentials

None is asked for, stored or passed: git runs as the user running the server, with that user's
ssh keys, ssh agent or git credential helper. A missing credential fails the sync instead of
waiting on a prompt: `GIT_TERMINAL_PROMPT=0`, and ssh runs with `BatchMode=yes` and a 30 s
connect timeout unless `GIT_SSH_COMMAND`, `GIT_SSH` or `core.sshCommand` is set. A new ssh host
needs its key in `~/.ssh/known_hosts` first (`ssh-keyscan github.com >> ~/.ssh/known_hosts`).

## The sync loop

`website-source-start-server`, a cli middleware (priority `51`) in front of `start-server`, reads
`--fonts` into `sandbox.Config.FontsDir` and starts one goroutine per record of `source` —
`frontend`, the website, and `agent`, the instructions of the [agent](../Agent/doc.md), cloned into
`<fonts>/agent` and set at `/admin/root/set-agent-source` — each living as long as the process, so a
slow clone of one never holds the other. Every 2 s each reads its record and syncs:

| When | Needs a sync time |
|---|---|
| the first tick, at start | yes |
| every sync time after the last sync, passed or failed | yes |
| the url, the branch or the sync time changed | yes |
| Sync asked for (button or API) — wakes the loop at once | no |

| `<fonts>/frontend` holds | Sync runs |
|---|---|
| no checkout | `git clone --depth 1 --single-branch --branch <b> -- <url> <dir>` |
| a checkout | `git remote set-url origin -- <url>` (when it changed), `git fetch --depth 1 origin +refs/heads/<b>:refs/remotes/origin/<b>`, `git checkout --force -B <b> refs/remotes/origin/<b>` |

A fetch that fails leaves the checkout, and the site, as they were. Every call carries
`core.symlinks=false` (a symlink of the repository is checked out as a plain file, so it never
points the site at a file of the server) and `protocol.ext.allow=never`. The page shows the last
sync: when, the commit, and the tail of git's output when it failed; stderr carries it too.

## How a path is answered

The `front` route (`GET /{*Rest}`, priority `1000`) reads the `frontend/` folder of the checkout,
`<fonts>/frontend/frontend/` — `repo/frontend/` below:

| Request | Answered with |
|---|---|
| `/` | `repo/frontend/index.html` |
| path ending in `/` (`/pagina/`) | `repo/frontend/<path>/index.html` |
| any other path (`/style/app.css`, `/pagina`) | `repo/frontend/<path>` when it is a file, else `repo/frontend/<path>.html` (`/pagina` → `pagina.html`) |
| not found | `404` with `repo/frontend/not-found.html`, else the embedded `assets/front/not-found.html` |

Every path goes through `OpinionatedAgnosFront.SafePath` first: no `..`, no empty segment, no
dotfile (`.git`) but `.well-known`. `/backoffice/*` files embedded under `assets/front/backoffice/`
(the backoffice script) are always the embedded ones. Content-Type comes from the extension,
`Cache-Control: no-cache`.

## Where it lives

| Path | What |
|---|---|
| `sandbox/internal/websitesource/` | settings and validation of every source (`websitesource.go`), one loop per source, Sync and git (`sync.go`), path resolution (`site.go`) |
| `sandbox/internal/databases/emissario_db/` | table `source`: `name` (`frontend`, `agent`), `repourl`, `branch`, `sync-time`; the store is `./data/emissariodb`, gitignored, in every backup |
| `sandbox/internal/commands/middleware/website_source_start_server/` | `--fonts`, starts the loop |
| `sandbox/internal/routes/backoffice/root/set_website_source_{page,form}/` | `GET` / `POST /admin/root/set-website-source` |
| `sandbox/internal/routes/backoffice/root/sync_website_source_form/` | `POST /admin/root/sync-website-source`, the Sync button |
| `sandbox/internal/routes/api/backoffice/root/api_sync_website_source/` | `POST /api/admin/root/sync-website-source` |
| `sandbox/internal/server/backoffice/backofficerender/website_source_form.go`, `assets/backoffice/website_source_form.html` | the page |
| `sandbox/internal/routes/front/handler.go` | the public website |
| `sandbox/api/projectconfig.go` | `Config.FontsDir`, `api.DefaultFontsDir` |
| `sandbox/deps/sleepdeps/`, `adapters/impls/stdsleep/` | `Deps.SleepDeps.Sleep`, the wait between two ticks |
| `sandbox/deps/rundeps/`, `adapters/impls/osexecrun/` | `Deps.RunDeps.Run`, how git is run |
