package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// VisualizationVisitorsPage is what backoffice/visualization_visitors.html is
// rendered with.
type VisualizationVisitorsPage struct {
	Viewer  Viewer
	Tabs    []VisualizationLink
	Filters VisualizationFilters
	// Error is why the filters were refused, shown above the bar.
	Error string
	Sorts []VisualizationLink
	Rows  []VisualizationVisitorLine
	Pager VisualizationPager
}

// VisualizationVisitorLine is one ip of the list.
type VisualizationVisitorLine struct {
	// Ip is masked to a viewer.
	Ip string
	// ViewsHref is the views of this ip, and BlockHref the blocks page
	// ready to block it, for a root alone.
	ViewsHref string
	BlockHref string
	Flag      string
	Country   string
	Device    string
	Browser   string
	Os        string
	Views     string
	Pages     string
	FirstAt   string
	LastAt    string
}

// RenderVisualizationVisitorsPage answers, under status, one page of the ips
// seen under the filters, sorted by sort; refused is why the filters did not
// read, "" when they did.
func RenderVisualizationVisitorsPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, raw visualization.RawFilter, filter visualization.Filter, refused string, sort string, rows []visualization.VisitorRow, paging visualization.Paging) error {
	viewer := viewerOf(sandbox, user)
	ok := refused == ""
	keep := visualizationKeep(sandbox, "sort", sort, "limit", limitParam(sandbox, paging.Limit))
	page := VisualizationVisitorsPage{
		Viewer:  viewer,
		Tabs:    visualizationTabs(sandbox, visualization.VisitorsPath, filter, ok, viewer.IsRoot),
		Filters: visualizationFilterBar(sandbox, visualization.VisitorsPath, raw, filter, ok, viewer.IsRoot, keep),
		Error:   refused,
	}
	if !ok {
		return RenderHTML(sandbox, response, status, "backoffice/visualization_visitors.html", page)
	}
	params := visualization.Params(sandbox, filter)
	page.Sorts = visualizationSorts(sandbox, visualization.VisitorsPath, append(params, keep...), sort, [][2]string{
		{visualization.SortViews, "Views"}, {visualization.SortLast, "Last seen"},
	})
	for _, row := range rows {
		line := VisualizationVisitorLine{
			Ip:      privacy.ShowIp(sandbox, row.Ip, true),
			Flag:    visualization.CountryFlag(sandbox, row.Country),
			Country: visualization.CountryName(sandbox, row.Country),
			Device:  visualizationDevice(sandbox, row.Device),
			Browser: row.Browser,
			Os:      row.Os,
			Views:   formatCount(sandbox, row.Views),
			Pages:   formatCount(sandbox, row.Pages),
			FirstAt: formatInstant(sandbox, row.FirstAt),
			LastAt:  formatInstant(sandbox, row.LastAt),
		}
		if viewer.IsRoot {
			line.ViewsHref = visualization.Href(sandbox, visualization.ViewsPath, visualization.SetParam(params, visualization.KeyIp, row.Ip))
			line.BlockHref = IpBlockHref(sandbox, row.Ip)
		}
		page.Rows = append(page.Rows, line)
	}
	page.Pager = visualizationPager(sandbox, visualization.VisitorsPath, append(params, keep...), paging, len(rows))
	return RenderHTML(sandbox, response, status, "backoffice/visualization_visitors.html", page)
}
