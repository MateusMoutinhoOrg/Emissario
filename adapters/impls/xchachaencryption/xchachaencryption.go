package xchachaencryption

import (
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/chacha20poly1305"

	encryptiondeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/encryptiondeps"

	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps"
)

// KeyBytes is how long a key is: 32 random bytes, never a passphrase.
const KeyBytes = 32

// header is the first byte of every sealed value.
const header byte = 0xE1

// keyIdBytes is how many bytes of the header name the key.
const keyIdBytes = 4

// overhead is everything a sealed value carries beside its plaintext: the
// header byte, the key id, the nonce and the tag.
const overhead = 1 + keyIdBytes + chacha20poly1305.NonceSizeX + chacha20poly1305.Overhead

// aadPrefix starts the additional data of every value, so a value of this
// format never opens as one of another.
const aadPrefix = "emissario/encryption/v1/"

// The purposes HKDF derives one subkey for each, so one key never does two
// jobs.
const (
	storagePurpose  = "emissario storage"
	keyCheckPurpose = "emissario key-check"
)

// keyring is one loaded key: its id and the cipher of its storage subkey.
type keyring struct {
	id   [keyIdBytes]byte
	aead cipher.AEAD
}

// mode is how the storage is armed.
type mode int

const (
	// unarmed refuses every call: nothing said yet whether the databases are
	// encrypted.
	unarmed mode = iota
	// plain writes and reads values as they are.
	plain
	// sealed seals every value under the current key.
	sealed
	// rotating opens values under the old key or the current one, and seals
	// under the current one.
	rotating
)

// state is the process-wide state of the encryption: one per process, since
// the storage every database goes through is one per process.
var state struct {
	sync.RWMutex
	mode    mode
	skip    int
	current *keyring
	old     *keyring
}

// errUnarmed is what every storage call answers before the storage is armed.
var errUnarmed = errors.New("the database encryption is not set up: no command may read or write a database before it is")

// derive builds the keyring of master: the storage subkey and the key id,
// both out of HKDF-SHA256, master itself never kept.
func derive(master []byte) (*keyring, error) {
	storage, err := hkdf.Key(sha256.New, master, nil, storagePurpose, chacha20poly1305.KeySize)
	if err != nil {
		return nil, err
	}
	check, err := hkdf.Key(sha256.New, master, nil, keyCheckPurpose, sha256.Size)
	if err != nil {
		return nil, err
	}
	aead, err := chacha20poly1305.NewX(storage)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, check)
	mac.Write([]byte("emissario"))
	ring := &keyring{aead: aead}
	copy(ring.id[:], mac.Sum(nil)[:keyIdBytes])
	clear(storage)
	clear(check)
	return ring, nil
}

// parseKey reads a key spelled in hex or in base64 (standard or url, padded
// or not), refusing anything that is not exactly KeyBytes bytes. The error
// never quotes the value.
func parseKey(name string, value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if len(value) == 2*KeyBytes {
		if decoded, err := hex.DecodeString(value); err == nil {
			return decoded, nil
		}
	}
	for _, encoding := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		decoded, err := encoding.DecodeString(value)
		if err == nil && len(decoded) == KeyBytes {
			return decoded, nil
		}
	}
	return nil, errors.New(name + " is not a key: it must be 32 random bytes in base64 or hex, never a passphrase; generate one with `emissario generate-encrypt-key`")
}

// load reads the key out of the environment variable name and removes the
// variable from the process environment, whatever it held, so no child
// process — git and its hooks — inherits it.
func load(name string) (*keyring, bool, error) {
	value := os.Getenv(name)
	os.Unsetenv(name)
	if strings.TrimSpace(value) == "" {
		return nil, false, nil
	}
	master, err := parseKey(name, value)
	if err != nil {
		return nil, true, err
	}
	ring, err := derive(master)
	clear(master)
	if err != nil {
		return nil, true, err
	}
	return ring, true, nil
}

// generateKey fills encryptiondeps.Contract.GenerateKey.
func generateKey() (string, error) {
	key := make([]byte, KeyBytes)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// loadKey fills encryptiondeps.Contract.LoadKey.
func loadKey(name string) (string, bool, error) {
	ring, present, err := load(name)
	if err != nil || !present {
		return "", present, err
	}
	state.Lock()
	state.current = ring
	state.Unlock()
	return hex.EncodeToString(ring.id[:]), true, nil
}

// loadOldKey fills encryptiondeps.Contract.LoadOldKey.
func loadOldKey(name string) (string, bool, error) {
	ring, present, err := load(name)
	if err != nil || !present {
		return "", present, err
	}
	state.Lock()
	state.old = ring
	state.Unlock()
	return hex.EncodeToString(ring.id[:]), true, nil
}

// arm sets the mode the storage runs in.
func arm(to mode, skip int) error {
	state.Lock()
	defer state.Unlock()
	if to == sealed && state.current == nil {
		return errors.New("no key was loaded to encrypt the databases with")
	}
	if to == rotating && (state.current == nil || state.old == nil) {
		return errors.New("a rotation needs both the old key and the new one")
	}
	if skip < 0 {
		skip = 0
	}
	state.mode = to
	state.skip = skip
	return nil
}

// keyIdOf fills encryptiondeps.Contract.KeyIdOf.
func keyIdOf(value []byte) (string, bool) {
	if len(value) < overhead || value[0] != header {
		return "", false
	}
	return hex.EncodeToString(value[1 : 1+keyIdBytes]), true
}

// Bind fills deps.Deps.EncryptionDeps. The storage it arms is the one Wrap
// hands encryptedkeep, which builds every database over it.
func Bind(deps *deps.Deps) {
	deps.EncryptionDeps = encryptiondeps.Contract{
		GenerateKey: generateKey,
		LoadKey:     loadKey,
		LoadOldKey:  loadOldKey,
		Encrypt: func(skip int) error {
			return arm(sealed, skip)
		},
		Plaintext: func() {
			arm(plain, 0)
		},
		Rotate: func(skip int) error {
			return arm(rotating, skip)
		},
		RotateFiles: rotateFiles,
		Reseal:      reseal,
		KeyIdOf:     keyIdOf,
	}
}
