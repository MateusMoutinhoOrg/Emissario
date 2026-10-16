package privacyadmin

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/contact_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/privacy_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/visualization_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/snapshots"
)

// Find and erase answers a person's access and erasure requests (LGPD arts.
// 18–19; GDPR arts. 15, 17): an IP — hashed with the current IP key to match
// the hashes — or a text (an email, a phone, a name) is looked up in the
// contact requests, the views, the agent usage and the agent summaries, and,
// for a reader, in the access log. Delete all deletes the matches but two
// kinds of rows, which it lists as kept: the access log rows, kept to meet a
// legal obligation (LGPD art. 16 I; GDPR art. 17(3)(b)) unless the root also
// deletes them after the erase-kept-v1 notice, and the rows on a legal hold,
// kept until their hold is released, whose reason is never shown here (Marco
// Civil art. 13 §4). Hashes made before a Rotate IP key no longer match
// (GDPR art. 11: no duty to re-identify).
//
// An erasure of contact requests, views, agent usage or summaries keeps the
// table and the ids deleted in privacy-erasure — no personal data, no user —
// so a restore can delete them again; an entry goes once no snapshot older
// than it remains. Deleting access log rows keeps nothing.

// The tables an erasure is kept for, so a restore deletes them again.
const (
	TableVisualizationView = "visualization-view"
	TableAgentUsage        = "agent-usage"
)

// MinSubjectText is the shortest text Find and erase looks for: a shorter
// one would match nearly every record.
const MinSubjectText = 3

// MaxSubjectText is the longest text Find and erase looks for.
const MaxSubjectText = 200

// ActionAccessLogFind is the action the inventory records when Find and
// erase shows access log rows to a reader.
const ActionAccessLogFind = "access-log-find"

// Subject is who Find and erase looks for: an IP, a text, or both.
type Subject struct {
	// Ip is the IP typed, in its canonical spelling, "" for none.
	Ip     string
	filter privacy.IpFilter
	// Text is the text typed, "" for none.
	Text  string
	lower string
}

// ParseSubject reads the ip and the text a root typed: one of them at least;
// the ip one address or one hash — a range would reach other people — and
// the text 3 to 200 characters.
func ParseSubject(sandbox *api.Sandbox, ip string, text string) (Subject, *privacy.Refusal) {
	strings := sandbox.Deps.StringsDeps
	subject := Subject{Text: strings.TrimSpace(text)}
	ip = strings.TrimSpace(ip)
	if ip == "" && subject.Text == "" {
		return subject, privacy.Refuse("ip", "Type an IP, or a text the person sent: an email, a phone, a name.")
	}
	if ip != "" {
		filter, refusal := privacy.ParseIpFilter(sandbox, "ip", ip)
		if refusal != nil {
			return subject, refusal
		}
		if filter.Range {
			return subject, privacy.Refuse("ip", "Find and erase looks for one person: one ip, not a range.")
		}
		subject.Ip = filter.Text
		subject.filter = filter
	}
	if subject.Text != "" {
		if len(subject.Text) < MinSubjectText || len(subject.Text) > MaxSubjectText {
			return subject, privacy.Refuse("text", "The text is 3 to 200 characters: an email, a phone, a name.")
		}
		subject.lower = strings.ToLower(subject.Text)
	}
	return subject, nil
}

// FilterJSON is the subject as the inventory records it.
func (subject Subject) FilterJSON(sandbox *api.Sandbox) string {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("ip", subject.Ip)
	document.AddItemToObject("text", subject.Text)
	return sandbox.Deps.SerializableDeps.SerializeToJson(document)
}

// matchesIp tells whether stored is the subject's ip, hashed or not.
func (subject Subject) matchesIp(sandbox *api.Sandbox, stored string) bool {
	return subject.Ip != "" && subject.filter.Matches(sandbox, stored)
}

// matchesText tells whether value holds the subject's text, case aside.
func (subject Subject) matchesText(sandbox *api.Sandbox, value string) bool {
	return subject.lower != "" && sandbox.Deps.StringsDeps.Contains(sandbox.Deps.StringsDeps.ToLower(value), subject.lower)
}

// Found is every record of a subject.
type Found struct {
	Contacts  []contact_db.ContactRequestRecord
	Views     []visualization_db.VisualizationViewRecord
	Usages    []agent_db.AgentUsageRecord
	Summaries []agent_db.AgentSummaryRecord
	// Access are the access log rows of the ip, for a reader alone;
	// AccessShown is false for a root who is not one.
	Access      []privacy_db.AccessLogRecord
	AccessShown bool
	// Held are the ids of each table on a legal hold, kept by Delete all.
	Held map[string]map[int64]bool
	// InsideTerm are the ids of the access log rows still inside their
	// term, kept by Delete all unless the root also deletes them.
	InsideTerm map[int64]bool
	// IpKeyMadeAt is when the IP key was made: hashes made before it no
	// longer match.
	IpKeyMadeAt int64
}

// Total is every record found.
func (found Found) Total() int {
	return len(found.Contacts) + len(found.Views) + len(found.Usages) + len(found.Summaries) + len(found.Access)
}

