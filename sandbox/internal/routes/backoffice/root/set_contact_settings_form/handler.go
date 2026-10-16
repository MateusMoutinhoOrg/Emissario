package set_contact_settings_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers POST /admin/root/set-contact-settings: the settings are
// checked and saved, and the browser is sent back to the page. A refused
// value answers the form again, filled with what was sent, under a 400 with
// the reason above it. Only a root reaches it: the backoffice-root-guard
// answers anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	sent := backofficerender.ContactSettingsForm{
		Enabled:       input.Body.Enabled == "1" || input.Body.Enabled == "on",
		MaxBytes:      input.Body.MaxBytes,
		RatePerMinute: input.Body.RatePerMinute,
		IpDailyLimit:  input.Body.IpDailyLimit,
		RetentionDays: input.Body.RetentionDays,
	}
	settingsInput, refusal := contact.InputOfForm(sandbox, contact.SettingsForm{
		Enabled:       input.Body.Enabled,
		MaxBytes:      input.Body.MaxBytes,
		RatePerMinute: input.Body.RatePerMinute,
		IpDailyLimit:  input.Body.IpDailyLimit,
		RetentionDays: input.Body.RetentionDays,
		Acknowledge:   input.Body.Acknowledge,
	})
	if refusal == nil {
		var err error
		_, refusal, err = contact.ApplySettings(sandbox, settingsInput)
		if err != nil {
			return err
		}
	}
	if refusal != nil && len(refusal.Notices) > 0 {
		notices, err := privacy.LoadNotices(sandbox, refusal.Notices, privacy.NoticeContext{Retention: settingsInput.RetentionDays})
		if err != nil {
			return err
		}
		sent.Notices = notices
		return backofficerender.RenderContactSettingsPage(sandbox, response, refusal.Status, props.User, sent, "", 0, "")
	}
	if refusal != nil {
		return backofficerender.RenderContactSettingsPage(sandbox, response, refusal.Status, props.User, sent, "", 0, refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, contact.SettingsPath+"?notice="+contact.NoticeSaved)
}
