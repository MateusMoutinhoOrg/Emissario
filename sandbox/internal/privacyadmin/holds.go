package privacyadmin

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// canHold refuses user a hold of table: one of access log rows is for its
// readers alone, as reading them is.
func canHold(sandbox *api.Sandbox, user *backoffice_db.BackofficeUserRecord, token *backoffice_db.ApiTokenRecord, table string) *privacy.Refusal {
	if table == privacy.TableAccessLog {
		return CanReadAccessLog(sandbox, user, token)
	}
	return nil
}

// untilOf reads the last day a hold keeps its rows, spelled
// visualization.DayLayout, as the instant it ends: the start of the next day.
func untilOf(sandbox *api.Sandbox, day string) (int64, *privacy.Refusal) {
	day = sandbox.Deps.StringsDeps.TrimSpace(day)
	if _, ok := visualization.ParseDay(sandbox, day); !ok {
		return 0, privacy.Refuse("until", "until is the last day of the hold, 2026-12-31.")
	}
	end, ok := visualization.ParseDay(sandbox, visualization.AddDays(sandbox, day, 1))
	if !ok {
		return 0, privacy.Refuse("until", "until is the last day of the hold, 2026-12-31.")
	}
	return end, nil
}

// HoldRecords holds ids of table until the end of the day until for reason,
// as user — a reader for access log rows.
func HoldRecords(sandbox *api.Sandbox, user *backoffice_db.BackofficeUserRecord, token *backoffice_db.ApiTokenRecord, table string, ids []int64, reason string, until string) (privacy.Hold, *privacy.Refusal, error) {
	if refusal := canHold(sandbox, user, token, table); refusal != nil {
		return privacy.Hold{}, refusal, nil
	}
	end, refusal := untilOf(sandbox, until)
	if refusal != nil {
		return privacy.Hold{}, refusal, nil
	}
	return privacy.AddHold(sandbox, table, ids, reason, end)
}

// ReleaseHold releases the hold of id as user — a reader for one of access
// log rows. found is false when no hold has that id.
func ReleaseHold(sandbox *api.Sandbox, user *backoffice_db.BackofficeUserRecord, token *backoffice_db.ApiTokenRecord, id int64) (found bool, refusal *privacy.Refusal, err error) {
	holds, err := privacy.Holds(sandbox)
	if err != nil {
		return false, nil, err
	}
	for _, hold := range holds {
		if hold.Id != id {
			continue
		}
		if refusal := canHold(sandbox, user, token, hold.Table); refusal != nil {
			return true, refusal, nil
		}
		released, err := privacy.ReleaseHold(sandbox, id)
		return released, nil, err
	}
	return false, nil, nil
}

// VisibleHolds is every hold user sees: those of access log rows to its
// readers alone.
func VisibleHolds(sandbox *api.Sandbox, user *backoffice_db.BackofficeUserRecord, token *backoffice_db.ApiTokenRecord) ([]privacy.Hold, error) {
	holds, err := privacy.Holds(sandbox)
	if err != nil {
		return nil, err
	}
	reader := CanReadAccessLog(sandbox, user, token) == nil
	visible := []privacy.Hold{}
	for _, hold := range holds {
		if hold.Table == privacy.TableAccessLog && !reader {
			continue
		}
		visible = append(visible, hold)
	}
	return visible, nil
}
