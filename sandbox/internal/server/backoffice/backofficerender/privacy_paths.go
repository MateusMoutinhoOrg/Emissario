package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// The paths of the legal holds, shared by every page that holds rows.
const (
	HoldPath        = "/admin/root/hold-privacy-records"
	ReleaseHoldPath = "/admin/root/release-privacy-hold"
)

// HoldBox is the form a root holds one record with, on its own page, and
// whether a hold keeps it already.
type HoldBox struct {
	Action string
	Table  string
	// Id is the id of the record, the one a hold keeps.
	Id   string
	Back string
	// Until is the end date the form defaults to: the 60 days of a Marco
	// Civil art. 15 §2 request.
	Until string
	// HeldUntil is the last day a hold keeps the record, "" while none does.
	HeldUntil string
}

// holdBoxOf is the HoldBox of the record id of table, whose page is back.
func holdBoxOf(sandbox *api.Sandbox, table string, id int64, back string) (HoldBox, error) {
	box := HoldBox{
		Action: HoldPath,
		Table:  table,
		Id:     sandbox.Deps.StringsDeps.FormatInt(id, 10),
		Back:   back,
		Until:  visualization.AddDays(sandbox, visualization.Today(sandbox), privacy.PreservationDays),
	}
	until, held, err := privacy.HeldUntil(sandbox, table, id, privacy.NowSeconds(sandbox))
	if err != nil {
		return box, err
	}
	if held {
		box.HeldUntil = visualization.DayOf(sandbox, until-1)
	}
	return box, nil
}

// holdNotice is how a record page words the outcome of a hold, Notice{}
// for any other code.
func holdNotice(sandbox *api.Sandbox, code string, reason string) Notice {
	switch code {
	case privacyadmin.NoticeHeld:
		return Notice{Text: "Held: the cleanup and Find and erase keep it until the end date.", Kind: "ok"}
	case privacyadmin.NoticeHoldRefused:
		return Notice{Text: "Not held: " + reason, Kind: "error"}
	}
	return Notice{}
}
