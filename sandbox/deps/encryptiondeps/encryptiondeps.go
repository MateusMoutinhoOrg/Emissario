package encryptiondeps

// This package is the sandbox's *copy* of the api the encryption at rest
// exposes — the same mechanic as hashdeps and jwtdeps, for the same reason: a
// cipher, a key derivation and the process environment are a library and an
// OS-bound effect, which may not appear inside the sandbox. The contract is
// restated here, and the adapter — which lives outside the sandbox — is what
// fills it.
//
// The encryption itself is Keep's: a database whose Props carries an
// EncryptKey has every value it writes sealed by Keep's CryptoDeps —
// AES-256-GCM, stored as nonce (12 bytes, random) | ciphertext | tag (16
// bytes) — and every entry of its unique indexes named by a SHA-256 keyed by
// that EncryptKey. The adapter of the databases sets the EncryptKey of every
// Props it opens from the key this contract loaded, and blanks it on the
// Props a database hands back.
//
// The key never crosses this boundary. The adapter reads it out of the
// environment, removes it from there, and keeps the EncryptKey it derives
// from it in memory; the sandbox only ever sees its id, the hex of four bytes
// derived from it, which reveals nothing about it. Until the databases are
// armed — Encrypt, EncryptOld or Plaintext — every database refuses to open:
// a command that forgot to decide never writes in clear by accident. Only
// builtin types cross this boundary.

// Contract is the encryption at rest injected whole as the
// Deps.EncryptionDeps field.
type Contract struct {
	// GenerateKey answers a fresh key: 32 random bytes in base64, the
	// spelling LoadKey reads back.
	GenerateKey func() (key string, err error)

	// LoadKey reads the key out of the environment variable name, removes
	// that variable from the process environment whatever it held, and
	// keeps it in memory as the one Encrypt arms the databases with.
	// present is false when the variable is unset or empty. The key is 32
	// bytes in base64 or hex; anything else — a passphrase — is an error
	// that never quotes the value. keyId is the hex of the four bytes that
	// name it in encryption.json and in every snapshot.
	LoadKey func(name string) (keyId string, present bool, err error)

	// LoadOldKey reads, the same way, the key a rotation moves away from:
	// the one EncryptOld arms the databases with, and Reseal and RotateFiles
	// open what they reseal under.
	LoadOldKey func(name string) (keyId string, present bool, err error)

	// Encrypt arms the databases: every one opened from then on is
	// encrypted by Keep under the key LoadKey read.
	Encrypt func() error

	// EncryptOld arms the databases under the key LoadOldKey read: what a
	// rotation opens them with before it reseals them.
	EncryptOld func() error

	// Plaintext arms the databases in clear, for --notencrypt: every one
	// opened from then on stores its values as they are.
	Plaintext func()

	// Opens reports whether value — a whole file as Keep wrote it — decrypts
	// under the key the databases are armed with. It is false in clear and
	// before the databases are armed.
	Opens func(value []byte) bool

	// Reseal answers value, sealed under the old key, sealed under the new
	// one: as it is, with changed false, when it already opens under the new
	// one. A value that opens under neither is an error that never quotes
	// it. It needs both LoadKey and LoadOldKey.
	Reseal func(value []byte) (resealed []byte, changed bool, err error)

	// RotateFiles walks every file below dir — the --database folder. A file
	// drop accepts — its path slash-separated and relative to dir — is
	// removed; every other one is resealed as Reseal does, written whole to
	// a temporary file and renamed over it. It answers how many files it
	// resealed and how many it removed. The store's temporary and lock
	// files, and any file whose name holds a '.', are left alone. Running it
	// again after an interruption reseals only what is left.
	RotateFiles func(dir string, drop func(path string) bool) (resealed int, dropped int, err error)
}
