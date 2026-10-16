package privacy_db

import (
	api "github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	databasedeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/databasedeps"
)

// AccessLogRecords is every access-log row, none of its fields read yet: no
// field of the access log is a key (a key's SHA-256 would be a file name), so
// every search scans, reading the at of each row and the rest only of the
// ones it keeps.
func AccessLogRecords(sandbox *api.Sandbox, db *PrivacyDb) ([]databasedeps.Record, error) {
	collection, err := sandbox.Deps.OpinionatedAgnosDatabase.Collection(db.database, "access-log")
	if err != nil {
		return nil, err
	}
	records, failure := collection.ListAll()
	if failure != nil {
		return nil, sandbox.Deps.OpinionatedAgnosDatabase.Fail(failure)
	}
	return records, nil
}

// PrivacyAccessRecords is every row of the inventory of the access log
// reads, none of its fields read yet.
func PrivacyAccessRecords(sandbox *api.Sandbox, db *PrivacyDb) ([]databasedeps.Record, error) {
	collection, err := sandbox.Deps.OpinionatedAgnosDatabase.Collection(db.database, "privacy-access")
	if err != nil {
		return nil, err
	}
	records, failure := collection.ListAll()
	if failure != nil {
		return nil, sandbox.Deps.OpinionatedAgnosDatabase.Fail(failure)
	}
	return records, nil
}
