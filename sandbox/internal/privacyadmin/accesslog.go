package privacyadmin

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/privacy_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeauth"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficehttp"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/totp"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// The access log is read as Decreto 8.771/2016 art. 13 asks: by named
// readers alone — roots a root marks access-log-reader (I) — signed in with
// two-factor by default, their API tokens made after it (II), every list,
// search and export inventoried in privacy-access (III), and encrypted with
// the rest (IV). Its rows are disclosed only on a court order (Marco Civil
// art. 10 §1).

// The actions the inventory records.
const (
	ActionAccessLogList   = "access-log-list"
	ActionAccessLogExport = "access-log-export"
)

// The formats an export of the access log is written in.
const (
	FormatJson = "json"
	FormatCsv  = "csv"
)

// MaxAccessExportRows bounds one export of the access log.
const MaxAccessExportRows = 100000

// DefaultAccessDays is the period the access log page shows by default.
const DefaultAccessDays = 7

// CanReadAccessLog refuses user — signed in with a session, or with token
// when token is not nil — the access log: 403 for a viewer, a root who is
// not a reader, and, while two-factor is required, a reader without it or a
// token made before it was turned on.
func CanReadAccessLog(sandbox *api.Sandbox, user *backoffice_db.BackofficeUserRecord, token *backoffice_db.ApiTokenRecord) *privacy.Refusal {
	forbidden := func(message string) *privacy.Refusal {
		return &privacy.Refusal{Status: api.StatusForbidden, Message: message}
	}
	if user == nil {
		return &privacy.Refusal{Status: api.StatusUnauthorized, Message: "no authenticated user"}
	}
	if backofficeauth.Role(user.Role) != backofficeauth.RoleRoot || user.AccessLogReader != 1 {
		return forbidden("Only the roots a root named readers read the access log (Decreto 8.771/2016 art. 13 I).")
	}
	settings, err := privacy.CurrentSettings(sandbox)
	if err != nil {
		return &privacy.Refusal{Status: api.StatusInternalServerError, Message: "the privacy settings could not be read"}
	}
	if !settings.AccessLogTotp {
		return nil
	}
	if !totp.Enabled(*user) {
		return forbidden("Turn on two-factor sign-in at " + privacy.TotpPath + " to read the access log (Decreto 8.771/2016 art. 13 II).")
	}
	if token != nil && token.CreatedAt < user.TotpEnabledAt {
		return forbidden("This API token was made before two-factor sign-in was turned on: make a new one to read the access log.")
	}
	return nil
}

// RawAccessFilter is an access log filter as a page or a JSON body sends it.
type RawAccessFilter struct {
	Ip   string
	From string
	To   string
	Kind string
}

// AccessFilter is an access log filter ready to apply.
type AccessFilter struct {
	// Ip is one ip or one CIDR range in its canonical spelling, "" for any.
	Ip string
	// From and To are the first and the last day, local, spelled
	// visualization.DayLayout.
	From string
	To   string
	// AtMin and AtMax bound the period in Unix milliseconds.
	AtMin int64
	AtMax int64
	// Kind is one of privacy.AccessKinds, "" for any.
	Kind string
}

// ParseAccessFilter reads raw: the last DefaultAccessDays days by default.
func ParseAccessFilter(sandbox *api.Sandbox, raw RawAccessFilter) (AccessFilter, *privacy.Refusal) {
	strings := sandbox.Deps.StringsDeps
	filter := AccessFilter{From: strings.TrimSpace(raw.From), To: strings.TrimSpace(raw.To)}
	if filter.To == "" {
		filter.To = visualization.Today(sandbox)
	}
	if filter.From == "" {
		filter.From = visualization.AddDays(sandbox, filter.To, -(DefaultAccessDays - 1))
	}
	start, ok := visualization.ParseDay(sandbox, filter.From)
	if !ok {
		return filter, privacy.Refuse("from", "from is a day, 2026-10-13.")
	}
	end, ok := visualization.ParseDay(sandbox, visualization.AddDays(sandbox, filter.To, 1))
	if !ok {
		return filter, privacy.Refuse("to", "to is a day, 2026-10-13.")
	}
	if end <= start {
		return filter, privacy.Refuse("to", "to is the same day as from or after it.")
	}
	filter.AtMin, filter.AtMax = start*1000, end*1000-1
	if ip := strings.TrimSpace(raw.Ip); ip != "" {
		canonical, ok := backofficehttp.CanonicalIpOrRange(sandbox, ip)
		if !ok {
			return filter, privacy.Refuse("ip", "ip is one ip (203.0.113.7) or one CIDR range (203.0.113.0/24).")
		}
		filter.Ip = canonical
	}
	if kind := strings.TrimSpace(raw.Kind); kind != "" {
		known := false
		for _, candidate := range privacy.AccessKinds {
			known = known || kind == candidate
		}
		if !known {
			return filter, privacy.Refuse("kind", "kind is view, contact-request or agent.")
		}
		filter.Kind = kind
	}
	return filter, nil
}

