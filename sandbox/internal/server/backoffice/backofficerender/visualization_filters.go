package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// What every page of the visualization shows around its content: the tabs
// between them, the filter bar with its quick periods and its chips, and the
// pager. Every link carries the filters of the page it is on, so moving from
// the dashboard to the views of one country keeps the period and the rest.

// VisualizationOption is one choice of a select of the filter bar.
type VisualizationOption struct {
	Value    string
	Label    string
	Selected bool
}

// VisualizationLink is one link of a row of links: a tab, a quick period, a
// sort order.
type VisualizationLink struct {
	Label  string
	Href   string
	Active bool
}

// VisualizationChip is one active filter, shown with an ✕ that removes it.
type VisualizationChip struct {
	Label string
	Value string
	// Remove is the page with every filter but this one.
	Remove string
}

// VisualizationHidden is one hidden input of the filter bar: what the page
// keeps across a new filter — a sort, a page size.
type VisualizationHidden struct {
	Name  string
	Value string
}

// VisualizationFilters is the filter bar of a visualization page.
type VisualizationFilters struct {
	// Action is the page the bar submits to.
	Action  string
	From    string
	To      string
	Path    string
	Prefix  string
	Ip      string
	Country string
	Bots    bool
	// ShowIp offers the ip filter: a root alone filters by it.
	ShowIp   bool
	Devices  []VisualizationOption
	Browsers []VisualizationOption
	Systems  []VisualizationOption
	Statuses []VisualizationOption
	// Quick is Today, 7 days, 30 days and 90 days.
	Quick  []VisualizationLink
	Chips  []VisualizationChip
	Hidden []VisualizationHidden
	// Clear is the page with no filter at all.
	Clear string
	// Period is the period shown, in words.
	Period string
	// Zone is the server's time zone the days are counted in.
	Zone string
}

// VisualizationPager is the pager of a listing.
type VisualizationPager struct {
	Total string
	Page  int
	Pages int
	// From and To number the first and the last row shown, counted from 1;
	// both 0 for an empty page.
	From     int
	To       int
	HasPrev  bool
	PrevHref string
	HasNext  bool
	NextHref string
	// Limits are the page sizes offered.
	Limits []LimitOption
}

// deviceLabels is how a device is named on a page.
var deviceLabels = map[string]string{
	visualization.DeviceDesktop: "Desktop",
	visualization.DeviceMobile:  "Mobile",
	visualization.DeviceTablet:  "Tablet",
	visualization.DeviceBot:     "Bot",
	visualization.DeviceOther:   "Other",
}

// visualizationDevice is how device is named on a page.
func visualizationDevice(sandbox *api.Sandbox, device string) string {
	if label, ok := deviceLabels[device]; ok {
		return label
	}
	return device
}

// visualizationOptions are values offered by a select, all of them first
// under allLabel, the one selected marked.
func visualizationOptions(sandbox *api.Sandbox, allLabel string, values []string, labels func(value string) string, selected string) []VisualizationOption {
	options := []VisualizationOption{{Value: "", Label: allLabel, Selected: selected == ""}}
	for _, value := range values {
		options = append(options, VisualizationOption{Value: value, Label: labels(value), Selected: value == selected})
	}
	return options
}

// visualizationStatus is how a status filter is named on a page.
func visualizationStatus(sandbox *api.Sandbox, status string) string {
	switch status {
	case "200":
		return "Found (200)"
	case "404":
		return "Not found (404)"
	}
	return status
}

