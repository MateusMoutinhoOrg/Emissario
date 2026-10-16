package backofficeapi

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	serializabledeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serializabledeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/websitesource"
)

// The documents of the agent api, the JSON twins of its pages. The api key
// and the values of the headers of a url are never in any of them.

// AgentSummaryJSON is one summary: {id, at, summary, messages, input_tokens,
// output_tokens} — preview in place of summary in a list.
func AgentSummaryJSON(sandbox *api.Sandbox, summary agent_db.AgentSummaryRecord, whole bool) *serializabledeps.SerializableObject {
	item := sandbox.Deps.SerializableDeps.CreateObject()
	item.AddItemToObject("id", summary.Ref)
	item.AddItemToObject("at", summary.At)
	if whole {
		item.AddItemToObject("summary", summary.Summary)
	} else {
		preview := summary.Summary
		if len([]rune(preview)) > 300 {
			preview = string([]rune(preview)[:300]) + "…"
		}
		item.AddItemToObject("preview", preview)
	}
	item.AddItemToObject("messages", summary.Messages)
	item.AddItemToObject("input_tokens", summary.InputTokens)
	item.AddItemToObject("output_tokens", summary.OutputTokens)
	return item
}

// AgentSummaryResponseJSON is {summary}.
func AgentSummaryResponseJSON(sandbox *api.Sandbox, summary agent_db.AgentSummaryRecord) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("summary", AgentSummaryJSON(sandbox, summary, true))
	return document
}

// AgentSummariesJSON is {summaries, from, to, total, page, pages, limit}.
func AgentSummariesJSON(sandbox *api.Sandbox, period agent.Period, summaries []agent_db.AgentSummaryRecord, paging visualization.Paging) *serializabledeps.SerializableObject {
	list := sandbox.Deps.SerializableDeps.CreateArray()
	for _, summary := range summaries {
		list.AddItemToArray(AgentSummaryJSON(sandbox, summary, false))
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("summaries", list)
	document.AddItemToObject("from", period.From)
	document.AddItemToObject("to", period.To)
	pagingJSON(sandbox, document, paging)
	return document
}

// AgentUsageJSON is {from, to, total, today, days, ips, at_limit}.
func AgentUsageJSON(sandbox *api.Sandbox, report agent.UsageReport) *serializabledeps.SerializableObject {
	json := sandbox.Deps.SerializableDeps
	today := json.CreateObject()
	today.AddItemToObject("day", report.Today.Day)
	today.AddItemToObject("total", report.Today.Total)
	today.AddItemToObject("daily_limit", report.Today.DailyLimit)
	today.AddItemToObject("ip_daily_limit", report.Today.IpDailyLimit)
	today.AddItemToObject("percent", report.Today.Percent)
	days := json.CreateArray()
	for _, day := range report.Days {
		item := json.CreateObject()
		item.AddItemToObject("day", day.Day)
		item.AddItemToObject("input_tokens", day.InputTokens)
		item.AddItemToObject("output_tokens", day.OutputTokens)
		item.AddItemToObject("total", day.Total)
		item.AddItemToObject("messages", day.Messages)
		item.AddItemToObject("summaries", day.Summaries)
		days.AddItemToArray(item)
	}
	ips := json.CreateArray()
	for _, ip := range report.Ips {
		item := json.CreateObject()
		item.AddItemToObject("ip", ip.Ip)
		item.AddItemToObject("calls", ip.Calls)
		item.AddItemToObject("input_tokens", ip.InputTokens)
		item.AddItemToObject("output_tokens", ip.OutputTokens)
		item.AddItemToObject("total", ip.Total)
		item.AddItemToObject("at_limit", ip.AtLimit)
		ips.AddItemToArray(item)
	}
	atLimit := json.CreateArray()
	for _, ip := range report.AtLimit {
		item := json.CreateObject()
		item.AddItemToObject("ip", ip.Ip)
		item.AddItemToObject("total", ip.Total)
		atLimit.AddItemToArray(item)
	}
	document := json.CreateObject()
	document.AddItemToObject("from", report.Period.From)
	document.AddItemToObject("to", report.Period.To)
	document.AddItemToObject("total", report.Total)
	document.AddItemToObject("today", today)
	document.AddItemToObject("days", days)
	document.AddItemToObject("ips", ips)
	document.AddItemToObject("at_limit", atLimit)
	return document
}

// AgentSettingsJSON is the settings, the api key as its hint alone.
func AgentSettingsJSON(sandbox *api.Sandbox, settings agent.Settings) *serializabledeps.SerializableObject {
	key, fromEnv := agent.ApiKey(sandbox, settings)
	item := sandbox.Deps.SerializableDeps.CreateObject()
	item.AddItemToObject("enabled", settings.Enabled)
	item.AddItemToObject("base_url", settings.BaseUrl)
	item.AddItemToObject("api_key_hint", agent.KeyHint(sandbox, key))
	item.AddItemToObject("api_key_from_env", fromEnv)
	item.AddItemToObject("model", settings.Model)
	item.AddItemToObject("max_tokens", settings.MaxTokens)
	item.AddItemToObject("daily_limit", settings.DailyLimit)
	item.AddItemToObject("ip_daily_limit", settings.IpDailyLimit)
	item.AddItemToObject("rate_per_minute", settings.RatePerMinute)
	item.AddItemToObject("summaries_enabled", settings.SummariesEnabled)
	item.AddItemToObject("summary_retention_days", RetentionJSON(sandbox, settings.SummaryRetentionDays))
	return item
}

// AgentSettingsResponseJSON is {settings}.
func AgentSettingsResponseJSON(sandbox *api.Sandbox, settings agent.Settings) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("settings", AgentSettingsJSON(sandbox, settings))
	return document
}

