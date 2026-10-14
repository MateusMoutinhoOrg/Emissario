package contact

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// The keys of the filters, in a query string and in a JSON body alike.
const (
	KeyFrom     = "from"
	KeyTo       = "to"
	KeyCountry  = "country"
	KeyContains = "contains"
	KeyStatus   = "status"
)

// RawFilter is the filters as a query string or a JSON body sends them, every
// one text.
type RawFilter struct {
	From     string
	To       string
	Country  string
	Contains string
	Status   string
}

// Filter is the filters read: the contact requests a listing keeps.
type Filter struct {
	// From and To are the first and the last day, YYYY-MM-DD, "" for no
	// bound.
	From, To string
	// AtMin and AtMax are the first and the last second of the period, 0
	// for no bound.
	AtMin, AtMax int64
	// Country is an ISO 3166-1 alpha-2 code, XX for an unknown one, "" for
	// every country.
	Country string
	// Contains is a text the JSON holds, any case, "" for every one.
	Contains string
	// Status is StatusInbox, StatusArchived or StatusAll.
	Status string
}

// ParseFilter reads raw as a Filter: from and to — YYYY-MM-DD, both included —
// each optional, a country of two letters, a text of at most
// MaxContainsLength characters and a status of Statuses, the inbox when
// blank. A value that does not read is refused, as is from after to.
func ParseFilter(sandbox *api.Sandbox, raw RawFilter) (Filter, *Refusal) {
	strings := sandbox.Deps.StringsDeps
	filter := Filter{
		From:     strings.TrimSpace(raw.From),
		To:       strings.TrimSpace(raw.To),
		Country:  strings.ToUpper(strings.TrimSpace(raw.Country)),
		Contains: strings.TrimSpace(raw.Contains),
		Status:   strings.ToLower(strings.TrimSpace(raw.Status)),
	}
	if filter.From != "" {
		start, ok := visualization.ParseDay(sandbox, filter.From)
		if !ok {
			return filter, refuse(KeyFrom, "from is a date spelled YYYY-MM-DD, such as 2026-10-14.")
		}
		filter.AtMin = start
	}
	if filter.To != "" {
		if _, ok := visualization.ParseDay(sandbox, filter.To); !ok {
			return filter, refuse(KeyTo, "to is a date spelled YYYY-MM-DD, such as 2026-10-14.")
		}
		end, _ := visualization.ParseDay(sandbox, visualization.AddDays(sandbox, filter.To, 1))
		filter.AtMax = end - 1
	}
	if filter.From != "" && filter.To != "" && visualization.DaysBetween(sandbox, filter.From, filter.To) < 0 {
		return filter, refuse(KeyFrom, "from is after to.")
	}
	if filter.Country != "" && !visualization.IsCountryCode(sandbox, filter.Country) {
		return filter, refuse(KeyCountry, "country is a two-letter ISO code, such as BR, or XX for an unknown one.")
	}
	if len([]rune(filter.Contains)) > MaxContainsLength {
		return filter, refuse(KeyContains, sandbox.Deps.StdDeps.Sprintf("contains is a text of at most %d characters.", MaxContainsLength))
	}
	switch filter.Status {
	case "":
		filter.Status = StatusInbox
	case StatusInbox, StatusArchived, StatusAll:
	default:
		return filter, refuse(KeyStatus, "status is inbox, archived or all.")
	}
	return filter, nil
}

// Params is filter as a query string spells it: every filter that is on, the
// status when it is not the inbox.
func Params(sandbox *api.Sandbox, filter Filter) []visualization.Param {
	params := []visualization.Param{}
	add := func(key string, value string) {
		if value != "" {
			params = append(params, visualization.Param{Key: key, Value: value})
		}
	}
	if filter.Status != StatusInbox {
		add(KeyStatus, filter.Status)
	}
	add(KeyFrom, filter.From)
	add(KeyTo, filter.To)
	add(KeyCountry, filter.Country)
	add(KeyContains, filter.Contains)
	return params
}

// keeps tells whether filter keeps request, the period aside: its status, its
// country and the text its JSON holds — needle, already lower case.
func keeps(sandbox *api.Sandbox, filter Filter, needle string, archived int64, country string, body string) bool {
	switch filter.Status {
	case StatusInbox:
		if archived != 0 {
			return false
		}
	case StatusArchived:
		if archived == 0 {
			return false
		}
	}
	if filter.Country != "" && country != filter.Country {
		return false
	}
	return needle == "" || Matches(sandbox, body, needle)
}
