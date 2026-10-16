# `deps.EncryptionDeps`

`sandbox/deps/encryptiondeps`

## `Contract`

Contract is the encryption at rest injected whole as the Deps.EncryptionDeps field.

| Field | Type | Description |
| --- | --- | --- |
| `GenerateKey` | `func() (key string, err error)` | GenerateKey answers a fresh key: 32 random bytes in base64, the spelling LoadKey reads back. |
| `LoadKey` | `func(name string) (keyId string, present bool, err error)` | LoadKey reads the key out of the environment variable name, removes that variable from the process environment whatever it held, and keeps the key in memory as the one every value is sealed under. present is false when the variable is unset or empty. The key is 32 bytes in base64 or hex; anything else — a passphrase — is an error that never quotes the value. keyId is the hex of the four bytes that name it in every value and in encryption.json. |
| `LoadOldKey` | `func(name string) (keyId string, present bool, err error)` | LoadOldKey reads, the same way, the key a rotation moves away from: values sealed under it are still opened once Rotate arms the storage, and resealed under the one LoadKey read. |
| `Encrypt` | `func(skip int) error` | Encrypt arms the storage: every value is sealed under the key LoadKey read, and a value read back that is not sealed under it is an error. skip is how many leading segments of every key name the --database folder, left out of the additional data. |
| `Plaintext` | `func()` | Plaintext arms the storage in clear, for --notencrypt: values are written and read as they are. |
| `Rotate` | `func(skip int) error` | Rotate arms the storage for a rotation: a value sealed under the old key or the new one opens, and every value written is sealed under the new one. |
| `RotateFiles` | `func(dir string) (resealed int, err error)` | RotateFiles reseals under the new key every file below dir — the --database folder — still sealed under the old one, each written whole to a temporary file and renamed over it, and answers how many it resealed. The store's temporary and lock files, and any file whose name holds a '.', are left alone. A file sealed under neither key is an error naming its path. It runs after Rotate, and running it again after an interruption reseals only what is left. |
| `Reseal` | `func(value []byte, path string) (resealed []byte, changed bool, err error)` | Reseal answers value, a whole file as the storage wrote it at path — slash-separated and relative to the --database folder — sealed under the new key: as it is, with changed false, when it already is. It runs after Rotate, on the files a snapshot keeps. |
| `KeyIdOf` | `func(value []byte) (keyId string, encrypted bool)` | KeyIdOf answers the id of the key value names in its header, and whether value carries one at all. |

[every contract](doc.md)
