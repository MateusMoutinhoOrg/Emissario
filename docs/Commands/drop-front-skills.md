# `drop-front-skills`

Write the skills that teach a coding agent to build the website emissario serves

```bash
emissario drop-front-skills [--path <path>] [--database <database>] [--help]
```

Writes the emissario-website skill into the website repository at --path, once for each coding agent: .claude/skills/ (Claude Code), .agents/skills/ (Codex) and .agent/skills/ (Antigravity). The skill holds the public interface the website needs and nothing else: how its frontend/ folder is served, the paths the server keeps, the contact form at POST /api/contact-requests and the chat of the agent at POST /api/agent/messages. A run removes the emissario-website folder of each agent and writes it again, so a second run changes nothing and a new version leaves no stale file; nothing else of the repository is touched. It opens no database.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--path` | string | `.` | the website repository to write the skills into (defaults to the current directory) | — |
| `--database` | string | `data` | the folder every database lives under, relative to where the command runs, lower-case letters, digits, - and _ (defaults to data) | [database-dir](database-dir.md) |
| `--help`, `-h` | boolean |  | Print the help of the command this command line is for | [help-flag](help-flag.md) |

| Runs in front of it | When |
| --- | --- |
| [`database-dir`](database-dir.md) | always |
| [`help-flag`](help-flag.md) | always |

```bash
emissario drop-front-skills --path .
emissario drop-front-skills --path ../my-site
```

Skills · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
