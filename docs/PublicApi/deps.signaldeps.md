# `deps.SignalDeps`

`sandbox/deps/signaldeps`

## `Contract`

Contract is the signal library injected whole as the Deps.SignalDeps field.

| Field | Type | Description |
| --- | --- | --- |
| `OnInterrupt` | `func(handler func())` | OnInterrupt calls handler once, on a goroutine of its own, the first time the process is asked to stop — an interrupt (Ctrl+C) or a termination request. It returns at once. A second request to stop, while handler is still running, ends the process the way it would have ended without OnInterrupt. |

[every contract](doc.md)