// visualizationFilterBar is the filter bar of the page at action. filter is
// what the request was read into, or — when ok is false, a filter that was
// refused — what raw sent, shown back as it is so it can be corrected. keep is
// what the page carries across a new filter.
func visualizationFilterBar(sandbox *api.Sandbox, action string, raw visualization.RawFilter, filter visualization.Filter, ok bool, root bool, keep []visualization.Param) VisualizationFilters {
	strings := sandbox.Deps.StringsDeps
	bar := VisualizationFilters{
		Action: action,
		ShowIp: root,
		Clear:  action,
		Zone:   visualization.ZoneName(sandbox),
	}
	for _, param := range keep {
		bar.Hidden = append(bar.Hidden, VisualizationHidden{Name: param.Key, Value: param.Value})
	}

	status := ""
	device, browser, system := "", "", ""
	params := []visualization.Param{}
	if ok {
		bar.From, bar.To = filter.From, filter.To
		bar.Path, bar.Prefix, bar.Ip, bar.Country = filter.Path, filter.Prefix, filter.Ip, filter.Country
		bar.Bots = filter.Bots
		device, browser, system = filter.Device, filter.Browser, filter.Os
		if filter.Status != 0 {
			status = strings.FormatInt(filter.Status, 10)
		}
		params = append(visualization.Params(sandbox, filter), keep...)
		bar.Period = filter.From + " → " + filter.To + " · " + plural(sandbox, filter.Days, "day", "days")
		if filter.From == filter.To {
			bar.Period = filter.From + " · 1 day"
		}
	} else {
		bar.From, bar.To = strings.TrimSpace(raw.From), strings.TrimSpace(raw.To)
		bar.Path, bar.Prefix, bar.Country = strings.TrimSpace(raw.Path), strings.TrimSpace(raw.Prefix), strings.TrimSpace(raw.Country)
		if root {
			bar.Ip = strings.TrimSpace(raw.Ip)
		}
		bar.Bots = raw.Bots == "1" || raw.Bots == "true"
		device, browser, system, status = raw.Device, raw.Browser, raw.Os, raw.Status
	}

	bar.Devices = visualizationOptions(sandbox, "Every device", visualization.Devices, func(value string) string { return visualizationDevice(sandbox, value) }, device)
	same := func(value string) string { return value }
	bar.Browsers = visualizationOptions(sandbox, "Every browser", visualization.Browsers, same, browser)
	bar.Systems = visualizationOptions(sandbox, "Every system", visualization.OperatingSystems, same, system)
	bar.Statuses = visualizationOptions(sandbox, "Every status", []string{"200", "404"}, func(value string) string { return visualizationStatus(sandbox, value) }, status)

	today := visualization.Today(sandbox)
	for _, quick := range []struct {
		label string
		days  int
	}{{"Today", 1}, {"7 days", 7}, {"30 days", 30}, {"90 days", 90}} {
		from := visualization.AddDays(sandbox, today, 1-quick.days)
		period := visualization.SetParam(visualization.SetParam(params, visualization.KeyFrom, from), visualization.KeyTo, today)
		if !ok {
			period = []visualization.Param{{Key: visualization.KeyFrom, Value: from}, {Key: visualization.KeyTo, Value: today}}
		}
		bar.Quick = append(bar.Quick, VisualizationLink{
			Label:  quick.label,
			Href:   visualization.Href(sandbox, action, period),
			Active: ok && filter.From == from && filter.To == today,
		})
	}
	if ok {
		bar.Chips = visualizationChips(sandbox, action, filter, params)
	}
	return bar
}

// visualizationChips is one chip per filter of filter that is on, each
// removing itself from params.
func visualizationChips(sandbox *api.Sandbox, action string, filter visualization.Filter, params []visualization.Param) []VisualizationChip {
	chips := []VisualizationChip{}
	without := func(keys ...string) string {
		kept := params
		for _, key := range keys {
			kept = visualization.SetParam(kept, key, "")
		}
		return visualization.Href(sandbox, action, kept)
	}
	defaultFrom := visualization.AddDays(sandbox, visualization.Today(sandbox), 1-visualization.DefaultPeriodDays)
	if filter.From != defaultFrom || filter.To != visualization.Today(sandbox) {
		chips = append(chips, VisualizationChip{Label: "Period", Value: filter.From + " → " + filter.To, Remove: without(visualization.KeyFrom, visualization.KeyTo)})
	}
	add := func(label string, value string, key string) {
		if value != "" {
			chips = append(chips, VisualizationChip{Label: label, Value: value, Remove: without(key)})
		}
	}
	add("Page", filter.Path, visualization.KeyPath)
	add("Pages starting with", filter.Prefix, visualization.KeyPrefix)
	add("Ip", filter.Ip, visualization.KeyIp)
	if filter.Country != "" {
		add("Country", visualization.CountryFlag(sandbox, filter.Country)+" "+visualization.CountryName(sandbox, filter.Country), visualization.KeyCountry)
	}
	if filter.Device != "" {
		add("Device", visualizationDevice(sandbox, filter.Device), visualization.KeyDevice)
	}
	add("Browser", filter.Browser, visualization.KeyBrowser)
	add("Os", filter.Os, visualization.KeyOs)
	if filter.Status != 0 {
		add("Status", visualizationStatus(sandbox, sandbox.Deps.StringsDeps.FormatInt(filter.Status, 10)), visualization.KeyStatus)
	}
	if filter.Bots && filter.Device != visualization.DeviceBot {
		add("Bots", "included", visualization.KeyBots)
	}
	return chips
}

