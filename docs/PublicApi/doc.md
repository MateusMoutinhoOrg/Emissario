# PublicApi

Every exported symbol of `github.com/MateusMoutinhoOrg/Emissario`, read straight from the contract sources on
every build: `sandbox/api/` is the surface `sandbox.New` returns, `sandbox/deps/`
the contracts an adapter fills and a caller may replace. Each description is the
doc comment of the declaration itself — change the comment, run `build`, and the page
follows.

One page per contract: the tables below say which page declares a symbol, so open that page
rather than reading the whole surface.

## Entry points

| Symbol | Signature |
| --- | --- |
| `sandbox.New` | `func(deps *deps.Deps) *api.Sandbox` |
| `standard.New` | `func() deps.Deps` (`adapters/bindings/standard`) |

Implementations live under `sandbox/internal` and are unreachable: every contract is a
struct of function fields, filled by a binder.

## The sandbox api

| Page | Declares |
| --- | --- |
| [`sandbox/api/generated.sandbox.go`](api.sandbox.md) | `Sandbox` |
| [`sandbox/api/backofficeconfig.go`](api.backofficeconfig.md) | `BackofficeConfig` |
| [`sandbox/api/generated.cli.go`](api.cli.md) | `ExitOk`, `ExitFailure`, `ExitUsage`, `Cli` |
| [`sandbox/api/generated.clisandbox.go`](api.clisandbox.md) | `CliSandbox` |
| [`sandbox/api/generated.command.go`](api.command.md) | `ArgString`, `ArgInteger`, `ArgNumber`, `ArgUuid`, `FlagString`, `FlagInteger`, `FlagNumber`, `FlagBoolean`, `FlagStringArray`, `FlagIntegerArray`, `FailureHandler`, `FailureNotFound`, `FailureBadUsage`, `FailureUnknownFlag`, `FailureUnexpectedArg`, `ArgType`, `CommandArg`, `FlagType`, `CommandFlag`, `CommandResponse`, `CommandFailureKind`, `CommandFailure`, `Command` |
| [`sandbox/api/generated.config.go`](api.config.md) | `Config` |
| [`sandbox/api/generated.databaseconfig.go`](api.databaseconfig.md) | `DefaultDatabaseDir`, `DatabaseConfig` |
| [`sandbox/api/generated.route.go`](api.route.md) | `PathString`, `PathInteger`, `PathNumber`, `PathUuid`, `SourceHeader`, `SourceQuery`, `SourceCookie`, `ParameterString`, `ParameterNumber`, `ParameterBoolean`, `ParameterDateTime`, `ParameterStringArray`, `ParameterInteger`, `ParameterIntegerArray`, `AnyMethod`, `PathType`, `Path`, `ParameterSource`, `ParameterType`, `Parameter`, `RouteBody`, `RouteFailure`, `Route` |
| [`sandbox/api/generated.server.go`](api.server.md) | `StatusOK`, `StatusCreated`, `StatusNoContent`, `StatusMovedPermanently`, `StatusFound`, `StatusSeeOther`, `StatusNotModified`, `StatusTemporaryRedirect`, `StatusPermanentRedirect`, `StatusBadRequest`, `StatusUnauthorized`, `StatusForbidden`, `StatusNotFound`, `StatusMethodNotAllowed`, `StatusConflict`, `StatusPayloadTooLarge`, `StatusUnsupportedMediaType`, `StatusUnprocessableEntity`, `StatusTooManyRequests`, `StatusInternalServerError`, `StatusServiceUnavailable`, `Server`, `ServeProps` |
| [`sandbox/api/generated.serversandbox.go`](api.serversandbox.md) | `ServerSandbox` |
| [`sandbox/api/generated.trigger.go`](api.trigger.md) | `TriggerEqual`, `TriggerPrefix`, `TriggerTextPrefix`, `TriggerSuffix`, `TriggerRegex`, `TriggerOneOf`, `TriggerType`, `Trigger` |
| [`sandbox/api/projectconfig.go`](api.projectconfig.md) | `DefaultFontsDir`, `ProjectConfig` |
| [`sandbox/api/projectsandbox.go`](api.projectsandbox.md) | `ProjectSandbox` |
| [`sandbox/api/visualizationconfig.go`](api.visualizationconfig.md) | `VisualizationConfig` |

## Dependency contracts

`deps.Deps` has one field per directory of `sandbox/deps/`, named by title-casing it. Each
field is that package's `Sandbox` struct, filled by `adapters/impls/<name>.Bind(&deps)`.

