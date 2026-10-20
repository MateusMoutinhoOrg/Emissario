package backoffice_db

import (
	api "github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	reindex "github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/reindex"
)

// RemoveBackofficeUserSession deletes one session of one backoffice-user
// record. The declaration generates Add and List for a nested collection but no
// Remove, so it is written here. Removing a session that is already gone is not
// an error.
func RemoveBackofficeUserSession(sandbox *api.Sandbox, db *BackofficeDb, userId int64, sessionId int64) error {
	collection, err := sandbox.Deps.OpinionatedAgnosDatabase.Collection(db.database, "backoffice-user")
	if err != nil {
		return err
	}
	user, ok, failure := collection.FindByID(userId)
	if failure != nil {
		return sandbox.Deps.OpinionatedAgnosDatabase.Fail(failure)
	}
	if !ok {
		return sandbox.Deps.StdDeps.Errorf("backoffice-user %d not found", userId)
	}
	sessions, failure := user.Nested("session")
	if failure != nil {
		return sandbox.Deps.OpinionatedAgnosDatabase.Fail(failure)
	}
	session, ok, failure := sessions.FindByID(sessionId)
	if failure != nil {
		return sandbox.Deps.OpinionatedAgnosDatabase.Fail(failure)
	}
	if !ok {
		return nil
	}
	return sandbox.Deps.OpinionatedAgnosDatabase.Fail(session.Remove())
}

// ReindexBackofficeDb gives every live record of every table of the database that
// indexes a key field back every index entry it is missing — the ones a key
// rotation removes included — through the store's Repair. No generated method
// reaches Repair, so it is written here.
func ReindexBackofficeDb(sandbox *api.Sandbox, db *BackofficeDb) error {
	return reindex.Database(sandbox, db.database)
}
