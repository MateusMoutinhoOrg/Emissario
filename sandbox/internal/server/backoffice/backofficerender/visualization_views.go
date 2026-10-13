package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/visualization_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// VisualizationViewsPage is what backoffice/visualization_views.html is
// rendered with.
type VisualizationViewsPage struct {
	Viewer  Viewer
	Tabs    []VisualizationLink
	Filters VisualizationFilters
	// Error is why the filters were refused, shown above the bar.
	Error string
	// Export is the CSV of these filters, for a root alone.
	Export string
	// Visitor is the summary of the one ip the views are filtered by.
	Visitor *VisualizationVisitorCard
	Rows    []VisualizationViewLine
	Pager   VisualizationPager
}

// VisualizationVisitorCard is the summary of one visitor.
type VisualizationVisitorCard struct {
	Ip         string
	FirstAt    string
	LastAt     string
	Views      string
	Pages      []VisualizationTopLine
	Countries  []VisualizationTopLine
	UserAgents []VisualizationTopLine
}

// VisualizationViewLine is one view of the list.
type VisualizationViewLine struct {
	At     string
	Path   string
	Status string
	// NotFound marks a not-found view.
	NotFound bool
	// Ip is masked to a viewer.
	Ip string
	// IpHref is the views of this ip, for a root alone.
	IpHref    string
	Flag      string
	Country   string
	Device    string
	Browser   string
	Os        string
	Referrer  string
	UserAgent string
}

// RenderVisualizationViewsPage answers, under status, one page of the views
// the filters keep, newest first, with the summary of the visitor when they
// keep one ip; refused is why the filters did not read, "" when they did.
func RenderVisualizationViewsPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, raw visualization.RawFilter, filter visualization.Filter, refused string, views []visualization_db.VisualizationViewRecord, paging visualization.Paging, visitor *visualization.VisitorSummary) error {
	viewer := viewerOf(sandbox, user)
	ok := refused == ""
	keep := visualizationKeep(sandbox, "limit", limitParam(sandbox, paging.Limit))
	page := VisualizationViewsPage{
		Viewer:  viewer,
		Tabs:    visualizationTabs(sandbox, visualization.ViewsPath, filter, ok, viewer.IsRoot),
		Filters: visualizationFilterBar(sandbox, visualization.ViewsPath, raw, filter, ok, viewer.IsRoot, keep),
		Error:   refused,
	}
	if !ok {
		return RenderHTML(sandbox, response, status, "backoffice/visualization_views.html", page)
	}
	params := visualization.Params(sandbox, filter)
	if viewer.IsRoot {
		page.Export = visualization.Href(sandbox, visualization.ExportPath, params)
	}
	for _, view := range views {
		line := VisualizationViewLine{
			At:        formatInstant(sandbox, view.At),
			Path:      view.Path,
			Status:    sandbox.Deps.StringsDeps.FormatInt(view.Status, 10),
			NotFound:  view.Status == visualization.StatusNotFound,
			Ip:        view.Ip,
			Flag:      visualization.CountryFlag(sandbox, view.Country),
			Country:   view.Country,
			Device:    visualizationDevice(sandbox, view.Device),
			Browser:   view.Browser,
			Os:        view.Os,
			Referrer:  view.Referrer,
			UserAgent: view.UserAgent,
		}
		if viewer.IsRoot && filter.Ip != view.Ip {
			line.IpHref = visualization.Href(sandbox, visualization.ViewsPath, visualization.SetParam(params, visualization.KeyIp, view.Ip))
		}
		page.Rows = append(page.Rows, line)
	}
	if visitor != nil {
		card := &VisualizationVisitorCard{
			Ip:      visitor.Ip,
			FirstAt: formatInstant(sandbox, visitor.FirstAt),
			LastAt:  formatInstant(sandbox, visitor.LastAt),
			Views:   formatCount(sandbox, visitor.Views),
		}
		card.Pages = topTable(sandbox, "", "", visitor.Pages, func(value string) string {
			return visualization.Href(sandbox, visualization.ViewsPath, visualization.SetParam(params, visualization.KeyPath, value))
		}, labelAsIs, "", "").Rows
		card.Countries = topTable(sandbox, "", "", visitor.Countries, func(value string) string { return "" }, labelCountry, "", "").Rows
		card.UserAgents = topTable(sandbox, "", "", visitor.UserAgents, func(value string) string { return "" }, labelAsIs, "", "").Rows
		page.Visitor = card
	}
	page.Pager = visualizationPager(sandbox, visualization.ViewsPath, append(params, keep...), paging, len(views))
	return RenderHTML(sandbox, response, status, "backoffice/visualization_views.html", page)
}
