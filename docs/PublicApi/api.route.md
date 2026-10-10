# `sandbox/api/generated.route.go`

| Constant | Value | Description |
| --- | --- | --- |
| `PathString` | `opinionatedagnosserver.PathString` | PathString takes any slice, bound as a string. |
| `PathInteger` | `opinionatedagnosserver.PathInteger` | PathInteger takes one segment reading as a whole number, bound as an int. |
| `PathNumber` | `opinionatedagnosserver.PathNumber` | PathNumber takes one segment reading as a number, bound as a float64. |
| `PathUuid` | `opinionatedagnosserver.PathUuid` | PathUuid takes one segment reading as a canonical uuid, bound as a string. |
| `SourceHeader` | `opinionatedagnosserver.SourceHeader` | SourceHeader reads a request header, matched without regard to case. |
| `SourceQuery` | `opinionatedagnosserver.SourceQuery` | SourceQuery reads a query-string parameter. |
| `SourceCookie` | `opinionatedagnosserver.SourceCookie` | SourceCookie reads a request cookie. |
| `ParameterString` | `opinionatedagnosserver.ParameterString` | ParameterString is bound as a string. |
| `ParameterNumber` | `opinionatedagnosserver.ParameterNumber` | ParameterNumber is bound as a float64. |
| `ParameterBoolean` | `opinionatedagnosserver.ParameterBoolean` | ParameterBoolean is bound as a bool: true/1 or false/0. |
| `ParameterDateTime` | `opinionatedagnosserver.ParameterDateTime` | ParameterDateTime is bound as a string that has to read as RFC 3339. |
| `ParameterStringArray` | `opinionatedagnosserver.ParameterStringArray` | ParameterStringArray is bound as a []string. |
| `ParameterInteger` | `opinionatedagnosserver.ParameterInteger` | ParameterInteger is bound as an int. |
| `ParameterIntegerArray` | `opinionatedagnosserver.ParameterIntegerArray` | ParameterIntegerArray is bound as a []int. |
| `AnyMethod` | `opinionatedagnosserver.AnyMethod` | AnyMethod is the one entry of Methods that accepts every http method. |

## `PathType`

PathType is what one segment a Path reads has to convert to.

`type PathType = opinionatedagnosserver.PathType`

## `Path`

Path is one entry of `paths` in route.yaml.

`type Path = opinionatedagnosserver.Path`

## `ParameterSource`

ParameterSource is one place of the request a Parameter is read from.

`type ParameterSource = opinionatedagnosserver.ParameterSource`

## `ParameterType`

ParameterType is the type a Parameter is converted to before it reaches Input.

`type ParameterType = opinionatedagnosserver.ParameterType`

## `Parameter`

Parameter is one entry of `parameters` in route.yaml.

`type Parameter = opinionatedagnosserver.Parameter`

## `RouteBody`

RouteBody is the request body a route declares — the parsed form of `body:` in its route.yaml.

`type RouteBody = opinionatedagnosserver.RouteBody`

## `RouteFailure`

RouteFailure is one way a request did not get answered; a Handle refuses a request by returning one, built by Deps.OpinionatedAgnosServer.Fail.

`type RouteFailure = opinionatedagnosserver.RouteFailure`

## `Route`

Route is one http route of the project: the whole of what its route.yaml declares, plus the handler behind it.

`type Route = opinionatedagnosserver.Route`

[every contract](doc.md)