// visualizationTabs are the tabs of the visualization pages, active the one
// shown, each carrying the filters of the page when they read.
func visualizationTabs(sandbox *api.Sandbox, active string, filter visualization.Filter, ok bool, root bool) []VisualizationLink {
	params := []visualization.Param{}
	if ok {
		params = visualization.Params(sandbox, filter)
	}
	tabs := []VisualizationLink{}
	for _, tab := range []struct {
		label string
		path  string
	}{
		{"Dashboard", visualization.DashboardPath},
		{"Pages", visualization.PagesPath},
		{"Visitors", visualization.VisitorsPath},
		{"Views", visualization.ViewsPath},
	} {
		tabs = append(tabs, VisualizationLink{Label: tab.label, Href: visualization.Href(sandbox, tab.path, params), Active: tab.path == active})
	}
	if root {
		tabs = append(tabs, VisualizationLink{Label: "Settings", Href: visualization.SettingsPath, Active: active == visualization.SettingsPath})
	}
	return tabs
}

// visualizationPager is the pager of the listing at action whose rows are
// shown counted rows at a time, params being its filters and what it keeps.
func visualizationPager(sandbox *api.Sandbox, action string, params []visualization.Param, paging visualization.Paging, shown int) VisualizationPager {
	pager := VisualizationPager{
		Total:   formatCount(sandbox, paging.Total),
		Page:    paging.Page,
		Pages:   paging.Pages,
		HasPrev: paging.Page > 1,
		HasNext: paging.Page < paging.Pages,
		Limits:  limitOptions(sandbox, paging.Limit),
	}
	params = visualization.SetParam(params, "limit", sandbox.Deps.StringsDeps.FormatInt(int64(paging.Limit), 10))
	if pager.HasPrev {
		pager.PrevHref = visualization.Href(sandbox, action, visualization.SetParam(params, "page", sandbox.Deps.StringsDeps.FormatInt(int64(paging.Page-1), 10)))
	}
	if pager.HasNext {
		pager.NextHref = visualization.Href(sandbox, action, visualization.SetParam(params, "page", sandbox.Deps.StringsDeps.FormatInt(int64(paging.Page+1), 10)))
	}
	if shown > 0 {
		pager.From = (paging.Page-1)*paging.Limit + 1
		pager.To = pager.From + shown - 1
	}
	return pager
}

// formatCount is count with its thousands set apart by commas: 12,345.
func formatCount(sandbox *api.Sandbox, count int) string {
	digits := sandbox.Deps.StringsDeps.FormatInt(int64(count), 10)
	sign := ""
	if count < 0 {
		sign, digits = "-", digits[1:]
	}
	grouped := ""
	for len(digits) > 3 {
		grouped = "," + digits[len(digits)-3:] + grouped
		digits = digits[:len(digits)-3]
	}
	return sign + digits + grouped
}

// formatDecimal is value with places decimals.
func formatDecimal(sandbox *api.Sandbox, value float64, places int) string {
	return sandbox.Deps.StringsDeps.FormatFloat(value, 'f', places, 64)
}

// formatInstant is the instant at in the server's time zone, "—" for 0.
func formatInstant(sandbox *api.Sandbox, at int64) string {
	if at <= 0 {
		return "—"
	}
	return visualization.FormatLocal(sandbox, at, visualization.TimeLayout)
}

// plural is count followed by one or many, as count calls for.
func plural(sandbox *api.Sandbox, count int, one string, many string) string {
	if count == 1 {
		return "1 " + one
	}
	return formatCount(sandbox, count) + " " + many
}

// visualizationKeep is what a page keeps across a new filter, given as key,
// value, key, value…: every pair whose value is not empty.
func visualizationKeep(sandbox *api.Sandbox, pairs ...string) []visualization.Param {
	keep := []visualization.Param{}
	for index := 0; index+1 < len(pairs); index += 2 {
		if pairs[index+1] != "" {
			keep = append(keep, visualization.Param{Key: pairs[index], Value: pairs[index+1]})
		}
	}
	return keep
}
