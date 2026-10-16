package visualization

// The visualization is the record of every page of the public website that is
// viewed — which page, when, and from which ip — and what the backoffice reads
// out of it at /admin/visualization. Everything is server side, read from the
// request the front route answered: no cookie, no script on the website, no
// analytics service. The front route only puts a Hit on an in-memory queue
// (queue.go) and returns; one goroutine, started by the
// visualization-start-server middleware, turns each one into a
// visualization-view of visualization-db, with its canonical page (page.go),
// country (geoip.go), device, browser and os (useragent.go) and referrer
// (referrer.go). The views of one period are searched and counted by
// filter.go, metrics.go and listings.go; the settings a root keeps are in
// settings.go and the views older than the retention are removed by
// retention.go. The ip is personal data: it is masked to a viewer (ip.go) and
// a root can remove the views at any time.

const (
	// DashboardPath is the dashboard of the backoffice.
	DashboardPath = "/admin/visualization"
	// PagesPath lists every canonical page viewed in a period.
	PagesPath = "/admin/list-visualization-pages"
	// VisitorsPath lists every ip seen in a period.
	VisitorsPath = "/admin/list-visualization-visitors"
	// ViewsPath lists the views themselves, newest first.
	ViewsPath = "/admin/list-visualization-views"
	// ExportPath answers the views of a period as CSV, for a root.
	ExportPath = "/admin/root/export-visualization-views"
	// SettingsPath is the settings page, and where its form posts.
	SettingsPath = "/admin/root/set-visualization-settings"
	// PurgePath is where the Delete views form posts.
	PurgePath = "/admin/root/purge-visualization-views"
)

const (
	// QueueSize is how many views wait in memory for the writer at most: a
	// view arriving with the queue full is dropped, never waited on.
	QueueSize = 10000
	// MaxPathLength is the longest path kept on a view.
	MaxPathLength = 512
	// MaxUserAgentLength is the longest User-Agent kept on a view.
	MaxUserAgentLength = 512
	// MaxPeriodDays is the longest period one query covers.
	MaxPeriodDays = 366
	// DefaultPeriodDays is the period a query names no date of: the last 7
	// days, today included.
	DefaultPeriodDays = 7
	// DefaultRetentionDays is how many days a view is kept before a root
	// saves the settings.
	DefaultRetentionDays = 90
	// MaxRetentionDays is the longest retention a root may set: a century,
	// as every retention of the privacy package. Past 760 days a save asks
	// for the retention notice; none is refused for being long.
	MaxRetentionDays = 36500
	// MaxIgnoredIps is how many ips and ranges the ignored ips hold at most.
	MaxIgnoredIps = 100
	// DefaultLimit is how many rows a listing page shows when none is asked.
	DefaultLimit = 20
	// MaxLimit is the most rows a listing page shows.
	MaxLimit = 100
	// DefaultTop is how many rows each top table of the metrics holds.
	DefaultTop = 10
	// MaxTop is the most rows a top table of the metrics holds.
	MaxTop = 100
	// MaxExportRows is the most views one CSV export holds, the newest ones.
	MaxExportRows = 100000
)

// The status a view is recorded with.
const (
	// StatusFound is a page the website answered.
	StatusFound = 200
	// StatusNotFound is a not-found view: a path the website answered with
	// its not-found page.
	StatusNotFound = 404
)

// The device a view was made from.
const (
	DeviceDesktop = "desktop"
	DeviceMobile  = "mobile"
	DeviceTablet  = "tablet"
	DeviceBot     = "bot"
	DeviceOther   = "other"
)

// Devices is every device a view is recorded with, in the order a filter
// offers them.
var Devices = []string{DeviceDesktop, DeviceMobile, DeviceTablet, DeviceBot, DeviceOther}

// Browsers is every browser a view is recorded with.
var Browsers = []string{"Chrome", "Firefox", "Safari", "Edge", "Opera", "Samsung Internet", "Other"}

// OperatingSystems is every operating system a view is recorded with.
var OperatingSystems = []string{"Windows", "macOS", "Linux", "Android", "iOS", "ChromeOS", "Other"}

// UnknownCountry is the country of a view whose ip no country is known for:
// a private, loopback or invalid ip, or one the database does not hold.
const UnknownCountry = "XX"

// The notice codes the settings page reads off ?notice=.
const (
	// NoticeSaved follows the settings saved.
	NoticeSaved = "saved"
	// NoticePurged follows views removed; the count follows it as
	// ?notice=purged&removed=<n>.
	NoticePurged = "purged"
)

// The sort orders of the listings.
const (
	SortViews  = "views"
	SortUnique = "unique"
	SortLast   = "last"
	SortPage   = "page"
)

// The buckets of the views over time.
const (
	BucketHour  = "hour"
	BucketDay   = "day"
	BucketMonth = "month"
)

// truncate is text cut to at most length bytes, never in the middle of a
// UTF-8 sequence.
func truncate(text string, length int) string {
	if len(text) <= length {
		return text
	}
	cut := length
	for cut > 0 && text[cut]&0xC0 == 0x80 {
		cut--
	}
	return text[:cut]
}

// oneOf tells whether value is one of values.
func oneOf(value string, values []string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
