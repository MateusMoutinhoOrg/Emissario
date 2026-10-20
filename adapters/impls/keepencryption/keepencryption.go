package keepencryption

import (
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"sync"

	aesgcmcrypto "github.com/MateusMoutinhoOrg/Keep/adapters/impls/aesgcmcrypto"
	keepdeps "github.com/MateusMoutinhoOrg/Keep/sandbox/deps"
	cryptodeps "github.com/MateusMoutinhoOrg/Keep/sandbox/deps/cryptodeps"

	encryptiondeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/encryptiondeps"

	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps"
)

// KeyBytes is how long a key is: 32 random bytes, never a passphrase.
const KeyBytes = 32

// keyIdBytes is how many bytes name a key in encryption.json.
const keyIdBytes = 4

// The purposes HKDF derives one value for each, so one key never does two
// jobs: what Keep encrypts under, and the id that names the key.
const (
	encryptKeyPurpose = "emissario keep encrypt-key"
	keyIdPurpose      = "emissario key-id"
)

// key is one loaded key: the Props.EncryptKey Keep encrypts under, and the
// id that names it. The key itself is never kept.
type key struct {
	encryptKey string
	id         string
}

// mode is how the databases are armed.
type mode int

const (
	// unarmed refuses to open any database: nothing said yet whether the
	// databases are encrypted.
	unarmed mode = iota
	// plain opens every database in clear.
	plain
	// sealed opens every database under the armed key.
	sealed
)

// state is the process-wide state of the encryption: one per process, since
// the databases every command opens are one per process.
var state struct {
	sync.RWMutex
	mode    mode
	armed   *key
	current *key
	old     *key
}

// crypto is Keep's own cipher — the CryptoDeps its aesgcmcrypto adapter
// fills, the one Keep's standard binding binds and every database runs over
// — so what Opens, Reseal and RotateFiles read is what the databases wrote.
var crypto = func() cryptodeps.Contract {
	remote := keepdeps.Deps{}
	aesgcmcrypto.Bind(&remote)
	return remote.CryptoDeps
}()

// errUnarmed is what EncryptKey answers before the databases are armed.
var errUnarmed = errors.New("the database encryption is not set up: no command may open a database before it is")

// EncryptKey answers the Props.EncryptKey every database opened now is built
// with: "" in clear, and an error before the databases are armed.
// encryptedkeep sets it on every Props.
func EncryptKey() (string, error) {
	state.RLock()
	defer state.RUnlock()
	switch state.mode {
	case plain:
		return "", nil
	case sealed:
		return state.armed.encryptKey, nil
	}
	return "", errUnarmed
}

// derive builds the key of master: its EncryptKey and its id, both out of
// HKDF-SHA256, master itself never kept.
func derive(master []byte) (*key, error) {
	encryptKey, err := hkdf.Key(sha256.New, master, nil, encryptKeyPurpose, sha256.Size)
	if err != nil {
		return nil, err
	}
	id, err := hkdf.Key(sha256.New, master, nil, keyIdPurpose, keyIdBytes)
	if err != nil {
		return nil, err
	}
	loaded := &key{encryptKey: base64.StdEncoding.EncodeToString(encryptKey), id: hex.EncodeToString(id)}
	clear(encryptKey)
	return loaded, nil
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
func load(name string) (*key, bool, error) {
	value := os.Getenv(name)
	os.Unsetenv(name)
	if strings.TrimSpace(value) == "" {
		return nil, false, nil
	}
	master, err := parseKey(name, value)
	if err != nil {
		return nil, true, err
	}
	loaded, err := derive(master)
	clear(master)
	if err != nil {
		return nil, true, err
	}
	return loaded, true, nil
}

// generateKey fills encryptiondeps.Contract.GenerateKey.
func generateKey() (string, error) {
	master := make([]byte, KeyBytes)
	if _, err := rand.Read(master); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(master), nil
}

// loadInto reads the key of the variable name into slot, which is
// state.current or state.old.
func loadInto(slot **key, name string) (string, bool, error) {
	loaded, present, err := load(name)
	if err != nil || !present {
		return "", present, err
	}
	state.Lock()
	*slot = loaded
	state.Unlock()
	return loaded.id, true, nil
}

// arm opens every database from now on under loaded, or in clear when to is
// plain.
func arm(to mode, loaded *key) {
	state.Lock()
	defer state.Unlock()
	state.mode = to
	state.armed = loaded
}

// armWith arms the databases under the key in slot, refusing one never
// loaded.
func armWith(slot **key, missing string) error {
	state.RLock()
	loaded := *slot
	state.RUnlock()
	if loaded == nil {
		return errors.New(missing)
	}
	arm(sealed, loaded)
	return nil
}

// opens fills encryptiondeps.Contract.Opens.
func opens(value []byte) bool {
	state.RLock()
	armed, mode := state.armed, state.mode
	state.RUnlock()
	if mode != sealed {
		return false
	}
	_, err := crypto.Decrypt(value, armed.encryptKey)
	return err == nil
}

// rotation is the old key and the new one, read once off the state.
func rotation() (*key, *key, error) {
	state.RLock()
	defer state.RUnlock()
	if state.old == nil || state.current == nil {
		return nil, nil, errors.New("a rotation needs both the old key and the new one")
	}
	return state.old, state.current, nil
}

// reseal fills encryptiondeps.Contract.Reseal.
func reseal(value []byte) ([]byte, bool, error) {
	oldKey, newKey, err := rotation()
	if err != nil {
		return nil, false, err
	}
	return resealWith(value, oldKey, newKey)
}

// resealWith answers value sealed under newKey, and whether that changed it.
func resealWith(value []byte, oldKey *key, newKey *key) ([]byte, bool, error) {
	if _, err := crypto.Decrypt(value, newKey.encryptKey); err == nil {
		return value, false, nil
	}
	opened, err := crypto.Decrypt(value, oldKey.encryptKey)
	if err != nil {
		return nil, false, errors.New("it decrypts under neither the old key nor the new one")
	}
	sealed, err := crypto.Encrypt(opened, newKey.encryptKey)
	clear(opened)
	if err != nil {
		return nil, false, err
	}
	return sealed, true, nil
}

// Bind fills deps.Deps.EncryptionDeps. The key it arms is the one
// EncryptKey hands encryptedkeep, which opens every database under it.
func Bind(deps *deps.Deps) {
	deps.EncryptionDeps = encryptiondeps.Contract{
		GenerateKey: generateKey,
		LoadKey: func(name string) (string, bool, error) {
			return loadInto(&state.current, name)
		},
		LoadOldKey: func(name string) (string, bool, error) {
			return loadInto(&state.old, name)
		},
		Encrypt: func() error {
			return armWith(&state.current, "no key was loaded to encrypt the databases with")
		},
		EncryptOld: func() error {
			return armWith(&state.old, "no old key was loaded to open the databases with")
		},
		Plaintext: func() {
			arm(plain, nil)
		},
		Opens:       opens,
		Reseal:      reseal,
		RotateFiles: rotateFiles,
	}
}
