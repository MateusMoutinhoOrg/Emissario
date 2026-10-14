# LibUsage

`emissario` is a Go module before it is anything else: every feature lives in `sandbox/`
and is reachable from any Go program that imports it.

```bash
go get github.com/MateusMoutinhoOrg/Emissario@latest
```

## Wiring

`sandbox/` performs no OS effects of its own — filesystem, clock, stdout, processes all
arrive through a `deps.Deps` struct. `adapters/bindings/standard` builds the ready-made
assembly, and `sandbox.New` turns it into the API object, which carries the deps on
`Sandbox.Deps` — so everything inside reaches them through the api it was handed.

```go
package main

import (
	"github.com/MateusMoutinhoOrg/Emissario/adapters/bindings/standard"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox"
)

func main() {
	deps := standard.New()    // every adapter lib bound
	lib := sandbox.New(&deps) // *api.Sandbox

	_ = lib
}
```

## What the sandbox exposes

`*api.Sandbox` is a flat struct, one field per contract declared in `sandbox/api/`.
Everything callable from Go is behind one of them.

| Field | Type |
| --- | --- |
| `lib.Cli` | `api.Cli` |
| `lib.Config` | `api.Config` |
| `lib.Server` | `api.Server` |

`lib.Cli.Commands` (`[]api.Command`) is the command surface itself: every command
the project declares, each carrying its flags, its args and the `Handler` that runs it.
`api.BindCommand(&command)` copies one into the command a single run binds to, so a caller
drives a command without a command line — bind the values into the copy's `Items` and call
`copy.Handler(copy)`.

`lib.Server.Routes` (`[]*api.Route`) is the http surface the same way:
every route the project declares, in run order — lowest `Priority` first — each carrying its
`paths`, its parameters, its body and the `Matches` / `Run` that match and
answer it. `api.BindRoute(route)` copies one into the route a single request runs on, so a
caller drives a route without a socket — set the copy's `Request` and `Response` and call
`copy.Run(copy)`, which returns the failure it did not answer itself and `nil`
otherwise. What it answered with is the status it wrote on the response, never what it returned.

[PublicApi](../PublicApi/doc.md) lists every one of them — signatures, props structs and
dependency contracts — generated from `sandbox/api/` itself on every build.

## Custom deps

Every sub-contract is a struct of function fields, so any of them can be swapped for a
test double, an in-memory implementation or an instrumented wrapper. Patch fields **before**
`sandbox.New(&deps)`: the constructors capture the pointer.

```go
deps := standard.New()

var out bytes.Buffer
deps.StdDeps.Printf = func(f string, a ...any) (int, error) {
	return fmt.Fprintf(&out, f, a...)
}

lib := sandbox.New(&deps)
```

The contracts available to patch:

| Field | Contract package |
| --- | --- |
| `deps.OpinionatedAgnosCli` | `sandbox/deps/OpinionatedAgnosCli` |
| `deps.OpinionatedAgnosDatabase` | `sandbox/deps/OpinionatedAgnosDatabase` |
| `deps.OpinionatedAgnosFront` | `sandbox/deps/OpinionatedAgnosFront` |
| `deps.OpinionatedAgnosServer` | `sandbox/deps/OpinionatedAgnosServer` |
| `deps.ArchiveDeps` | `sandbox/deps/archivedeps` |
| `deps.ArgvDeps` | `sandbox/deps/argvdeps` |
| `deps.DatabaseDeps` | `sandbox/deps/databasedeps` |
| `deps.EmbedDeps` | `sandbox/deps/embeddeps` |
| `deps.EnvDeps` | `sandbox/deps/envdeps` |
| `deps.GeoipDeps` | `sandbox/deps/geoipdeps` |
| `deps.HashDeps` | `sandbox/deps/hashdeps` |
| `deps.IoDeps` | `sandbox/deps/iodeps` |
| `deps.JwtDeps` | `sandbox/deps/jwtdeps` |
| `deps.LocaltimeDeps` | `sandbox/deps/localtimedeps` |
| `deps.PasswordDeps` | `sandbox/deps/passworddeps` |
| `deps.RandDeps` | `sandbox/deps/randdeps` |
| `deps.RatelimitDeps` | `sandbox/deps/ratelimitdeps` |
| `deps.RequestDeps` | `sandbox/deps/requestdeps` |
| `deps.RunDeps` | `sandbox/deps/rundeps` |
| `deps.SerializableDeps` | `sandbox/deps/serializabledeps` |
| `deps.ServerDeps` | `sandbox/deps/serverdeps` |
| `deps.SignalDeps` | `sandbox/deps/signaldeps` |
| `deps.SleepDeps` | `sandbox/deps/sleepdeps` |
| `deps.SortDeps` | `sandbox/deps/sortdeps` |
| `deps.StdDeps` | `sandbox/deps/stddeps` |
| `deps.StringsDeps` | `sandbox/deps/stringsdeps` |
| `deps.TimeDeps` | `sandbox/deps/timedeps` |

