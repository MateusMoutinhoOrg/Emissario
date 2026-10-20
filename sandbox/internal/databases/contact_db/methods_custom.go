package contact_db

import (
	api "github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	databasedeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/databasedeps"
	reindex "github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/reindex"
)

// Only a `key` field is indexed, so every query by time scans the table. The
// generated List methods read every field of every record before they filter;
// the ones below read the `at` of each record first and touch the rest only
// for the records they keep.

// ListContactRequestsBetween reads every contact-request whose at is within
// atMin and atMax, both inclusive — a bound of 0 is no bound — in the order
// the store holds them, not by time.
func ListContactRequestsBetween(sandbox *api.Sandbox, db *ContactDb, atMin int64, atMax int64) ([]ContactRequestRecord, error) {
	records, err := listCollection(sandbox, db, "contact-request")
	if err != nil {
		return nil, err
	}
	kept := []ContactRequestRecord{}
	for _, record := range records {
		inside, err := atWithin(sandbox, record, atMin, atMax)
		if err != nil {
			return nil, err
		}
		if !inside {
			continue
		}
		built, err := buildContactRequestRecord(sandbox, record)
		if err != nil {
			return nil, err
		}
		kept = append(kept, built)
	}
	return kept, nil
}

// CountContactRequests answers how many contact-requests are not archived,
// and how many arrived at since or after, reading only their at and archived.
func CountContactRequests(sandbox *api.Sandbox, db *ContactDb, since int64) (int, int, error) {
	records, err := listCollection(sandbox, db, "contact-request")
	if err != nil {
		return 0, 0, err
	}
	inbox, recent := 0, 0
	for _, record := range records {
		archived, err := sandbox.Deps.OpinionatedAgnosDatabase.ReadInt(record, "archived")
		if err != nil {
			return 0, 0, err
		}
		if archived == 0 {
			inbox++
		}
		at, err := sandbox.Deps.OpinionatedAgnosDatabase.ReadInt(record, "at")
		if err != nil {
			return 0, 0, err
		}
		if at >= since {
			recent++
		}
	}
	return inbox, recent, nil
}

// RemoveContactRequestsBefore removes every contact-request whose at is
// before before, and answers how many it removed. A record written while it
// runs may be left, never one in the range it already read.
func RemoveContactRequestsBefore(sandbox *api.Sandbox, db *ContactDb, before int64) (int, error) {
	records, err := listCollection(sandbox, db, "contact-request")
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, record := range records {
		at, err := sandbox.Deps.OpinionatedAgnosDatabase.ReadInt(record, "at")
		if err != nil {
			return removed, err
		}
		if at >= before {
			continue
		}
		if err := sandbox.Deps.OpinionatedAgnosDatabase.Fail(record.Remove()); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

// atWithin tells whether the at of record is within atMin and atMax, both
// inclusive; a bound of 0 is no bound.
func atWithin(sandbox *api.Sandbox, record databasedeps.Record, atMin int64, atMax int64) (bool, error) {
	if atMin == 0 && atMax == 0 {
		return true, nil
	}
	at, err := sandbox.Deps.OpinionatedAgnosDatabase.ReadInt(record, "at")
	if err != nil {
		return false, err
	}
	return (atMin == 0 || at >= atMin) && (atMax == 0 || at <= atMax), nil
}

// listCollection is every record of collection, none of its fields read yet.
func listCollection(sandbox *api.Sandbox, db *ContactDb, collection string) ([]databasedeps.Record, error) {
	found, err := sandbox.Deps.OpinionatedAgnosDatabase.Collection(db.database, collection)
	if err != nil {
		return nil, err
	}
	records, failure := found.ListAll()
	if failure != nil {
		return nil, sandbox.Deps.OpinionatedAgnosDatabase.Fail(failure)
	}
	return records, nil
}

// ContactRequestRecords is every contact-request, none of its fields read
// yet: the retention cleanup reads the at of each, and the rest only of the
// ones it touches.
func ContactRequestRecords(sandbox *api.Sandbox, db *ContactDb) ([]databasedeps.Record, error) {
	return listCollection(sandbox, db, "contact-request")
}

// ReindexContactDb gives every live record of every table of the database that
// indexes a key field back every index entry it is missing — the ones a key
// rotation removes included — through the store's Repair. No generated method
// reaches Repair, so it is written here.
func ReindexContactDb(sandbox *api.Sandbox, db *ContactDb) error {
	return reindex.Database(sandbox, db.database)
}
