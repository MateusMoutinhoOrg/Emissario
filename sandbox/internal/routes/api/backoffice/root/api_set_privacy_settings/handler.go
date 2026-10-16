package api_set_privacy_settings

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/set-privacy-settings, the JSON twin of
// the Privacy settings form: the settings are checked and saved, and
// answered as {settings}. A refused value is answered 400 with its field,
// and a save that needs notices 409 acknowledgement-required with them,
// until every id is in acknowledge. Only a root reaches it: the
// backoffice-api-root-guard answers anyone else under /api/admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	settings, refusal, err := privacyadmin.ApplySettings(sandbox, privacyadmin.SettingsInput{
		IpMode:                 input.Body.IpMode,
		ViewerIp:               input.Body.ViewerIp,
		AccessLog:              input.Body.AccessLog,
		AccessLogRetentionDays: int64(input.Body.AccessLogRetentionDays),
		AccessLogTotp:          input.Body.AccessLogTotp,
		AccessLogInventory:     input.Body.AccessLogInventory,
		IpRetentionDays:        int64(input.Body.IpRetentionDays),
		BackupRetentionDays:    int64(input.Body.BackupRetentionDays),
		NoticeDraft:            input.Body.NoticeDraft,
		Acknowledge:            input.Body.Acknowledge,
	})
	if err != nil {
		return err
	}
	if refusal != nil && len(refusal.Notices) > 0 {
		notices, err := privacyadmin.LoadSettingsNotices(sandbox, settings, refusal.Notices)
		if err != nil {
			return err
		}
		return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusConflict, backofficeapi.AcknowledgementRequiredJSON(sandbox, notices))
	}
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.PrivacySettingsResponseJSON(sandbox, settings))
}
