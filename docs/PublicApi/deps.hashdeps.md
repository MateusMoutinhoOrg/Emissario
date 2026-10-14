# `deps.HashDeps`

`sandbox/deps/hashdeps`

## `Contract`

Contract is the hashing library injected whole as the Deps.HashDeps field.

| Field | Type | Description |
| --- | --- | --- |
| `Sha256Hex` | `func(content []byte) string` | Sha256Hex returns the SHA-256 digest of content, lower-case hexadecimal. It is what every recorded example tree is compared by, so the encoding is part of the golden and may not change. |
| `HmacSha256Hex` | `func(key []byte, content []byte) string` | HmacSha256Hex returns the HMAC-SHA256 of content under key, lower-case hexadecimal. |
| `Equal` | `func(a []byte, b []byte) bool` | Equal tells whether a and b hold the same bytes, in a time that does not depend on where they differ — how a signature is checked. |

[every contract](doc.md)
