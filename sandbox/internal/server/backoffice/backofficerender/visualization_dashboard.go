package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// VisualizationDashboardPage is what backoffice/visualization_dashboard.html
// is rendered with.
type VisualizationDashboardPage struct {
	Viewer  Viewer
	Tabs    []VisualizationLink
	Filters VisualizationFilters
	// Error is why the filters were refused, shown above the bar; the
	// metrics are left out meanwhile.
	Error           string
	Views           string
	Unique          string
	ViewsPerVisitor string
	NotFound        string
	Bots            string
	// BotsIncluded is true when the bots are counted in the rest too.
	BotsIncluded bool
	Chart        VisualizationChart
	Tables       []VisualizationTopTable
	// Credit shows the credit the embedded country database asks for.
	Credit    bool
	CreditUrl string
}

// VisualizationChart is the views over time, drawn on the server as an
// inline SVG: the backoffice runs no inline script.
type VisualizationChart struct {
	// Empty is true when the period holds no view.
	Empty bool
	// Width and Height are the size of its viewBox.
	Width  int
	Height int
	// Every says what one bar stands for: an hour, a day, a month.
	Every  string
	Bars   []VisualizationBar
	Grid   []VisualizationGridLine
	Labels []VisualizationAxisLabel
}

// VisualizationBar is one bar of the chart.
type VisualizationBar struct {
	X      string
	Y      string
	Width  string
	Height string
	// Unique is the bar of the unique visitors, beside the views'.
	Unique bool
	// Title is shown on hover.
	Title string
}

// VisualizationGridLine is one horizontal line of the chart and its value.
type VisualizationGridLine struct {
	Y      string
	LabelY string
	Label  string
}

// VisualizationAxisLabel is one label under the chart.
type VisualizationAxisLabel struct {
	X    string
	Text string
}

// VisualizationTopTable is one top table of the dashboard.
type VisualizationTopTable struct {
	Title string
	// Column names what Rows count: Page, Country, Device…
	Column string
	Rows   []VisualizationTopLine
	// SeeAll is the listing holding every row, with the same filters.
	SeeAll      string
	SeeAllLabel string
}

// VisualizationTopLine is one row of a top table.
type VisualizationTopLine struct {
	Label string
	// Flag is the flag of a country row.
	Flag string
	// Muted shows Label greyed: Direct, Unknown.
	Muted bool
	// Href adds the filter of the row to the dashboard, "" for a row that
	// is no filter.
	Href    string
	Views   string
	Unique  string
	Percent string
	// Width is Percent as a number, the width of its bar.
	Width string
}

// The geometry of the chart, in units of its viewBox.
const (
	chartWidth  = 960
	chartHeight = 260
	chartLeft   = 52
	chartRight  = 12
	chartTop    = 14
	chartBottom = 34
	// chartLabels is about how many labels the axis under the chart holds.
	chartLabels = 8
)

// RenderVisualizationDashboardPage answers, under status, the dashboard of
// the views for user: metrics when the filters read, the reason they did not
// above the filter bar otherwise.
func RenderVisualizationDashboardPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, raw visualization.RawFilter, filter visualization.Filter, refused string, metrics visualization.Metrics) error {
	viewer := viewerOf(sandbox, user)
	ok := refused == ""
	page := VisualizationDashboardPage{
		Viewer:    viewer,
		Tabs:      visualizationTabs(sandbox, visualization.DashboardPath, filter, ok, viewer.IsRoot),
		Filters:   visualizationFilterBar(sandbox, visualization.DashboardPath, raw, filter, ok, viewer.IsRoot, nil),
		Error:     refused,
		CreditUrl: visualization.GeoipCreditUrl,
	}
	geoip := visualization.CurrentGeoip(sandbox)
	page.Credit = !geoip.Open || geoip.Source == visualization.GeoipSourceEmbedded
	if !ok {
		return RenderHTML(sandbox, response, status, "backoffice/visualization_dashboard.html", page)
	}

	summary := metrics.Summary
	page.Views = formatCount(sandbox, summary.Views)
	page.Unique = formatCount(sandbox, summary.Unique)
	page.ViewsPerVisitor = formatDecimal(sandbox, summary.ViewsPerVisitor, 2)
	page.NotFound = formatCount(sandbox, summary.NotFound)
	page.Bots = formatCount(sandbox, summary.Bots)
	page.BotsIncluded = filter.Bots
	page.Chart = visualizationChart(sandbox, metrics)

	params := visualization.Params(sandbox, filter)
	link := func(key string) func(value string) string {
		return func(value string) string {
			return visualization.Href(sandbox, visualization.DashboardPath, visualization.SetParam(params, key, value))
		}
	}
	none := func(value string) string { return "" }
	views := visualization.Href(sandbox, visualization.ViewsPath, params)
	top := metrics.Top
	visitorLink := none
	if viewer.IsRoot {
		visitorLink = link(visualization.KeyIp)
	}
	page.Tables = []VisualizationTopTable{
		topTable(sandbox, "Pages", "Page", top.Pages, link(visualization.KeyPath), labelAsIs, visualization.Href(sandbox, visualization.PagesPath, params), "Every page"),
		topTable(sandbox, "Countries", "Country", top.Countries, link(visualization.KeyCountry), labelCountry, views, "Every view"),
		topTable(sandbox, "Devices", "Device", top.Devices, link(visualization.KeyDevice), labelDevice, views, "Every view"),
		topTable(sandbox, "Browsers", "Browser", top.Browsers, link(visualization.KeyBrowser), labelAsIs, views, "Every view"),
		topTable(sandbox, "Operating systems", "Os", top.Os, link(visualization.KeyOs), labelAsIs, views, "Every view"),
		topTable(sandbox, "Referrers", "Referrer", top.Referrers, none, labelReferrer, views, "Every view"),
		topTable(sandbox, "Not-found paths", "Path", top.NotFound, func(value string) string {
			return visualization.Href(sandbox, visualization.DashboardPath, visualization.SetParam(visualization.SetParam(params, visualization.KeyPath, value), visualization.KeyStatus, "404"))
		}, labelAsIs, visualization.Href(sandbox, visualization.PagesPath, visualization.SetParam(params, visualization.KeyStatus, "404")), "Every not-found page"),
		topTable(sandbox, "Top visitors", "Visitor", top.Visitors, visitorLink, labelAsIs, visualization.Href(sandbox, visualization.VisitorsPath, params), "Every visitor"),
	}
	return RenderHTML(sandbox, response, status, "backoffice/visualization_dashboard.html", page)
}

