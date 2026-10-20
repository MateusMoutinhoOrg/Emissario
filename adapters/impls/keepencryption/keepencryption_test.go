package keepencryption

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const keyA = "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8="

const keyAHex = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

const keyB = "202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f"

// reset puts the process-wide state back to unarmed with no key.
func reset(t *testing.T) {
	t.Helper()
	state.Lock()
	state.mode = unarmed
	state.armed = nil
	state.current = nil
	state.old = nil
	state.Unlock()
}

// loaded loads value into slot through the environment, answering the key.
func loaded(t *testing.T, slot **key, value string) *key {
	t.Helper()
	t.Setenv("TEST_ENCRYPT_KEY", value)
	if _, present, err := loadInto(slot, "TEST_ENCRYPT_KEY"); err != nil || !present {
		t.Fatalf("loadInto: present %v err %v", present, err)
	}
	return *slot
}

func TestLoadKeyUnsetsTheEnvironment(t *testing.T) {
	reset(t)
	t.Setenv("TEST_ENCRYPT_KEY", keyA)
	id, present, err := loadInto(&state.current, "TEST_ENCRYPT_KEY")
	if err != nil || !present || len(id) != 2*keyIdBytes {
		t.Fatalf("loadInto: id %q present %v err %v", id, present, err)
	}
	if _, set := os.LookupEnv("TEST_ENCRYPT_KEY"); set {
		t.Fatal("the key is still in the environment")
	}
}

func TestLoadKeyRefusesAPassphraseWithoutQuotingIt(t *testing.T) {
	reset(t)
	t.Setenv("TEST_ENCRYPT_KEY", "correct horse battery staple")
	_, present, err := loadInto(&state.current, "TEST_ENCRYPT_KEY")
	if err == nil || !present {
		t.Fatalf("a passphrase was accepted: present %v err %v", present, err)
	}
	if strings.Contains(err.Error(), "horse") {
		t.Fatalf("the error quotes the value: %v", err)
	}
}

func TestHexAndBase64SpellTheSameKey(t *testing.T) {
	reset(t)
	fromBase64 := loaded(t, &state.current, keyA)
	fromHex := loaded(t, &state.old, keyAHex)
	if fromBase64.id != fromHex.id || fromBase64.encryptKey != fromHex.encryptKey {
		t.Fatal("the hex and the base64 spelling of one key load as two keys")
	}
}

func TestEncryptKeyFollowsTheArming(t *testing.T) {
	reset(t)
	if _, err := EncryptKey(); err == nil {
		t.Fatal("an unarmed EncryptKey answered a key")
	}
	current := loaded(t, &state.current, keyA)
	if err := armWith(&state.current, "missing"); err != nil {
		t.Fatal(err)
	}
	encryptKey, err := EncryptKey()
	if err != nil || encryptKey != current.encryptKey || encryptKey == keyA {
		t.Fatalf("sealed EncryptKey: %q %v", encryptKey, err)
	}
	arm(plain, nil)
	if encryptKey, err := EncryptKey(); err != nil || encryptKey != "" {
		t.Fatalf("plain EncryptKey: %q %v", encryptKey, err)
	}
	if err := armWith(&state.old, "no old key"); err == nil {
		t.Fatal("armed under an old key never loaded")
	}
}

func TestOpensOnlyUnderTheArmedKey(t *testing.T) {
	reset(t)
	current := loaded(t, &state.current, keyA)
	other := loaded(t, &state.old, keyB)
	sealed, err := crypto.Encrypt([]byte("value"), current.encryptKey)
	if err != nil {
		t.Fatal(err)
	}
	if opens(sealed) {
		t.Fatal("opens before the databases are armed")
	}
	armWith(&state.current, "missing")
	if !opens(sealed) {
		t.Fatal("a value sealed under the armed key does not open")
	}
	foreign, _ := crypto.Encrypt([]byte("value"), other.encryptKey)
	if opens(foreign) || opens([]byte("value")) {
		t.Fatal("a value sealed under another key, or in clear, opens")
	}
}

func TestReseal(t *testing.T) {
	reset(t)
	if _, _, err := reseal([]byte("x")); err == nil {
		t.Fatal("resealed without both keys")
	}
	oldKey := loaded(t, &state.old, keyA)
	newKey := loaded(t, &state.current, keyB)
	sealed, _ := crypto.Encrypt([]byte("value"), oldKey.encryptKey)

	resealed, changed, err := reseal(sealed)
	if err != nil || !changed {
		t.Fatalf("reseal: %v %v", changed, err)
	}
	if opened, err := crypto.Decrypt(resealed, newKey.encryptKey); err != nil || string(opened) != "value" {
		t.Fatalf("resealed value: %q %v", opened, err)
	}
	if again, changed, err := reseal(resealed); err != nil || changed || !bytes.Equal(again, resealed) {
		t.Fatalf("a value already under the new key was resealed: %v %v", changed, err)
	}
	if _, _, err := reseal([]byte("in clear, under no key at all")); err == nil {
		t.Fatal("a value under neither key was resealed")
	}
}

func TestRotateFiles(t *testing.T) {
	reset(t)
	oldKey := loaded(t, &state.old, keyA)
	newKey := loaded(t, &state.current, keyB)
	dir := t.TempDir()
	write := func(path string, value []byte) {
		t.Helper()
		full := filepath.Join(dir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, value, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	seal := func(value string) []byte {
		sealed, err := crypto.Encrypt([]byte(value), oldKey.encryptKey)
		if err != nil {
			t.Fatal(err)
		}
		return sealed
	}
	values := map[string]string{
		"db/user/1/values/name":      "ana",
		"db/user/size":               "1",
		"backup/blob/1/values/value": "a snapshot file",
	}
	for path, value := range values {
		write(path, seal(value))
	}
	write("db/user/keys/name/aa", seal("1"))
	write("db/user/1/values/name.keeplock", []byte("lock"))
	write("encryption.json", []byte("{}"))

	drop := func(path string) bool {
		return strings.Contains(path, "/keys/")
	}
	resealed, dropped, err := rotateFiles(dir, drop)
	if err != nil || resealed != len(values) || dropped != 1 {
		t.Fatalf("rotateFiles: resealed %d dropped %d err %v", resealed, dropped, err)
	}
	for path, value := range values {
		onDisk, _ := os.ReadFile(filepath.Join(dir, filepath.FromSlash(path)))
		opened, err := crypto.Decrypt(onDisk, newKey.encryptKey)
		if err != nil || string(opened) != value {
			t.Fatalf("%s after the rotation: %q %v", path, opened, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "db", "user", "keys", "name", "aa")); !os.IsNotExist(err) {
		t.Fatal("the index entry is still there")
	}
	if lock, _ := os.ReadFile(filepath.Join(dir, "db", "user", "1", "values", "name.keeplock")); string(lock) != "lock" {
		t.Fatal("a lock file was touched")
	}

	again, droppedAgain, err := rotateFiles(dir, drop)
	if err != nil || again != 0 || droppedAgain != 0 {
		t.Fatalf("a second rotateFiles resealed %d and dropped %d: %v", again, droppedAgain, err)
	}

	write("db/user/2/values/name", []byte("in clear"))
	if _, _, err := rotateFiles(dir, drop); err == nil || !strings.Contains(err.Error(), "db/user/2/values/name") {
		t.Fatalf("a file under neither key: %v", err)
	}
}
