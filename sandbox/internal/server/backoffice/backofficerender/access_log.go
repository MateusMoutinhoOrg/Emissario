package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/privacy_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// The paths of the access log.
const (
	AccessLogPath       = privacy.AccessLogPath
	AccessLogExportPath = "/admin/root/export-access-log"
	AccessLogDeletePath = "/admin/root/delete-access-log"
)

// The outcome codes the access log page shows after a form.
const (
	AccessNoticeDeleted = "deleted"
	AccessNoticeHeld    = "held"
)

// AccessLogPage is what backoffice/access_log.html is rendered with.
type AccessLogPage struct {
	Viewer  Viewer
	Notice  Notice
	Error   string
	Notices []privacy.Notice
	// Ip, From, To and Kind refill the filter form.
	Ip    string
	From  string
	To    string
	Kind  string
	Kinds []string
	Zone  string
	Rows  []AccessLogLine
	Pager VisualizationPager
	// ExportJson and ExportCsv export every row the filter keeps.
	ExportJson   string
	ExportCsv    string
	DeleteAction string
	HoldAction   string
	// HoldUntil is the end date a hold of the rows shown defaults to.
	HoldUntil string
	Action    string
	// AccessLog is false while the access log is off.
	AccessLog bool
	// Confirm are the ids a deletion waiting for its notice holds.
	Confirm []string
}

// AccessLogLine is one row of the access log as the page shows it.
type AccessLogLine struct {
	Id      string
	AtUtc   string
	AtLocal string
	Ip      string
	Port    string
	Kind    string
	Ref     string
	RefHref string
}

// accessParams is filter as the query string of the page spells it.
func accessParams(raw privacyadmin.RawAccessFilter, filter privacyadmin.AccessFilter, ok bool) []visualization.Param {
	ip, from, to, kind := raw.Ip, raw.From, raw.To, raw.Kind
	if ok {
		ip, from, to, kind = filter.Ip, filter.From, filter.To, filter.Kind
	}
	params := []visualization.Param{{Key: "from", Value: from}, {Key: "to", Value: to}}
	if ip != "" {
		params = append(params, visualization.Param{Key: "ip", Value: ip})
	}
	if kind != "" {
		params = append(params, visualization.Param{Key: "kind", Value: kind})
	}
	return params
}

// accessRef is where the record an access log row wrote is shown, "" for
// none.
func accessRef(row privacy_db.AccessLogRecord) string {
	if row.Ref == "" {
		return ""
	}
	switch row.Kind {
	case privacy.AccessContactRequest:
		return contact.RequestPath + "?id=" + row.Ref
	case privacy.AccessAgent:
		return agent.SummaryPath + "?id=" + row.Ref
	}
	return ""
}

// AccessLogNotice is how the page words an outcome code.
func AccessLogNotice(sandbox *api.Sandbox, code string, count int) Notice {
	switch code {
	case AccessNoticeDeleted:
		return Notice{Text: plural(sandbox, count, "row", "rows") + " deleted.", Kind: "ok"}
	case AccessNoticeHeld:
		return Notice{Text: plural(sandbox, count, "row", "rows") + " put on hold.", Kind: "ok"}
	}
	return Notice{}
}

// RenderAccessLogPage answers the access log page under status: rows, one
// page of what filter keeps — refused, the raw filter and message — and the
// notices a deletion needs.
func RenderAccessLogPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, raw privacyadmin.RawAccessFilter, filter privacyadmin.AccessFilter, ok bool, rows []privacy_db.AccessLogRecord, paging visualization.Paging, notice Notice, message string, notices []privacy.Notice, confirm []int64) error {
	settings, _ := privacy.CurrentSettings(sandbox)
	params := accessParams(raw, filter, ok)
	page := AccessLogPage{
		Viewer:       viewerOf(sandbox, user),
		Notice:       notice,
		Error:        message,
		Notices:      notices,
		Ip:           raw.Ip,
		From:         raw.From,
		To:           raw.To,
		Kind:         raw.Kind,
		Kinds:        privacy.AccessKinds,
		Zone:         visualization.ZoneName(sandbox),
		DeleteAction: AccessLogDeletePath,
		HoldAction:   HoldPath,
		HoldUntil:    visualization.AddDays(sandbox, visualization.Today(sandbox), privacy.PreservationDays),
		Action:       AccessLogPath,
		AccessLog:    settings.AccessLog,
	}
	if ok {
		page.Ip, page.From, page.To, page.Kind = filter.Ip, filter.From, filter.To, filter.Kind
		page.ExportJson = visualization.Href(sandbox, AccessLogExportPath, append(params, visualization.Param{Key: "format", Value: privacyadmin.FormatJson}))
		page.ExportCsv = visualization.Href(sandbox, AccessLogExportPath, append(params, visualization.Param{Key: "format", Value: privacyadmin.FormatCsv}))
	}
	for _, row := range rows {
		port := ""
		if row.Port > 0 {
			port = sandbox.Deps.StringsDeps.FormatInt(row.Port, 10)
		}
		page.Rows = append(page.Rows, AccessLogLine{
			Id:      sandbox.Deps.StringsDeps.FormatInt(row.Id, 10),
			AtUtc:   privacyadmin.FormatUtc(sandbox, row.At),
			AtLocal: visualization.FormatLocal(sandbox, row.At/1000, visualization.TimeLayout),
			Ip:      row.Ip,
			Port:    port,
			Kind:    row.Kind,
			Ref:     row.Ref,
			RefHref: accessRef(row),
		})
	}
	for _, id := range confirm {
		page.Confirm = append(page.Confirm, sandbox.Deps.StringsDeps.FormatInt(id, 10))
	}
	page.Pager = visualizationPager(sandbox, AccessLogPath, params, paging, len(rows))
	return RenderHTML(sandbox, response, status, "backoffice/access_log.html", page)
}
