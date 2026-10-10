# `deps.PasswordDeps`

`sandbox/deps/passworddeps`

## `Contract`

Contract is the password hashing library injected whole as the Deps.PasswordDeps field.

| Field | Type | Description |
| --- | --- | --- |
| `Hash` | `func(password string) (string, error)` | Hash derives a hash of password to store, over a fresh random salt, spelled with everything Verify needs to check a password against it. The error reports a random source that could not be read. |
| `Verify` | `func(hash string, password string) (bool, error)` | Verify tells whether password is the one hash was derived from, comparing in constant time. The error reports a hash it cannot read. |

[every contract](doc.md)
