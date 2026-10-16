package xchachaencryption

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MateusMoutinhoOrg/Keep/adapters/impls/filestorage"
	storagedeps "github.com/MateusMoutinhoOrg/Keep/sandbox/deps/storagedeps"
)

const keyA = "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8="

const keyB = "202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f"

// reset puts the process-wide state back to unarmed with no key.
func reset(t *testing.T) {
	t.Helper()
	state.Lock()
	state.mode = unarmed
	state.skip = 0
	state.current = nil
	state.old = nil
	state.Unlock()
}

// armed loads key as the current one and arms the storage over a fresh
// folder, answering the wrapped store, the raw one and the folder.
func armed(t *testing.T, key string) (storagedeps.Contract, storagedeps.Contract, string) {
	t.Helper()
	reset(t)
	t.Setenv("TEST_ENCRYPT_KEY", key)
	if _, present, err := loadKey("TEST_ENCRYPT_KEY"); err != nil || !present {
		t.Fatalf("loadKey: present %v err %v", present, err)
	}
	if err := arm(sealed, 1); err != nil {
		t.Fatal(err)
	}
	base := t.TempDir()
	raw := filestorage.New(base)
	return Wrap(raw), raw, base
}

func TestLoadKeyUnsetsTheEnvironment(t *testing.T) {
	reset(t)
	t.Setenv("TEST_ENCRYPT_KEY", keyA)
	id, present, err := loadKey("TEST_ENCRYPT_KEY")
	if err != nil || !present || len(id) != 8 {
		t.Fatalf("loadKey: id %q present %v err %v", id, present, err)
	}
	if _, set := os.LookupEnv("TEST_ENCRYPT_KEY"); set {
		t.Fatal("the key is still in the environment")
	}
}

func TestLoadKeyRefusesAPassphraseWithoutQuotingIt(t *testing.T) {
	reset(t)
	t.Setenv("TEST_ENCRYPT_KEY", "correct horse battery staple")
	_, present, err := loadKey("TEST_ENCRYPT_KEY")
	if err == nil || !present {
		t.Fatal("a passphrase was taken for a key")
	}
	if strings.Contains(err.Error(), "horse") {
		t.Fatal("the error quotes the value")
	}
	if _, set := os.LookupEnv("TEST_ENCRYPT_KEY"); set {
		t.Fatal("a refused key is still in the environment")
	}
}

