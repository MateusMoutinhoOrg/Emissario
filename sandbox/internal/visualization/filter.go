package visualization

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/visualization_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficehttp"
)

// Every page and every api of the visualization reads the same filters, under
// the same keys: the query string of a page, the JSON body of an api.

// The keys of the filters, in the order a query string spells them.
const (
	KeyFrom    = "from"
	KeyTo      = "to"
	KeyPath    = "path"
	KeyPrefix  = "prefix"
	KeyIp      = "ip"
	KeyCountry = "country"
	KeyDevice  = "device"
	KeyBrowser = "browser"
	KeyOs      = "os"
	KeyStatus  = "status"
	KeyBots    = "bots"
)

// RawFilter is the filters as a request spells them, every one a string: ""
// is a filter left off.
type RawFilter struct {
	From    string
	To      string
	Path    string
	Prefix  string
	Ip      string
	Country string
	Device  string
	Browser string
	Os      string
	Status  string
	Bots    string
}

// Filter is a RawFilter read and checked.
type Filter struct {
	// From and To are the first and the last local day of the period, both
	// included, spelled DayLayout: the last DefaultPeriodDays when the
	// request named neither.
	From string
	To   string
	// AtMin and AtMax are the first and the last instant of the period.
	AtMin int64
	AtMax int64
	// Days is how many days the period covers.
	Days int
	// Path is one canonical page, exactly; "" for every one.
	Path string
	// Prefix keeps the pages starting with it; "" for every one.
	Prefix string
	// Ip is one ip or one CIDR range, in its canonical spelling; "" for
	// every one. Only a root filters by it.
	Ip string
	// IpRange is true when Ip is a CIDR range.
	IpRange bool
	// Country is one ISO code, upper case.
	Country string
	// Device, Browser and Os are one value of Devices, Browsers and
	// OperatingSystems.
	Device  string
	Browser string
	Os      string
	// Status is StatusFound or StatusNotFound; 0 for both.
	Status int64
	// Bots includes the views of bots, left out of everything otherwise. A
	// filter on the bot device includes them.
	Bots bool
}

// Refusal is why a filter is refused: Status is api.StatusBadRequest for a
// value that does not read, api.StatusForbidden for one the reader may not
// send; Field is the key it is about.
type Refusal struct {
	Status  int
	Field   string
	Message string
}

// refuse is a Refusal of field under a 400.
func refuse(field string, message string) *Refusal {
	return &Refusal{Status: api.StatusBadRequest, Field: field, Message: message}
}

