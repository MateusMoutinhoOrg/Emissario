package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// The pages of the agent every user reads — the overview, the summaries, one
// summary and the usage — share the tabs of agentTabs: a root also sees
// Settings, URLs and Source, the pages of agent_settings.go.

// agentTabs are the tabs of every agent page, active the one at path.
func agentTabs(sandbox *api.Sandbox, active string, root bool) []VisualizationLink {
	tabs := []VisualizationLink{}
	for _, tab := range []struct {
		label string
		path  string
		root  bool
	}{
		{"Overview", agent.DashboardPath, false},
		{"Summaries", agent.SummariesPath, false},
		{"Usage", agent.UsagePath, false},
		{"Settings", agent.SettingsPath, true},
		{"URLs", agent.UrlsPath, true},
		{"Source", agent.SourcePath, true},
	} {
		if tab.root && !root {
			continue
		}
		tabs = append(tabs, VisualizationLink{Label: tab.label, Href: tab.path, Active: tab.path == active})
	}
	return tabs
}

// agentNotice is the sentence a notice code of an agent page reads as.
func agentNotice(sandbox *api.Sandbox, code string, removed int) Notice {
	switch code {
	case agent.NoticeSaved:
		return Notice{Text: "Saved.", Kind: "ok"}
	case agent.NoticeAdded:
		return Notice{Text: "Url added: the agent can call it from the next message.", Kind: "ok"}
	case agent.NoticeRemoved:
		if removed >= 0 {
			return Notice{Text: plural(sandbox, removed, "record", "records") + " removed.", Kind: "ok"}
		}
		return Notice{Text: "Removed.", Kind: "ok"}
	}
	return Notice{}
}

// AgentSummaryLine is one summary of a list.
type AgentSummaryLine struct {
	Id       string
	Href     string
	At       string
	Preview  string
	Messages string
	Tokens   string
}

// agentSummaryLine is summary as a list shows it.
func agentSummaryLine(sandbox *api.Sandbox, summary agent_db.AgentSummaryRecord) AgentSummaryLine {
	preview := summary.Summary
	if len([]rune(preview)) > 220 {
		preview = string([]rune(preview)[:220]) + "…"
	}
	return AgentSummaryLine{
		Id:       summary.Ref,
		Href:     agent.SummaryPath + "?id=" + summary.Ref,
		At:       formatInstant(sandbox, summary.At),
		Preview:  preview,
		Messages: formatCount(sandbox, int(summary.Messages)),
		Tokens:   formatCount(sandbox, int(summary.InputTokens+summary.OutputTokens)),
	}
}

// AgentDashboardPage is what backoffice/agent_dashboard.html is rendered with.
type AgentDashboardPage struct {
	Viewer Viewer
	Tabs   []VisualizationLink
	// Answering is true when the chat answers, Problem why not.
	Answering bool
	Problem   string
	// Model and BaseUrl are shown to a root alone.
	Model   string
	BaseUrl string
	Readme  bool
	Summary bool
	Running bool
	Dropped string
	// The tokens of today, against the limit.
	TodayTokens  string
	DailyLimit   string
	Percent      string
	HasLimit     bool
	IpDailyLimit string
	AtLimit      string
	// Source is what the last sync of the agent source did, for a root.
	SourceSaved   bool
	SourceSynced  string
	SourceOk      bool
	SourceCommit  string
	SourceSyncing bool
	Latest        []AgentSummaryLine
	Zone          string
}

// RenderAgentDashboardPage answers the overview of the agent: whether the chat
// answers and why not, the tokens of today, the latest summaries.
func RenderAgentDashboardPage(sandbox *api.Sandbox, response *serverdeps.Response, user *backoffice_db.BackofficeUserRecord) error {
	viewer := viewerOf(sandbox, user)
	status, settings, err := agent.CurrentStatus(sandbox)
	if err != nil {
		return err
	}
	period, _ := agent.ParsePeriod(sandbox, visualization.Today(sandbox), visualization.Today(sandbox))
	report, err := agent.Usage(sandbox, period, viewer.IsRoot)
	if err != nil {
		return err
	}
	latest, err := agent.LatestSummaries(sandbox, 5)
	if err != nil {
		return err
	}
	page := AgentDashboardPage{
		Viewer:        viewer,
		Tabs:          agentTabs(sandbox, agent.DashboardPath, viewer.IsRoot),
		Answering:     status.Answering,
		Problem:       status.Problem,
		Readme:        status.Readme,
		Summary:       status.Summary,
		Running:       status.Running,
		Dropped:       formatCount(sandbox, int(status.Dropped)),
		TodayTokens:   formatCount(sandbox, int(report.Today.Total)),
		HasLimit:      settings.DailyLimit > 0,
		DailyLimit:    limitText(sandbox, settings.DailyLimit),
		IpDailyLimit:  limitText(sandbox, settings.IpDailyLimit),
		Percent:       formatDecimal(sandbox, min(report.Today.Percent, 100), 0),
		AtLimit:       formatCount(sandbox, len(report.AtLimit)),
		SourceSaved:   status.SourceSaved,
		SourceOk:      status.Source.Ok,
		SourceCommit:  status.Source.Commit,
		SourceSyncing: status.Source.Pending || status.Source.Syncing,
		Zone:          visualization.ZoneName(sandbox),
	}
	if viewer.IsRoot {
		page.Model = settings.Model
		page.BaseUrl = settings.BaseUrl
	}
	if status.Source.AttemptedAt > 0 {
		page.SourceSynced = formatInstant(sandbox, status.Source.AttemptedAt)
	}
	for _, summary := range latest {
		page.Latest = append(page.Latest, agentSummaryLine(sandbox, summary))
	}
	return RenderHTML(sandbox, response, api.StatusOK, "backoffice/agent_dashboard.html", page)
}

