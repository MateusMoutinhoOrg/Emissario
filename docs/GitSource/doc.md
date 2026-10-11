# GitSource

The public website is a git repository, served from its root. A root sets the repository,
branch and sync time in the backoffice; `start-server` clones it into `<fonts>/frontend` and keeps
pulling the branch.

## Run it

```bash
emissario start-server --fonts EmissarioFonts   # the default; clones into ./EmissarioFonts/frontend
```

Then, as a root, open `/admin/root/set-git-source` (the **Git source** link of the backoffice).

| Field | Rule |
|---|---|
| Repository URL | `https://`, `http://`, `ssh://`, `git://`, `file://` or `user@host:path`; no `::` remote helper, no leading `-`, no password or token in it |
| Branch | letters, digits and `. _ / -`, as git accepts a branch name (`main`, `release/v2`) |
| Sync time | one or more `<n>s`, `<n>m`, `<n>h` (`30s`, `10m`, `1h30m`), 10s to 720h; defaults to `10m` in the form; blank stops syncing |

## Credentials

None is asked for, stored or passed: git runs as the user running the server, with that user's
ssh keys, ssh agent or git credential helper. A missing credential fails the sync instead of
waiting on a prompt: `GIT_TERMINAL_PROMPT=0`, and ssh runs with `BatchMode=yes` and a 30 s
connect timeout unless `GIT_SSH_COMMAND`, `GIT_SSH` or `core.sshCommand` is set. A new ssh host
needs its key in `~/.ssh/known_hosts` first (`ssh-keyscan github.com >> ~/.ssh/known_hosts`).

## The sync loop

`git-source-start-server`, a cli middleware (priority `51`) in front of `start-server`, reads
`--fonts` into `sandbox.Config.FontsDir` and starts one goroutine, which lives as long as the
process. Every 2 s it reads the `frontend` record of `source` and, when a sync time is set, syncs:

| When | |
|---|---|
| the first tick | at start |
| every sync time | after the last sync, whether it passed or failed |
| the url or the branch changed | at once |
| the form is saved | at once |

| `<fonts>/frontend` holds | Sync runs |
|---|---|
| no checkout | `git clone --depth 1 --single-branch --branch <b> -- <url> <dir>` |
| a checkout | `git remote set-url origin -- <url>` (when it changed), `git fetch --depth 1 origin +refs/heads/<b>:refs/remotes/origin/<b>`, `git checkout --force -B <b> refs/remotes/origin/<b>` |

A fetch that fails leaves the checkout, and the site, as they were. Every call carries
`core.symlinks=false` (a symlink of the repository is checked out as a plain file, so it never
points the site at a file of the server) and `protocol.ext.allow=never`. The page shows the last
sync: when, the commit, and the tail of git's output when it failed; stderr carries it too.

## How a path is answered

The `front` route (`GET /{*Rest}`, priority `1000`) reads the checkout, `<fonts>/frontend/`:

| Request | Answered with |
|---|---|
| `/` | `<fonts>/frontend/index.html` |
| path ending in `/` (`/blog/`) | `<fonts>/frontend/<path>/index.html` |
| any other path (`/blog/app.css`) | `<fonts>/frontend/<path>` — a folder without its `/` is not found |
| not found | `404` with `<fonts>/frontend/not-found.html`, else the embedded `assets/front/not-found.html` |

Every path goes through `OpinionatedAgnosFront.SafePath` first: no `..`, no empty segment, no
dotfile (`.git`) but `.well-known`. `/backoffice/*` files embedded under `assets/front/backoffice/`
(the backoffice script) are always the embedded ones. Content-Type comes from the extension,
`Cache-Control: no-cache`.

## Where it lives

| Path | What |
|---|---|
| `sandbox/internal/gitsource/` | settings and validation (`gitsource.go`), the loop and git (`sync.go`), path resolution (`site.go`) |
| `sandbox/internal/databases/emissario_db/` | table `source`: `name`, `repourl`, `branch`, `sync-time`; the store is `./data/emissariodb`, gitignored, in every backup |
| `sandbox/internal/commands/middleware/git_source_start_server/` | `--fonts`, starts the loop |
| `sandbox/internal/routes/backoffice/root/set_git_source_{page,form}/` | `GET` / `POST /admin/root/set-git-source`, root only |
| `sandbox/internal/server/backoffice/backofficerender/git_source_form.go`, `assets/backoffice/git_source_form.html` | the page |
| `sandbox/internal/routes/front/handler.go` | the public website |
| `sandbox/api/projectconfig.go` | `Config.FontsDir`, `api.DefaultFontsDir` |
| `sandbox/deps/sleepdeps/`, `adapters/impls/stdsleep/` | `Deps.SleepDeps.Sleep`, the wait between two ticks |
| `sandbox/deps/rundeps/`, `adapters/impls/osexecrun/` | `Deps.RunDeps.Run`, how git is run |
