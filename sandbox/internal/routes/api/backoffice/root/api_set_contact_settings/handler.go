package api_set_contact_settings

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/set-contact-settings, the JSON twin of
// the settings form: the settings are checked and saved, and answered. Only a
// root reaches it: the backoffice-api-root-guard answers anyone else under
// /api/admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	settings, refusal, err := contact.ApplySettings(sandbox, contact.SettingsInput{
		Enabled:       input.Body.Enabled,
		MaxBytes:      int64(input.Body.MaxBytes),
		RatePerMinute: int64(input.Body.RatePerMinute),
		IpDailyLimit:  int64(input.Body.IpDailyLimit),
		RetentionDays: int64(input.Body.RetentionDays),
		Acknowledge:   input.Body.Acknowledge,
	})
	if err != nil {
		return err
	}
	if refusal != nil {
		return backofficeapi.FailPrivacy(sandbox, response, &privacy.Refusal{Status: refusal.Status, Field: refusal.Field, Message: refusal.Message, Notices: refusal.Notices}, privacy.NoticeContext{Retention: int64(input.Body.RetentionDays)})
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.ContactSettingsResponseJSON(sandbox, settings))
}
