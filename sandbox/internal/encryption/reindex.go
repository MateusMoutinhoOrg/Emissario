package encryption

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backup"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/blocker_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/contact_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/emissario_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/privacy_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/visualization_db"
)

// digestDigits is how long the name of an index entry is: the lowercase hex
// of a SHA-256.
const digestDigits = 64

// IndexEntry tells whether the file at path — slash-separated and relative
// to the --database folder — is an entry of one of Keep's unique indexes,
// <table>/keys/<field>/<digest>: its name is a SHA-256 keyed by the
// EncryptKey, so the name itself depends on the key. A rotation removes such
// a file rather than reseal it, and Reindex writes it back under the new key.
// No other file of the store's layout has a 64-digit hex name two folders
// below a keys folder: ids and positions are decimal, and every other name is
// a word.
func IndexEntry(sandbox *api.Sandbox, path string) bool {
	segments := sandbox.Deps.StringsDeps.Split(path, "/")
	count := len(segments)
	if count < 3 || segments[count-3] != "keys" {
		return false
	}
	digest := segments[count-1]
	if len(digest) != digestDigits {
		return false
	}
	for index := 0; index < len(digest); index++ {
		character := digest[index]
		if !(character >= '0' && character <= '9') && !(character >= 'a' && character <= 'f') {
			return false
		}
	}
	return true
}

// Reindex gives every live record of every database of the project, the
// backup database included, back every index entry it is missing, named
// under the key the databases are armed with — Keep's Repair over every table
// that indexes a key field: how the entries a rotation removed come back, and
// how a snapshot a rotation left them out of is made whole once restored. A
// database added to the project is added here.
func Reindex(sandbox *api.Sandbox) error {
	databases := []struct {
		name    string
		reindex func() error
	}{
		{"agent-db", func() error { return agent_db.ReindexAgentDb(sandbox, agent_db.New(sandbox)) }},
		{"backoffice-db", func() error { return backoffice_db.ReindexBackofficeDb(sandbox, backoffice_db.New(sandbox)) }},
		{"backup", func() error { return backup.ReindexBackup(sandbox, backup.New(sandbox)) }},
		{"blocker-db", func() error { return blocker_db.ReindexBlockerDb(sandbox, blocker_db.New(sandbox)) }},
		{"contact-db", func() error { return contact_db.ReindexContactDb(sandbox, contact_db.New(sandbox)) }},
		{"emissario-db", func() error { return emissario_db.ReindexEmissarioDb(sandbox, emissario_db.New(sandbox)) }},
		{"privacy-db", func() error { return privacy_db.ReindexPrivacyDb(sandbox, privacy_db.New(sandbox)) }},
		{"visualization-db", func() error { return visualization_db.ReindexVisualizationDb(sandbox, visualization_db.New(sandbox)) }},
	}
	for _, database := range databases {
		if err := database.reindex(); err != nil {
			return sandbox.Deps.StdDeps.Errorf("reindexing %s: %s", database.name, err.Error())
		}
	}
	return nil
}
