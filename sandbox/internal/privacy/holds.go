package privacy

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/privacy_db"
)

// A legal hold keeps rows past their retention and out of Find and erase:
// an authority asked to keep access logs longer (Marco Civil art. 15 §2), or
// a lawsuit needs them (LGPD art. 7 VI; GDPR art. 17(3)(e)). A root holds
// contact requests or agent summaries, a reader rows of the access log, with
// a reason and an end date; a hold is released by hand or on its end date,
// and released it is deleted, leaving nothing behind. Its reason is shown on
// the list of holds alone — never in Find and erase — since a preservation
// request is confidential (Marco Civil art. 13 §4).

// The tables a hold keeps rows of.
const (
	TableContactRequest = "contact-request"
	TableAgentSummary   = "agent-summary"
	TableAccessLog      = "access-log"
)

// HoldTables is every table a hold keeps rows of.
var HoldTables = []string{TableContactRequest, TableAgentSummary, TableAccessLog}

// PreservationDays is the end date a hold for a request of Marco Civil art.
// 15 §2 defaults to: the 60 days the authority has to go to court (art. 13
// §3), which a court order extends.
const PreservationDays = 60

// MaxHoldIds bounds the rows one hold keeps.
const MaxHoldIds = 10000

// Hold is one legal hold.
type Hold struct {
	Id     int64
	Reason string
	// Until is the end of the hold, in Unix seconds.
	Until int64
	Table string
	Ids   []int64
}

// Holds answers every hold, released ones — their end date past — included.
func Holds(sandbox *api.Sandbox) ([]Hold, error) {
	records, err := privacy_db.New(sandbox).ListPrivacyHolds(privacy_db.PrivacyHoldFilter{})
	if err != nil {
		return nil, err
	}
	holds := []Hold{}
	for _, record := range records {
		holds = append(holds, Hold{Id: record.Id, Reason: record.Reason, Until: record.Until, Table: record.Table, Ids: decodeIds(sandbox, record.Ids)})
	}
	return holds, nil
}

// HeldIds answers the ids of table every hold in force at now keeps.
func HeldIds(sandbox *api.Sandbox, table string, now int64) (map[int64]bool, error) {
	holds, err := Holds(sandbox)
	if err != nil {
		return nil, err
	}
	held := map[int64]bool{}
	for _, hold := range holds {
		if hold.Table != table || hold.Until <= now {
			continue
		}
		for _, id := range hold.Ids {
			held[id] = true
		}
	}
	return held, nil
}

// validTable tells whether table is one a hold keeps rows of.
func validTable(table string) bool {
	for _, known := range HoldTables {
		if table == known {
			return true
		}
	}
	return false
}

// AddHold holds ids of table until until — Unix seconds, in the future —
// for reason.
func AddHold(sandbox *api.Sandbox, table string, ids []int64, reason string, until int64) (Hold, *Refusal, error) {
	strings := sandbox.Deps.StringsDeps
	reason = strings.TrimSpace(reason)
	switch {
	case !validTable(table):
		return Hold{}, refuse("table", "A hold keeps contact requests, agent summaries or access log rows."), nil
	case len(ids) == 0:
		return Hold{}, refuse("ids", "Select the rows to hold."), nil
	case len(ids) > MaxHoldIds:
		return Hold{}, refuse("ids", "One hold keeps at most 10000 rows."), nil
	case reason == "" || len(reason) > 500:
		return Hold{}, refuse("reason", "The reason is the request the hold answers — an ofício, a court order, a lawsuit — up to 500 characters."), nil
	case until <= NowSeconds(sandbox):
		return Hold{}, refuse("until", "The end date of a hold is in the future."), nil
	}
	record, err := privacy_db.New(sandbox).AddPrivacyHold(privacy_db.PrivacyHoldInput{
		Reason: reason,
		Until:  until,
		Table:  table,
		Ids:    encodeIds(sandbox, ids),
	})
	if err != nil {
		return Hold{}, nil, err
	}
	return Hold{Id: record.Id, Reason: reason, Until: until, Table: table, Ids: ids}, nil, nil
}

// ReleaseHold releases the hold of id: deleted, its rows follow their
// retention again. false when no hold has that id.
func ReleaseHold(sandbox *api.Sandbox, id int64) (bool, error) {
	db := privacy_db.New(sandbox)
	if _, ok := db.FindPrivacyHoldById(id); !ok {
		return false, nil
	}
	return true, db.RemovePrivacyHold(id)
}

// ReleaseExpired releases every hold whose end date passed by now, and
// answers how many.
func ReleaseExpired(sandbox *api.Sandbox, now int64) (int, error) {
	holds, err := Holds(sandbox)
	if err != nil {
		return 0, err
	}
	released := 0
	for _, hold := range holds {
		if hold.Until > now {
			continue
		}
		if err := privacy_db.New(sandbox).RemovePrivacyHold(hold.Id); err != nil {
			return released, err
		}
		released++
	}
	return released, nil
}

// encodeIds spells ids as a JSON array.
func encodeIds(sandbox *api.Sandbox, ids []int64) string {
	list := sandbox.Deps.SerializableDeps.CreateArray()
	for _, id := range ids {
		list.AddItemToArray(id)
	}
	return sandbox.Deps.SerializableDeps.SerializeToJson(list)
}

// EncodeIds spells ids as a JSON array.
func EncodeIds(sandbox *api.Sandbox, ids []int64) string {
	return encodeIds(sandbox, ids)
}

// decodeIds reads a JSON array of ids; what does not read is left out.
func decodeIds(sandbox *api.Sandbox, text string) []int64 {
	ids := []int64{}
	document, err := sandbox.Deps.SerializableDeps.ParseJson(text)
	if err != nil || !document.IsArray() {
		return ids
	}
	size, err := document.GetArraySize()
	if err != nil {
		return ids
	}
	for index := 0; index < size; index++ {
		item := document.GetArrayItem(index)
		if item == nil || !item.IsInt() {
			continue
		}
		if id, err := item.GetInt(); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

// DecodeIds reads a JSON array of ids.
func DecodeIds(sandbox *api.Sandbox, text string) []int64 {
	return decodeIds(sandbox, text)
}
