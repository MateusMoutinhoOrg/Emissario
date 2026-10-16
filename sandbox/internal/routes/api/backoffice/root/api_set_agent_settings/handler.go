package api_set_agent_settings

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/set-agent-settings, the JSON twin of
// the settings form: the settings are checked and saved, a blank api key
// keeping the one saved, and answered as get-agent-settings answers them. A
// refused value is answered 400 with its field, and a save that needs notices
// 409 acknowledgement-required with them. Only a root reaches it: the
// backoffice-api-root-guard answers anyone else under /api/admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	settings, refusal, err := agent.ApplySettings(sandbox, agent.SettingsInput{
		Enabled:              input.Body.Enabled,
		BaseUrl:              input.Body.BaseUrl,
		ApiKey:               input.Body.ApiKey,
		ClearApiKey:          input.Body.ClearApiKey,
		Model:                input.Body.Model,
		MaxTokens:            int64(input.Body.MaxTokens),
		DailyLimit:           int64(input.Body.DailyLimit),
		IpDailyLimit:         int64(input.Body.IpDailyLimit),
		RatePerMinute:        int64(input.Body.RatePerMinute),
		SummariesEnabled:     input.Body.SummariesEnabled,
		SummaryRetentionDays: int64(input.Body.SummaryRetentionDays),
		Acknowledge:          input.Body.Acknowledge,
	})
	if err != nil {
		return err
	}
	if refusal != nil {
		return backofficeapi.FailPrivacy(sandbox, response, &privacy.Refusal{Status: refusal.Status, Field: refusal.Field, Message: refusal.Message, Notices: refusal.Notices}, privacy.NoticeContext{Host: agent.ProviderHost(sandbox, settings), Retention: settings.SummaryRetentionDays})
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.AgentSettingsResponseJSON(sandbox, settings))
}
