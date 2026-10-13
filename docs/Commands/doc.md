# Commands

`emissario <command> [args] [flags]`. `emissario help <command>`, or `emissario <command> --help`,
prints the same for one command; an empty command line prints the general help and exits 0.
A command declaring a `--help` flag of its own keeps it, and is described through `help` alone.

One page per command, each rendered from that command's `command.yaml`
([CommandYaml](../CommandYaml/doc.md)) on each build — open the one you need rather than this
whole page. Hidden commands are not listed. The args are the leading words of the command line;
the flags follow them, in any order. A `repeatable` flag is given once per value. A command's page
lists the flags of the middlewares in front of it too.

## Backoffice

| Command | Does |
| --- | --- |
| [`add-backoffice-user`](add-backoffice-user.md) | Create a backoffice user with a generated password, printed once |

## Info

| Command | Does |
| --- | --- |
| [`help`](help.md) | Display help for a command |
| [`version`](version.md) | Print the installed version |

## Server

| Command | Does |
| --- | --- |
| [`start-server`](start-server.md) | Start the http server |

## Middlewares

Run in front of the commands they match, lowest `priority` first; typed by nobody.

| Middleware | Runs before | Priority | Flags it adds |
| --- | --- | --- | --- |
| [`backoffice-start-server`](backoffice-start-server.md) | `start-server` | 50 | `--addr`, `--allow-x-forwarded-for`, `--insecure-http` |
| [`database-dir`](database-dir.md) | `*` | 6 | `--database` |
| [`help-flag`](help-flag.md) | `*` | 5 | `--help`, `-h` |
| [`visualization-start-server`](visualization-start-server.md) | `start-server` | 52 | `--visualization-geoip` |
| [`website-source-start-server`](website-source-start-server.md) | `start-server` | 51 | `--fonts` |

Output channels and exit codes are in [Rules](../Rules/doc.md#output-channels).
