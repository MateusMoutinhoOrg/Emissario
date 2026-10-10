# Workflow

Every change this project takes and the command that makes it. `agnos` owns every generated
file; what stays hand-written is listed in [GeneratedFiles](../GeneratedFiles/doc.md), and the
rules each recipe holds to are in [Rules](../Rules/doc.md).

## The loop

```bash
agnos build      # verify + regenerate every generated file + go mod tidy + compile
agnos verify     # the schema check alone, writes nothing
```

`build` is the only thing that regenerates the wiring,
`README.md` and `docs/`, so run it after every hand edit. It is idempotent: a second run leaves
the tree unchanged. Every command below takes `--path <dir>` (default `.`) and `-q`, and runs
`build` for you.

No recipe below asks for a Go file to be created by hand except the cases listed under
[Hand-written code](#hand-written-code).

## Choose what agnos generates

```bash
agnos list-extensions             # every generation mechanic and whether it is on
agnos enable-extension readme     # start generating README.md again
agnos disable-extension doc       # stop generating docs/, keep what is there
```

`AgnosConfig/extensions.yaml` is what `build` reads to decide what to render. Turning a
mechanic off stops the generation and removes nothing: the files stay, and they are yours to
edit. Every key is in [Extensions](../Extensions/doc.md).

## Add the CLI layer

```bash
agnos cli-init     # sandbox/internal/cli, cmd/main, help, version and help-flag, stddeps + argvdeps + stringsdeps + OpinionatedAgnosCli
```

From there `agnos add-command <name> --summary "..." --category "..."` declares a command and
`agnos add-flag` / `add-arg` its fields. `agnos cli-purge` removes the layer again.


## Add the server layer

```bash
agnos server-init      # serverdeps, signaldeps, sandbox/internal/server, the health and openapi routes, start-server
emissario start-server  # listens on the first free port of 3000..4000
```

From there `agnos add-route <name> --pattern '/<path>/{id}'` declares a
route and `agnos add-path` / `add-parameter` / `add-body-field` what it reads. A
project with no CLI gets one first: a server needs a command that starts it.
`agnos server-purge` removes the layer again.


## Add the front layer

```bash
agnos front-init      # the OpinionatedAgnosFront lib, the front route, assets/front/{index,404}.html
emissario start-server  # serves every file of assets/front
```

From there any file under `assets/front/` is served; `agnos add-page <name>`
scaffolds an html one and `remove-page` deletes it. A project with no server layer gets one
first: the front is answered over http. `agnos front-purge` removes the layer
again, leaving `assets/front/` alone.

## Add the database layer

```bash
agnos database-init                     # the store contract, the OpinionatedAgnosDatabase lib, the mechanic on
agnos add-database app-database         # the first database
agnos add-table url --database app-database
```

From there `add-table-field` declares what a table holds and every method it generates is
written for you. `agnos database-purge` removes the layer again.

## Add the backoffice

```bash
agnos backoffice-init   # /admin pages, /api/admin, users, API tokens, backoffice-db
```

It installs the server, front and database layers it is missing, and writes every file once.
`start-server` then reads the session secret from `EMISSARIO_BACKOFFICE_SECRET`, or generates one per run when
it is unset. `agnos backoffice-purge` removes it again.
## Add reusable logic

`sandbox/internal/<pkg>/`, one directory per concern, imported by whatever needs it. No
declaration, no generated counterpart — write the package and run `build`.

## Add a surface to the sandbox api

The api is what a Go caller gets back from `sandbox.New` (see [LibUsage](../LibUsage/doc.md)).
Two hand-written places, then `build` regenerates `sandbox/api/generated.sandbox.go` and
`sandbox/generated.new.go` around them:

1. `sandbox/api/<x>.go` — the contract: `type <X> struct { ... }` of function fields, named
   after the file, every declaration doc-commented (those comments render
   [PublicApi](../PublicApi/doc.md)). It becomes the `api.Sandbox` field `<X>`.
2. `sandbox/internal/<x>/new.go` — `func New<X>(sandbox *api.Sandbox) api.<X>`, assigning each
   field of the contract, with the implementation beside it.

`build` then writes `sandbox/constructors/<x>/constructor.go` — `sandbox.<X> =
<x>.New<X>(sandbox)` — **once**, and `sandbox/generated.new.go` calls it. From there the constructor is
yours: wrap the implementation, decorate the contract, or build a different one entirely.

## Construct a field yourself

`sandbox/generated.new.go` is one `<x>.Constructor(&self)` per directory of `sandbox/constructors/`, so
adding a directory is adding a call. Write `sandbox/constructors/<x>/constructor.go` with
`func Constructor(sandbox *api.Sandbox)` in `package <x>`, run `build`, and it is wired — the
same way a generated one is, and with no generated file to fight over. Editing a constructor
`build` wrote earlier works for the same reason: nothing rewrites it.

## Add a dependency

Everything the sandbox is not allowed to do itself — filesystem, clock, network, subprocess —
arrives through `sandbox.Deps`. Install a ready-made one:

```bash
agnos list-deps                 # every installable contract
agnos add-dep <dep>             # sandbox/deps/<dep>/ + its default adapter + the go.mod require
agnos add-dep <dep> --adapter <adapter>
agnos remove-dep <dep> [--with-adapters]
```

[DepList](../DepList/doc.md) is the catalogue. For one of your own, write the two halves:

1. `sandbox/deps/<x>deps/<x>deps.go` — `type Contract struct { ... }` of function fields, no import at all.
2. `adapters/impls/<impl><x>/<impl><x>.go` — `func Bind(deps *deps.Deps) { deps.<X>Deps = <x>deps.Contract{...} }`,
   any import allowed, beside an `adapter.yaml` saying `dep: <x>deps`.

Then bind it: add `<impl><x>` to `adapters/bindings/standard/binding.yaml`, or let
`agnos add-dep` do both for a dep of the catalogue. Reach it as `sandbox.Deps.<X>Deps`
from anywhere inside `sandbox/`.

One contract may have several adapters — see [Adapters](../Adapters/doc.md).

This project has no `sandbox/deps/` yet: `agnos deps-init` creates it (`deps-purge` removes it).

## Add a doc

```bash
agnos add-doc <Name> --theme <id> --description "one line"    # themes: AgnosConfig/themes.yaml
agnos add-doc <Name>/<Sub> --description "one line"           # sub-doc, no theme
agnos remove-doc <Name>
```

Write `docs/<Name>/doc.md`; `README.md`'s index, and the parent `Index.md` of a sub-doc, are
regenerated. Describe any new path worth naming in `AgnosConfig/structure.yaml` — that
file is what renders [Structure](../Structure/doc.md).

## Add an example

```bash
agnos add-lib-example <name>       # examples/lib/<name>/example.go
agnos run-examples                    # run them all, check each against its golden
agnos run-examples --only <name>      # one example, both sides
agnos update-example <name>           # rewrite that one golden with what it produces now
agnos run-examples --update           # rewrite every golden at once
agnos remove-lib-example <name>
```

Write the example itself, ending with the copy out of `test-dir` into `assert-dir` that says what
it asserts: `result.yaml` records `assert-dir`, and an example that copies nothing out fails.
The golden is written by the first `run-examples` and refreshed with `update-example <name>`, which
prints what it changes before writing. Details in [LibExamples](../LibExamples/doc.md).

## Hand-written code

| File | Written when |
| --- | --- |
| `sandbox/internal/<pkg>/*.go` (never named `generated.*`) | logic worth reusing |
| `sandbox/api/<x>.go` + `sandbox/internal/<x>/new.go` | a new api surface |
| `sandbox/constructors/<x>/constructor.go` | how a field of the `Sandbox` is built |
| `sandbox/deps/<x>/<x>.go` + `adapters/impls/<x>/<x>.go` + its `adapter.yaml` | a new dependency |

Everything else is regenerated over. Two more files are yours: `AgnosConfig/docs/ReadmeHeader.md`
is the whole of `README.md` above the documentation index, and `LICENSE` is pasted verbatim into
its License section — put whatever license you want there.

A project built before the `OpinionatedAgnos<X>` libs keeps hand-written files written against
the generated packages they replaced. `add-dep` the lib of every mechanic that is on (`verify`
names the missing ones); the next `build` removes `sandbox/internal/generated/{cliio,trigger,routeio,frontio,databaseio}`,
`cli/command`, `server/route`, `main.go` and `main.go`; then `verify` names every
hand-written import of them with its replacement — `cliio.Fail(sandbox, …)` is
`sandbox.Deps.OpinionatedAgnosCli.Fail(…)`, `routeio.RequestOf(route)` is `route.Request`, a `Handle*`
logs and then calls `OpinionatedAgnosServer.WriteError(sandbox.Deps.SerializableDeps, …)`, and
`start-server` calls `sandbox.Server.Serve` rather than `server.Main`.

A project built before the `generated.` prefix needs nothing done by hand: its next `build`
writes every generated Go file as `generated.<name>.go` (a remote dep's copy and shim included),
drops the old name beside it, moves `sandbox/internal/generated/{cli,server,config}` to
`sandbox/internal/<x>/` and rewrites the import of each in the project's own files.

## Ship

This project has no `cmd/main` to compile: it ships as the Go module other programs import
(see [LibUsage](../LibUsage/doc.md)). Bump `version` in `AgnosConfig/project.yaml` and tag
the repository; `agnos cli-init` adds a binary if you want one.
