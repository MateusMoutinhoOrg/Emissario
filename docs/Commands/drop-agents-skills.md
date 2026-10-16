# `drop-agents-skills`

Write the skills that teach a coding agent to write the instructions of the agent of emissario

```bash
emissario drop-agents-skills [--path <path>] [--database <database>] [--help]
```

Writes the emissario-agent skill into the agent instructions repository at --path, once for each coding agent: .claude/skills/ (Claude Code), .agents/skills/ (Codex) and .agent/skills/ (Antigravity). The skill holds what the repository needs to know of the agent and nothing else: instructions/README.md read first at every message, instructions/SUMMARY.md for the summaries, every other file of instructions/ read on demand through the tools of the model, the limits, and the mandatory rules that come before the instructions. A run removes the emissario-agent folder of each agent and writes it again, so a second run changes nothing and a new version leaves no stale file; nothing else of the repository is touched. It opens no database.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--path` | string | `.` | the agent instructions repository to write the skills into (defaults to the current directory) | — |
| `--database` | string | `data` | the folder every database lives under, relative to where the command runs, lower-case letters, digits, - and _ (defaults to data) | [database-dir](database-dir.md) |
| `--help`, `-h` | boolean |  | Print the help of the command this command line is for | [help-flag](help-flag.md) |

| Runs in front of it | When |
| --- | --- |
| [`database-dir`](database-dir.md) | always |
| [`help-flag`](help-flag.md) | always |

```bash
emissario drop-agents-skills --path .
emissario drop-agents-skills --path ../my-agent
```

Skills · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