// limitText is a limit as a page shows it: "No limit" for 0.
func limitText(sandbox *api.Sandbox, limit int64) string {
	if limit <= 0 {
		return "No limit"
	}
	return formatCount(sandbox, int(limit))
}

// AgentPeriodBar is the period filter of the summaries and the usage.
type AgentPeriodBar struct {
	Action string
	From   string
	To     string
	// Search is the text searched, on the summaries alone.
	Search     string
	ShowSearch bool
	Quick      []VisualizationLink
	Zone       string
}

// agentPeriodBar is the period bar of the page at action.
func agentPeriodBar(sandbox *api.Sandbox, action string, from string, to string, search string, showSearch bool) AgentPeriodBar {
	bar := AgentPeriodBar{Action: action, From: from, To: to, Search: search, ShowSearch: showSearch, Zone: visualization.ZoneName(sandbox)}
	today := visualization.Today(sandbox)
	for _, quick := range []struct {
		label string
		days  int
	}{{"Today", 1}, {"7 days", 7}, {"30 days", 30}, {"90 days", 90}} {
		start := visualization.AddDays(sandbox, today, 1-quick.days)
		params := []visualization.Param{{Key: "from", Value: start}, {Key: "to", Value: today}}
		if search != "" {
			params = append(params, visualization.Param{Key: "q", Value: search})
		}
		bar.Quick = append(bar.Quick, VisualizationLink{
			Label:  quick.label,
			Href:   visualization.Href(sandbox, action, params),
			Active: from == start && to == today,
		})
	}
	return bar
}

// AgentSummariesPage is what backoffice/agent_summaries.html is rendered with.
type AgentSummariesPage struct {
	Viewer Viewer
	Tabs   []VisualizationLink
	Bar    AgentPeriodBar
	Notice Notice
	// Error is why the period was refused.
	Error string
	Rows  []AgentSummaryLine
	Pager VisualizationPager
	// RemoveAction is where a root removes the summaries before a date.
	RemoveAction string
	Today        string
}

// RenderAgentSummariesPage answers, under status, one page of the summaries of
// the period holding search, newest first; refused is why the period did not
// read, "" when it did.
func RenderAgentSummariesPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, from string, to string, search string, refused string, summaries []agent_db.AgentSummaryRecord, paging visualization.Paging, notice string, removed int) error {
	viewer := viewerOf(sandbox, user)
	page := AgentSummariesPage{
		Viewer: viewer,
		Tabs:   agentTabs(sandbox, agent.SummariesPath, viewer.IsRoot),
		Bar:    agentPeriodBar(sandbox, agent.SummariesPath, from, to, search, true),
		Notice: agentNotice(sandbox, notice, removed),
		Error:  refused,
		Today:  visualization.Today(sandbox),
	}
	if viewer.IsRoot {
		page.RemoveAction = agent.RemoveSummariesPath
	}
	for _, summary := range summaries {
		page.Rows = append(page.Rows, agentSummaryLine(sandbox, summary))
	}
	params := []visualization.Param{{Key: "from", Value: from}, {Key: "to", Value: to}}
	if search != "" {
		params = append(params, visualization.Param{Key: "q", Value: search})
	}
	page.Pager = visualizationPager(sandbox, agent.SummariesPath, params, paging, len(summaries))
	return RenderHTML(sandbox, response, status, "backoffice/agent_summaries.html", page)
}

// AgentSummaryPage is what backoffice/agent_summary.html is rendered with.
type AgentSummaryPage struct {
	Viewer       Viewer
	Tabs         []VisualizationLink
	Id           string
	At           string
	Summary      string
	Messages     string
	InputTokens  string
	OutputTokens string
	// RemoveAction is where a root removes it.
	RemoveAction string
}

