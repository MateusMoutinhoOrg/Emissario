# `deps.TotpDeps`

`sandbox/deps/totpdeps`

## `Contract`

Contract is the one-time password library injected whole as the Deps.TotpDeps field.

| Field | Type | Description |
| --- | --- | --- |
| `NewSecret` | `func() (secret string, err error)` | NewSecret answers a fresh secret: 20 random bytes in base32, upper case, no padding. |
| `Code` | `func(secret string, unixSeconds int64) (code string, err error)` | Code answers the 6-digit code of secret for the 30-second step the Unix instant unixSeconds falls in. The error reports a secret that is not base32. |
| `Verify` | `func(secret string, code string, unixSeconds int64, skew int) (step int64, ok bool)` | Verify tells whether code is the one of secret for the step of unixSeconds, or of one of the skew steps before or after it — the clock of a phone drifts — and answers the step it matched, which the caller keeps so the same code is never accepted twice. It compares in constant time. |
| `Uri` | `func(issuer string, account string, secret string) string` | Uri is the otpauth://totp/ URI an authenticator app reads secret from, labelled issuer:account. |
| `QrPng` | `func(text string) (png []byte, err error)` | QrPng answers text drawn as a QR code, a PNG image: what the setup page shows, from a route of its own the backoffice serves. |

[every contract](doc.md)