// FindSubject looks for subject in every table of personal data, as user —
// signed in with token, nil for a session. The access log is searched by ip
// for a reader alone, and that search is inventoried.
func FindSubject(sandbox *api.Sandbox, user *backoffice_db.BackofficeUserRecord, token *backoffice_db.ApiTokenRecord, subject Subject) (Found, error) {
	started := privacy.NowMilliseconds(sandbox)
	now := privacy.NowSeconds(sandbox)
	found := Found{Held: map[string]map[int64]bool{}, InsideTerm: map[int64]bool{}}
	for _, table := range privacy.HoldTables {
		held, err := privacy.HeldIds(sandbox, table, now)
		if err != nil {
			return found, err
		}
		found.Held[table] = held
	}
	made, err := privacy.IpKeyRotatedAt(sandbox)
	if err != nil {
		return found, err
	}
	found.IpKeyMadeAt = made

	contacts := contact_db.New(sandbox)
	for offset := 0; ; offset += pageSize {
		page, err := contacts.ListContactRequestsPage(offset, pageSize)
		if err != nil {
			return found, err
		}
		for _, request := range page {
			if subject.matchesIp(sandbox, request.Ip) || subject.matchesText(sandbox, request.Body) {
				found.Contacts = append(found.Contacts, request)
			}
		}
		if len(page) < pageSize {
			break
		}
	}

	agents := agent_db.New(sandbox)
	for offset := 0; ; offset += pageSize {
		page, err := agents.ListAgentSummariesPage(offset, pageSize)
		if err != nil {
			return found, err
		}
		for _, summary := range page {
			if subject.matchesText(sandbox, summary.Summary) {
				found.Summaries = append(found.Summaries, summary)
			}
		}
		if len(page) < pageSize {
			break
		}
	}

	if subject.Ip == "" {
		return found, nil
	}
	views := visualization_db.New(sandbox)
	for offset := 0; ; offset += pageSize {
		page, err := views.ListVisualizationViewsPage(offset, pageSize)
		if err != nil {
			return found, err
		}
		for _, view := range page {
			if subject.matchesIp(sandbox, view.Ip) {
				found.Views = append(found.Views, view)
			}
		}
		if len(page) < pageSize {
			break
		}
	}
	for offset := 0; ; offset += pageSize {
		page, err := agents.ListAgentUsagesPage(offset, pageSize)
		if err != nil {
			return found, err
		}
		for _, usage := range page {
			if subject.matchesIp(sandbox, usage.Ip) {
				found.Usages = append(found.Usages, usage)
			}
		}
		if len(page) < pageSize {
			break
		}
	}

	if CanReadAccessLog(sandbox, user, token) != nil || privacy.IsHash(sandbox, subject.Ip) {
		return found, nil
	}
	found.AccessShown = true
	settings, err := privacy.CurrentSettings(sandbox)
	if err != nil {
		return found, err
	}
	rows, err := SelectAccess(sandbox, AccessFilter{Ip: subject.Ip})
	if err != nil {
		return found, err
	}
	for _, row := range rows {
		found.Access = append(found.Access, row)
		if InsideTerm(sandbox, settings, row.At) {
			found.InsideTerm[row.Id] = true
		}
	}
	return found, Inventory(sandbox, user, ActionAccessLogFind, subject.FilterJSON(sandbox), len(found.Access), started)
}

// Erased is what Delete all deleted, and what it kept.
type Erased struct {
	Contacts  int
	Views     int
	Usages    int
	Summaries int
	Access    int
	// Held is how many rows a legal hold kept, KeptByLaw how many access log
	// rows were kept inside their term.
	Held      int
	KeptByLaw int
}

// Total is every row deleted.
func (erased Erased) Total() int {
	return erased.Contacts + erased.Views + erased.Usages + erased.Summaries + erased.Access
}

