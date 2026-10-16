package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// The paths of Find and erase.
const (
	PrivacyErasePath  = "/admin/root/erase-privacy-subject"
	PrivacyExportPath = "/admin/root/export-privacy-subject"
)

// PrivacyNoticeErased is the outcome code of an erasure.
const PrivacyNoticeErased = "erased"

// SubjectLine is one record of a person on the Find and erase page.
type SubjectLine struct {
	Table string
	At    string
	What  string
	Href  string
	// Kept says why Delete all leaves it, "" when it deletes it.
	Kept string
}

// PrivacySubjectPage is what backoffice/privacy_subject.html is rendered
// with.
type PrivacySubjectPage struct {
	Viewer Viewer
	Notice Notice
	Error  string
	Action string
	Ip     string
	Text   string
	// Searched is true once a person was looked for.
	Searched bool
	Lines    []SubjectLine
	Total    string
	// Deleted and Kept count what Delete all would delete and leave.
	Deleted     int
	Kept        int
	AccessShown bool
	HasAccess   bool
	EraseAction string
	ExportHref  string
	// Notices is the erase-kept-v1 notice a deletion of access log rows
	// waits for.
	Notices    []privacy.Notice
	AlsoAccess bool
	// IpKeyMade is the day the IP key was made: hashes before it no longer
	// match.
	IpKeyMade   string
	PrivacyHref string
}

// RenderPrivacySubjectPage answers, under status, the Find and erase page:
// the records of subject — found, nil before a search — with message above
// it, or the outcome code notice with how many rows it deleted and kept.
func RenderPrivacySubjectPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, ip string, text string, found *privacyadmin.Found, notices []privacy.Notice, alsoAccess bool, notice string, count int, kept int, message string) error {
	page := PrivacySubjectPage{
		Viewer:      viewerOf(sandbox, user),
		Error:       message,
		Action:      privacy.FindPath,
		Ip:          ip,
		Text:        text,
		EraseAction: PrivacyErasePath,
		ExportHref:  visualization.Href(sandbox, PrivacyExportPath, []visualization.Param{{Key: "ip", Value: ip}, {Key: "text", Value: text}}),
		Notices:     notices,
		AlsoAccess:  alsoAccess,
		PrivacyHref: privacy.SettingsPath,
	}
	if notice == PrivacyNoticeErased {
		page.Notice = Notice{Text: plural(sandbox, count, "row", "rows") + " deleted; " + plural(sandbox, kept, "row", "rows") + " kept, by law or by a legal hold. Snapshots taken before still hold the deleted rows until the backup retention removes them; a restore deletes them again.", Kind: "ok"}
	}
	if made, err := privacy.IpKeyRotatedAt(sandbox); err == nil && made > 0 {
		page.IpKeyMade = visualization.DayOf(sandbox, made)
	}
	if found == nil {
		return RenderHTML(sandbox, response, status, "backoffice/privacy_subject.html", page)
	}
	page.Searched = true
	page.AccessShown = found.AccessShown
	page.Total = formatCount(sandbox, found.Total())
	add := func(line SubjectLine) {
		if line.Kept == "" {
			page.Deleted++
		} else {
			page.Kept++
		}
		page.Lines = append(page.Lines, line)
	}
	held := func(table string, id int64) string {
		if found.Held[table][id] {
			return "legal hold"
		}
		return ""
	}
	for _, request := range found.Contacts {
		add(SubjectLine{Table: "contact request", At: formatInstant(sandbox, request.At), What: truncateText(request.Body, 140), Href: contact.RequestPath + "?id=" + request.Ref, Kept: held(privacy.TableContactRequest, request.Id)})
	}
	for _, summary := range found.Summaries {
		add(SubjectLine{Table: "agent summary", At: formatInstant(sandbox, summary.At), What: truncateText(summary.Summary, 140), Href: agent.SummaryPath + "?id=" + summary.Ref, Kept: held(privacy.TableAgentSummary, summary.Id)})
	}
	for _, view := range found.Views {
		add(SubjectLine{Table: "view", At: formatInstant(sandbox, view.At), What: view.Path})
	}
	for _, usage := range found.Usages {
		add(SubjectLine{Table: "agent usage", At: formatInstant(sandbox, usage.At), What: usage.Kind})
	}
	for _, row := range found.Access {
		page.HasAccess = true
		kept := held(privacy.TableAccessLog, row.Id)
		if kept == "" {
			kept = "access log"
			if found.InsideTerm[row.Id] {
				kept = "access log, inside its term"
			}
		}
		add(SubjectLine{Table: "access log", At: privacyadmin.FormatUtc(sandbox, row.At), What: row.Kind + " " + row.Ref, Kept: kept})
	}
	return RenderHTML(sandbox, response, status, "backoffice/privacy_subject.html", page)
}

// truncateText is text cut to at most limit bytes on a rune boundary, with
// an ellipsis when it was cut.
func truncateText(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	cut := limit
	for cut > 0 && text[cut]&0xC0 == 0x80 {
		cut--
	}
	return text[:cut] + "…"
}
