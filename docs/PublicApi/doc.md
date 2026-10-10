# PublicApi

Every exported symbol of `github.com/MateusMoutinhoOrg/Emissario`, read straight from the contract sources on
every build: `sandbox/api/` is the surface `sandbox.New` returns. Each description is the
doc comment of the declaration itself — change the comment, run `build`, and the page
follows.

One page per contract: the tables below say which page declares a symbol, so open that page
rather than reading the whole surface.

## Entry points

| Symbol | Signature |
| --- | --- |
| `sandbox.New` | `func() *api.Sandbox` |

Implementations live under `sandbox/internal` and are unreachable: every contract is a
struct of function fields, filled by a binder.

## The sandbox api

| Page | Declares |
| --- | --- |
| [`sandbox/api/generated.sandbox.go`](api.sandbox.md) | `Sandbox` |
| [`sandbox/api/generated.config.go`](api.config.md) | `Config` |
| [`sandbox/api/projectconfig.go`](api.projectconfig.md) | `ProjectConfig` |
| [`sandbox/api/projectsandbox.go`](api.projectsandbox.md) | `ProjectSandbox` |