Each one is filled by a matching implementation under `adapters/impls/`, every package
exposing the same `Bind(deps *deps.Deps)` entry point:

| Adapter lib | Binder |
| --- | --- |
| `adapters/impls/OpinionatedAgnosCli` | `OpinionatedAgnosCli.Bind(&deps)` |
| `adapters/impls/OpinionatedAgnosDatabase` | `OpinionatedAgnosDatabase.Bind(&deps)` |
| `adapters/impls/OpinionatedAgnosFront` | `OpinionatedAgnosFront.Bind(&deps)` |
| `adapters/impls/OpinionatedAgnosServer` | `OpinionatedAgnosServer.Bind(&deps)` |
| `adapters/impls/cryptorand` | `cryptorand.Bind(&deps)` |
| `adapters/impls/databasedeps` | `databasedeps.Bind(&deps)` |
| `adapters/impls/goembed` | `goembed.Bind(&deps)` |
| `adapters/impls/golangjwt` | `golangjwt.Bind(&deps)` |
| `adapters/impls/maxmindgeoip` | `maxmindgeoip.Bind(&deps)` |
| `adapters/impls/memoryratelimit` | `memoryratelimit.Bind(&deps)` |
| `adapters/impls/nethttprequest` | `nethttprequest.Bind(&deps)` |
| `adapters/impls/nethttpserver` | `nethttpserver.Bind(&deps)` |
| `adapters/impls/osenv` | `osenv.Bind(&deps)` |
| `adapters/impls/osexecrun` | `osexecrun.Bind(&deps)` |
| `adapters/impls/osio` | `osio.Bind(&deps)` |
| `adapters/impls/ossignal` | `ossignal.Bind(&deps)` |
| `adapters/impls/osstd` | `osstd.Bind(&deps)` |
| `adapters/impls/pbkdf2password` | `pbkdf2password.Bind(&deps)` |
| `adapters/impls/sha256hash` | `sha256hash.Bind(&deps)` |
| `adapters/impls/stdargv` | `stdargv.Bind(&deps)` |
| `adapters/impls/stdlocaltime` | `stdlocaltime.Bind(&deps)` |
| `adapters/impls/stdserializable` | `stdserializable.Bind(&deps)` |
| `adapters/impls/stdsleep` | `stdsleep.Bind(&deps)` |
| `adapters/impls/stdsort` | `stdsort.Bind(&deps)` |
| `adapters/impls/stdstrings` | `stdstrings.Bind(&deps)` |
| `adapters/impls/stdtime` | `stdtime.Bind(&deps)` |
| `adapters/impls/ziparchive` | `ziparchive.Bind(&deps)` |

Starting from `standard.New()` is the safe default: an unfilled field is a nil func that
panics on first call. For a permanent mix, write your own
`adapters/bindings/<name>/generated.new.go` binding only the libs you want — `standard/generated.new.go` is
regenerated on every build, while other directories under `bindings/` are left alone.

`sandbox/api` is pure contract and `sandbox/` never touches the OS, so both are safe to import
anywhere; the rest of the rules a caller can count on are in [Rules](../Rules/doc.md#layers),
and [DepList](../DepList/doc.md) lists every contract that can be added.
