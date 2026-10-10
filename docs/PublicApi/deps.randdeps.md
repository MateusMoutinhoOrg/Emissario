# `deps.RandDeps`

`sandbox/deps/randdeps`

## `Contract`

Contract is the random source injected whole as the Deps.RandDeps field. Every byte it hands out comes from the operating system's cryptographically secure generator, so what it answers may be used as a secret.

| Field | Type | Description |
| --- | --- | --- |
| `Hex` | `func(bytes int) (string, error)` | Hex returns bytes random bytes, lower-case hexadecimal — a string twice bytes long. The error reports a generator that could not be read. |

[every contract](doc.md)
