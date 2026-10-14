package agent

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// What the backoffice reads of the agent: the summaries, searched by period
// and text, and the usage in tokens — today against the limits, by day, by ip
// and the ips at their limit. A viewer reads every ip masked.

// Period is a range of whole days of the server, both included.
type Period struct {
	// From and To are the first and the last day, YYYY-MM-DD.
	From, To string
	// AtMin and AtMax are the first and the last second of the period.
	AtMin, AtMax int64
	// Days is how many days it covers.
	Days int
}

// ParsePeriod reads from and to — YYYY-MM-DD, both included — as a period: the
// last DefaultPeriodDays days when both are blank, the other derived when one
// is; refused when a date is not one, from is after to, or the period covers
// more than MaxPeriodDays days.
func ParsePeriod(sandbox *api.Sandbox, from string, to string) (Period, *Refusal) {
	strings := sandbox.Deps.StringsDeps
	from, to = strings.TrimSpace(from), strings.TrimSpace(to)
	switch {
	case from == "" && to == "":
		to = visualization.Today(sandbox)
		from = visualization.AddDays(sandbox, to, 1-DefaultPeriodDays)
	case to == "":
		to = visualization.Today(sandbox)
	case from == "":
		from = visualization.AddDays(sandbox, to, 1-DefaultPeriodDays)
	}
	period := Period{From: from, To: to}
	start, ok := visualization.ParseDay(sandbox, from)
	if !ok {
		return period, refuse("from", "from is a date spelled YYYY-MM-DD, such as 2026-10-14.")
	}
	if _, ok := visualization.ParseDay(sandbox, to); !ok {
		return period, refuse("to", "to is a date spelled YYYY-MM-DD, such as 2026-10-14.")
	}
	end, _ := visualization.ParseDay(sandbox, visualization.AddDays(sandbox, to, 1))
	period.Days = visualization.DaysBetween(sandbox, from, to) + 1
	if period.Days < 1 {
		return period, refuse("from", "from is after to.")
	}
	if period.Days > MaxPeriodDays {
		return period, refuse("from", sandbox.Deps.StdDeps.Sprintf("A period covers at most %d days.", MaxPeriodDays))
	}
	period.AtMin, period.AtMax = start, end-1
	return period, nil
}

// ListSummaries is one page of the summaries of period holding search — in
// their text or their id, any case — newest first.
func ListSummaries(sandbox *api.Sandbox, period Period, search string, page int, limit int) ([]agent_db.AgentSummaryRecord, visualization.Paging, error) {
	strings := sandbox.Deps.StringsDeps
	all, err := agent_db.ListAgentSummariesBetween(sandbox, agent_db.New(sandbox), period.AtMin, period.AtMax)
	if err != nil {
		return nil, visualization.Paging{}, err
	}
	search = strings.ToLower(strings.TrimSpace(search))
	kept := []agent_db.AgentSummaryRecord{}
	for _, summary := range all {
		if search == "" || strings.Contains(strings.ToLower(summary.Summary), search) || strings.Contains(summary.Ref, search) {
			kept = append(kept, summary)
		}
	}
	sandbox.Deps.SortDeps.SliceStable(kept, func(i int, j int) bool {
		if kept[i].At != kept[j].At {
			return kept[i].At > kept[j].At
		}
		return kept[i].Id > kept[j].Id
	})
	paging, from, to := visualization.Paginate(sandbox, len(kept), page, limit)
	return kept[from:to], paging, nil
}

// LatestSummaries is the count newest summaries of the last DefaultPeriodDays
// days.
func LatestSummaries(sandbox *api.Sandbox, count int) ([]agent_db.AgentSummaryRecord, error) {
	period, _ := ParsePeriod(sandbox, "", "")
	summaries, _, err := ListSummaries(sandbox, period, "", 1, count)
	return summaries, err
}

// FindSummary is the summary whose id is ref, false when there is none.
func FindSummary(sandbox *api.Sandbox, ref string) (agent_db.AgentSummaryRecord, bool) {
	ref = sandbox.Deps.StringsDeps.ToLower(sandbox.Deps.StringsDeps.TrimSpace(ref))
	if ref == "" {
		return agent_db.AgentSummaryRecord{}, false
	}
	return agent_db.New(sandbox).FindAgentSummaryByRef(ref)
}

// RemoveSummaries removes the summary whose id is ref, or every summary
// before the day before — exactly one of the two — and answers how many it
// removed: 404 for an id no summary holds.
func RemoveSummaries(sandbox *api.Sandbox, ref string, before string) (int, *Refusal, error) {
	strings := sandbox.Deps.StringsDeps
	ref, before = strings.TrimSpace(ref), strings.TrimSpace(before)
	if (ref == "") == (before == "") {
		return 0, refuse("id", "Send the id of one summary, or a date to remove every summary before it."), nil
	}
	if ref != "" {
		summary, ok := FindSummary(sandbox, ref)
		if !ok {
			return 0, &Refusal{Status: api.StatusNotFound, Field: "id", Message: "No summary has this id."}, nil
		}
		return 1, nil, agent_db.New(sandbox).RemoveAgentSummary(summary.Id)
	}
	cutoff, ok := visualization.ParseDay(sandbox, before)
	if !ok {
		return 0, refuse("before", "before is a date spelled YYYY-MM-DD, such as 2026-10-14."), nil
	}
	removed, err := agent_db.RemoveAgentSummariesBefore(sandbox, agent_db.New(sandbox), cutoff)
	return removed, nil, err
}

