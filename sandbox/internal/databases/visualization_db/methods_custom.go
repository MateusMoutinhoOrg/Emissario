package visualization_db

import (
	api "github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	databasedeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/databasedeps"
	reindex "github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/reindex"
)

// Only a `key` field is indexed, so every query by time scans the table. The
// generated ListVisualizationViews reads every field of every view before it
// filters; the three below read the `at` of each view first and touch the
// rest only for the views they keep, so a query for one day of a table
// holding ninety reads a ninetieth of it whole.

// ListVisualizationViewsBetween reads every visualization-view whose at is
// within atMin and atMax, both inclusive, in the order the store holds them —
// not by time: a removal moves the last view into the place it freed.
func ListVisualizationViewsBetween(sandbox *api.Sandbox, db *VisualizationDb, atMin int64, atMax int64) ([]VisualizationViewRecord, error) {
	records, err := listViews(sandbox, db)
	if err != nil {
		return nil, err
	}
	kept := []VisualizationViewRecord{}
	for _, record := range records {
		at, err := sandbox.Deps.OpinionatedAgnosDatabase.ReadInt(record, "at")
		if err != nil {
			return nil, err
		}
		if at < atMin || at > atMax {
			continue
		}
		built, err := buildVisualizationViewRecord(sandbox, record)
		if err != nil {
			return nil, err
		}
		kept = append(kept, built)
	}
	return kept, nil
}

// RemoveVisualizationViewsBefore removes every visualization-view whose at is
// before before, every view when before is 0 or less, and answers how many it
// removed. A view written while it runs may be left, never one in the range it
// already read.
func RemoveVisualizationViewsBefore(sandbox *api.Sandbox, db *VisualizationDb, before int64) (int, error) {
	records, err := listViews(sandbox, db)
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, record := range records {
		if before > 0 {
			at, err := sandbox.Deps.OpinionatedAgnosDatabase.ReadInt(record, "at")
			if err != nil {
				return removed, err
			}
			if at >= before {
				continue
			}
		}
		if err := sandbox.Deps.OpinionatedAgnosDatabase.Fail(record.Remove()); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

// OldestVisualizationViewAt is the at of the oldest visualization-view, false
// when the table holds none.
func OldestVisualizationViewAt(sandbox *api.Sandbox, db *VisualizationDb) (int64, bool, error) {
	records, err := listViews(sandbox, db)
	if err != nil {
		return 0, false, err
	}
	oldest, found := int64(0), false
	for _, record := range records {
		at, err := sandbox.Deps.OpinionatedAgnosDatabase.ReadInt(record, "at")
		if err != nil {
			return 0, false, err
		}
		if at > 0 && (!found || at < oldest) {
			oldest, found = at, true
		}
	}
	return oldest, found, nil
}

// listViews is every record of the visualization-view collection, none of its
// fields read yet.
func listViews(sandbox *api.Sandbox, db *VisualizationDb) ([]databasedeps.Record, error) {
	collection, err := sandbox.Deps.OpinionatedAgnosDatabase.Collection(db.database, "visualization-view")
	if err != nil {
		return nil, err
	}
	records, failure := collection.ListAll()
	if failure != nil {
		return nil, sandbox.Deps.OpinionatedAgnosDatabase.Fail(failure)
	}
	return records, nil
}

// VisualizationViewRecords is every visualization-view, none of its fields
// read yet: the retention cleanup reads the at of each, and the rest only of
// the ones it touches.
func VisualizationViewRecords(sandbox *api.Sandbox, db *VisualizationDb) ([]databasedeps.Record, error) {
	return listViews(sandbox, db)
}

// ReindexVisualizationDb gives every live record of every table of the database that
// indexes a key field back every index entry it is missing — the ones a key
// rotation removes included — through the store's Repair. No generated method
// reaches Repair, so it is written here.
func ReindexVisualizationDb(sandbox *api.Sandbox, db *VisualizationDb) error {
	return reindex.Database(sandbox, db.database)
}
