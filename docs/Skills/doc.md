# Skills

Two repositories work with emissario besides its own: the **website** it serves (its `frontend/`
folder, see [Website Source](../WebsiteSource/doc.md)) and the **agent instructions** its public agent
reads (its `instructions/` folder, see [Agent](../Agent/doc.md)). Two commands write into each one a
skill: a folder of Markdown a coding agent loads when the task needs it, so that Claude Code, Codex or
Antigravity changing that repository knows the side of emissario it talks to, and nothing more.

## Run it

```bash
emissario drop-front-skills  --path ../my-site    # in the website repository
emissario drop-agents-skills --path ../my-agent   # in the agent instructions repository
```

`--path` defaults to the current directory, which must exist. Neither command opens a database, so
neither needs `EMISSARIO_ENCRYPT_KEY`. Each prints the path of every file it wrote, relative to `--path`,
and warns on stderr when the repository does not look like one it is meant for: a website with no
`frontend/` folder, an agent repository with no `instructions/README.md`. It writes anyway. Commit the
folders it writes, so every clone of the repository has them.

When the agent source is the website repository, run both: the two skills have different names and
live side by side.

## Where they go

The same skill is written once for each agent, in the folder that agent reads a project's skills from,
in the [Agent Skills](https://agentskills.io) layout (`<name>/SKILL.md` and its `references/`):

| Agent | Folder |
|---|---|
| Claude Code | `.claude/skills/<name>/` |
| Codex | `.agents/skills/<name>/` |
| Antigravity | `.agent/skills/<name>/` |

A run removes `<folder>/<name>/` and writes it again. A second run changes nothing, a newer
emissario leaves no stale file, and an edit made by hand to these folders is lost. Nothing else of the
repository is touched: no `AGENTS.md`, no `CLAUDE.md`, no other skill. Run the command again after
upgrading emissario.

## What each one holds

Only the public interface that repository needs: what its pages send and receive, or what the model is
given. Never how the server works, its backoffice, its databases or its source.

### `emissario-website`, from `drop-front-skills`

| File | Holds |
|---|---|
| `SKILL.md` | only `frontend/` is served, as committed, with no build step; how a path is answered and the `404` page; dotfiles; the headers every file gets; the paths the server keeps (`/admin`, `/api/…`, `/backoffice/`, `/health`, `/openapi.json`); same origin, no CORS; no analytics script needed; what the privacy notice must say |
| `references/contact-form.md` | `POST /api/contact-requests`: the body, `201 {id}`, every error and what to tell the visitor, the default limits, a form that sends it |
| `references/agent-chat.md` | the AI disclosure before the chat; `POST /api/agent/messages` and its signed history; the limits; every error and what the page does; a client keeping the history in `localStorage`; `POST /api/agent/summaries` |

### `emissario-agent`, from `drop-agents-skills`

| File | Holds |
|---|---|
| `SKILL.md` | `instructions/README.md` read first at every message and required; `instructions/SUMMARY.md` for the summaries; the other files read on demand through `list_instructions`, `read_instruction` and `call_url`; the limits; how to write the README as an index; what must never be in the repository; publishing by `git push` |
| `references/mandatory-rules.md` | the embedded `assets/agent/system.md` and `assets/agent/summary.md`, word for word, and what they mean for the instructions |
| `references/readme-template.md` | a starting point for `instructions/README.md` |

The limits a skill quotes (body sizes, messages, characters, tool rounds, rates) are rendered from the
constants the server enforces, and the mandatory rules from the files it embeds, so a skill never
drifts from the binary that wrote it.

## Where it lives

| Path | What |
|---|---|
| `assets/skills/front/`, `assets/skills/agents/` | the skills, one folder each, rendered as text/template |
| `sandbox/internal/skills/` | `Drop`: render every file, then remove and write the folder of each agent |
| `sandbox/internal/commands/skills/` | the two commands |
| `examples/cli/drop-skills/` | both commands run twice on two repositories |