// ParseFilter reads raw into a Filter for a reader who is a root when root is
// true. A value that does not read is refused under a 400 — a bad date, from
// after to, a period over MaxPeriodDays days, a bad ip or range, a value no
// view is recorded with — and an ip sent by a viewer under a 403.
func ParseFilter(sandbox *api.Sandbox, raw RawFilter, root bool) (Filter, *Refusal) {
	strings := sandbox.Deps.StringsDeps
	filter := Filter{}

	from, to := strings.TrimSpace(raw.From), strings.TrimSpace(raw.To)
	switch {
	case from == "" && to == "":
		to = Today(sandbox)
		from = AddDays(sandbox, to, 1-DefaultPeriodDays)
	case to == "":
		to = Today(sandbox)
	case from == "":
		from = AddDays(sandbox, to, 1-DefaultPeriodDays)
	}
	start, ok := ParseDay(sandbox, from)
	if !ok {
		return filter, refuse(KeyFrom, "from is a date spelled YYYY-MM-DD, such as 2026-10-13.")
	}
	if _, valid := wallDay(sandbox, to); !valid {
		return filter, refuse(KeyTo, "to is a date spelled YYYY-MM-DD, such as 2026-10-13.")
	}
	end, _ := ParseDay(sandbox, AddDays(sandbox, to, 1))
	filter.From, filter.To = from, to
	filter.Days = DaysBetween(sandbox, from, to) + 1
	if filter.Days < 1 {
		return filter, refuse(KeyFrom, "from is after to.")
	}
	if filter.Days > MaxPeriodDays {
		return filter, refuse(KeyFrom, sandbox.Deps.StdDeps.Sprintf("A period covers at most %d days.", MaxPeriodDays))
	}
	filter.AtMin, filter.AtMax = start, end-1

	filter.Path = strings.TrimSpace(raw.Path)
	if filter.Path != "" && (!strings.HasPrefix(filter.Path, "/") || len(filter.Path) > MaxPathLength) {
		return filter, refuse(KeyPath, "path is one page of the website, starting with / (/blog/).")
	}
	filter.Prefix = strings.TrimSpace(raw.Prefix)
	if filter.Prefix != "" && (!strings.HasPrefix(filter.Prefix, "/") || len(filter.Prefix) > MaxPathLength) {
		return filter, refuse(KeyPrefix, "prefix is the start of the pages to keep, starting with / (/blog/).")
	}

	if ip := strings.TrimSpace(raw.Ip); ip != "" {
		if !root {
			return filter, &Refusal{Status: api.StatusForbidden, Field: KeyIp, Message: "Only a root filters the views by ip."}
		}
		canonical, ok := backofficehttp.CanonicalIpOrRange(sandbox, ip)
		if !ok {
			return filter, refuse(KeyIp, "ip is one ip (203.0.113.7) or one CIDR range (203.0.113.0/24).")
		}
		filter.Ip, filter.IpRange = canonical, strings.Contains(canonical, "/")
	}

	if country := strings.ToUpper(strings.TrimSpace(raw.Country)); country != "" {
		if !isCountryCode(country) {
			return filter, refuse(KeyCountry, "country is a two-letter ISO code, such as BR, or XX for an unknown one.")
		}
		filter.Country = country
	}
	var refusal *Refusal
	if filter.Device, refusal = pick(sandbox, KeyDevice, raw.Device, Devices); refusal != nil {
		return filter, refusal
	}
	if filter.Browser, refusal = pick(sandbox, KeyBrowser, raw.Browser, Browsers); refusal != nil {
		return filter, refusal
	}
	if filter.Os, refusal = pick(sandbox, KeyOs, raw.Os, OperatingSystems); refusal != nil {
		return filter, refusal
	}

	switch strings.TrimSpace(raw.Status) {
	case "":
	case "200":
		filter.Status = StatusFound
	case "404":
		filter.Status = StatusNotFound
	default:
		return filter, refuse(KeyStatus, "status is 200 or 404.")
	}

	switch strings.ToLower(strings.TrimSpace(raw.Bots)) {
	case "", "0", "false":
	case "1", "true":
		filter.Bots = true
	default:
		return filter, refuse(KeyBots, "bots is 1 to include the bots.")
	}
	if filter.Device == DeviceBot {
		filter.Bots = true
	}
	return filter, nil
}

// pick is the one of values text names, compared without regard to case, ""
// for an empty text; a text naming none of them is refused under key.
func pick(sandbox *api.Sandbox, key string, text string, values []string) (string, *Refusal) {
	text = sandbox.Deps.StringsDeps.TrimSpace(text)
	if text == "" {
		return "", nil
	}
	for _, value := range values {
		if sandbox.Deps.StringsDeps.ToLower(value) == sandbox.Deps.StringsDeps.ToLower(text) {
			return value, nil
		}
	}
	return "", refuse(key, key+" is one of "+sandbox.Deps.StringsDeps.Join(values, ", ")+".")
}

// Selection is the views one filter keeps.
type Selection struct {
	Filter Filter
	// Views are the views every metric counts, newest first: the bots left
	// out unless Filter.Bots.
	Views []visualization_db.VisualizationViewRecord
	// Bots is how many views of bots the other filters keep, whether or not
	// Views holds them.
	Bots int
}

// Select reads the views of the period of filter and keeps the ones the rest
// of it matches.
func Select(sandbox *api.Sandbox, filter Filter) (Selection, error) {
	records, err := visualization_db.ListVisualizationViewsBetween(sandbox, visualization_db.New(sandbox), filter.AtMin, filter.AtMax)
	if err != nil {
		return Selection{}, err
	}
	selection := Selection{Filter: filter, Views: []visualization_db.VisualizationViewRecord{}}
	for _, view := range records {
		if !matches(sandbox, filter, view) {
			continue
		}
		bot := view.Device == DeviceBot
		if bot {
			selection.Bots++
		}
		if bot && !filter.Bots {
			continue
		}
		selection.Views = append(selection.Views, view)
	}
	views := selection.Views
	sandbox.Deps.SortDeps.Slice(views, func(i int, j int) bool {
		if views[i].At != views[j].At {
			return views[i].At > views[j].At
		}
		return views[i].Id > views[j].Id
	})
	return selection, nil
}

