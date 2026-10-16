package visualization

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/visualization_db"
)

// The ip is personal data, so a view is kept RetentionDays at most: the
// retention cleanup of privacyadmin, which runs every hour in start-server,
// removes the views older than that, and a root can remove them at any time
// through Purge.

// Purge removes, at a root's asking, every view before the local day before
// (DayLayout) — or every view, with all — and answers how many it removed.
// field and message are why the request is refused, "" when it is not.
func Purge(sandbox *api.Sandbox, before string, all bool) (removed int, field string, message string, err error) {
	before = sandbox.Deps.StringsDeps.TrimSpace(before)
	cutoff := int64(0)
	if !all {
		if before == "" {
			return 0, "before", "Pick a date to delete the views before, or delete every view.", nil
		}
		start, ok := ParseDay(sandbox, before)
		if !ok {
			return 0, "before", "before is a date spelled YYYY-MM-DD, such as 2026-10-13.", nil
		}
		cutoff = start
	}
	removed, err = visualization_db.RemoveVisualizationViewsBefore(sandbox, visualization_db.New(sandbox), cutoff)
	return removed, "", "", err
}

// StoredViews is how many views the table holds, and the instant of the
// oldest — 0 when it holds none.
func StoredViews(sandbox *api.Sandbox) (int, int64, error) {
	db := visualization_db.New(sandbox)
	count, err := db.CountVisualizationView()
	if err != nil {
		return 0, 0, err
	}
	oldest, _, err := visualization_db.OldestVisualizationViewAt(sandbox, db)
	return count, oldest, err
}
