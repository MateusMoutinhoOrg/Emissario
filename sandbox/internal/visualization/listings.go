package visualization

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/visualization_db"
)

// PageRow is one canonical page of the Pages listing.
type PageRow struct {
	Page string
	// Views is how many views it had in the period.
	Views int
	// Unique is how many distinct ips viewed it.
	Unique int
	// NotFound is true when it answered the not-found page.
	NotFound bool
	// LastAt is when it was last viewed in the period.
	LastAt int64
}

// VisitorRow is one ip of the Visitors listing.
type VisitorRow struct {
	// Ip is masked to a viewer.
	Ip string
	// Country is the one of its last view.
	Country string
	// Device, Browser and Os are the ones it used most.
	Device  string
	Browser string
	Os      string
	// Views is how many views it made in the period.
	Views int
	// Pages is how many distinct pages it saw.
	Pages int
	// FirstAt and LastAt are its first and last view in the period.
	FirstAt int64
	LastAt  int64
}

// Paging is where one page of a listing sits among every page of it.
type Paging struct {
	// Total is how many rows the listing holds.
	Total int
	// Page is the page shown, counted from 1.
	Page int
	// Pages is how many pages the listing holds, 1 at least.
	Pages int
	// Limit is how many rows a page holds.
	Limit int
}

// Paginate is the page of a listing of total rows the request asks — the
// first for one below 1, the last for one past it — with a limit of
// DefaultLimit when below 1 and MaxLimit at most, and the rows it holds,
// from (included) to (excluded).
func Paginate(sandbox *api.Sandbox, total int, page int, limit int) (paging Paging, from int, to int) {
	paging.Total = total
	paging.Limit = limit
	if paging.Limit < 1 {
		paging.Limit = DefaultLimit
	}
	paging.Limit = min(paging.Limit, MaxLimit)
	paging.Pages = max(1, (total+paging.Limit-1)/paging.Limit)
	paging.Page = min(max(page, 1), paging.Pages)
	from = (paging.Page - 1) * paging.Limit
	return paging, from, min(from+paging.Limit, total)
}

// ListPages is one page of every canonical page selection holds, sorted by
// views (the default), unique, last or page. Another sort is refused.
func ListPages(sandbox *api.Sandbox, selection Selection, sort string, page int, limit int) ([]PageRow, Paging, *Refusal) {
	sort = sandbox.Deps.StringsDeps.TrimSpace(sort)
	if sort == "" {
		sort = SortViews
	}
	if !oneOf(sort, []string{SortViews, SortUnique, SortLast, SortPage}) {
		return nil, Paging{}, refuse("sort", "sort is views, unique, last or page.")
	}
	rows := []PageRow{}
	position := map[string]int{}
	ips := []map[string]bool{}
	for _, view := range selection.Views {
		at, ok := position[view.Page]
		if !ok {
			at = len(rows)
			position[view.Page] = at
			rows = append(rows, PageRow{Page: view.Page})
			ips = append(ips, map[string]bool{})
		}
		rows[at].Views++
		ips[at][view.Ip] = true
		rows[at].NotFound = rows[at].NotFound || view.Status == StatusNotFound
		rows[at].LastAt = max(rows[at].LastAt, view.At)
	}
	for at := range rows {
		rows[at].Unique = len(ips[at])
	}
	sandbox.Deps.SortDeps.Slice(rows, func(i int, j int) bool {
		switch sort {
		case SortUnique:
			if rows[i].Unique != rows[j].Unique {
				return rows[i].Unique > rows[j].Unique
			}
		case SortLast:
			if rows[i].LastAt != rows[j].LastAt {
				return rows[i].LastAt > rows[j].LastAt
			}
		case SortPage:
			return rows[i].Page < rows[j].Page
		}
		if rows[i].Views != rows[j].Views {
			return rows[i].Views > rows[j].Views
		}
		return rows[i].Page < rows[j].Page
	})
	paging, from, to := Paginate(sandbox, len(rows), page, limit)
	return rows[from:to], paging, nil
}

