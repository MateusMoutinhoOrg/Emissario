package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// VisualizationPagesPage is what backoffice/visualization_pages.html is
// rendered with.
type VisualizationPagesPage struct {
	Viewer  Viewer
	Tabs    []VisualizationLink
	Filters VisualizationFilters
	// Error is why the filters were refused, shown above the bar.
	Error string
	// Sorts are the orders the list is offered in.
	Sorts []VisualizationLink
	Rows  []VisualizationPageLine
	Pager VisualizationPager
}

// VisualizationPageLine is one canonical page of the list.
type VisualizationPageLine struct {
	Page string
	// Dashboard is the dashboard of this page alone.
	Dashboard string
	// ViewsHref is the views of this page.
	ViewsHref string
	Views     string
	Unique    string
	NotFound  bool
	LastAt    string
}

// visualizationSorts are the sort orders of a listing at action, each a link
// keeping params, the one shown marked.
func visualizationSorts(sandbox *api.Sandbox, action string, params []visualization.Param, current string, orders [][2]string) []VisualizationLink {
	if current == "" {
		current = visualization.SortViews
	}
	links := []VisualizationLink{}
	for _, order := range orders {
		links = append(links, VisualizationLink{
			Label:  order[1],
			Href:   visualization.Href(sandbox, action, visualization.SetParam(params, "sort", order[0])),
			Active: order[0] == current,
		})
	}
	return links
}

// RenderVisualizationPagesPage answers, under status, one page of the
// canonical pages viewed under the filters, sorted by sort; refused is why the
// filters did not read, "" when they did.
func RenderVisualizationPagesPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, raw visualization.RawFilter, filter visualization.Filter, refused string, sort string, rows []visualization.PageRow, paging visualization.Paging) error {
	viewer := viewerOf(sandbox, user)
	ok := refused == ""
	keep := visualizationKeep(sandbox, "sort", sort, "limit", limitParam(sandbox, paging.Limit))
	page := VisualizationPagesPage{
		Viewer:  viewer,
		Tabs:    visualizationTabs(sandbox, visualization.PagesPath, filter, ok, viewer.IsRoot),
		Filters: visualizationFilterBar(sandbox, visualization.PagesPath, raw, filter, ok, viewer.IsRoot, keep),
		Error:   refused,
	}
	if !ok {
		return RenderHTML(sandbox, response, status, "backoffice/visualization_pages.html", page)
	}
	params := visualization.Params(sandbox, filter)
	page.Sorts = visualizationSorts(sandbox, visualization.PagesPath, append(params, keep...), sort, [][2]string{
		{visualization.SortViews, "Views"}, {visualization.SortUnique, "Unique visitors"}, {visualization.SortLast, "Last view"}, {visualization.SortPage, "Page"},
	})
	for _, row := range rows {
		page.Rows = append(page.Rows, VisualizationPageLine{
			Page:      row.Page,
			Dashboard: visualization.Href(sandbox, visualization.DashboardPath, visualization.SetParam(params, visualization.KeyPath, row.Page)),
			ViewsHref: visualization.Href(sandbox, visualization.ViewsPath, visualization.SetParam(params, visualization.KeyPath, row.Page)),
			Views:     formatCount(sandbox, row.Views),
			Unique:    formatCount(sandbox, row.Unique),
			NotFound:  row.NotFound,
			LastAt:    formatInstant(sandbox, row.LastAt),
		})
	}
	page.Pager = visualizationPager(sandbox, visualization.PagesPath, append(params, keep...), paging, len(rows))
	return RenderHTML(sandbox, response, status, "backoffice/visualization_pages.html", page)
}

// limitParam is limit as a query string spells it, "" for the default one.
func limitParam(sandbox *api.Sandbox, limit int) string {
	if limit == 0 || limit == visualization.DefaultLimit {
		return ""
	}
	return sandbox.Deps.StringsDeps.FormatInt(int64(limit), 10)
}
