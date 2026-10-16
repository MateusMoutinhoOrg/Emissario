package privacy

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/snapshots"
)

// The privacy mechanic makes the server private by default and gives a root
// what they need to run it under the LGPD, the Marco Civil da Internet and the
// GDPR: IPs hashed in the records, a Marco Civil access log a root turns on,
// retentions with their references, notices before every risky choice, and
// presets. This package is its core — the settings, the IP hashes, the
// retention rules, the notices and the access log — which the visualization,
// the contact requests and the agent call; what reaches across all of them
// (the cleanup, find and erase, the presets) lives in privacyadmin.
//
// It keeps no record of who chose what: no setting, hold, erasure or notice
// is ever tied to a user. The one log of people is the inventory of the
// access log reads Decreto 8.771/2016 art. 13 III asks of a Brazilian company,
// and a root may turn it off.

// The paths of the Privacy pages.
const (
	SettingsPath     = "/admin/root/set-privacy-settings"
	FindPath         = "/admin/root/find-privacy-subject"
	AccessLogPath    = "/admin/root/access-log"
	PresetsPath      = "/admin/root/privacy-presets"
	PresetPreviewPth = "/admin/root/preview-privacy-preset"
	TotpPath         = "/admin/set-totp"
)

// Refusal is why a request is not answered as asked: its status, the field it
// blames ("" for none) and the message. Notices, when it holds any, are the
// ids of the notices the request has to acknowledge first: it is then a 409.
type Refusal struct {
	Status  int
	Field   string
	Message string
	Notices []string
}

// refuse is a 400 blaming field.
func refuse(field string, message string) *Refusal {
	return &Refusal{Status: api.StatusBadRequest, Field: field, Message: message}
}

// Refuse is a 400 blaming field.
func Refuse(field string, message string) *Refusal {
	return refuse(field, message)
}

// NowSeconds is the current instant, in Unix seconds.
func NowSeconds(sandbox *api.Sandbox) int64 {
	return sandbox.Deps.StdDeps.Now() / 1000000000
}

// NowMilliseconds is the current instant, in Unix milliseconds.
func NowMilliseconds(sandbox *api.Sandbox) int64 {
	return sandbox.Deps.StdDeps.Now() / 1000000
}

// HoldForBackups waits while a backup is taken or restored, as every writer
// does, so what it writes is in a snapshot whole or not at all.
func HoldForBackups(sandbox *api.Sandbox) {
	for snapshots.Creating(sandbox) || snapshots.Restoring(sandbox) {
		sandbox.Deps.SleepDeps.Sleep(500)
	}
}