// SelectAccess answers every access log row filter keeps, newest first.
func SelectAccess(sandbox *api.Sandbox, filter AccessFilter) ([]privacy_db.AccessLogRecord, error) {
	records, err := privacy_db.New(sandbox).ListAccessLogs(privacy_db.AccessLogFilter{AtMin: filter.AtMin, AtMax: filter.AtMax, KindEquals: filter.Kind})
	if err != nil {
		return nil, err
	}
	kept := []privacy_db.AccessLogRecord{}
	for _, record := range records {
		if filter.Ip != "" && !backofficehttp.IpMatches(sandbox, filter.Ip, record.Ip) {
			continue
		}
		kept = append(kept, record)
	}
	sandbox.Deps.SortDeps.SliceStable(kept, func(i int, j int) bool {
		if kept[i].At != kept[j].At {
			return kept[i].At > kept[j].At
		}
		return kept[i].Id > kept[j].Id
	})
	return kept, nil
}

// FilterJSON is filter as the inventory records it.
func FilterJSON(sandbox *api.Sandbox, filter AccessFilter) string {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("from", filter.From)
	document.AddItemToObject("to", filter.To)
	document.AddItemToObject("ip", filter.Ip)
	document.AddItemToObject("kind", filter.Kind)
	return sandbox.Deps.SerializableDeps.SerializeToJson(document)
}

// Inventory records one read of the access log by user, started at
// startedMs, while the inventory is on: what was done, with which filter,
// and how many rows came back. It is the one log of people the privacy
// mechanic keeps (Decreto 8.771/2016 art. 13 III).
func Inventory(sandbox *api.Sandbox, user *backoffice_db.BackofficeUserRecord, action string, filter string, rows int, startedMs int64) error {
	settings, err := privacy.CurrentSettings(sandbox)
	if err != nil || !settings.AccessLogInventory || user == nil {
		return err
	}
	_, err = privacy_db.New(sandbox).AddPrivacyAccess(privacy_db.PrivacyAccessInput{
		At:       startedMs,
		EndedAt:  privacy.NowMilliseconds(sandbox),
		UserId:   user.Id,
		Username: user.Username,
		Action:   action,
		Filter:   filter,
		Rows:     int64(rows),
	})
	return err
}

// InsideTerm tells whether the access log row at atMs is younger than the
// access log retention of settings: kept under its legal obligation.
func InsideTerm(sandbox *api.Sandbox, settings privacy.Settings, atMs int64) bool {
	if settings.AccessLogRetentionDays == privacy.Forever {
		return true
	}
	return atMs > (privacy.NowSeconds(sandbox)-settings.AccessLogRetentionDays*86400)*1000
}

// DeleteResult is what DeleteAccess did.
type DeleteResult struct {
	Deleted int
	// Held is how many rows a hold kept.
	Held int
	// Missing is how many ids no row held.
	Missing int
}

// DeleteAccess deletes the access log rows of ids. A row inside its term
// asks for the erase-kept-v1 notice at every use — the duty to keep it may
// not apply to this website — and a row on a legal hold is never deleted
// before its hold is released. Nothing records the deletion.
func DeleteAccess(sandbox *api.Sandbox, ids []int64, acknowledge []string) (DeleteResult, *privacy.Refusal, error) {
	result := DeleteResult{}
	settings, err := privacy.CurrentSettings(sandbox)
	if err != nil {
		return result, nil, err
	}
	db := privacy_db.New(sandbox)
	held, err := privacy.HeldIds(sandbox, privacy.TableAccessLog, privacy.NowSeconds(sandbox))
	if err != nil {
		return result, nil, err
	}
	targets := []int64{}
	inside := false
	for _, id := range ids {
		record, ok := db.FindAccessLogById(id)
		if !ok {
			result.Missing++
			continue
		}
		if held[id] {
			result.Held++
			continue
		}
		inside = inside || InsideTerm(sandbox, settings, record.At)
		targets = append(targets, id)
	}
	if inside {
		if refusal := privacy.NeedsAcknowledgement([]string{privacy.NoticeEraseKept}, acknowledge); refusal != nil {
			return DeleteResult{}, refusal, nil
		}
	}
	privacy.HoldForBackups(sandbox)
	for _, id := range targets {
		if err := db.RemoveAccessLog(id); err != nil {
			return result, nil, err
		}
		result.Deleted++
	}
	return result, nil, nil
}