| Page | Declares |
| --- | --- |
| [`deps.OpinionatedAgnosCli`](deps.OpinionatedAgnosCli.md) | `TriggerEqual`, `TriggerPrefix`, `TriggerTextPrefix`, `TriggerSuffix`, `TriggerRegex`, `TriggerOneOf`, `ArgString`, `ArgInteger`, `ArgNumber`, `ArgUuid`, `FlagString`, `FlagInteger`, `FlagNumber`, `FlagBoolean`, `FlagStringArray`, `FlagIntegerArray`, `FailureHandler`, `FailureNotFound`, `FailureBadUsage`, `FailureUnknownFlag`, `FailureUnexpectedArg`, `ExitOk`, `ExitFailure`, `ExitUsage`, `TriggerType`, `Trigger`, `ArgType`, `CommandArg`, `FlagType`, `CommandFlag`, `CommandResponse`, `CommandFailureKind`, `CommandFailure`, `Command`, `Cli`, `MainProps`, `Contract`, `Error` |
| [`deps.OpinionatedAgnosDatabase`](deps.OpinionatedAgnosDatabase.md) | `Contract` |
| [`deps.OpinionatedAgnosFront`](deps.OpinionatedAgnosFront.md) | `Root`, `Index`, `NotFound`, `RevalidateCache`, `Contract` |
| [`deps.OpinionatedAgnosServer`](deps.OpinionatedAgnosServer.md) | `PathString`, `PathInteger`, `PathNumber`, `PathUuid`, `SourceHeader`, `SourceQuery`, `SourceCookie`, `ParameterString`, `ParameterNumber`, `ParameterBoolean`, `ParameterDateTime`, `ParameterStringArray`, `ParameterInteger`, `ParameterIntegerArray`, `AnyMethod`, `StatusOK`, `StatusCreated`, `StatusNoContent`, `StatusMovedPermanently`, `StatusFound`, `StatusSeeOther`, `StatusNotModified`, `StatusTemporaryRedirect`, `StatusPermanentRedirect`, `StatusBadRequest`, `StatusUnauthorized`, `StatusForbidden`, `StatusNotFound`, `StatusMethodNotAllowed`, `StatusConflict`, `StatusPayloadTooLarge`, `StatusUnsupportedMediaType`, `StatusUnprocessableEntity`, `StatusTooManyRequests`, `StatusInternalServerError`, `StatusServiceUnavailable`, `PathType`, `Path`, `ParameterSource`, `ParameterType`, `Parameter`, `RouteBody`, `RouteFailure`, `Route`, `Server`, `ServeProps`, `MainProps`, `Contract`, `Error` |
| [`deps.ArchiveDeps`](deps.archivedeps.md) | `File`, `Contract` |
| [`deps.ArgvDeps`](deps.argvdeps.md) | `Contract`, `Parser` |
| [`deps.DatabaseDeps`](deps.databasedeps.md) | `Config`, `Key`, `Int`, `Nested`, `Float`, `String`, `Link`, `Bytes`, `KeyConflict`, `NoValue`, `MissingField`, `InvalidField`, `Internal`, `Removed`, `InvalidSchema`, `InvalidArgument`, `Field`, `Schema`, `Props`, `Error`, `Record`, `Collection`, `Database`, `Databases`, `Info`, `ProjectConfig`, `ProjectSandbox`, `Sandbox` |
| [`deps.EmbedDeps`](deps.embeddeps.md) | `Contract` |
| [`deps.EnvDeps`](deps.envdeps.md) | `Contract` |
| [`deps.GeoipDeps`](deps.geoipdeps.md) | `Contract`, `Database` |
| [`deps.HashDeps`](deps.hashdeps.md) | `Contract` |
| [`deps.IoDeps`](deps.iodeps.md) | `Contract` |
| [`deps.JwtDeps`](deps.jwtdeps.md) | `Contract`, `Claims` |
| [`deps.LocaltimeDeps`](deps.localtimedeps.md) | `Contract` |
| [`deps.PasswordDeps`](deps.passworddeps.md) | `Contract` |
| [`deps.RandDeps`](deps.randdeps.md) | `Contract` |
| [`deps.RatelimitDeps`](deps.ratelimitdeps.md) | `Contract` |
| [`deps.RequestDeps`](deps.requestdeps.md) | `Contract`, `Request`, `Response` |
| [`deps.RunDeps`](deps.rundeps.md) | `Contract`, `RunProps`, `Result` |
| [`deps.SerializableDeps`](deps.serializabledeps.md) | `SerializableObject`, `Contract` |
| [`deps.ServerDeps`](deps.serverdeps.md) | `Contract`, `ServerProps`, `Server`, `Request`, `Response` |
| [`deps.SignalDeps`](deps.signaldeps.md) | `Contract` |
| [`deps.SleepDeps`](deps.sleepdeps.md) | `Contract` |
| [`deps.SortDeps`](deps.sortdeps.md) | `Contract` |
| [`deps.StdDeps`](deps.stddeps.md) | `Contract` |
| [`deps.StringsDeps`](deps.stringsdeps.md) | `Contract` |
| [`deps.TimeDeps`](deps.timedeps.md) | `Contract` |