// ListVisitors is one page of every ip selection holds, sorted by views (the
// default) or last, each ip masked unless root. Another sort is refused.
func ListVisitors(sandbox *api.Sandbox, selection Selection, sort string, page int, limit int, root bool) ([]VisitorRow, Paging, *Refusal) {
	sort = sandbox.Deps.StringsDeps.TrimSpace(sort)
	if sort == "" {
		sort = SortViews
	}
	if !oneOf(sort, []string{SortViews, SortLast}) {
		return nil, Paging{}, refuse("sort", "sort is views or last.")
	}
	type tally struct {
		row      VisitorRow
		pages    map[string]bool
		devices  map[string]int
		browsers map[string]int
		systems  map[string]int
	}
	tallies := []*tally{}
	position := map[string]int{}
	for _, view := range selection.Views {
		at, ok := position[view.Ip]
		if !ok {
			at = len(tallies)
			position[view.Ip] = at
			tallies = append(tallies, &tally{
				row:   VisitorRow{Ip: view.Ip, FirstAt: view.At, LastAt: view.At, Country: view.Country},
				pages: map[string]bool{}, devices: map[string]int{}, browsers: map[string]int{}, systems: map[string]int{},
			})
		}
		current := tallies[at]
		current.row.Views++
		current.pages[view.Page] = true
		current.devices[view.Device]++
		current.browsers[view.Browser]++
		current.systems[view.Os]++
		if view.At < current.row.FirstAt {
			current.row.FirstAt = view.At
		}
		if view.At > current.row.LastAt {
			current.row.LastAt = view.At
			current.row.Country = view.Country
		}
	}
	rows := []VisitorRow{}
	for _, current := range tallies {
		current.row.Pages = len(current.pages)
		current.row.Device = mostUsed(sandbox, current.devices)
		current.row.Browser = mostUsed(sandbox, current.browsers)
		current.row.Os = mostUsed(sandbox, current.systems)
		rows = append(rows, current.row)
	}
	sandbox.Deps.SortDeps.Slice(rows, func(i int, j int) bool {
		if sort == SortLast && rows[i].LastAt != rows[j].LastAt {
			return rows[i].LastAt > rows[j].LastAt
		}
		if rows[i].Views != rows[j].Views {
			return rows[i].Views > rows[j].Views
		}
		if rows[i].LastAt != rows[j].LastAt {
			return rows[i].LastAt > rows[j].LastAt
		}
		return rows[i].Ip < rows[j].Ip
	})
	paging, from, to := Paginate(sandbox, len(rows), page, limit)
	shown := rows[from:to]
	for index := range shown {
		shown[index].Ip = ShownIp(sandbox, shown[index].Ip, root)
	}
	return shown, paging, nil
}

// mostUsed is the key of counts with the highest count, the first by name
// on a tie.
func mostUsed(sandbox *api.Sandbox, counts map[string]int) string {
	best, most := "", -1
	for key, count := range counts {
		if count > most || count == most && key < best {
			best, most = key, count
		}
	}
	return best
}

// ListViews is one page of the views of selection, newest first, each ip
// masked unless root.
func ListViews(sandbox *api.Sandbox, selection Selection, page int, limit int, root bool) ([]visualization_db.VisualizationViewRecord, Paging) {
	paging, from, to := Paginate(sandbox, len(selection.Views), page, limit)
	views := append([]visualization_db.VisualizationViewRecord{}, selection.Views[from:to]...)
	for index := range views {
		views[index].Ip = ShownIp(sandbox, views[index].Ip, root)
	}
	return views, paging
}

// VisitorSummary is everything one ip did in the period.
type VisitorSummary struct {
	Ip      string
	FirstAt int64
	LastAt  int64
	Views   int
	// Pages is every page it saw, the most viewed first.
	Pages []TopRow
	// Countries is every country it came from.
	Countries []TopRow
	// UserAgents is every user agent it came with.
	UserAgents []TopRow
}

// SummarizeVisitor is the VisitorSummary of the one ip selection is filtered
// by; false when it is filtered by no ip, or by a range.
func SummarizeVisitor(sandbox *api.Sandbox, selection Selection) (VisitorSummary, bool) {
	if selection.Filter.Ip == "" || selection.Filter.IpRange {
		return VisitorSummary{}, false
	}
	summary := VisitorSummary{Ip: selection.Filter.Ip, Views: len(selection.Views)}
	for _, view := range selection.Views {
		if summary.FirstAt == 0 || view.At < summary.FirstAt {
			summary.FirstAt = view.At
		}
		summary.LastAt = max(summary.LastAt, view.At)
	}
	summary.Pages = rank(sandbox, selection.Views, 0, func(view visualization_db.VisualizationViewRecord) string { return view.Page })
	summary.Countries = rank(sandbox, selection.Views, 0, func(view visualization_db.VisualizationViewRecord) string { return view.Country })
	summary.UserAgents = rank(sandbox, selection.Views, 0, func(view visualization_db.VisualizationViewRecord) string { return view.UserAgent })
	return summary, true
}

// HomeCounts is how many views — bots left out — the website had today and
// in the last DefaultPeriodDays days, for the home page of the backoffice.
func HomeCounts(sandbox *api.Sandbox) (int, int, error) {
	filter, refusal := ParseFilter(sandbox, RawFilter{}, true)
	if refusal != nil {
		return 0, 0, sandbox.Deps.StdDeps.Errorf("%s", refusal.Message)
	}
	selection, err := Select(sandbox, filter)
	if err != nil {
		return 0, 0, err
	}
	todayStart, _ := ParseDay(sandbox, filter.To)
	today := 0
	for _, view := range selection.Views {
		if view.At >= todayStart {
			today++
		}
	}
	return today, len(selection.Views), nil
}
