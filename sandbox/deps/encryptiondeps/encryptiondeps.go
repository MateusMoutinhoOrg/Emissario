package encryptiondeps

// This package is the sandbox's *copy* of the api the encryption at rest
// exposes — the same mechanic as hashdeps and jwtdeps, for the same reason: a
// cipher, a key derivation and the process environment are a library and an
// OS-bound effect, which may not appear inside the sandbox. The contract is
// restated here, and the adapter — which lives outside the sandbox — is what
// fills it.
//
// The key never crosses this boundary. The adapter reads it out of the
// environment, removes it from there, and keeps it for the storage every
// database is written through; the sandbox only ever sees its id, the hex of
// four bytes derived from it, which reveals nothing about it. Until the
// storage is armed — Encrypt, Plaintext or Rotate — every read and write of
// every database fails: a command that forgot to decide never writes in
// clear by accident.
//
// Every value the storage writes is sealed whole with XChaCha20-Poly1305 and
// carries a header naming the key it was sealed under:
//
//	0xE1 | key id (4 bytes) | nonce (24 bytes, random) | ciphertext and tag
//
// Its additional data is the path the value lives at relative to the
// --database folder, so a value moved to another field or another record
// fails to open, while a snapshot restored into another folder opens as it
// was. Only builtin types cross this boundary.

// Contract is the encryption at rest injected whole as the
// Deps.EncryptionDeps field.
type Contract struct {
	// GenerateKey answers a fresh key: 32 random bytes in base64, the
	// spelling LoadKey reads back.
	GenerateKey func() (key string, err error)

	// LoadKey reads the key out of the environment variable name, removes
	// that variable from the process environment whatever it held, and
	// keeps the key in memory as the one every value is sealed under.
	// present is false when the variable is unset or empty. The key is 32
	// bytes in base64 or hex; anything else — a passphrase — is an error
	// that never quotes the value. keyId is the hex of the four bytes that
	// name it in every value and in encryption.json.
	LoadKey func(name string) (keyId string, present bool, err error)

	// LoadOldKey reads, the same way, the key a rotation moves away from:
	// values sealed under it are still opened once Rotate arms the storage,
	// and resealed under the one LoadKey read.
	LoadOldKey func(name string) (keyId string, present bool, err error)

	// Encrypt arms the storage: every value is sealed under the key LoadKey
	// read, and a value read back that is not sealed under it is an error.
	// skip is how many leading segments of every key name the --database
	// folder, left out of the additional data.
	Encrypt func(skip int) error

	// Plaintext arms the storage in clear, for --notencrypt: values are
	// written and read as they are.
	Plaintext func()

	// Rotate arms the storage for a rotation: a value sealed under the old
	// key or the new one opens, and every value written is sealed under the
	// new one.
	Rotate func(skip int) error

	// RotateFiles reseals under the new key every file below dir — the
	// --database folder — still sealed under the old one, each written
	// whole to a temporary file and renamed over it, and answers how many
	// it resealed. The store's temporary and lock files, and any file
	// whose name holds a '.', are left alone. A file sealed under neither
	// key is an error naming its path. It runs after Rotate, and running
	// it again after an interruption reseals only what is left.
	RotateFiles func(dir string) (resealed int, err error)

	// Reseal answers value, a whole file as the storage wrote it at path
	// — slash-separated and relative to the --database folder — sealed
	// under the new key: as it is, with changed false, when it already is.
	// It runs after Rotate, on the files a snapshot keeps.
	Reseal func(value []byte, path string) (resealed []byte, changed bool, err error)

	// KeyIdOf answers the id of the key value names in its header, and
	// whether value carries one at all.
	KeyIdOf func(value []byte) (keyId string, encrypted bool)
}
