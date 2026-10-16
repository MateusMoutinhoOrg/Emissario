package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/contact_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// The pages of the contact requests every user reads — the list, as Inbox,
// Archived or All, and one request whole — share the tabs of contactTabs: a
// root also sees Settings, the page of contact_settings.go, and the buttons
// that archive, unarchive and delete.

// contactTabs are the tabs of every contact page: one per status, each
// carrying params — the filters of the list shown — and, for a root,
// Settings. active is the status shown, or contact.SettingsPath.
func contactTabs(sandbox *api.Sandbox, active string, params []visualization.Param, root bool) []VisualizationLink {
	tabs := []VisualizationLink{}
	for _, tab := range []struct {
		label  string
		status string
	}{
		{"Inbox", contact.StatusInbox},
		{"Archived", contact.StatusArchived},
		{"All", contact.StatusAll},
	} {
		status := tab.status
		if status == contact.StatusInbox {
			status = ""
		}
		href := visualization.Href(sandbox, contact.RequestsPath, visualization.SetParam(params, contact.KeyStatus, status))
		tabs = append(tabs, VisualizationLink{Label: tab.label, Href: href, Active: tab.status == active})
	}
	if root {
		tabs = append(tabs, VisualizationLink{Label: "Settings", Href: contact.SettingsPath, Active: active == contact.SettingsPath})
	}
	return tabs
}

// contactNotice is the sentence a notice code of a contact page reads as,
// count being how many contact requests it touched.
func contactNotice(sandbox *api.Sandbox, code string, count int) Notice {
	switch code {
	case contact.NoticeSaved:
		return Notice{Text: "Saved.", Kind: "ok"}
	case contact.NoticeArchived:
		return Notice{Text: plural(sandbox, count, "contact request", "contact requests") + " archived.", Kind: "ok"}
	case contact.NoticeUnarchived:
		return Notice{Text: plural(sandbox, count, "contact request", "contact requests") + " moved back to the inbox.", Kind: "ok"}
	case contact.NoticeRemoved:
		return Notice{Text: plural(sandbox, count, "contact request", "contact requests") + " deleted.", Kind: "ok"}
	}
	return Notice{}
}

// ContactRequestLine is one contact request of the list.
type ContactRequestLine struct {
	Id      string
	Href    string
	At      string
	Ip      string
	Country string
	Flag    string
	// CountryHref is the list shown, kept to this country.
	CountryHref string
	Size        string
	Preview     string
	Archived    bool
}

// contactRequestLine is request as the list at params shows it.
func contactRequestLine(sandbox *api.Sandbox, request contact_db.ContactRequestRecord, params []visualization.Param, root bool) ContactRequestLine {
	return ContactRequestLine{
		Id:          request.Ref,
		Href:        contact.RequestPath + "?id=" + request.Ref,
		At:          formatInstant(sandbox, request.At),
		Ip:          privacy.ShowIp(sandbox, request.Ip, root),
		Country:     visualization.CountryName(sandbox, request.Country),
		Flag:        visualization.CountryFlag(sandbox, request.Country),
		CountryHref: visualization.Href(sandbox, contact.RequestsPath, visualization.SetParam(visualization.SetParam(params, contact.KeyCountry, request.Country), "page", "")),
		Size:        formatBytes(sandbox, request.Size),
		Preview:     contact.Preview(sandbox, request.Body, contact.PreviewLength),
		Archived:    request.Archived != 0,
	}
}

// formatBytes is size as a page shows it: 512 B, 2.4 KB, 1.0 MB.
func formatBytes(sandbox *api.Sandbox, size int64) string {
	switch {
	case size < 1024:
		return formatCount(sandbox, int(size)) + " B"
	case size < 1024*1024:
		return formatDecimal(sandbox, float64(size)/1024, 1) + " KB"
	}
	return formatDecimal(sandbox, float64(size)/(1024*1024), 1) + " MB"
}

// ContactFilterBar is the filter bar of the list.
type ContactFilterBar struct {
	Action   string
	From     string
	To       string
	Country  string
	Contains string
	// Quick is Today, 7 days, 30 days, 90 days and All time.
	Quick  []VisualizationLink
	Chips  []VisualizationChip
	Hidden []VisualizationHidden
	// Clear is the list of the status shown with no filter at all.
	Clear string
	Zone  string
}

