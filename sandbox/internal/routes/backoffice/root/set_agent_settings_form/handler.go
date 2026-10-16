package set_agent_settings_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers POST /admin/root/set-agent-settings: the settings are
// checked and saved, a blank api key keeping the one saved, and the browser
// is sent back to the page. A refused value answers the form again, filled
// with what was sent, under a 400 with the reason above it; a save that needs
// notices answers it under a 409 with the notices inside it, to confirm. Only
// a root reaches it: the backoffice-root-guard answers anyone else under
// /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	sent := backofficerender.AgentSettingsForm{
		Enabled:              input.Body.Enabled == "1" || input.Body.Enabled == "on",
		BaseUrl:              input.Body.BaseUrl,
		Model:                input.Body.Model,
		MaxTokens:            input.Body.MaxTokens,
		DailyLimit:           input.Body.DailyLimit,
		IpDailyLimit:         input.Body.IpDailyLimit,
		RatePerMinute:        input.Body.RatePerMinute,
		SummariesEnabled:     input.Body.SummariesEnabled == "1" || input.Body.SummariesEnabled == "on",
		SummaryRetentionDays: input.Body.SummaryRetentionDays,
	}
	settingsInput, refusal := agent.InputOfForm(sandbox, agent.SettingsForm{
		Enabled:              input.Body.Enabled,
		BaseUrl:              input.Body.BaseUrl,
		ApiKey:               input.Body.ApiKey,
		ClearApiKey:          input.Body.ClearApiKey,
		Model:                input.Body.Model,
		MaxTokens:            input.Body.MaxTokens,
		DailyLimit:           input.Body.DailyLimit,
		IpDailyLimit:         input.Body.IpDailyLimit,
		RatePerMinute:        input.Body.RatePerMinute,
		SummariesEnabled:     input.Body.SummariesEnabled,
		SummaryRetentionDays: input.Body.SummaryRetentionDays,
		Acknowledge:          input.Body.Acknowledge,
	})
	if refusal == nil {
		var err error
		var settings agent.Settings
		settings, refusal, err = agent.ApplySettings(sandbox, settingsInput)
		if err != nil {
			return err
		}
		if refusal != nil && len(refusal.Notices) > 0 {
			notices, err := privacy.LoadNotices(sandbox, refusal.Notices, privacy.NoticeContext{Host: agent.ProviderHost(sandbox, settings), Retention: settings.SummaryRetentionDays})
			if err != nil {
				return err
			}
			sent.Notices = notices
			sent.RetypeKey = input.Body.ApiKey != ""
			return backofficerender.RenderAgentSettingsPage(sandbox, response, refusal.Status, props.User, sent, "", "")
		}
	}
	if refusal != nil {
		return backofficerender.RenderAgentSettingsPage(sandbox, response, refusal.Status, props.User, sent, "", refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, agent.SettingsPath+"?notice="+agent.NoticeSaved)
}
