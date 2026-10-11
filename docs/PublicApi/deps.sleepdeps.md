# `deps.SleepDeps`

`sandbox/deps/sleepdeps`

## `Contract`

Contract is the waiting half of a clock, injected whole as the Deps.SleepDeps field. The git source sync loop is what reaches for it, to wait between two looks at the sources it keeps pulled.

| Field | Type | Description |
| --- | --- | --- |
| `Sleep` | `func(milliseconds int64)` | Sleep pauses the calling goroutine for at least milliseconds. Zero or a negative value returns at once. |

[every contract](doc.md)