// contactFilterBar is the filter bar of the list for raw — as it was sent —
// and filter, as it read; params is the list shown.
func contactFilterBar(sandbox *api.Sandbox, raw contact.RawFilter, filter contact.Filter, ok bool, params []visualization.Param, limit string) ContactFilterBar {
	status := ""
	if filter.Status != contact.StatusInbox {
		status = filter.Status
	}
	bar := ContactFilterBar{
		Action:   contact.RequestsPath,
		From:     raw.From,
		To:       raw.To,
		Country:  raw.Country,
		Contains: raw.Contains,
		Clear:    visualization.Href(sandbox, contact.RequestsPath, visualizationKeep(sandbox, contact.KeyStatus, status)),
		Zone:     visualization.ZoneName(sandbox),
		Hidden:   []VisualizationHidden{},
	}
	if status != "" {
		bar.Hidden = append(bar.Hidden, VisualizationHidden{Name: contact.KeyStatus, Value: status})
	}
	if limit != "" {
		bar.Hidden = append(bar.Hidden, VisualizationHidden{Name: "limit", Value: limit})
	}
	today := visualization.Today(sandbox)
	for _, quick := range []struct {
		label string
		days  int
	}{{"Today", 1}, {"7 days", 7}, {"30 days", 30}, {"90 days", 90}, {"All time", 0}} {
		start, end := "", ""
		if quick.days > 0 {
			start, end = visualization.AddDays(sandbox, today, 1-quick.days), today
		}
		kept := visualization.SetParam(visualization.SetParam(visualization.SetParam(params, contact.KeyFrom, start), contact.KeyTo, end), "page", "")
		bar.Quick = append(bar.Quick, VisualizationLink{
			Label:  quick.label,
			Href:   visualization.Href(sandbox, contact.RequestsPath, kept),
			Active: ok && filter.From == start && filter.To == end,
		})
	}
	if !ok {
		return bar
	}
	without := func(keys ...string) string {
		kept := visualization.SetParam(params, "page", "")
		for _, key := range keys {
			kept = visualization.SetParam(kept, key, "")
		}
		return visualization.Href(sandbox, contact.RequestsPath, kept)
	}
	switch {
	case filter.From != "" && filter.To != "":
		bar.Chips = append(bar.Chips, VisualizationChip{Label: "Period", Value: filter.From + " → " + filter.To, Remove: without(contact.KeyFrom, contact.KeyTo)})
	case filter.From != "":
		bar.Chips = append(bar.Chips, VisualizationChip{Label: "Since", Value: filter.From, Remove: without(contact.KeyFrom)})
	case filter.To != "":
		bar.Chips = append(bar.Chips, VisualizationChip{Label: "Until", Value: filter.To, Remove: without(contact.KeyTo)})
	}
	if filter.Country != "" {
		bar.Chips = append(bar.Chips, VisualizationChip{Label: "Country", Value: visualization.CountryFlag(sandbox, filter.Country) + " " + visualization.CountryName(sandbox, filter.Country), Remove: without(contact.KeyCountry)})
	}
	if filter.Contains != "" {
		bar.Chips = append(bar.Chips, VisualizationChip{Label: "Contains", Value: filter.Contains, Remove: without(contact.KeyContains)})
	}
	return bar
}

// ContactRequestsPage is what backoffice/contact_requests.html is rendered
// with.
type ContactRequestsPage struct {
	Viewer Viewer
	Tabs   []VisualizationLink
	Bar    ContactFilterBar
	Notice Notice
	// Error is why the filters were refused.
	Error string
	// Status is the status shown: inbox, archived or all.
	Status string
	Rows   []ContactRequestLine
	Pager  VisualizationPager
	// ArchiveAction and RemoveAction are where a root archives and deletes
	// the contact requests selected; Back is the list shown, where they
	// come back to. All three are "" for a viewer.
	ArchiveAction string
	RemoveAction  string
	// ExportHref downloads every contact request the filter keeps as JSON.
	ExportHref string
	Back       string
	// ShowArchive and ShowUnarchive offer the buttons the status calls for.
	ShowArchive   bool
	ShowUnarchive bool
	Today         string
}