// SetReaders makes the roots of ids the readers of the access log, and no
// one else: an id of a viewer is refused, since only a root reads it.
func SetReaders(sandbox *api.Sandbox, ids []int64) *privacy.Refusal {
	db := backoffice_db.New(sandbox)
	users, err := db.ListBackofficeUsers(backoffice_db.BackofficeUserFilter{})
	if err != nil {
		return &privacy.Refusal{Status: api.StatusInternalServerError, Message: "the backoffice users could not be read"}
	}
	wanted := map[int64]bool{}
	for _, id := range ids {
		wanted[id] = true
	}
	byId := map[int64]backoffice_db.BackofficeUserRecord{}
	for _, user := range users {
		byId[user.Id] = user
	}
	for id := range wanted {
		user, ok := byId[id]
		if !ok {
			return privacy.Refuse("ids", "No backoffice user has the id "+sandbox.Deps.StringsDeps.FormatInt(id, 10)+".")
		}
		if backofficeauth.Role(user.Role) != backofficeauth.RoleRoot {
			return privacy.Refuse("ids", user.Username+" is a viewer: only a root reads the access log.")
		}
	}
	for _, user := range users {
		value := int64(0)
		if wanted[user.Id] {
			value = 1
		}
		if user.AccessLogReader != value {
			if err := db.SetBackofficeUserAccessLogReader(user.Id, value); err != nil {
				return &privacy.Refusal{Status: api.StatusInternalServerError, Message: "the readers could not be saved"}
			}
		}
	}
	InvalidateBanners(sandbox)
	return nil
}

// Readers answers every root, and whether each reads the access log.
func Readers(sandbox *api.Sandbox) ([]backoffice_db.BackofficeUserRecord, error) {
	users, err := backoffice_db.New(sandbox).ListBackofficeUsers(backoffice_db.BackofficeUserFilter{})
	if err != nil {
		return nil, err
	}
	roots := []backoffice_db.BackofficeUserRecord{}
	for _, user := range users {
		if backofficeauth.Role(user.Role) == backofficeauth.RoleRoot {
			roots = append(roots, user)
		}
	}
	return roots, nil
}

// FormatUtc is an instant in Unix milliseconds, in UTC with milliseconds:
// what a court order asks for, with the time zone.
func FormatUtc(sandbox *api.Sandbox, atMs int64) string {
	seconds := atMs / 1000
	millis := atMs % 1000
	text := sandbox.Deps.TimeDeps.FormatUnix(seconds, "2006-01-02T15:04:05")
	digits := sandbox.Deps.StringsDeps.FormatInt(1000+millis, 10)[1:]
	return text + "." + digits + "Z"
}

// AccessCsv is rows as a CSV: the time in UTC and in the server's time zone,
// the ip, the source port, the kind and the ref.
func AccessCsv(sandbox *api.Sandbox, rows []privacy_db.AccessLogRecord) string {
	lines := []string{"time_utc,time_local,time_zone,ip,port,kind,ref"}
	zone := visualization.ZoneName(sandbox)
	for _, row := range rows {
		port := ""
		if row.Port > 0 {
			port = sandbox.Deps.StringsDeps.FormatInt(row.Port, 10)
		}
		lines = append(lines, sandbox.Deps.StringsDeps.Join([]string{
			FormatUtc(sandbox, row.At),
			visualization.FormatLocal(sandbox, row.At/1000, visualization.TimeLayout),
			csvField(sandbox, zone),
			csvField(sandbox, row.Ip),
			port,
			csvField(sandbox, row.Kind),
			csvField(sandbox, row.Ref),
		}, ","))
	}
	return sandbox.Deps.StringsDeps.Join(lines, "\n") + "\n"
}

// csvField quotes value when it holds a comma, a quote or a line break.
func csvField(sandbox *api.Sandbox, value string) string {
	strings := sandbox.Deps.StringsDeps
	if !strings.ContainsAny(value, ",\"\n\r") {
		return value
	}
	return "\"" + strings.ReplaceAll(value, "\"", "\"\"") + "\""
}