// AgentSettingsStateJSON is {settings, status}.
func AgentSettingsStateJSON(sandbox *api.Sandbox, settings agent.Settings, state agent.Status) *serializabledeps.SerializableObject {
	status := sandbox.Deps.SerializableDeps.CreateObject()
	status.AddItemToObject("answering", state.Answering)
	status.AddItemToObject("problem", state.Problem)
	status.AddItemToObject("readme", state.Readme)
	status.AddItemToObject("summary", state.Summary)
	status.AddItemToObject("running", state.Running)
	status.AddItemToObject("dropped", state.Dropped)
	host := agent.ProviderHost(sandbox, settings)
	status.AddItemToObject("provider_host", host)
	status.AddItemToObject("third_party", host != "" && !agent.ProviderIsLocal(sandbox, host))
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("settings", AgentSettingsJSON(sandbox, settings))
	document.AddItemToObject("status", status)
	return document
}

// AgentUrlJSON is one url: {id, name, description, method, url,
// header_names} — the values of its headers never.
func AgentUrlJSON(sandbox *api.Sandbox, url agent_db.AgentUrlRecord) *serializabledeps.SerializableObject {
	names := sandbox.Deps.SerializableDeps.CreateArray()
	for _, name := range agent.HeaderNames(sandbox, url.Headers) {
		names.AddItemToArray(name)
	}
	item := sandbox.Deps.SerializableDeps.CreateObject()
	item.AddItemToObject("id", url.Id)
	item.AddItemToObject("name", url.Name)
	item.AddItemToObject("description", url.Description)
	item.AddItemToObject("method", url.Method)
	item.AddItemToObject("url", url.Url)
	item.AddItemToObject("header_names", names)
	return item
}

// AgentUrlResponseJSON is {url}.
func AgentUrlResponseJSON(sandbox *api.Sandbox, url agent_db.AgentUrlRecord) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("url", AgentUrlJSON(sandbox, url))
	return document
}

// AgentUrlsJSON is {urls}.
func AgentUrlsJSON(sandbox *api.Sandbox, urls []agent_db.AgentUrlRecord) *serializabledeps.SerializableObject {
	list := sandbox.Deps.SerializableDeps.CreateArray()
	for _, url := range urls {
		list.AddItemToArray(AgentUrlJSON(sandbox, url))
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("urls", list)
	return document
}

// RemovedJSON is {removed}.
func RemovedJSON(sandbox *api.Sandbox, removed int) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("removed", removed)
	return document
}

// AgentSourceJSON is {source: {repo_url, branch, sync_time}}.
func AgentSourceJSON(sandbox *api.Sandbox, settings websitesource.Settings) *serializabledeps.SerializableObject {
	source := sandbox.Deps.SerializableDeps.CreateObject()
	source.AddItemToObject("repo_url", settings.RepoUrl)
	source.AddItemToObject("branch", settings.Branch)
	source.AddItemToObject("sync_time", settings.SyncTime)
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("source", source)
	return document
}