// UsageDay is the usage of one day.
type UsageDay struct {
	Day          string
	InputTokens  int64
	OutputTokens int64
	Total        int64
	// Messages and Summaries are how many provider calls each kind made.
	Messages  int
	Summaries int
}

// UsageIp is the usage of one ip — an IPv6 one by its /64 — masked to a
// viewer. AtLimit is true when it reached the ip daily limit today.
type UsageIp struct {
	Ip           string
	InputTokens  int64
	OutputTokens int64
	Total        int64
	Calls        int
	AtLimit      bool
}

// UsageToday is the tokens of today against the limits.
type UsageToday struct {
	Day          string
	Total        int64
	DailyLimit   int64
	IpDailyLimit int64
	// Percent is Total of DailyLimit, 0 with no limit.
	Percent float64
}

// UsageReport is the usage of a period.
type UsageReport struct {
	Period Period
	Today  UsageToday
	// Days is every day of the period, newest first, the empty ones included.
	Days []UsageDay
	// Ips are the ips of the period, the most tokens first, at most
	// MaxIpRows; summaries hold no ip and are not among them.
	Ips []UsageIp
	// AtLimit are the ips of today at the ip daily limit.
	AtLimit []UsageIp
	// Total is every token of the period.
	Total int64
}

// Usage is the usage of period, its ips masked unless root.
func Usage(sandbox *api.Sandbox, period Period, root bool) (UsageReport, error) {
	settings, err := ReadSettings(sandbox)
	if err != nil {
		return UsageReport{}, err
	}
	rows, err := agent_db.ListAgentUsagesBetween(sandbox, agent_db.New(sandbox), period.AtMin, period.AtMax)
	if err != nil {
		return UsageReport{}, err
	}
	report := UsageReport{Period: period}
	days := map[string]*UsageDay{}
	ips := map[string]*UsageIp{}
	for _, row := range rows {
		used := row.InputTokens + row.OutputTokens
		report.Total += used
		dayName := visualization.DayOf(sandbox, row.At)
		day, ok := days[dayName]
		if !ok {
			day = &UsageDay{Day: dayName}
			days[dayName] = day
		}
		day.InputTokens += row.InputTokens
		day.OutputTokens += row.OutputTokens
		day.Total += used
		if row.Kind == KindSummary {
			day.Summaries++
		} else {
			day.Messages++
		}
		if row.Ip == "" {
			continue
		}
		key := IpKey(sandbox, row.Ip)
		ip, ok := ips[key]
		if !ok {
			ip = &UsageIp{Ip: key}
			ips[key] = ip
		}
		ip.InputTokens += row.InputTokens
		ip.OutputTokens += row.OutputTokens
		ip.Total += used
		ip.Calls++
	}
	for day := period.To; ; day = visualization.AddDays(sandbox, day, -1) {
		found, ok := days[day]
		if !ok {
			found = &UsageDay{Day: day}
		}
		report.Days = append(report.Days, *found)
		if day == period.From || len(report.Days) >= MaxPeriodDays {
			break
		}
	}

	tally := todayTally(sandbox)
	report.Today = UsageToday{Day: tally.Day, Total: tally.Total, DailyLimit: settings.DailyLimit, IpDailyLimit: settings.IpDailyLimit}
	if settings.DailyLimit > 0 {
		report.Today.Percent = float64(tally.Total) * 100 / float64(settings.DailyLimit)
	}
	atLimit := map[string]bool{}
	if settings.IpDailyLimit > 0 {
		for key, used := range tally.ByIp {
			if used >= settings.IpDailyLimit {
				atLimit[key] = true
				report.AtLimit = append(report.AtLimit, UsageIp{Ip: ShownKey(sandbox, key, root), Total: used, AtLimit: true})
			}
		}
		sandbox.Deps.SortDeps.Slice(report.AtLimit, func(i int, j int) bool { return report.AtLimit[i].Total > report.AtLimit[j].Total })
	}

	for key, ip := range ips {
		ip.AtLimit = atLimit[key]
		report.Ips = append(report.Ips, *ip)
	}
	sandbox.Deps.SortDeps.Slice(report.Ips, func(i int, j int) bool {
		if report.Ips[i].Total != report.Ips[j].Total {
			return report.Ips[i].Total > report.Ips[j].Total
		}
		return report.Ips[i].Ip < report.Ips[j].Ip
	})
	if len(report.Ips) > MaxIpRows {
		report.Ips = report.Ips[:MaxIpRows]
	}
	for index := range report.Ips {
		report.Ips[index].Ip = ShownKey(sandbox, report.Ips[index].Ip, root)
	}
	return report, nil
}

// todayTally is the tally the limits hold to when the agent runs in this
// process — it counts the writes still queued — and the one agent-usage holds
// otherwise.
func todayTally(sandbox *api.Sandbox) Tally {
	if Running(sandbox) {
		return TodayTally(sandbox)
	}
	day := visualization.Today(sandbox)
	tally := Tally{Day: day, ByIp: map[string]int64{}}
	start, _ := visualization.ParseDay(sandbox, day)
	rows, err := agent_db.ListAgentUsagesBetween(sandbox, agent_db.New(sandbox), start, NowSeconds(sandbox)+60)
	if err != nil {
		return tally
	}
	for _, row := range rows {
		used := row.InputTokens + row.OutputTokens
		tally.Total += used
		if row.Ip != "" {
			tally.ByIp[IpKey(sandbox, row.Ip)] += used
		}
	}
	return tally
}
