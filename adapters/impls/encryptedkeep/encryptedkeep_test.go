package encryptedkeep

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	keepencryption "github.com/MateusMoutinhoOrg/Emissario/adapters/impls/keepencryption"
	contract "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/databasedeps"
	encryptiondeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/encryptiondeps"

	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps"
)

const key = "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8="

var props = contract.Props{
	Path: "data/testdb",
	Schemas: []contract.Schema{
		{Name: "user", Fields: []contract.Field{
			{Name: "email", Type: contract.Key, Required: true},
			{Name: "note", Type: contract.String},
		}},
	},
}

// bound answers the deps both adapters fill, run from an empty folder: Keep's
// file storage writes under the working directory.
func bound(t *testing.T) (contract.Databases, encryptiondeps.Contract) {
	t.Helper()
	t.Chdir(t.TempDir())
	filled := deps.Deps{}
	Bind(&filled)
	keepencryption.Bind(&filled)
	return filled.DatabaseDeps.Databases, filled.EncryptionDeps
}

func TestEveryDatabaseIsEncryptedByKeep(t *testing.T) {
	databases, encryption := bound(t)
	if _, failure := databases.New(props); failure == nil || failure.Type != contract.Internal {
		t.Fatalf("a database opened before the encryption was armed: %v", failure)
	}

	t.Setenv("TEST_ENCRYPT_KEY", key)
	if _, _, err := encryption.LoadKey("TEST_ENCRYPT_KEY"); err != nil {
		t.Fatal(err)
	}
	if err := encryption.Encrypt(); err != nil {
		t.Fatal(err)
	}
	database, failure := databases.New(props)
	if failure != nil {
		t.Fatal(failure.Message)
	}
	if database.Props.EncryptKey != "" {
		t.Fatal("the EncryptKey reached the Props a database hands back")
	}
	users, _ := database.Collection("user")
	if _, failure := users.Insert(map[string]any{"email": "ana@example.com", "note": "secret note"}); failure != nil {
		t.Fatal(failure.Message)
	}
	if _, found, failure := users.FindByKey("email", "ANA@example.com"); failure != nil || !found {
		t.Fatalf("FindByKey under the key: found %v %v", found, failure)
	}

	err := filepath.WalkDir("data", func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		content, _ := os.ReadFile(path)
		if strings.Contains(string(content), "ana@example.com") || strings.Contains(string(content), "secret note") {
			t.Errorf("%s holds a value in clear", path)
		}
		if !encryption.Opens(content) {
			t.Errorf("%s does not open under the key", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	encryption.Plaintext()
	inClear, _ := databases.New(props)
	users, _ = inClear.Collection("user")
	if _, found, _ := users.FindByKey("email", "ana@example.com"); found {
		t.Fatal("a database opened in clear found a record through an encrypted index")
	}
}
