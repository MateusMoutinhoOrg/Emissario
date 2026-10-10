# `deps.JwtDeps`

`sandbox/deps/jwtdeps`

## `Contract`

Contract is the JSON Web Token library injected whole as the Deps.JwtDeps field. Every token it signs and every token it accepts is HS256, keyed by the secret handed to the call.

| Field | Type | Description |
| --- | --- | --- |
| `Sign` | `func(claims Claims, secret string) (string, error)` | Sign returns the compact HS256 token carrying claims, signed with secret. The error reports an empty secret or a claims set the library could not encode. |
| `Parse` | `func(token string, secret string) (Claims, error)` | Parse checks token against secret and returns the claims it carries. The error reports a token that is malformed, signed by another key or another algorithm than HS256, or whose ExpiresAt has passed — so a nil error is a token that is valid right now. |

## `Claims`

Claims is the set of registered claims a token carries, plus the private `ip` claim.

| Field | Type | Description |
| --- | --- | --- |
| `Id` | `string` | Id is the `jti` claim: the one token among every token issued. |
| `Subject` | `string` | Subject is the `sub` claim: who the token was issued for. |
| `IssuedAt` | `int64` | IssuedAt is the `iat` claim, in seconds since the Unix epoch. |
| `ExpiresAt` | `int64` | ExpiresAt is the `exp` claim, in seconds since the Unix epoch. A token is refused by Parse from that instant on. |
| `Ip` | `string` | Ip is the private `ip` claim: the client ip the token was issued to. |

[every contract](doc.md)
