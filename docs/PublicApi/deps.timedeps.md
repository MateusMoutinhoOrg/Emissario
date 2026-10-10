# `deps.TimeDeps`

`sandbox/deps/timedeps`

## `Contract`

Contract is the calendar library injected whole as the Deps.TimeDeps field.

| Field | Type | Description |
| --- | --- | --- |
| `FormatUnix` | `func(seconds int64, layout string) string` | FormatUnix returns the instant seconds, in UTC, spelled by layout. |
| `ParseUnix` | `func(layout string, value string) (int64, error)` | ParseUnix reads value as a UTC date spelled by layout and returns it in seconds since the Unix epoch. The error reports a value that does not fit layout. |

[every contract](doc.md)