// RenderAgentSummaryPage answers one summary whole.
func RenderAgentSummaryPage(sandbox *api.Sandbox, response *serverdeps.Response, user *backoffice_db.BackofficeUserRecord, summary agent_db.AgentSummaryRecord) error {
	viewer := viewerOf(sandbox, user)
	page := AgentSummaryPage{
		Viewer:       viewer,
		Tabs:         agentTabs(sandbox, agent.SummariesPath, viewer.IsRoot),
		Id:           summary.Ref,
		At:           formatInstant(sandbox, summary.At),
		Summary:      summary.Summary,
		Messages:     formatCount(sandbox, int(summary.Messages)),
		InputTokens:  formatCount(sandbox, int(summary.InputTokens)),
		OutputTokens: formatCount(sandbox, int(summary.OutputTokens)),
	}
	if viewer.IsRoot {
		page.RemoveAction = agent.RemoveSummariesPath
	}
	return RenderHTML(sandbox, response, api.StatusOK, "backoffice/agent_summary.html", page)
}

// AgentUsageDayLine is one day of the usage.
type AgentUsageDayLine struct {
	Day       string
	Input     string
	Output    string
	Total     string
	Messages  string
	Summaries string
	// Share is the width of its bar, 0 to 100.
	Share string
}

// AgentUsageIpLine is one ip of the usage.
type AgentUsageIpLine struct {
	Ip      string
	Calls   string
	Input   string
	Output  string
	Total   string
	AtLimit bool
}

// AgentUsagePage is what backoffice/agent_usage.html is rendered with.
type AgentUsagePage struct {
	Viewer Viewer
	Tabs   []VisualizationLink
	Bar    AgentPeriodBar
	Error  string
	// The tokens of today, against the limits.
	TodayTokens  string
	DailyLimit   string
	IpDailyLimit string
	Percent      string
	HasLimit     bool
	PeriodTotal  string
	Days         []AgentUsageDayLine
	Ips          []AgentUsageIpLine
	AtLimit      []AgentUsageIpLine
}

// RenderAgentUsagePage answers, under status, the usage of the period;
// refused is why the period did not read, "" when it did.
func RenderAgentUsagePage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, from string, to string, refused string, report agent.UsageReport) error {
	viewer := viewerOf(sandbox, user)
	page := AgentUsagePage{
		Viewer: viewer,
		Tabs:   agentTabs(sandbox, agent.UsagePath, viewer.IsRoot),
		Bar:    agentPeriodBar(sandbox, agent.UsagePath, from, to, "", false),
		Error:  refused,
	}
	if refused != "" {
		return RenderHTML(sandbox, response, status, "backoffice/agent_usage.html", page)
	}
	page.TodayTokens = formatCount(sandbox, int(report.Today.Total))
	page.DailyLimit = limitText(sandbox, report.Today.DailyLimit)
	page.IpDailyLimit = limitText(sandbox, report.Today.IpDailyLimit)
	page.HasLimit = report.Today.DailyLimit > 0
	page.Percent = formatDecimal(sandbox, min(report.Today.Percent, 100), 0)
	page.PeriodTotal = formatCount(sandbox, int(report.Total))
	most := int64(1)
	for _, day := range report.Days {
		most = max(most, day.Total)
	}
	for _, day := range report.Days {
		page.Days = append(page.Days, AgentUsageDayLine{
			Day:       day.Day,
			Input:     formatCount(sandbox, int(day.InputTokens)),
			Output:    formatCount(sandbox, int(day.OutputTokens)),
			Total:     formatCount(sandbox, int(day.Total)),
			Messages:  formatCount(sandbox, day.Messages),
			Summaries: formatCount(sandbox, day.Summaries),
			Share:     formatDecimal(sandbox, float64(day.Total)*100/float64(most), 1),
		})
	}
	for _, ip := range report.Ips {
		page.Ips = append(page.Ips, agentUsageIpLine(sandbox, ip))
	}
	for _, ip := range report.AtLimit {
		page.AtLimit = append(page.AtLimit, agentUsageIpLine(sandbox, ip))
	}
	return RenderHTML(sandbox, response, status, "backoffice/agent_usage.html", page)
}

// agentUsageIpLine is ip as the usage shows it.
func agentUsageIpLine(sandbox *api.Sandbox, ip agent.UsageIp) AgentUsageIpLine {
	return AgentUsageIpLine{
		Ip:      ip.Ip,
		Calls:   formatCount(sandbox, ip.Calls),
		Input:   formatCount(sandbox, int(ip.InputTokens)),
		Output:  formatCount(sandbox, int(ip.OutputTokens)),
		Total:   formatCount(sandbox, int(ip.Total)),
		AtLimit: ip.AtLimit,
	}
}