// RenderContactRequestsPage answers, under status, one page of the contact
// requests filter keeps, newest first; raw is the filters as they were sent,
// refused why they did not read, "" when they did.
func RenderContactRequestsPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, raw contact.RawFilter, filter contact.Filter, refused string, requests []contact_db.ContactRequestRecord, paging visualization.Paging, notice string, count int) error {
	viewer := viewerOf(sandbox, user)
	ok := refused == ""
	params := []visualization.Param{}
	if ok {
		params = contact.Params(sandbox, filter)
	}
	limit := limitParam(sandbox, paging.Limit)
	if limit != "" {
		params = append(params, visualization.Param{Key: "limit", Value: limit})
	}
	shown := filter.Status
	if !ok {
		shown = contact.StatusInbox
	}
	page := ContactRequestsPage{
		Viewer: viewer,
		Tabs:   contactTabs(sandbox, shown, visualization.SetParam(params, "page", ""), viewer.IsRoot),
		Bar:    contactFilterBar(sandbox, raw, filter, ok, params, limit),
		Notice: contactNotice(sandbox, notice, count),
		Error:  refused,
		Status: shown,
		Today:  visualization.Today(sandbox),
	}
	for _, request := range requests {
		page.Rows = append(page.Rows, contactRequestLine(sandbox, request, params, viewer.IsRoot))
	}
	page.Pager = visualizationPager(sandbox, contact.RequestsPath, params, paging, len(requests))
	if viewer.IsRoot {
		page.ArchiveAction = contact.ArchivePath
		page.RemoveAction = contact.RemovePath
		page.ExportHref = visualization.Href(sandbox, "/admin/root/export-contact-requests", params)
		back := params
		if paging.Page > 1 {
			back = visualization.SetParam(back, "page", sandbox.Deps.StringsDeps.FormatInt(int64(paging.Page), 10))
		}
		page.Back = visualization.Href(sandbox, contact.RequestsPath, back)
		page.ShowArchive = shown != contact.StatusArchived
		page.ShowUnarchive = shown != contact.StatusInbox
	}
	return RenderHTML(sandbox, response, status, "backoffice/contact_requests.html", page)
}

// ContactRequestPage is what backoffice/contact_request.html is rendered with.
type ContactRequestPage struct {
	Viewer      Viewer
	Tabs        []VisualizationLink
	Notice      Notice
	Id          string
	At          string
	Ip          string
	Country     string
	Flag        string
	CountryHref string
	Size        string
	Archived    bool
	// Json is the body laid out on lines, as it was sent.
	Json string
	// ListHref is the list the request is in.
	ListHref string
	// ArchiveAction and RemoveAction are where a root archives and deletes
	// it; Back is this page. All three are "" for a viewer.
	ArchiveAction string
	RemoveAction  string
	Back          string
	// Hold is the form a root holds the request with, nil for a viewer.
	Hold *HoldBox
	// BlockHref is the blocks page ready to block the ip, "" for a viewer.
	BlockHref string
}

// RenderContactRequestPage answers one contact request whole, with the
// sentence of notice above it.
func RenderContactRequestPage(sandbox *api.Sandbox, response *serverdeps.Response, user *backoffice_db.BackofficeUserRecord, request contact_db.ContactRequestRecord, notice string, count int, reason string) error {
	viewer := viewerOf(sandbox, user)
	status := contact.StatusInbox
	statusParams := []visualization.Param{}
	if request.Archived != 0 {
		status = contact.StatusArchived
		statusParams = append(statusParams, visualization.Param{Key: contact.KeyStatus, Value: status})
	}
	page := ContactRequestPage{
		Viewer:      viewer,
		Tabs:        contactTabs(sandbox, status, nil, viewer.IsRoot),
		Notice:      contactNotice(sandbox, notice, count),
		Id:          request.Ref,
		At:          formatInstant(sandbox, request.At),
		Ip:          privacy.ShowIp(sandbox, request.Ip, viewer.IsRoot),
		Country:     visualization.CountryName(sandbox, request.Country),
		Flag:        visualization.CountryFlag(sandbox, request.Country),
		CountryHref: visualization.Href(sandbox, contact.RequestsPath, visualization.SetParam(statusParams, contact.KeyCountry, request.Country)),
		Size:        formatBytes(sandbox, request.Size),
		Archived:    request.Archived != 0,
		Json:        contact.Indent(sandbox, request.Body),
		ListHref:    visualization.Href(sandbox, contact.RequestsPath, statusParams),
	}
	if viewer.IsRoot {
		page.ArchiveAction = contact.ArchivePath
		page.RemoveAction = contact.RemovePath
		page.Back = contact.RequestPath + "?id=" + request.Ref
		page.BlockHref = IpBlockHref(sandbox, request.Ip)
		hold, err := holdBoxOf(sandbox, privacy.TableContactRequest, request.Id, page.Back)
		if err != nil {
			return err
		}
		page.Hold = &hold
	}
	if held := holdNotice(sandbox, notice, reason); held.Text != "" {
		page.Notice = held
	}
	return RenderHTML(sandbox, response, api.StatusOK, "backoffice/contact_request.html", page)
}

// RenderContactRequestsRefusal answers the inbox under status, with message —
// why a form or an id was refused — above it in place of the list.
func RenderContactRequestsRefusal(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, message string) error {
	filter, _ := contact.ParseFilter(sandbox, contact.RawFilter{})
	return RenderContactRequestsPage(sandbox, response, status, user, contact.RawFilter{}, filter, message, nil, visualization.Paging{}, "", 0)
}
