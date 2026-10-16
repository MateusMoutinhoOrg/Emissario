package agent_db

import (
	api "github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	databasedeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/databasedeps"
)

// Only a `key` field is indexed, so every query by time scans the table. The
// generated List methods read every field of every record before they filter;
// the ones below read the `at` of each record first and touch the rest only
// for the records they keep.

// ListAgentUsagesBetween reads every agent-usage whose at is within atMin and
// atMax, both inclusive, in the order the store holds them — not by time.
func ListAgentUsagesBetween(sandbox *api.Sandbox, db *AgentDb, atMin int64, atMax int64) ([]AgentUsageRecord, error) {
	records, err := listCollection(sandbox, db, "agent-usage")
	if err != nil {
		return nil, err
	}
	kept := []AgentUsageRecord{}
	for _, record := range records {
		inside, err := atWithin(sandbox, record, atMin, atMax)
		if err != nil {
			return nil, err
		}
		if !inside {
			continue
		}
		built, err := buildAgentUsageRecord(sandbox, record)
		if err != nil {
			return nil, err
		}
		kept = append(kept, built)
	}
	return kept, nil
}

// ListAgentSummariesBetween reads every agent-summary whose at is within
// atMin and atMax, both inclusive, in the order the store holds them — not by
// time.
func ListAgentSummariesBetween(sandbox *api.Sandbox, db *AgentDb, atMin int64, atMax int64) ([]AgentSummaryRecord, error) {
	records, err := listCollection(sandbox, db, "agent-summary")
	if err != nil {
		return nil, err
	}
	kept := []AgentSummaryRecord{}
	for _, record := range records {
		inside, err := atWithin(sandbox, record, atMin, atMax)
		if err != nil {
			return nil, err
		}
		if !inside {
			continue
		}
		built, err := buildAgentSummaryRecord(sandbox, record)
		if err != nil {
			return nil, err
		}
		kept = append(kept, built)
	}
	return kept, nil
}

// RemoveAgentUsagesBefore removes every agent-usage whose at is before
// before, and answers how many it removed.
func RemoveAgentUsagesBefore(sandbox *api.Sandbox, db *AgentDb, before int64) (int, error) {
	return removeBefore(sandbox, db, "agent-usage", before)
}

// RemoveAgentSummariesBefore removes every agent-summary whose at is before
// before, every summary when before is 0 or less, and answers how many it
// removed.
func RemoveAgentSummariesBefore(sandbox *api.Sandbox, db *AgentDb, before int64) (int, error) {
	return removeBefore(sandbox, db, "agent-summary", before)
}

// removeBefore removes every record of collection whose at is before before,
// every record when before is 0 or less. A record written while it runs may
// be left, never one in the range it already read.
func removeBefore(sandbox *api.Sandbox, db *AgentDb, collection string, before int64) (int, error) {
	records, err := listCollection(sandbox, db, collection)
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

// atWithin tells whether the at of record is within atMin and atMax, both
// inclusive.
func atWithin(sandbox *api.Sandbox, record databasedeps.Record, atMin int64, atMax int64) (bool, error) {
	at, err := sandbox.Deps.OpinionatedAgnosDatabase.ReadInt(record, "at")
	if err != nil {
		return false, err
	}
	return at >= atMin && at <= atMax, nil
}

// listCollection is every record of collection, none of its fields read yet.
func listCollection(sandbox *api.Sandbox, db *AgentDb, collection string) ([]databasedeps.Record, error) {
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

// AgentSummaryRecords is every agent-summary, none of its fields read yet.
func AgentSummaryRecords(sandbox *api.Sandbox, db *AgentDb) ([]databasedeps.Record, error) {
	return listCollection(sandbox, db, "agent-summary")
}

// AgentUsageRecords is every agent-usage, none of its fields read yet.
func AgentUsageRecords(sandbox *api.Sandbox, db *AgentDb) ([]databasedeps.Record, error) {
	return listCollection(sandbox, db, "agent-usage")
}
