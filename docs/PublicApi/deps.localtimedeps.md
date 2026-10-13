# `deps.LocaltimeDeps`

`sandbox/deps/localtimedeps`

## `Contract`

Contract is the local time zone of the process, injected whole as the Deps.LocaltimeDeps field. It is named after the capability, not after whoever reads it, so any part of the project that counts days the way the server's clock does reuses it.

| Field | Type | Description |
| --- | --- | --- |
| `Offset` | `func(seconds int64) int64` | Offset answers how many seconds the server's time zone is ahead of UTC at the instant seconds — -10800 for America/Sao_Paulo, 3600 for Europe/Paris in winter and 7200 in summer: a daylight-saving change is in it. |
| `Name` | `func(seconds int64) string` | Name answers the abbreviation of the server's time zone at the instant seconds ("-03", "CEST", "UTC"). |

[every contract](doc.md)
