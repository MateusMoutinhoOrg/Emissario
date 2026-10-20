# `deps.EncryptionDeps`

`sandbox/deps/encryptiondeps`

## `Contract`

Contract is the encryption at rest injected whole as the Deps.EncryptionDeps field.

| Field | Type | Description |
| --- | --- | --- |
| `GenerateKey` | `func() (key string, err error)` | GenerateKey answers a fresh key: 32 random bytes in base64, the spelling LoadKey reads back. |
| `LoadKey` | `func(name string) (keyId string, present bool, err error)` | LoadKey reads the key out of the environment variable name, removes that variable from the process environment whatever it held, and keeps it in memory as the one Encrypt arms the databases with. present is false when the variable is unset or empty. The key is 32 bytes in base64 or hex; anything else — a passphrase — is an error that never quotes the value. keyId is the hex of the four bytes that name it in encryption.json and in every snapshot. |
| `LoadOldKey` | `func(name string) (keyId string, present bool, err error)` | LoadOldKey reads, the same way, the key a rotation moves away from: the one EncryptOld arms the databases with, and Reseal and RotateFiles open what they reseal under. |
| `Encrypt` | `func() error` | Encrypt arms the databases: every one opened from then on is encrypted by Keep under the key LoadKey read. |
| `EncryptOld` | `func() error` | EncryptOld arms the databases under the key LoadOldKey read: what a rotation opens them with before it reseals them. |
| `Plaintext` | `func()` | Plaintext arms the databases in clear, for --notencrypt: every one opened from then on stores its values as they are. |
| `Opens` | `func(value []byte) bool` | Opens reports whether value — a whole file as Keep wrote it — decrypts under the key the databases are armed with. It is false in clear and before the databases are armed. |
| `Reseal` | `func(value []byte) (resealed []byte, changed bool, err error)` | Reseal answers value, sealed under the old key, sealed under the new one: as it is, with changed false, when it already opens under the new one. A value that opens under neither is an error that never quotes it. It needs both LoadKey and LoadOldKey. |
| `RotateFiles` | `func(dir string, drop func(path string) bool) (resealed int, dropped int, err error)` | RotateFiles walks every file below dir — the --database folder. A file drop accepts — its path slash-separated and relative to dir — is removed; every other one is resealed as Reseal does, written whole to a temporary file and renamed over it. It answers how many files it resealed and how many it removed. The store's temporary and lock files, and any file whose name holds a '.', are left alone. Running it again after an interruption reseals only what is left. |

[every contract](doc.md)
