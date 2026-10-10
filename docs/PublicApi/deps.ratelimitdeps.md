# `deps.RatelimitDeps`

`sandbox/deps/ratelimitdeps`

## `Contract`

Contract is the rate limiter injected whole as the Deps.RatelimitDeps field. Every function is safe to call from requests served concurrently.

| Field | Type | Description |
| --- | --- | --- |
| `Hit` | `func(key string, windowSeconds int64) int` | Hit records one hit under key, in a window of windowSeconds, and returns how many hits the window holds with it included. |
| `Count` | `func(key string, windowSeconds int64) int` | Count returns how many hits key holds in its open window, 0 when it has none or its window of windowSeconds closed. It records nothing. |
| `Reset` | `func(key string)` | Reset forgets every hit of key. |
| `Undo` | `func(key string)` | Undo takes back one hit of key from its open window, so an attempt recorded with Hit before it was checked — the only way a limit holds against requests served at the same time — can be withdrawn once it turned out not to count. A key with no open window, or none left, stays as it is. |

[every contract](doc.md)
