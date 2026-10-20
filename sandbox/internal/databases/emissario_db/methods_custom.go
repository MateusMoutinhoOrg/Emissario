package emissario_db

import (
	api "github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	reindex "github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/reindex"
)

// ReindexEmissarioDb gives every live record of every table of the database that
// indexes a key field back every index entry it is missing — the ones a key
// rotation removes included — through the store's Repair. No generated method
// reaches Repair, so it is written here.
func ReindexEmissarioDb(sandbox *api.Sandbox, db *EmissarioDb) error {
	return reindex.Database(sandbox, db.database)
}