// labelAsIs is a value as it is.
func labelAsIs(sandbox *api.Sandbox, value string) (string, string, bool) {
	return value, "", false
}

// labelCountry is a country by its name and flag.
func labelCountry(sandbox *api.Sandbox, value string) (string, string, bool) {
	return visualization.CountryName(sandbox, value), visualization.CountryFlag(sandbox, value), value == visualization.UnknownCountry
}

// labelDevice is a device by its name.
func labelDevice(sandbox *api.Sandbox, value string) (string, string, bool) {
	return visualizationDevice(sandbox, value), "", false
}

// labelReferrer is a referrer, Direct for none.
func labelReferrer(sandbox *api.Sandbox, value string) (string, string, bool) {
	if value == "" {
		return "Direct", "", true
	}
	return value, "", false
}

// topTable is one top table of the dashboard: each row labelled by label and
// linked by href.
func topTable(sandbox *api.Sandbox, title string, column string, rows []visualization.TopRow, href func(value string) string, label func(sandbox *api.Sandbox, value string) (string, string, bool), seeAll string, seeAllLabel string) VisualizationTopTable {
	table := VisualizationTopTable{Title: title, Column: column, SeeAll: seeAll, SeeAllLabel: seeAllLabel}
	for _, row := range rows {
		text, flag, muted := label(sandbox, row.Value)
		table.Rows = append(table.Rows, VisualizationTopLine{
			Label:   text,
			Flag:    flag,
			Muted:   muted,
			Href:    href(row.Value),
			Views:   formatCount(sandbox, row.Views),
			Unique:  formatCount(sandbox, row.Unique),
			Percent: formatDecimal(sandbox, row.Percent, 1),
			Width:   formatDecimal(sandbox, row.Percent, 1),
		})
	}
	return table
}

// visualizationChart lays out the views over time of metrics: for each bucket
// a bar of its views and one of its unique visitors, under a scale rounded up
// by niceCeiling.
func visualizationChart(sandbox *api.Sandbox, metrics visualization.Metrics) VisualizationChart {
	chart := VisualizationChart{Width: chartWidth, Height: chartHeight, Empty: metrics.Summary.Views == 0}
	switch metrics.Bucket {
	case visualization.BucketHour:
		chart.Every = "per hour"
	case visualization.BucketMonth:
		chart.Every = "per month"
	default:
		chart.Every = "per day"
	}
	count := len(metrics.Timeline)
	if chart.Empty || count == 0 {
		return chart
	}

	most := 1
	for _, bucket := range metrics.Timeline {
		most = max(most, bucket.Views)
	}
	scale := niceCeiling(most)
	plotWidth := float64(chartWidth - chartLeft - chartRight)
	plotHeight := float64(chartHeight - chartTop - chartBottom)
	slot := plotWidth / float64(count)
	barWidth := min(slot*0.38, 26)
	number := func(value float64) string { return formatDecimal(sandbox, value, 1) }

	for _, step := range []int{0, scale / 2, scale} {
		y := float64(chartTop) + plotHeight - plotHeight*float64(step)/float64(scale)
		chart.Grid = append(chart.Grid, VisualizationGridLine{Y: number(y), LabelY: number(y + 4), Label: formatCount(sandbox, step)})
	}

	every := max(1, (count+chartLabels-1)/chartLabels)
	for index, bucket := range metrics.Timeline {
		center := float64(chartLeft) + slot*float64(index) + slot/2
		label := visualization.BucketLabel(sandbox, bucket.Start, metrics.Bucket)
		for _, bar := range []struct {
			value  int
			x      float64
			unique bool
			title  string
		}{
			{bucket.Views, center - barWidth - 1, false, label + ": " + plural(sandbox, bucket.Views, "view", "views")},
			{bucket.Unique, center + 1, true, label + ": " + plural(sandbox, bucket.Unique, "unique visitor", "unique visitors")},
		} {
			height := plotHeight * float64(bar.value) / float64(scale)
			chart.Bars = append(chart.Bars, VisualizationBar{
				X:      number(bar.x),
				Y:      number(float64(chartTop) + plotHeight - height),
				Width:  number(barWidth),
				Height: number(height),
				Unique: bar.unique,
				Title:  bar.title,
			})
		}
		if index%every == 0 {
			chart.Labels = append(chart.Labels, VisualizationAxisLabel{X: number(center), Text: label})
		}
	}
	return chart
}

// niceCeiling is the least of 1, 2, 4 or 5 times a power of ten that is at
// least value, and even, so its half is a whole number too.
func niceCeiling(value int) int {
	if value <= 2 {
		return 2
	}
	for power := 1; ; power *= 10 {
		for _, factor := range []int{1, 2, 4, 5} {
			if candidate := factor * power; candidate >= value && candidate%2 == 0 {
				return candidate
			}
		}
	}
}