// matches tells whether view passes every filter but the period, which the
// read already kept, and the bots, which Select counts apart.
func matches(sandbox *api.Sandbox, filter Filter, view visualization_db.VisualizationViewRecord) bool {
	switch {
	case filter.Path != "" && view.Page != filter.Path:
		return false
	case filter.Prefix != "" && !sandbox.Deps.StringsDeps.HasPrefix(view.Page, filter.Prefix):
		return false
	case filter.Country != "" && view.Country != filter.Country:
		return false
	case filter.Device != "" && view.Device != filter.Device:
		return false
	case filter.Browser != "" && view.Browser != filter.Browser:
		return false
	case filter.Os != "" && view.Os != filter.Os:
		return false
	case filter.Status != 0 && view.Status != filter.Status:
		return false
	case filter.Ip != "" && !backofficehttp.IpMatches(sandbox, filter.Ip, view.Ip):
		return false
	}
	return true
}

// Param is one key and value of a query string.
type Param struct {
	Key   string
	Value string
}

// Params is filter as a query string spells it: the period always, then
// every other filter that is on.
func Params(sandbox *api.Sandbox, filter Filter) []Param {
	params := []Param{{KeyFrom, filter.From}, {KeyTo, filter.To}}
	add := func(key string, value string) {
		if value != "" {
			params = append(params, Param{key, value})
		}
	}
	add(KeyPath, filter.Path)
	add(KeyPrefix, filter.Prefix)
	add(KeyIp, filter.Ip)
	add(KeyCountry, filter.Country)
	add(KeyDevice, filter.Device)
	add(KeyBrowser, filter.Browser)
	add(KeyOs, filter.Os)
	if filter.Status != 0 {
		add(KeyStatus, sandbox.Deps.StringsDeps.FormatInt(filter.Status, 10))
	}
	if filter.Bots && filter.Device != DeviceBot {
		add(KeyBots, "1")
	}
	return params
}

// SetParam is params with key set to value — replaced where it is, added at
// the end otherwise — and taken off for an empty value.
func SetParam(params []Param, key string, value string) []Param {
	set := []Param{}
	done := false
	for _, param := range params {
		if param.Key != key {
			set = append(set, param)
			continue
		}
		if !done && value != "" {
			set = append(set, Param{key, value})
		}
		done = true
	}
	if !done && value != "" {
		set = append(set, Param{key, value})
	}
	return set
}

// Href is base with params as its query string, each value escaped.
func Href(sandbox *api.Sandbox, base string, params []Param) string {
	if len(params) == 0 {
		return base
	}
	parts := []string{}
	for _, param := range params {
		parts = append(parts, QueryEscape(sandbox, param.Key)+"="+QueryEscape(sandbox, param.Value))
	}
	return base + "?" + sandbox.Deps.StringsDeps.Join(parts, "&")
}

// QueryEscape is text escaped for a query string: every byte but a letter, a
// digit and - _ . ~ / spelled %XX.
func QueryEscape(sandbox *api.Sandbox, text string) string {
	const hex = "0123456789ABCDEF"
	escaped := make([]byte, 0, len(text))
	for index := 0; index < len(text); index++ {
		char := text[index]
		switch {
		case char >= 'a' && char <= 'z', char >= 'A' && char <= 'Z', char >= '0' && char <= '9',
			char == '-', char == '_', char == '.', char == '~', char == '/':
			escaped = append(escaped, char)
		default:
			escaped = append(escaped, '%', hex[char>>4], hex[char&0x0f])
		}
	}
	return string(escaped)
}

// StatusParam is the status filter a JSON body sends, spelled the way a query
// string spells it: "" for 0, the filter left off.
func StatusParam(sandbox *api.Sandbox, status int) string {
	if status == 0 {
		return ""
	}
	return sandbox.Deps.StringsDeps.FormatInt(int64(status), 10)
}

// BotsParam is the bots filter a JSON body sends, spelled the way a query
// string spells it.
func BotsParam(sandbox *api.Sandbox, bots bool) string {
	if bots {
		return "1"
	}
	return ""
}