// EraseSubject deletes every record of subject but the held ones and — unless
// alsoAccess — the access log rows. Deleting access log rows asks for the
// erase-kept-v1 notice every time: until acknowledge holds it the refusal is
// a 409, and nothing is deleted. Nothing records who erased.
func EraseSubject(sandbox *api.Sandbox, user *backoffice_db.BackofficeUserRecord, token *backoffice_db.ApiTokenRecord, subject Subject, alsoAccess bool, acknowledge []string) (Erased, *privacy.Refusal, error) {
	erased := Erased{}
	found, err := FindSubject(sandbox, user, token, subject)
	if err != nil {
		return erased, nil, err
	}
	accessIds := []int64{}
	for _, row := range found.Access {
		switch {
		case found.Held[privacy.TableAccessLog][row.Id]:
			erased.Held++
		case alsoAccess:
			accessIds = append(accessIds, row.Id)
		default:
			erased.KeptByLaw++
		}
	}
	if len(accessIds) > 0 {
		if refusal := privacy.NeedsAcknowledgement([]string{privacy.NoticeEraseKept}, acknowledge); refusal != nil {
			return Erased{}, refusal, nil
		}
	}

	privacy.HoldForBackups(sandbox)
	now := privacy.NowSeconds(sandbox)
	contacts := contact_db.New(sandbox)
	ids := []int64{}
	for _, request := range found.Contacts {
		if found.Held[privacy.TableContactRequest][request.Id] {
			erased.Held++
			continue
		}
		if err := contacts.RemoveContactRequest(request.Id); err != nil {
			return erased, nil, err
		}
		ids = append(ids, request.Id)
	}
	erased.Contacts = len(ids)
	if err := keepErasure(sandbox, privacy.TableContactRequest, ids, now); err != nil {
		return erased, nil, err
	}

	agents := agent_db.New(sandbox)
	ids = []int64{}
	for _, summary := range found.Summaries {
		if found.Held[privacy.TableAgentSummary][summary.Id] {
			erased.Held++
			continue
		}
		if err := agents.RemoveAgentSummary(summary.Id); err != nil {
			return erased, nil, err
		}
		ids = append(ids, summary.Id)
	}
	erased.Summaries = len(ids)
	if err := keepErasure(sandbox, privacy.TableAgentSummary, ids, now); err != nil {
		return erased, nil, err
	}

	views := visualization_db.New(sandbox)
	ids = []int64{}
	for _, view := range found.Views {
		if err := views.RemoveVisualizationView(view.Id); err != nil {
			return erased, nil, err
		}
		ids = append(ids, view.Id)
	}
	erased.Views = len(ids)
	if err := keepErasure(sandbox, TableVisualizationView, ids, now); err != nil {
		return erased, nil, err
	}

	ids = []int64{}
	for _, usage := range found.Usages {
		if err := agents.RemoveAgentUsage(usage.Id); err != nil {
			return erased, nil, err
		}
		ids = append(ids, usage.Id)
	}
	erased.Usages = len(ids)
	if err := keepErasure(sandbox, TableAgentUsage, ids, now); err != nil {
		return erased, nil, err
	}

	logs := privacy_db.New(sandbox)
	for _, id := range accessIds {
		if err := logs.RemoveAccessLog(id); err != nil {
			return erased, nil, err
		}
		erased.Access++
	}
	return erased, nil, nil
}

// keepErasure keeps ids of table in privacy-erasure, so a restore deletes
// them again; nothing for no id.
func keepErasure(sandbox *api.Sandbox, table string, ids []int64, now int64) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := privacy_db.New(sandbox).AddPrivacyErasure(privacy_db.PrivacyErasureInput{At: now, Table: table, Ids: privacy.EncodeIds(sandbox, ids)})
	return err
}

// ReplayErasures deletes again every row an erasure kept that a restore
// brought back, and answers how many. Ids are never reused, so a row already
// gone is left as it is.
func ReplayErasures(sandbox *api.Sandbox) (int, error) {
	erasures, err := privacy_db.New(sandbox).ListPrivacyErasures(privacy_db.PrivacyErasureFilter{})
	if err != nil {
		return 0, err
	}
	contacts := contact_db.New(sandbox)
	views := visualization_db.New(sandbox)
	agents := agent_db.New(sandbox)
	deleted := 0
	for _, erasure := range erasures {
		for _, id := range privacy.DecodeIds(sandbox, erasure.Ids) {
			var remove func() error
			switch erasure.Table {
			case privacy.TableContactRequest:
				if _, ok := contacts.FindContactRequestById(id); ok {
					remove = func() error { return contacts.RemoveContactRequest(id) }
				}
			case privacy.TableAgentSummary:
				if _, ok := agents.FindAgentSummaryById(id); ok {
					remove = func() error { return agents.RemoveAgentSummary(id) }
				}
			case TableVisualizationView:
				if _, ok := views.FindVisualizationViewById(id); ok {
					remove = func() error { return views.RemoveVisualizationView(id) }
				}
			case TableAgentUsage:
				if _, ok := agents.FindAgentUsageById(id); ok {
					remove = func() error { return agents.RemoveAgentUsage(id) }
				}
			}
			if remove == nil {
				continue
			}
			if err := remove(); err != nil {
				return deleted, err
			}
			deleted++
		}
	}
	return deleted, nil
}

// ForgetErasures deletes every privacy-erasure older than the oldest
// snapshot — none can bring its rows back anymore — and answers how many.
func ForgetErasures(sandbox *api.Sandbox) (int, error) {
	listed, err := snapshots.List(sandbox, "")
	if err != nil {
		return 0, err
	}
	oldest := int64(0)
	for _, snapshot := range listed {
		if oldest == 0 || snapshot.Data < oldest {
			oldest = snapshot.Data
		}
	}
	db := privacy_db.New(sandbox)
	erasures, err := db.ListPrivacyErasures(privacy_db.PrivacyErasureFilter{})
	if err != nil {
		return 0, err
	}
	forgotten := 0
	for _, erasure := range erasures {
		if oldest != 0 && erasure.At >= oldest {
			continue
		}
		if err := db.RemovePrivacyErasure(erasure.Id); err != nil {
			return forgotten, err
		}
		forgotten++
	}
	return forgotten, nil
}