func TestHexAndBase64SpellTheSameKey(t *testing.T) {
	reset(t)
	t.Setenv("TEST_ENCRYPT_KEY", "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	hexId, _, _ := loadKey("TEST_ENCRYPT_KEY")
	t.Setenv("TEST_ENCRYPT_KEY", keyA)
	base64Id, _, _ := loadKey("TEST_ENCRYPT_KEY")
	if hexId == "" || hexId != base64Id {
		t.Fatalf("hex %q, base64 %q", hexId, base64Id)
	}
}

func TestUnarmedFailsClosed(t *testing.T) {
	reset(t)
	store := Wrap(filestorage.New(t.TempDir()))
	if err := store.Write([]string{"db", "t", "1"}, []byte("x")); err == nil {
		t.Fatal("an unarmed storage wrote")
	}
	if _, _, err := store.Read([]string{"db", "t", "1"}); err == nil {
		t.Fatal("an unarmed storage read")
	}
	if _, err := store.Exists([]string{"db"}); err == nil {
		t.Fatal("an unarmed storage answered Exists")
	}
}

func TestSealedValuesOnDisk(t *testing.T) {
	store, _, base := armed(t, keyA)
	key := []string{"data", "contactdb", "contact-request", "1", "values", "body"}
	secret := []byte(`{"email":"someone@example.com"}`)
	if err := store.Write(key, secret); err != nil {
		t.Fatal(err)
	}
	onDisk, err := os.ReadFile(filepath.Join(base, "data", "contactdb", "contact-request", "1", "values", "body"))
	if err != nil {
		t.Fatal(err)
	}
	if onDisk[0] != header || len(onDisk) != overhead+len(secret) {
		t.Fatalf("not the sealed format: % x", onDisk[:8])
	}
	if bytes.Contains(onDisk, []byte("someone@example.com")) {
		t.Fatal("the plaintext is on disk")
	}
	state.RLock()
	id := hex.EncodeToString(state.current.id[:])
	state.RUnlock()
	if got, encrypted := keyIdOf(onDisk); !encrypted || got != id {
		t.Fatalf("key id %q, want %q", got, id)
	}
	read, found, err := store.Read(key)
	if err != nil || !found || !bytes.Equal(read, secret) {
		t.Fatalf("read %q found %v err %v", read, found, err)
	}
}

func TestAdditionalDataIsThePathBelowTheDatabaseFolder(t *testing.T) {
	store, raw, base := armed(t, keyA)
	key := []string{"data", "Contact.DB", "a/b", "", "values", "ip"}
	if err := store.Write(key, []byte("203.0.113.9")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(base, "data", filepath.FromSlash(relativePath(key, 1)))); err != nil {
		t.Fatalf("encodeSegment differs from filestorage's: %v", err)
	}

	// Moved to another field, it does not open.
	stored, _, _ := raw.Read(key)
	moved := []string{"data", "Contact.DB", "a/b", "", "values", "country"}
	if err := raw.Write(moved, stored); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Read(moved); err == nil {
		t.Fatal("a value moved to another field opened")
	}

	// Under another --database folder, it opens.
	other := []string{"var", "app", "Contact.DB", "a/b", "", "values", "ip"}
	if err := raw.Write(other, stored); err != nil {
		t.Fatal(err)
	}
	if err := arm(sealed, 2); err != nil {
		t.Fatal(err)
	}
	if read, _, err := store.Read(other); err != nil || string(read) != "203.0.113.9" {
		t.Fatalf("read %q err %v", read, err)
	}
}

func TestAPlaintextValueIsAnError(t *testing.T) {
	store, raw, _ := armed(t, keyA)
	key := []string{"data", "db", "t", "1", "values", "f"}
	if err := raw.Write(key, []byte("in clear")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Read(key); err == nil {
		t.Fatal("a plaintext value was read in an encrypted database")
	}
}

func TestAnotherKeyIsRefused(t *testing.T) {
	store, _, base := armed(t, keyA)
	key := []string{"data", "db", "t", "1", "values", "f"}
	if err := store.Write(key, []byte("value")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_ENCRYPT_KEY", keyB)
	if _, _, err := loadKey("TEST_ENCRYPT_KEY"); err != nil {
		t.Fatal(err)
	}
	other := Wrap(filestorage.New(base))
	if _, _, err := other.Read(key); err == nil || !strings.Contains(err.Error(), "another key") {
		t.Fatalf("err %v", err)
	}
}

func TestConditionalWritesComparePlaintext(t *testing.T) {
	store, _, _ := armed(t, keyA)
	key := []string{"data", "db", "t", "last-id"}
	if written, err := store.WriteIfAbsent(key, []byte("1")); err != nil || !written {
		t.Fatalf("WriteIfAbsent: %v %v", written, err)
	}
	if written, err := store.WriteIfAbsent(key, []byte("9")); err != nil || written {
		t.Fatalf("WriteIfAbsent over a value: %v %v", written, err)
	}
	if written, err := store.WriteIfValueEquals(key, []byte("2"), []byte("1")); err != nil || !written {
		t.Fatalf("WriteIfValueEquals on the plaintext: %v %v", written, err)
	}
	if written, err := store.WriteIfValueEquals(key, []byte("3"), []byte("1")); err != nil || written {
		t.Fatalf("WriteIfValueEquals on a stale value: %v %v", written, err)
	}
	if read, _, _ := store.Read(key); string(read) != "2" {
		t.Fatalf("read %q", read)
	}
}

func TestAppendInsertAtReadAt(t *testing.T) {
	store, _, _ := armed(t, keyA)
	key := []string{"data", "db", "t", "1", "values", "log"}
	if err := store.Append(key, []byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(key, []byte(" world")); err != nil {
		t.Fatal(err)
	}
	if err := store.InsertAt(key, 5, []byte(",")); err != nil {
		t.Fatal(err)
	}
	if read, _, _ := store.Read(key); string(read) != "hello, world" {
		t.Fatalf("read %q", read)
	}
	if part, found, err := store.ReadAt(key, 7, 3); err != nil || !found || string(part) != "wor" {
		t.Fatalf("ReadAt %q %v %v", part, found, err)
	}
	if part, _, _ := store.ReadAt(key, 10, 100); string(part) != "ld" {
		t.Fatalf("ReadAt past the end %q", part)
	}
	if err := store.InsertAt(key, 99, []byte("x")); err == nil {
		t.Fatal("InsertAt past the end")
	}
}

func TestPlaintextModePassesValuesThrough(t *testing.T) {
	reset(t)
	arm(plain, 0)
	base := t.TempDir()
	store := Wrap(filestorage.New(base))
	key := []string{"data", "db", "t", "1", "values", "f"}
	if err := store.Write(key, []byte("in clear")); err != nil {
		t.Fatal(err)
	}
	onDisk, _ := os.ReadFile(filepath.Join(base, "data", "db", "t", "1", "values", "f"))
	if string(onDisk) != "in clear" {
		t.Fatalf("on disk %q", onDisk)
	}
}

func TestRotation(t *testing.T) {
	store, _, base := armed(t, keyA)
	keys := [][]string{
		{"data", "db", "t", "1", "values", "f"},
		{"data", "db", "t", "size"},
		{"data", "backup", "blob", "1", "values", "value"},
	}
	for index, key := range keys {
		if err := store.Write(key, []byte{byte('a' + index)}); err != nil {
			t.Fatal(err)
		}
	}
	snapshotFile, _ := os.ReadFile(filepath.Join(base, "data", "db", "t", "1", "values", "f"))

	t.Setenv("TEST_OLD_KEY", keyA)
	t.Setenv("TEST_ENCRYPT_KEY", keyB)
	if _, _, err := loadOldKey("TEST_OLD_KEY"); err != nil {
		t.Fatal(err)
	}
	newId, _, err := loadKey("TEST_ENCRYPT_KEY")
	if err != nil {
		t.Fatal(err)
	}
	if err := arm(rotating, 1); err != nil {
		t.Fatal(err)
	}
	count, err := rotateFiles(filepath.Join(base, "data"))
	if err != nil || count != len(keys) {
		t.Fatalf("rotateFiles: %d %v", count, err)
	}
	again, err := rotateFiles(filepath.Join(base, "data"))
	if err != nil || again != 0 {
		t.Fatalf("a second rotateFiles resealed %d: %v", again, err)
	}
	resealed, changed, err := reseal(snapshotFile, "db/t/1/values/f")
	if err != nil || !changed {
		t.Fatalf("reseal: %v %v", changed, err)
	}
	if id, _ := keyIdOf(resealed); id != newId {
		t.Fatalf("resealed under %q, want %q", id, newId)
	}

	state.Lock()
	state.old = nil
	state.Unlock()
	if err := arm(sealed, 1); err != nil {
		t.Fatal(err)
	}
	for index, key := range keys {
		read, _, err := store.Read(key)
		if err != nil || !bytes.Equal(read, []byte{byte('a' + index)}) {
			t.Fatalf("after rotation %v: %q %v", key, read, err)
		}
	}
}
